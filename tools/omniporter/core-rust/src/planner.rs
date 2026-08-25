use std::fs;
use std::path::{Path, PathBuf};

use crate::{VERSION, json, json_opt, usage};

#[derive(Default)]
struct PortSignals {
    source_files: usize,
    java_files: usize,
    kotlin_files: usize,
    json_files: usize,
    mixin_configs: usize,
    mixin_sources: usize,
    data_components: Vec<String>,
    modern_networking: Vec<String>,
    architectury: Vec<String>,
    registries: Vec<String>,
    datagen: Vec<String>,
    access_wideners: Vec<String>,
    access_transformers: Vec<String>,
    class_tweakers: Vec<String>,
    resources: usize,
    generated_resources: usize,
}

fn walk_files(root: &Path, out: &mut Vec<PathBuf>) {
    let Ok(entries) = fs::read_dir(root) else { return; };
    for entry in entries.flatten() {
        let p = entry.path();
        if p.is_dir() {
            let name = p.file_name().and_then(|s| s.to_str()).unwrap_or("");
            if matches!(name, ".git" | ".gradle" | "build" | "target" | "run" | "runs") { continue; }
            walk_files(&p, out);
        } else if p.is_file() { out.push(p); }
    }
}

fn rel(root: &Path, p: &Path) -> String {
    p.strip_prefix(root).unwrap_or(p).to_string_lossy().replace('\\', "/")
}

fn push_once(v: &mut Vec<String>, path: String) {
    if !v.iter().any(|x| x == &path) { v.push(path); }
}

fn contains_any(text: &str, needles: &[&str]) -> bool { needles.iter().any(|n| text.contains(n)) }

fn scan_port_signals(root: &Path) -> PortSignals {
    let mut files = Vec::new();
    walk_files(root, &mut files);
    let mut s = PortSignals { source_files: files.len(), ..PortSignals::default() };
    for p in files {
        let path = rel(root, &p);
        let lower = path.to_ascii_lowercase();
        let ext = p.extension().and_then(|x| x.to_str()).unwrap_or("").to_ascii_lowercase();
        match ext.as_str() { "java" => s.java_files += 1, "kt" | "kts" => s.kotlin_files += 1, "json" => s.json_files += 1, _ => {} }
        if lower.ends_with(".accesswidener") { push_once(&mut s.access_wideners, path.clone()); }
        if lower.ends_with("accesstransformer.cfg") { push_once(&mut s.access_transformers, path.clone()); }
        if lower.ends_with(".classtweaker") { push_once(&mut s.class_tweakers, path.clone()); }
        if lower.ends_with(".mixins.json") || (lower.contains("mixin") && lower.ends_with(".json")) { s.mixin_configs += 1; }
        if lower.contains("src/main/resources") || lower.contains("src/main/generated") { s.resources += 1; }
        if lower.contains("src/main/generated") { s.generated_resources += 1; }
        if !matches!(ext.as_str(), "java" | "kt" | "kts" | "gradle" | "json" | "toml" | "properties") { continue; }
        let Ok(text) = fs::read_to_string(&p) else { continue; };
        if contains_any(&text, &["net.minecraft.component", "DataComponentType", "DataComponents", "DataComponentTypes", "ComponentMap", "AttributeModifiersComponent", "CustomData", "SpellDataComponents"]) { push_once(&mut s.data_components, path.clone()); }
        if contains_any(&text, &["CustomPayload", "CustomPacketPayload", "StreamCodec", "PacketCodec", "RegistryFriendlyByteBuf", "PayloadTypeRegistry", "ClientPlayNetworking", "ServerPlayNetworking"]) { push_once(&mut s.modern_networking, path.clone()); }
        if contains_any(&text, &["dev.architectury", "ExpectPlatform", "architectury {"]) || lower.starts_with("common/") || lower.starts_with("fabric/") || lower.starts_with("neoforge/") { push_once(&mut s.architectury, path.clone()); }
        if contains_any(&text, &["@Mixin(", "org.spongepowered.asm.mixin"]) { s.mixin_sources += 1; }
        if contains_any(&text, &["DeferredRegister", "RegistrationProvider", "RegistrySupplier", "BuiltInRegistries", "Registries.", "Registry.register", "RegisterEvent"]) { push_once(&mut s.registries, path.clone()); }
        if contains_any(&text, &["DataGenerator", "Datagen", "DataProvider", "FabricDataGenerator", "GatherDataEvent"]) || lower.contains("generated/") { push_once(&mut s.datagen, path.clone()); }
    }
    s
}

fn read_properties(root: &Path) -> Vec<(String, String)> {
    let mut candidates = vec![root.join("gradle.properties")];
    for module in ["common", "fabric", "forge", "neoforge"] { candidates.push(root.join(module).join("gradle.properties")); }
    let mut out = Vec::new();
    for p in candidates {
        let Ok(text) = fs::read_to_string(p) else { continue; };
        for raw in text.lines() {
            let line = raw.trim();
            if line.is_empty() || line.starts_with('#') { continue; }
            if let Some((k, v)) = line.split_once('=') {
                let k = k.trim().to_string(); let v = v.trim().to_string();
                if !out.iter().any(|(ek, _): &(String, String)| ek == &k) { out.push((k, v)); }
            }
        }
    }
    out
}

fn prop<'a>(props: &'a [(String, String)], keys: &[&str]) -> Option<&'a str> {
    for key in keys { if let Some((_, v)) = props.iter().find(|(k, _)| k == key) { return Some(v.as_str()); } }
    None
}

fn jarray(items: &[String]) -> String {
    let mut out = String::from("[");
    for (i, item) in items.iter().enumerate() { if i > 0 { out.push(','); } out.push_str(&json(item)); }
    out.push(']'); out
}

mod semantic;
use semantic::semantic_rules_json;

fn migration_risk(signals: &PortSignals, source_mc: Option<&str>, target_mc: &str) -> &'static str {
    if !signals.data_components.is_empty() && target_mc == "1.20.1" { return "critical-semantic-adaptation"; }
    if source_mc.is_some_and(|v| v != target_mc) && (!signals.modern_networking.is_empty() || signals.mixin_sources > 0) { return "high"; }
    if source_mc.is_some_and(|v| v != target_mc) { return "medium"; }
    "low"
}

fn port_plan_json(source: &Path, baseline: Option<&Path>, target_mc: &str, target_loader: &str) -> String {
    let props = read_properties(source);
    let signals = scan_port_signals(source);
    let source_mc = prop(&props, &["minecraft_version", "minecraft_version_range"]);
    let mod_version = prop(&props, &["mod_version", "version"]);
    let archive = prop(&props, &["archives_base_name", "mod_id"]);
    let risk = migration_risk(&signals, source_mc, target_mc);
    let baseline_status = baseline.map(|b| if b.is_dir() { "available" } else { "missing" }).unwrap_or("not-supplied");
    let baseline_path = baseline.map(|b| b.to_string_lossy().into_owned());
    let component_strategy = if target_mc == "1.20.1" && !signals.data_components.is_empty() { "translate-item-components-to-versioned-nbt-and-runtime-adapters; preserve serialization/network semantics" } else { "native-or-not-required" };
    let loader_strategy = if target_loader.eq_ignore_ascii_case("forge") && !signals.architectury.is_empty() { "emit-forge-1.20.1-platform-adapter-from-common-semantic-contract" } else { "target-native" };
    let semantic_rules = semantic_rules_json(&signals, target_mc, target_loader);
    format!("{{\"schema\":2,\"engine\":\"OmniPorter/{VERSION}\",\"source\":{{\"path\":{},\"minecraft\":{},\"modVersion\":{},\"archive\":{}}},\"target\":{{\"minecraft\":{},\"loader\":{}}},\"baseline\":{{\"status\":{},\"path\":{}}},\"signals\":{{\"files\":{},\"java\":{},\"kotlin\":{},\"json\":{},\"resources\":{},\"generatedResources\":{},\"mixinConfigs\":{},\"mixinSources\":{},\"dataComponents\":{},\"modernNetworking\":{},\"architectury\":{},\"registries\":{},\"datagen\":{},\"accessWideners\":{},\"accessTransformers\":{},\"classTweakers\":{}}},\"semanticRules\":{},\"portPolicy\":{{\"risk\":{},\"componentStrategy\":{},\"loaderStrategy\":{},\"preserve\":[\"registry-ids\",\"gameplay-semantics\",\"assets\",\"animations\",\"configs\",\"network-behavior\",\"persistence\",\"dependencies\"],\"acceptance\":\"build-plus-clean-client-server-runtime-and-behavior-diff\"}}}}", json(&source.to_string_lossy()), json_opt(source_mc), json_opt(mod_version), json_opt(archive), json(target_mc), json(target_loader), json(baseline_status), json_opt(baseline_path.as_deref()), signals.source_files, signals.java_files, signals.kotlin_files, signals.json_files, signals.resources, signals.generated_resources, signals.mixin_configs, signals.mixin_sources, jarray(&signals.data_components), jarray(&signals.modern_networking), jarray(&signals.architectury), jarray(&signals.registries), jarray(&signals.datagen), jarray(&signals.access_wideners), jarray(&signals.access_transformers), jarray(&signals.class_tweakers), semantic_rules, json(risk), json(component_strategy), json(loader_strategy))
}

fn parse_flag_value(args: &[String], flag: &str) -> Option<String> { args.windows(2).find(|w| w[0] == flag).map(|w| w[1].clone()) }
fn write_or_print(text: &str, out: Option<String>) { if let Some(path) = out { if let Some(parent) = Path::new(&path).parent() { let _ = fs::create_dir_all(parent); } if let Err(e) = fs::write(&path, text) { eprintln!("failed to write {path}: {e}"); std::process::exit(10); } println!("{path}"); } else { println!("{text}"); } }

pub(crate) fn port_plan(args: &[String]) {
    if args.len() < 3 { usage(); }
    let source = PathBuf::from(&args[0]);
    if !source.is_dir() { eprintln!("not a source directory: {}", source.display()); std::process::exit(3); }
    let baseline_arg = parse_flag_value(args, "--baseline");
    let baseline = baseline_arg.as_deref().map(Path::new);
    let out = parse_flag_value(args, "--out");
    let plan = port_plan_json(&source, baseline, &args[1], &args[2]);
    write_or_print(&(plan + "\n"), out);
}

fn find_project_roots(suite: &Path) -> Vec<PathBuf> {
    let mut roots = Vec::new();
    let Ok(level1) = fs::read_dir(suite) else { return roots; };
    for e in level1.flatten() {
        let p = e.path(); if !p.is_dir() { continue; }
        if p.join("gradle.properties").is_file() || p.join("settings.gradle").is_file() || p.join("settings.gradle.kts").is_file() { roots.push(p); continue; }
        if let Ok(level2) = fs::read_dir(&p) { for e2 in level2.flatten() { let p2 = e2.path(); if p2.is_dir() && (p2.join("gradle.properties").is_file() || p2.join("settings.gradle").is_file() || p2.join("settings.gradle.kts").is_file()) { roots.push(p2); } } }
    }
    roots.sort(); roots
}

fn suite_priority(name: &str) -> u8 {
    let n = name.to_ascii_lowercase();
    if n.contains("spellpower") || n.contains("spell-power") { 0 }
    else if n.contains("rangedweapon") || n.contains("ranged-weapon") { 1 }
    else if n.contains("spellengine") || n.contains("spell-engine") { 2 }
    else if n.contains("more-rpg-library") || n.contains("morerpglibrary") { 3 }
    else if ["wizards", "paladins", "rogues", "archers"].iter().any(|x| n.contains(x)) { 4 }
    else { 5 }
}

pub(crate) fn suite_plan(args: &[String]) {
    if args.len() < 3 { usage(); }
    let suite = PathBuf::from(&args[0]);
    if !suite.is_dir() { eprintln!("not a suite directory: {}", suite.display()); std::process::exit(3); }
    let mut roots = find_project_roots(&suite);
    roots.sort_by_key(|p| (suite_priority(&p.to_string_lossy()), p.to_string_lossy().to_string()));
    let names: Vec<String> = roots.iter().map(|p| p.file_name().and_then(|x| x.to_str()).unwrap_or("unknown").to_string()).collect();
    let lower_names: Vec<String> = names.iter().map(|x| x.to_ascii_lowercase()).collect();
    let missing_spell_power = !lower_names.iter().any(|x| x.contains("spellpower") || x.contains("spell-power")) && roots.iter().any(|p| { let mut files = Vec::new(); walk_files(p, &mut files); files.iter().any(|f| fs::read_to_string(f).map(|t| t.contains("spell_power") || t.contains("spell-power") || t.contains("spell_power_version")).unwrap_or(false)) });
    let mut project_json = String::from("[");
    for (i, root) in roots.iter().enumerate() { if i > 0 { project_json.push(','); } project_json.push_str(&port_plan_json(root, None, &args[1], &args[2])); }
    project_json.push(']');
    let missing = if missing_spell_power { "[\"SpellPower\"]" } else { "[]" };
    let out_json = format!("{{\"schema\":2,\"engine\":\"OmniPorter/{VERSION}\",\"suite\":{},\"target\":{{\"minecraft\":{},\"loader\":{}}},\"dependencyOrder\":{},\"missingRequiredSourceLayers\":{},\"projects\":{},\"policy\":\"dependency-first; baseline-branch semantic transplantation; no feature deletion to satisfy build\"}}\n", json(&suite.to_string_lossy()), json(&args[1]), json(&args[2]), jarray(&names), missing, project_json);
    write_or_print(&out_json, parse_flag_value(args, "--out"));
}
