use super::{PortSignals, jarray};

fn has_path_fragment(paths: &[String], fragment: &str) -> bool {
    paths.iter().any(|p| p.contains(fragment))
}

pub(super) fn semantic_rules_json(
    signals: &PortSignals,
    target_mc: &str,
    target_loader: &str,
) -> String {
    let mut rules: Vec<String> = Vec::new();

    if target_mc == "1.20.1" && !signals.data_components.is_empty() {
        rules.push(format!(
            "{{\"id\":\"mc-1.20.5-data-components-to-1.20.1\",\"status\":\"required\",\"kind\":\"semantic-storage-backport\",\"source\":\"ItemStack data components\",\"target\":\"1.20.1 ItemStack NBT + target-native APIs\",\"preserve\":[\"field-values\",\"defaults\",\"save-reload\",\"client-server-sync\",\"item-copy-semantics\"],\"verification\":[\"roundtrip-original-target-reload\",\"stack-copy-split-merge\",\"server-client-inventory-sync\"],\"evidenceFiles\":{}}}",
            jarray(&signals.data_components)
        ));
    }

    if target_mc == "1.20.1"
        && has_path_fragment(&signals.data_components, "SpellDataComponents.java")
    {
        rules.push(
            "{\"id\":\"spell-engine-spell-container-component-to-nbt\",\"status\":\"proven-target-anchor\",\"kind\":\"semantic-storage-backport\",\"source\":\"SpellDataComponents.SPELL_CONTAINER\",\"target\":\"SpellContainerHelper NBT_KEY_CONTAINER=spell_container\",\"mapping\":{\"content\":\"content-discriminator\",\"is_proxy\":\"is_proxy\",\"pool\":\"pool\",\"max_spell_count\":\"max_spell_count\",\"spell_ids\":\"spell_ids-string-list\"},\"preserve\":[\"container-validity\",\"pool-binding\",\"spell-ordering\",\"proxy-semantics\",\"loot-bound-containers\",\"save-reload\"],\"targetEvidence\":\"SpellEngine-1.20.1-native-SpellContainerHelper-NBT-codec\"}".to_string(),
        );
        rules.push(
            "{\"id\":\"spell-engine-spell-choice-component-to-nbt\",\"status\":\"suite-safe-plus-generic-fallback\",\"kind\":\"semantic-storage-backport\",\"source\":\"SpellDataComponents.SPELL_CHOICE\",\"target\":\"spell_engine:spell_choice compatibility NBT\",\"mapping\":{\"pool\":\"pool\",\"apply_on_choice\":\"versioned-per-spell-component-change-payload\"},\"suiteEvidence\":\"audited-RPGSeries-sources-do-not-populate-apply_on_choice-outside-Spell-Engine-builder-API\",\"policy\":\"known-vanilla-component-changes-get-explicit-1.20.1-adapters; unknown-patches-stay-preserved-as-opaque-payload-and-block-fidelity-certification-until-adapted\"}".to_string(),
        );
        rules.push(
            "{\"id\":\"spell-engine-equipment-set-component\",\"status\":\"required\",\"kind\":\"identifier-storage-backport\",\"source\":\"SpellDataComponents.EQUIPMENT_SET\",\"target\":\"versioned-item-NBT-identifier-plus-equipment-resolver\",\"preserve\":[\"equipment-set-identity\",\"armor-model-selection\",\"save-reload\"]}".to_string(),
        );
        rules.push(
            "{\"id\":\"spell-engine-item-model-component\",\"status\":\"required\",\"kind\":\"render-data-backport\",\"source\":\"SpellDataComponents.ITEM_MODEL\",\"target\":\"1.20.1-model-or-CMD-compatibility-metadata-plus-render-resolver\",\"preserve\":[\"spell-book-model-identity\",\"scroll-model-identity\",\"advancement-display-stacks\"]}".to_string(),
        );
    }

    if target_mc == "1.20.1"
        && has_path_fragment(&signals.data_components, "RangedWeaponProperties.java")
    {
        rules.push(
            "{\"id\":\"ranged-weapon-properties-component-to-ranged-config\",\"status\":\"proven-target-anchor\",\"kind\":\"weapon-physics-backport\",\"source\":\"ranged_weapon:properties DataComponent<RangedWeaponProperties>\",\"target\":\"1.20.1 CustomRangedWeapon + RangedConfig contract\",\"mapping\":{\"pull_time\":\"RangedConfig.pull_time\",\"damage\":\"RangedConfig.damage / target baseline override\",\"velocity\":\"RangedConfig.velocity / target baseline override\"},\"preserve\":[\"pull-time\",\"damage-baseline\",\"velocity-baseline\",\"bow-crossbow-defaults\",\"attribute-scaling\"],\"targetEvidence\":\"RangedWeaponAPI-1.20.1-RangedConfig-and-CustomRangedWeapon\"}".to_string(),
        );
    }

    if target_mc == "1.20.1" && !signals.modern_networking.is_empty() {
        rules.push(format!(
            "{{\"id\":\"modern-payload-network-to-1.20.1\",\"status\":\"required\",\"kind\":\"network-protocol-backport\",\"source\":\"CustomPayload-or-PacketCodec-era\",\"target\":\"target-era-packet-buffers-and-channels\",\"preserve\":[\"packet-direction\",\"logical-side\",\"ordering\",\"validation\",\"server-authority\"],\"evidenceFiles\":{}}}",
            jarray(&signals.modern_networking)
        ));
    }

    if target_loader.eq_ignore_ascii_case("forge") && !signals.architectury.is_empty() {
        rules.push(
            "{\"id\":\"architectury-to-forge-1.20.1\",\"status\":\"required\",\"kind\":\"loader-platform-backport\",\"source\":\"common-fabric-neoforge-platform-split\",\"target\":\"Forge-47.4.x-plus-Java-17\",\"policy\":\"preserve-common-semantic-contract-and-emit-Forge-native-lifecycle-registry-network-render-hooks\"}".to_string(),
        );
    }

    if signals.mixin_sources > 0 || signals.mixin_configs > 0 {
        rules.push(format!(
            "{{\"id\":\"mixin-semantic-retarget\",\"status\":\"required\",\"kind\":\"mixin-backport\",\"mixins\":{},\"configs\":{},\"policy\":\"retarget-by-owner-name-descriptor-injection-point-and-control-flow-fingerprint; refmap-or-name-remap-alone-is-not-acceptance\"}}",
            signals.mixin_sources, signals.mixin_configs
        ));
    }

    if !signals.access_wideners.is_empty()
        || !signals.access_transformers.is_empty()
        || !signals.class_tweakers.is_empty()
    {
        rules.push(format!(
            "{{\"id\":\"access-intent-translation\",\"status\":\"required\",\"kind\":\"access-semantic-backport\",\"accessWideners\":{},\"accessTransformers\":{},\"classTweakers\":{},\"policy\":\"map-member-owner-name-descriptor-first-then-translate-visibility-extendability-mutability-intent\"}}",
            jarray(&signals.access_wideners),
            jarray(&signals.access_transformers),
            jarray(&signals.class_tweakers)
        ));
    }

    format!("[{}]", rules.join(","))
}
