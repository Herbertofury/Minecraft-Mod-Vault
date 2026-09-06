package com.herbertofury.spellbrook.broom;

import java.util.Arrays;

public enum CoreVariant {
    SPELLSPHERE("spellsphere", "spellsphere_core_crystal", 75019, 1.00f, 1.05f, true),
    AMETHYST("amethyst", "core_crystal_amethyst", 75020, 0.98f, 1.12f, true),
    AQUAMARINE("aquamarine", "core_crystal_aquamarine", 75021, 0.96f, 1.18f, true),
    BLOODSTONE("bloodstone", "core_crystal_bloodstone", 75022, 1.12f, 0.94f, true),
    CRYSTAL("crystal", "core_crystal_crystal", 75023, 1.00f, 1.00f, true),
    DIAMOND("diamond", "core_crystal_diamond", 75024, 1.08f, 1.05f, true),
    EMERALD("emerald", "core_crystal_emerald", 75025, 1.04f, 1.08f, true),
    GARNET("garnet", "core_crystal_garnet", 75026, 1.08f, 1.00f, true),
    MOONSTONE("moonstone", "core_crystal_moonstone", 75027, 1.04f, 1.16f, true),
    OPAL("opal", "core_crystal_opal", 75028, 1.02f, 1.12f, true),
    PEARL("pearl", "core_crystal_pearl", 75029, 0.98f, 1.16f, true),
    RUBY("ruby", "core_crystal_ruby", 75030, 1.12f, 0.98f, true),
    SAPPHIRE("sapphire", "core_crystal_sapphire", 75031, 1.04f, 1.10f, true),
    TOPAZ("topaz", "core_crystal_topaz", 75032, 1.08f, 1.04f, true),
    FOUNDER("founder", "founder_crystal_core", 75039, 1.15f, 1.15f, true),
    SOCIETY("society", "society_core_crystal", 75041, 1.10f, 1.14f, true),
    WISP("wisp", "wisp_core", 75044, 1.08f, 1.22f, true),
    MUSHROOM("mushroom", "mushroom_core", 75046, 1.02f, 1.18f, true),
    MUSHROOM_SHINY("mushroom_shiny", "mushroom_shiny_core", 75048, 1.08f, 1.20f, true),
    DEBUG("debug", "debug_core", 75018, 1.00f, 1.00f, false);

    private final String id;
    private final String model;
    private final int customModelData;
    private final float speedMultiplier;
    private final float handlingMultiplier;
    private final boolean publicVariant;

    CoreVariant(String id, String model, int customModelData, float speedMultiplier, float handlingMultiplier, boolean publicVariant) {
        this.id=id; this.model=model; this.customModelData=customModelData; this.speedMultiplier=speedMultiplier; this.handlingMultiplier=handlingMultiplier; this.publicVariant=publicVariant;
    }
    public String id(){return id;} public String model(){return model;} public int customModelData(){return customModelData;}
    public float speedMultiplier(){return speedMultiplier;} public float handlingMultiplier(){return handlingMultiplier;} public boolean publicVariant(){return publicVariant;}
    public static CoreVariant byId(String id){ return Arrays.stream(values()).filter(v -> v.id.equalsIgnoreCase(id)).findFirst().orElse(CRYSTAL); }
}
