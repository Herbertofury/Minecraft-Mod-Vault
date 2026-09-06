package com.herbertofury.spellbrook.broom;

import java.util.Arrays;

public enum BroomVariant {
    FELINE_FRIEND("feline_friend", "feline_friend_broomstick", 75000, 1.05f, 1.05f, true),
    PICNIC("picnic", "picnic_broom", 75001, 0.98f, 1.12f, true),
    SAKURA("sakura", "sakura_broom", 75002, 1.02f, 1.15f, true),
    SHODDY_1("shoddy_1", "shoddy_old_broomstick_1", 75003, 0.72f, 0.78f, true),
    SHODDY_2("shoddy_2", "shoddy_old_broomstick_2", 75004, 0.76f, 0.80f, true),
    SHODDY_3("shoddy_3", "shoddy_old_broomstick_3", 75005, 0.80f, 0.82f, true),
    CLASSIC_BLUE("classic_blue", "classic_broom_blue", 75006, 1.00f, 1.00f, true),
    CLASSIC_DARK("classic_dark", "classic_broom_dark", 75007, 1.00f, 1.00f, true),
    CLASSIC_GREEN("classic_green", "classic_broom_green", 75008, 1.00f, 1.00f, true),
    CLASSIC_LIGHT("classic_light", "classic_broom_light", 75009, 1.00f, 1.00f, true),
    CLASSIC_ORANGE("classic_orange", "classic_broom_orange", 75010, 1.00f, 1.00f, true),
    CLASSIC_PINK("classic_pink", "classic_broom_pink", 75011, 1.00f, 1.00f, true),
    CLASSIC_PURPLE("classic_purple", "classic_broom_purple", 75012, 1.00f, 1.00f, true),
    CLASSIC_RED("classic_red", "classic_broom_red", 75013, 1.00f, 1.00f, true),
    CLASSIC_RICH("classic_rich", "classic_broom_rich", 75014, 1.04f, 0.98f, true),
    CLASSIC_YELLOW("classic_yellow", "classic_broom_yellow", 75015, 1.00f, 1.00f, true),
    DEFAULT("default", "default", 75016, 0.92f, 1.00f, true),
    DEBUG("debug", "debug_broomstick", 75017, 1.00f, 1.00f, false),
    MOSS("moss", "moss_brooms", 75034, 0.96f, 1.20f, true),
    HEART("heart", "heart_brooms", 75035, 1.04f, 1.12f, true),
    GOTHIC("gothic", "gothic_brooms", 75036, 1.12f, 0.96f, true),
    DRAGON("dragon", "dragon_brooms", 75037, 1.23f, 0.90f, true),
    WING("wing", "wing_brooms", 75038, 1.18f, 1.12f, true),
    FOUNDER("founder", "founder_brooms", 75040, 1.18f, 1.18f, true),
    SOCIETY("society", "society_broom", 75042, 1.10f, 1.16f, true),
    WISP("wisp", "wisp", 75043, 1.12f, 1.24f, true),
    MUSHROOM_MAGE("mushroom_mage", "mushroom_mage_broom", 75045, 1.02f, 1.18f, true),
    MUSHROOM_SHINY_MAGE("mushroom_shiny_mage", "mushroom_shiny_mage_broom", 75047, 1.08f, 1.20f, true);

    private final String id;
    private final String model;
    private final int customModelData;
    private final float speed;
    private final float handling;
    private final boolean publicVariant;

    BroomVariant(String id, String model, int customModelData, float speed, float handling, boolean publicVariant) {
        this.id = id; this.model = model; this.customModelData = customModelData; this.speed = speed; this.handling = handling; this.publicVariant = publicVariant;
    }
    public String id() { return id; }
    public String model() { return model; }
    public int customModelData() { return customModelData; }
    public float speed() { return speed; }
    public float handling() { return handling; }
    public boolean publicVariant() { return publicVariant; }
    public static BroomVariant byId(String id) { return Arrays.stream(values()).filter(v -> v.id.equalsIgnoreCase(id)).findFirst().orElse(DEFAULT); }
}
