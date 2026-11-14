#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_enchantment_wave7.py <prepared-port-root>')

root = Path(sys.argv[1]).resolve()
forge_mod = root / 'forge/src/main/java/net/more_rpg_classes/forge/ForgeMod.java'
mrpg_mod = root / 'common/src/main/java/net/more_rpg_classes/MRPGCMod.java'
compat_java = root / 'common/src/main/java/net/more_rpg_classes/compat/MoreRpg1201Enchantments.java'

for path in (forge_mod, mrpg_mod):
    if not path.is_file():
        raise SystemExit(f'More RPG enchantment wave7 input missing: {path}')
if compat_java.exists():
    raise SystemExit(f'More RPG enchantment wave7 unexpectedly already present: {compat_java}')

forge = forge_mod.read_text()
common = mrpg_mod.read_text()
forge_contracts = {
    'import net.more_rpg_classes.MRPGCMod;': 1,
    'event.register(RegistryKeys.ITEM, helper -> MRPGCMod.registerItems());': 1,
    'event.register(RegistryKeys.STATUS_EFFECT, helper -> MRPGCMod.registerEffects());': 1,
}
common_contracts = {
    'import net.more_rpg_classes.compat.CriticalStrikeCompat;': 1,
    'MoreSpellSchools.initialize();': 1,
}
for needle, expected in forge_contracts.items():
    actual = forge.count(needle)
    if actual != expected:
        raise SystemExit(f'More RPG enchantment wave7 Forge precondition drifted for {needle!r}: expected {expected}, found {actual}')
for needle, expected in common_contracts.items():
    actual = common.count(needle)
    if actual != expected:
        raise SystemExit(f'More RPG enchantment wave7 common precondition drifted for {needle!r}: expected {expected}, found {actual}')
for forbidden in ('MoreRpg1201Enchantments', 'RegistryKeys.ENCHANTMENT'):
    if forbidden in forge or forbidden in common:
        raise SystemExit(f'More RPG enchantment wave7 unexpectedly already wired: {forbidden}')

forge = forge.replace(
    'import net.more_rpg_classes.MRPGCMod;\n',
    'import net.more_rpg_classes.MRPGCMod;\nimport net.more_rpg_classes.compat.MoreRpg1201Enchantments;\n',
    1,
)
forge = forge.replace(
    '        event.register(RegistryKeys.ITEM, helper -> MRPGCMod.registerItems());\n'
    '        event.register(RegistryKeys.STATUS_EFFECT, helper -> MRPGCMod.registerEffects());\n',
    '        event.register(RegistryKeys.ITEM, helper -> MRPGCMod.registerItems());\n'
    '        // 1.21 made More RPG enchantments data-driven. Forge 1.20.1 still requires an\n'
    '        // ENCHANTMENT registry entry, so recreate only those two upstream 2.7.2 definitions.\n'
    '        event.register(RegistryKeys.ENCHANTMENT, helper -> MoreRpg1201Enchantments.register());\n'
    '        event.register(RegistryKeys.STATUS_EFFECT, helper -> MRPGCMod.registerEffects());\n',
    1,
)
common = common.replace(
    'import net.more_rpg_classes.compat.CriticalStrikeCompat;\n',
    'import net.more_rpg_classes.compat.CriticalStrikeCompat;\nimport net.more_rpg_classes.compat.MoreRpg1201Enchantments;\n',
    1,
)
common = common.replace(
    '\t\t\tMoreSpellSchools.initialize();\n',
    '\t\t\tMoreSpellSchools.initialize();\n\t\t\tMoreRpg1201Enchantments.attachSchoolSources();\n',
    1,
)
forge_mod.write_text(forge)
mrpg_mod.write_text(common)

compat_java.parent.mkdir(parents=True, exist_ok=True)
compat_java.write_text(r'''package net.more_rpg_classes.compat;

import net.minecraft.enchantment.Enchantment;
import net.minecraft.enchantment.EnchantmentTarget;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.item.ItemStack;
import net.minecraft.registry.Registries;
import net.minecraft.registry.Registry;
import net.minecraft.util.Identifier;
import net.more_rpg_classes.MRPGCMod;
import net.more_rpg_classes.custom.MoreSpellSchools;
import net.spell_power.SpellPowerMod;
import net.spell_power.api.SpellSchool;
import net.spell_power.config.EnchantmentsConfig;
import net.spell_power.internals.AmplifierEnchantment;
import net.spell_power.internals.SchoolFilteredEnchantment;

import java.util.Set;

/**
 * Forge 1.20.1 downgrade for the two enchantments that More RPG Library 2.7.2 defines as
 * 1.21 data-driven enchantments. Values are copied from the pinned 2.7.2 JSON authority:
 * Typhoon = Air + Water, Stonebloom = Earth + Nature, +3% base power per level, max level 5.
 */
public final class MoreRpg1201Enchantments {
    private static final EquipmentSlot[] ARMOR = {
            EquipmentSlot.HEAD, EquipmentSlot.CHEST, EquipmentSlot.LEGS, EquipmentSlot.FEET
    };

    public static final Identifier TYPHOON_ID = MRPGCMod.id("typhoon");
    public static final Identifier STONEBLOOM_ID = MRPGCMod.id("stonebloom");

    public static final SchoolFilteredEnchantment TYPHOON = specialized(
            Set.of(MoreSpellSchools.AIR, MoreSpellSchools.WATER),
            MRPGCMod.id("enchantable/typhoon"));
    public static final SchoolFilteredEnchantment STONEBLOOM = specialized(
            Set.of(MoreSpellSchools.EARTH, MoreSpellSchools.NATURE),
            MRPGCMod.id("enchantable/stonebloom"));

    private static boolean schoolSourcesAttached;

    private MoreRpg1201Enchantments() {}

    private static SchoolFilteredEnchantment specialized(Set<SpellSchool> schools, Identifier requiredTag) {
        return new ModernSpecializedEnchantment(schools, requiredTag);
    }

    public static void register() {
        Registry.register(Registries.ENCHANTMENT, TYPHOON_ID, TYPHOON);
        Registry.register(Registries.ENCHANTMENT, STONEBLOOM_ID, STONEBLOOM);
    }

    /** Mirrors Spell Power 1.6 specialized-enchantment power semantics on 1.20.1. */
    public static void attachSchoolSources() {
        if (schoolSourcesAttached) {
            return;
        }
        schoolSourcesAttached = true;
        attach(TYPHOON, MoreSpellSchools.AIR);
        attach(TYPHOON, MoreSpellSchools.WATER);
        attach(STONEBLOOM, MoreSpellSchools.EARTH);
        attach(STONEBLOOM, MoreSpellSchools.NATURE);
    }

    private static void attach(SchoolFilteredEnchantment enchantment, SpellSchool school) {
        school.addSource(SpellSchool.Trait.POWER, new SpellSchool.Source(SpellSchool.Apply.ADD, query -> {
            int level = SpellPowerMod.armorEnchantmentLevel(enchantment, query.entity());
            return school.attribute.getDefaultValue() * enchantment.config.bonus_per_level * level;
        }));
    }

    /**
     * Avoid depending on Spell Power's generated EnchantmentRestriction helper. The public sealed
     * SchoolFilteredEnchantment API already exposes the required-tag and relevant-school checks.
     */
    private static final class ModernSpecializedEnchantment extends SchoolFilteredEnchantment {
        private ModernSpecializedEnchantment(Set<SpellSchool> schools, Identifier requiredTag) {
            super(Enchantment.Rarity.RARE,
                    AmplifierEnchantment.Operation.ADD,
                    new EnchantmentsConfig.PowerEnchantmentConfig(true, 5, 1, 11, 12, 11, 0.03F),
                    schools,
                    EnchantmentTarget.BREAKABLE,
                    ARMOR);
            requireTag(requiredTag);
        }

        @Override
        public boolean isAcceptableItem(ItemStack stack) {
            boolean matchingAttributesRequired =
                    SpellPowerMod.attributesConfig.value.enchantments_require_matching_attribute;
            boolean schoolMatches = !matchingAttributesRequired
                    || SchoolFilteredEnchantment.schoolsIntersect(poweredSchools(), stack);
            return matchesRequiredTag(stack) && schoolMatches && super.isAcceptableItem(stack);
        }
    }
}
''')

forge_after = forge_mod.read_text()
common_after = mrpg_mod.read_text()
compat_after = compat_java.read_text()
contracts = {
    'RegistryKeys.ENCHANTMENT': (forge_after.count('RegistryKeys.ENCHANTMENT'), 1),
    'MoreRpg1201Enchantments.register()': (forge_after.count('MoreRpg1201Enchantments.register()'), 1),
    'MoreRpg1201Enchantments.attachSchoolSources()': (common_after.count('MoreRpg1201Enchantments.attachSchoolSources()'), 1),
    'MRPGCMod.id("typhoon")': (compat_after.count('MRPGCMod.id("typhoon")'), 1),
    'MRPGCMod.id("stonebloom")': (compat_after.count('MRPGCMod.id("stonebloom")'), 1),
    'Set.of(MoreSpellSchools.AIR, MoreSpellSchools.WATER)': (compat_after.count('Set.of(MoreSpellSchools.AIR, MoreSpellSchools.WATER)'), 1),
    'Set.of(MoreSpellSchools.EARTH, MoreSpellSchools.NATURE)': (compat_after.count('Set.of(MoreSpellSchools.EARTH, MoreSpellSchools.NATURE)'), 1),
    'new EnchantmentsConfig.PowerEnchantmentConfig(true, 5, 1, 11, 12, 11, 0.03F)': (compat_after.count('new EnchantmentsConfig.PowerEnchantmentConfig(true, 5, 1, 11, 12, 11, 0.03F)'), 1),
    'SpellPowerMod.armorEnchantmentLevel': (compat_after.count('SpellPowerMod.armorEnchantmentLevel'), 1),
    'isAcceptableItem(ItemStack stack)': (compat_after.count('isAcceptableItem(ItemStack stack)'), 1),
}
for needle, (actual, expected) in contracts.items():
    if actual != expected:
        raise SystemExit(f'More RPG enchantment wave7 contract failed for {needle!r}: expected {expected}, found {actual}')
if compat_after.count('attach(TYPHOON,') != 2 or compat_after.count('attach(STONEBLOOM,') != 2:
    raise SystemExit('More RPG enchantment wave7 did not attach exactly two schools to each enchantment')

print('[More RPG 2.7.2] ENCHANTMENT_DATA_1201_DOWNGRADE_PASS '
      'typhoon=air+water stonebloom=earth+nature bonus=0.03 max_level=5 armor_only=true specialized_exclusion=true')
