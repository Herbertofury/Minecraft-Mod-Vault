package com.bert.slrsharedmana;

import net.minecraftforge.common.ForgeConfigSpec;

public final class BridgeConfig {
    public static final ForgeConfigSpec SPEC;
    public static final ForgeConfigSpec.BooleanValue ENABLED;
    public static final ForgeConfigSpec.DoubleValue SLR_MP_PER_IRON_MANA;
    public static final ForgeConfigSpec.BooleanValue SUPPRESS_ISS_PASSIVE_REGEN;
    public static final ForgeConfigSpec.BooleanValue ACCEPT_POSITIVE_ISS_CREDITS;
    public static final ForgeConfigSpec.IntValue MANA_REFRESH_TICKS;
    public static final ForgeConfigSpec.BooleanValue HIDE_ISS_MANA_HUD;
    public static final ForgeConfigSpec.BooleanValue LOG_COMPAT_SUMMARY;

    static {
        ForgeConfigSpec.Builder b = new ForgeConfigSpec.Builder();
        b.comment(
                "SLR Shared Mana - Forge 1.20.1",
                "SLR MP is the authoritative mana pool. Iron's Spells reads and spends that pool on demand.",
                "There is deliberately no player-tick mirror and no nearby-block scan in this mod.")
         .push("shared_mana");

        ENABLED = b.comment(
                "Master switch. When false, Iron's Spells returns to its native mana behavior.")
                .define("enabled", true);

        SLR_MP_PER_IRON_MANA = b.comment(
                "Conversion scale. 10 means 1 Iron mana == 10 SLR MP.",
                "The original 1.21 shared-mana bridge used 10 and this remains the compatibility default.")
                .defineInRange("slrMpPerIronMana", 10.0D, 0.001D, 1_000_000.0D);

        SUPPRESS_ISS_PASSIVE_REGEN = b.comment(
                "Disable Iron's passive mana regeneration while sharing is active.",
                "Keep this true so SLR's native regen/cooldown rules remain authoritative.",
                "Active positive mana grants (potions, integrations, Iron's Botany ISS-primary conversion) are NOT blocked.")
                .define("suppressIronPassiveRegen", true);

        ACCEPT_POSITIVE_ISS_CREDITS = b.comment(
                "Translate intentional positive Iron mana changes into SLR MP.",
                "Recommended true. This is what makes Iron's Botany ISS_PRIMARY/bidirectional conversion feed the shared pool.")
                .define("acceptPositiveIronManaCredits", true);

        MANA_REFRESH_TICKS = b.comment(
                "SLR native mana_refresh cooldown applied after an Iron mana spend.",
                "40 ticks (2 seconds) matches the original shared-mana bridge default.")
                .defineInRange("slrManaRefreshTicks", 40, 0, 20 * 60);

        b.pop();
        b.comment(
                "Iron's Botany compatibility",
                "The bridge intentionally does NOT duplicate or replace Iron's Botany's mana router.",
                "Iron's Botany keeps full control of manaUnificationMode:",
                "  HYBRID          - Botania and the SLR-backed ISS side cooperate exactly as Botany defines.",
                "  BOTANIA_PRIMARY - Botany may pay a cast from Botania instead of ISS; SLR is not double-charged.",
                "  ISS_PRIMARY     - Botany credits ISS; those credits become SLR MP when enabled above.",
                "  SEPARATE        - Botania remains separate; the ISS side is simply backed by SLR MP.",
                "  DISABLED        - Botany does no routing; SLR <-> ISS sharing still works.",
                "Use Iron's Botany's own config to select those modes. This avoids conflicting duplicate mode knobs.")
         .push("irons_botany");

        LOG_COMPAT_SUMMARY = b.comment(
                "Log the detected Iron's Botany mode once when a server starts.")
                .define("logDetectedModeOnServerStart", true);
        b.pop();

        b.comment("Client presentation").push("client");
        HIDE_ISS_MANA_HUD = b.comment(
                "Hide Iron's mana bar so the SLR MP bar is the single source of truth.",
                "Spell bars and all other Iron's UI remain untouched.")
                .define("hideIronsManaHud", true);
        b.pop();

        SPEC = b.build();
    }

    private BridgeConfig() {}
}
