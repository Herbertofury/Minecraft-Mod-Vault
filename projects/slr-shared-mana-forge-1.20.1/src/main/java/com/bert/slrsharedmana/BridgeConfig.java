package com.bert.slrsharedmana;

import net.minecraftforge.common.ForgeConfigSpec;

public final class BridgeConfig {
    public static final ForgeConfigSpec SPEC;
    public static final ForgeConfigSpec.BooleanValue ENABLED;
    public static final ForgeConfigSpec.BooleanValue SEPARATE_POOLS_BORROW_FROM_IRON;
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
                "Default behavior is the original shared pool: SLR MP is authoritative and Iron uses it on demand.",
                "An opt-in separate-pools mode can instead keep both native pools and let SLR borrow only a real shortage from Iron.",
                "There is deliberately no player-tick mirror and no nearby-block scan in this mod.")
         .push("shared_mana");

        ENABLED = b.comment(
                "Master switch. Default ON for backwards compatibility.",
                "When false, both the classic shared bridge and separate-pool borrowing are disabled.")
                .define("enabled", true);

        SEPARATE_POOLS_BORROW_FROM_IRON = b.comment(
                "Opt-in alternate mode. Default OFF.",
                "When enabled=true and this is true, SLR and Iron keep separate native mana pools.",
                "SLR affordability/spend paths may use Iron mana only for the missing deficit after SLR MP is exhausted.",
                "Iron keeps its native HUD, passive regeneration, events, caps, and normal mana behavior.",
                "No mana is moved while idle, and positive SLR gains never leak into Iron.")
                .define("separatePoolsBorrowFromIron", false);

        SLR_MP_PER_IRON_MANA = b.comment(
                "Conversion scale. 10 means 1 Iron mana == 10 SLR MP.",
                "Used by classic sharing and by the opt-in deficit borrower.",
                "The original 1.21 shared-mana bridge used 10 and this remains the compatibility default.")
                .defineInRange("slrMpPerIronMana", 10.0D, 0.001D, 1_000_000.0D);

        SUPPRESS_ISS_PASSIVE_REGEN = b.comment(
                "Disable Iron's passive mana regeneration only while classic shared mode is active.",
                "Keep this true so SLR's native regen/cooldown rules remain authoritative in classic shared mode.",
                "It has no effect in separate-pools mode; Iron regeneration stays native there.")
                .define("suppressIronPassiveRegen", true);

        ACCEPT_POSITIVE_ISS_CREDITS = b.comment(
                "In classic shared mode, translate intentional positive Iron mana changes into SLR MP.",
                "Recommended true. This lets Iron's Botany ISS_PRIMARY/bidirectional conversion feed the shared pool.",
                "It has no effect in separate-pools mode because Iron keeps its own pool there.")
                .define("acceptPositiveIronManaCredits", true);

        MANA_REFRESH_TICKS = b.comment(
                "SLR native mana_refresh cooldown applied after an Iron mana spend in classic shared mode.",
                "40 ticks (2 seconds) matches the original shared-mana bridge default.")
                .defineInRange("slrManaRefreshTicks", 40, 0, 20 * 60);

        b.pop();
        b.comment(
                "Iron's Botany compatibility",
                "The bridge intentionally does NOT duplicate or replace Iron's Botany's mana router.",
                "In classic shared mode, Iron's Botany keeps full control of its normal manaUnificationMode.",
                "In separate-pools mode, Iron's side remains native and SLR only borrows a shortage when an SLR spend actually occurs.")
         .push("irons_botany");

        LOG_COMPAT_SUMMARY = b.comment(
                "Log the active bridge mode and detected Iron's Botany mode once when a server starts.")
                .define("logDetectedModeOnServerStart", true);
        b.pop();

        b.comment("Client presentation").push("client");
        HIDE_ISS_MANA_HUD = b.comment(
                "Hide Iron's mana bar only in classic shared mode so the SLR MP bar is the single source of truth.",
                "In separate-pools mode Iron's native mana HUD is always retained.")
                .define("hideIronsManaHud", true);
        b.pop();

        SPEC = b.build();
    }

    public static boolean classicSharedMode() {
        return ENABLED.get() && !SEPARATE_POOLS_BORROW_FROM_IRON.get();
    }

    public static boolean separateBorrowMode() {
        return ENABLED.get() && SEPARATE_POOLS_BORROW_FROM_IRON.get();
    }

    public static String modeName() {
        if (!ENABLED.get()) return "DISABLED";
        return SEPARATE_POOLS_BORROW_FROM_IRON.get() ? "SEPARATE_BORROW" : "CLASSIC_SHARED";
    }

    private BridgeConfig() {}
}
