use std::env;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const VERSION: &str = "0.2.0";

fn usage() -> ! {
    eprintln!(
        "omniporter {VERSION}\n\nUSAGE:\n  omniporter doctor\n  omniporter fingerprint <file>\n  omniporter identify <file>\n  omniporter ferium scan <directory> <game_version> <loader> [platform]\n  omniporter ferium upgrade <directory> <game_version> <loader> [platform]\n  omniporter resolve modrinth <project> [game_version] [loader]\n  omniporter resolve github <owner/repo>\n  omniporter resolve curseforge <mod_id>\n\nENVIRONMENT:\n  OMNIPORTER_FERIUM   Optional explicit path to Ferium.\n  CURSEFORGE_API_KEY  Optional CurseForge API key; never persisted.\n  GITHUB_TOKEN        Optional GitHub token; never persisted.\n\nOmniPorter treats provider matches as evidence, not proof. Cross-provider identity must agree before an automatic port/update is accepted."
    );
    std::process::exit(2);
}

fn have(cmd: &str) -> bool {
    Command::new("sh")
        .args(["-lc", &format!("command -v {} >/dev/null 2>&1", cmd)])
        .status()
        .map(|s| s.success())
        .unwrap_or(false)
}

fn version_line(cmd: &str, args: &[&str]) -> String {
    match Command::new(cmd).args(args).stderr(Stdio::piped()).output() {
        Ok(o) => {
            let s = if o.stdout.is_empty() {
                String::from_utf8_lossy(&o.stderr).into_owned()
            } else {
                String::from_utf8_lossy(&o.stdout).into_owned()
            };
            s.lines().next().unwrap_or("unknown").to_string()
        }
        Err(_) => "missing".to_string(),
    }
}

fn ferium_path() -> Option<PathBuf> {
    if let Some(explicit) = env::var_os("OMNIPORTER_FERIUM") {
        let p = PathBuf::from(explicit);
        if p.is_file() {
            return Some(p);
        }
    }
    if have("ferium") {
        return Some(PathBuf::from("ferium"));
    }
    for p in [
        "/mnt/data/ferium-bin/ferium",
        "./tools/ferium/ferium",
        "./ferium",
    ] {
        let path = PathBuf::from(p);
        if path.is_file() {
            return Some(path);
        }
    }
    None
}

fn doctor() {
    println!("OmniPorter Dev Kit Doctor");
    println!("version: {VERSION}");
    println!("rustc: {}", version_line("rustc", &["--version"]));
    println!("cargo: {}", version_line("cargo", &["--version"]));
    println!("java: {}", version_line("java", &["-version"]));
    for c in ["curl", "git", "unzip", "sha1sum", "sha256sum", "sha512sum"] {
        println!("{c}: {}", if have(c) { "ok" } else { "missing" });
    }
    match ferium_path() {
        Some(p) => println!(
            "ferium: {} [{}]",
            version_line(p.to_string_lossy().as_ref(), &["--version"]),
            p.display()
        ),
        None => println!("ferium: missing"),
    }
    println!(
        "curseforge_api_key: {}",
        if env::var_os("CURSEFORGE_API_KEY").is_some() {
            "configured"
        } else {
            "not configured (optional)"
        }
    );
    println!(
        "github_token: {}",
        if env::var_os("GITHUB_TOKEN").is_some() {
            "configured"
        } else {
            "not configured (optional)"
        }
    );
}

fn command_hash(program: &str, path: &Path) -> Option<String> {
    let out = Command::new(program).arg(path).output().ok()?;
    if !out.status.success() {
        return None;
    }
    Some(
        String::from_utf8_lossy(&out.stdout)
            .split_whitespace()
            .next()
            .unwrap_or("")
            .to_string(),
    )
}

fn artifact_hashes(path: &Path) -> (Option<String>, Option<String>, Option<String>) {
    (
        command_hash("sha1sum", path),
        command_hash("sha256sum", path),
        command_hash("sha512sum", path),
    )
}

fn fingerprint(path: &str) {
    let p = Path::new(path);
    if !p.is_file() {
        eprintln!("not a file: {path}");
        std::process::exit(3);
    }
    let size = fs::metadata(p).map(|m| m.len()).unwrap_or(0);
    let (sha1, sha256, sha512) = artifact_hashes(p);
    println!(
        "{{\"path\":{},\"size\":{},\"sha1\":{},\"sha256\":{},\"sha512\":{}}}",
        json(path),
        size,
        json_opt(sha1.as_deref()),
        json_opt(sha256.as_deref()),
        json_opt(sha512.as_deref())
    );
}

fn json_opt(s: Option<&str>) -> String {
    s.map(json).unwrap_or_else(|| "null".to_string())
}

fn json(s: &str) -> String {
    let mut out = String::with_capacity(s.len() + 2);
    out.push('"');
    for c in s.chars() {
        match c {
            '\\' => out.push_str("\\\\"),
            '"' => out.push_str("\\\""),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            c if c.is_control() => out.push_str(&format!("\\u{:04x}", c as u32)),
            c => out.push(c),
        }
    }
    out.push('"');
    out
}

fn curl_json_soft(url: &str, headers: &[String]) -> Result<String, String> {
    let mut c = Command::new("curl");
    c.args([
        "-fsSL",
        "--connect-timeout",
        "10",
        "--max-time",
        "30",
        "-H",
        &format!(
            "User-Agent: OmniPorter/{VERSION} (+https://github.com/Herbertofury/Minecraft-Mod-Vault)"
        ),
    ]);
    for h in headers {
        c.args(["-H", h]);
    }
    c.arg(url);
    let o = c.output().map_err(|e| e.to_string())?;
    if !o.status.success() {
        return Err(String::from_utf8_lossy(&o.stderr).trim().to_string());
    }
    String::from_utf8(o.stdout).map_err(|e| e.to_string())
}

fn curl_json(url: &str, headers: &[String]) -> String {
    match curl_json_soft(url, headers) {
        Ok(s) => s,
        Err(e) => {
            eprintln!("request failed: {url}\n{e}");
            std::process::exit(5);
        }
    }
}

fn print_normalized(provider: &str, query: &str, endpoint: &str, raw: &str) {
    println!(
        "{{\"provider\":{},\"query\":{},\"endpoint\":{},\"fetched\":true,\"raw\":{}}}",
        json(provider),
        json(query),
        json(endpoint),
        json(raw)
    );
}

fn identify(path: &str) {
    let p = Path::new(path);
    if !p.is_file() {
        eprintln!("not a file: {path}");
        std::process::exit(3);
    }
    let size = fs::metadata(p).map(|m| m.len()).unwrap_or(0);
    let (sha1, sha256, sha512) = artifact_hashes(p);
    let mut modrinth_status = "unavailable".to_string();
    let mut modrinth_raw: Option<String> = None;
    if let Some(hash) = sha512.as_deref() {
        let endpoint = format!("https://api.modrinth.com/v2/version_file/{hash}?algorithm=sha512");
        match curl_json_soft(&endpoint, &[]) {
            Ok(raw) => {
                modrinth_status = "matched".to_string();
                modrinth_raw = Some(raw);
            }
            Err(_) => modrinth_status = "no-match-or-network-error".to_string(),
        }
    }
    println!(
        "{{\"path\":{},\"size\":{},\"hashes\":{{\"sha1\":{},\"sha256\":{},\"sha512\":{}}},\"providers\":{{\"modrinth\":{{\"status\":{},\"raw\":{}}},\"curseforge\":{{\"status\":\"defer-to-ferium/fingerprint-worker\"}},\"github\":{{\"status\":\"requires-source-link-correlation\"}}}},\"verdict\":\"identity-evidence-only\"}}",
        json(path),
        size,
        json_opt(sha1.as_deref()),
        json_opt(sha256.as_deref()),
        json_opt(sha512.as_deref()),
        json(&modrinth_status),
        json_opt(modrinth_raw.as_deref())
    );
}

fn resolve_modrinth(args: &[String]) {
    if args.is_empty() {
        usage();
    }
    let project = &args[0];
    let mut url = format!(
        "https://api.modrinth.com/v2/project/{}/version?include_changelog=false",
        project
    );
    if let Some(game) = args.get(1) {
        url.push_str(&format!("&game_versions=%5B%22{}%22%5D", game));
    }
    if let Some(loader) = args.get(2) {
        url.push_str(&format!("&loaders=%5B%22{}%22%5D", loader));
    }
    let raw = curl_json(&url, &[]);
    print_normalized("modrinth", project, &url, &raw);
}

fn resolve_github(args: &[String]) {
    if args.is_empty() || !args[0].contains('/') {
        usage();
    }
    let repo = &args[0];
    let url = format!("https://api.github.com/repos/{}/releases/latest", repo);
    let mut headers = vec![
        "Accept: application/vnd.github+json".to_string(),
        "X-GitHub-Api-Version: 2022-11-28".to_string(),
    ];
    if let Ok(token) = env::var("GITHUB_TOKEN") {
        headers.push(format!("Authorization: Bearer {token}"));
    }
    let raw = curl_json(&url, &headers);
    print_normalized("github", repo, &url, &raw);
}

fn resolve_curseforge(args: &[String]) {
    if args.is_empty() {
        usage();
    }
    let id = &args[0];
    let key = match env::var("CURSEFORGE_API_KEY") {
        Ok(k) if !k.trim().is_empty() => k,
        _ => {
            eprintln!(
                "CurseForge's official API requires x-api-key. Configure CURSEFORGE_API_KEY locally; do not place the key in Drive, GitHub, logs, or manifests. Ferium can also be used as the provider worker when a key is supplied to it at runtime."
            );
            std::process::exit(6);
        }
    };
    let url = format!("https://api.curseforge.com/v1/mods/{id}");
    let headers = vec![
        format!("x-api-key: {key}"),
        "Accept: application/json".to_string(),
    ];
    let raw = curl_json(&url, &headers);
    print_normalized("curseforge", id, &url, &raw);
}

fn resolve(args: &[String]) {
    if args.len() < 2 {
        usage();
    }
    match args[0].as_str() {
        "modrinth" => resolve_modrinth(&args[1..]),
        "github" => resolve_github(&args[1..]),
        "curseforge" => resolve_curseforge(&args[1..]),
        _ => usage(),
    }
}

fn normalize_loader(loader: &str) -> &str {
    match loader {
        "neoforge" => "neo-forge",
        other => other,
    }
}

fn ferium_profile_args(directory: &str, game: &str, loader: &str) -> Vec<String> {
    vec![
        "profile".to_string(),
        "create".to_string(),
        "--name".to_string(),
        "omniporter-intake".to_string(),
        "--game-version".to_string(),
        game.to_string(),
        "--mod-loader".to_string(),
        normalize_loader(loader).to_string(),
        "--output-dir".to_string(),
        directory.to_string(),
    ]
}

fn run_ferium(args: &[String], config: &Path) -> std::process::Output {
    let ferium = ferium_path().unwrap_or_else(|| {
        eprintln!("Ferium is not available. Set OMNIPORTER_FERIUM or add it to the Dev Kit.");
        std::process::exit(7);
    });
    Command::new(ferium)
        .arg("--config-file")
        .arg(config)
        .args(args)
        .output()
        .unwrap_or_else(|e| {
            eprintln!("failed to launch Ferium: {e}");
            std::process::exit(7);
        })
}

fn ferium_workflow(args: &[String], upgrade: bool) {
    if args.len() < 3 {
        usage();
    }
    let directory = &args[0];
    let game = &args[1];
    let loader = &args[2];
    let platform = args.get(3).map(String::as_str).unwrap_or("modrinth");
    if !Path::new(directory).is_dir() {
        eprintln!("not a directory: {directory}");
        std::process::exit(3);
    }

    let temp = env::temp_dir().join(format!(
        "omniporter-ferium-{}-{}",
        std::process::id(),
        if upgrade { "upgrade" } else { "scan" }
    ));
    let _ = fs::create_dir_all(&temp);
    let config = temp.join("config.json");

    let create = run_ferium(&ferium_profile_args(directory, game, loader), &config);
    if !create.status.success() {
        eprintln!(
            "Ferium profile creation failed:\n{}",
            String::from_utf8_lossy(&create.stderr)
        );
        std::process::exit(8);
    }

    let scan_args = vec![
        "scan".to_string(),
        "--directory".to_string(),
        directory.to_string(),
        "--platform".to_string(),
        platform.to_string(),
    ];
    let scan = run_ferium(&scan_args, &config);
    let scan_out = String::from_utf8_lossy(&scan.stdout).into_owned();
    let scan_err = String::from_utf8_lossy(&scan.stderr).into_owned();
    if !scan.status.success() {
        eprintln!("Ferium scan failed:\n{scan_err}");
        std::process::exit(8);
    }

    let mut upgrade_out = String::new();
    let mut upgrade_err = String::new();
    let mut upgrade_status = "not-requested";
    if upgrade {
        let up = run_ferium(&["upgrade".to_string()], &config);
        upgrade_out = String::from_utf8_lossy(&up.stdout).into_owned();
        upgrade_err = String::from_utf8_lossy(&up.stderr).into_owned();
        if !up.status.success() {
            upgrade_status = "failed";
        } else {
            upgrade_status = "completed";
        }
    }

    println!(
        "{{\"worker\":\"ferium\",\"mode\":{},\"directory\":{},\"gameVersion\":{},\"loader\":{},\"preferredPlatform\":{},\"scanStdout\":{},\"scanStderr\":{},\"upgradeStatus\":{},\"upgradeStdout\":{},\"upgradeStderr\":{},\"policy\":\"results-require-omniporter-provenance-validation\"}}",
        json(if upgrade { "scan+upgrade" } else { "scan" }),
        json(directory),
        json(game),
        json(loader),
        json(platform),
        json(&scan_out),
        json(&scan_err),
        json(upgrade_status),
        json(&upgrade_out),
        json(&upgrade_err)
    );

    let _ = fs::remove_dir_all(&temp);
    if upgrade && upgrade_status == "failed" {
        std::process::exit(9);
    }
}

fn ferium(args: &[String]) {
    if args.is_empty() {
        usage();
    }
    match args[0].as_str() {
        "scan" => ferium_workflow(&args[1..], false),
        "upgrade" => ferium_workflow(&args[1..], true),
        _ => usage(),
    }
}

fn main() {
    let args: Vec<String> = env::args().skip(1).collect();
    if args.is_empty() {
        usage();
    }
    match args[0].as_str() {
        "doctor" => doctor(),
        "fingerprint" if args.len() == 2 => fingerprint(&args[1]),
        "identify" if args.len() == 2 => identify(&args[1]),
        "ferium" => ferium(&args[1..]),
        "resolve" => resolve(&args[1..]),
        "--version" | "-V" => println!("omniporter {VERSION}"),
        _ => usage(),
    }
}
