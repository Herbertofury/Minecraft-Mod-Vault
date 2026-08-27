#!/usr/bin/env python3
from pathlib import Path
import re
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: compat_api_1_20_1.py <generated-wizards-root>')

root = Path(sys.argv[1]).resolve()

def file_at(rel: str) -> Path:
    path = root / rel
    if not path.is_file():
        raise SystemExit(f'Wizards compatibility source missing: {rel}')
    return path

def replace_exact(rel: str, old: str, new: str, expected: int = 1) -> None:
    path = file_at(rel)
    text = path.read_text()
    actual = text.count(old)
    if actual != expected:
        raise SystemExit(f'Wizards compatibility stale transform in {rel}: expected {expected} occurrences, found {actual}: {old[:120]!r}')
    path.write_text(text.replace(old, new))

def add_import(rel: str, import_line: str) -> None:
    path = file_at(rel)
    text = path.read_text()
    if import_line in text:
        return
    match = re.search(r'^package [^;]+;\n', text)
    if not match:
        raise SystemExit(f'Wizards compatibility could not locate package declaration in {rel}')
    insert_at = match.end()
    text = text[:insert_at] + '\n' + import_line + '\n' + text[insert_at:]
    path.write_text(text)

def insert_before_outer_close(rel: str, block: str) -> None:
    path = file_at(rel)
    text = path.read_text().rstrip()
    if not text.endswith('}'):
        raise SystemExit(f'Wizards compatibility could not locate outer class close in {rel}')
    path.write_text(text[:-1].rstrip() + '\n\n' + block.rstrip() + '\n}\n')

# ---------------------------------------------------------------------------
# Rendering contracts changed between 1.20.1 and 1.21.1.
# ---------------------------------------------------------------------------
for rel in (
    'common/src/main/java/net/wizards/client/entity/FrostElementalRenderer.java',
    'common/src/main/java/net/wizards/client/entity/ArcaneEmitterRenderer.java',
):
    replace_exact(
        rel,
        'protected void setupTransforms(',
        'protected void setupTransforms(',
        1,
    )
    path = file_at(rel)
    text = path.read_text()
    old_sig = re.compile(r'protected void setupTransforms\(([^\n]+), float animationProgress, float bodyYaw, float tickDelta, float scale\) \{')
    text, n = old_sig.subn(r'protected void setupTransforms(\1, float animationProgress, float bodyYaw, float tickDelta) {', text)
    if n != 1:
        raise SystemExit(f'Wizards compatibility expected one 1.21 setupTransforms signature in {rel}, found {n}')
    text, n = re.subn(
        r'super\.setupTransforms\(([^\n]+), animationProgress, bodyYaw, tickDelta, scale\);',
        r'super.setupTransforms(\1, animationProgress, bodyYaw, tickDelta);',
        text,
    )
    if n != 1:
        raise SystemExit(f'Wizards compatibility expected one 1.21 setupTransforms super call in {rel}, found {n}')
    path.write_text(text)

model_roots = {
    'common/src/main/java/net/wizards/client/entity/FrostElementalModel.java': 'root',
    'common/src/main/java/net/wizards/client/entity/FireHydraModel.java': 'root',
    'common/src/main/java/net/wizards/client/entity/ArcaneEmitterModel.java': 'arcane_missile_small_portal',
}
for rel, field in model_roots.items():
    replace_exact(
        rel,
        'public void render(MatrixStack matrices, VertexConsumer vertices, int light, int overlay, int color) {',
        'public void render(MatrixStack matrices, VertexConsumer vertices, int light, int overlay, float red, float green, float blue, float alpha) {',
    )
    replace_exact(
        rel,
        f'{field}.render(matrices, vertices, light, overlay, color);',
        f'{field}.render(matrices, vertices, light, overlay, red, green, blue, alpha);',
    )

for rel in (
    'common/src/main/java/net/wizards/client/entity/FrostElementalGlowFeatureRenderer.java',
    'common/src/main/java/net/wizards/client/entity/FireHydraGlowFeatureRenderer.java',
):
    replace_exact(
        rel,
        'getContextModel().render(matrices, vertexConsumer, 15728640, OverlayTexture.DEFAULT_UV);',
        'getContextModel().render(matrices, vertexConsumer, 15728640, OverlayTexture.DEFAULT_UV, 1F, 1F, 1F, 1F);',
    )

# ---------------------------------------------------------------------------
# 1.20.1 Tameable still exposes the intermediary-named EntityView bridge.
# Keep the already-graduated Spell Engine artifact untouched and satisfy the
# leaf contract in Wizards' three current summoned entities.
# ---------------------------------------------------------------------------
for class_name in ('FrostElementalEntity', 'FireHydraEntity', 'ArcaneEmitterEntity'):
    rel = f'common/src/main/java/net/wizards/entity/{class_name}.java'
    add_import(rel, 'import net.minecraft.world.EntityView;')
    path = file_at(rel)
    text = path.read_text()
    if 'EntityView method_48926()' not in text:
        insert_before_outer_close(rel, '''    @Override
    public EntityView method_48926() {
        return getWorld();
    }''')

# ---------------------------------------------------------------------------
# Attribute IDs + modifier operation names.
# ---------------------------------------------------------------------------
effects_rel = 'common/src/main/java/net/wizards/effect/WizardsEffects.java'
add_import(effects_rel, 'import net.minecraft.registry.Registries;')
effects = file_at(effects_rel)
text = effects.read_text()
text, ids = re.subn(
    r'EntityAttributes\.([A-Z0-9_]+)\.getIdAsString\(\)',
    r'Registries.ATTRIBUTE.getId(EntityAttributes.\1).toString()',
    text,
)
if ids != 4:
    raise SystemExit(f'Wizards compatibility expected 4 vanilla attribute ID rewrites in WizardsEffects, found {ids}')
text = text.replace('EntityAttributeModifier.Operation.ADD_MULTIPLIED_BASE', 'EntityAttributeModifier.Operation.MULTIPLY_BASE')
jump_block = '''                            new AttributeModifier(
                                    Registries.ATTRIBUTE.getId(EntityAttributes.GENERIC_JUMP_STRENGTH).toString(),
                                    -10,
                                    EntityAttributeModifier.Operation.MULTIPLY_BASE
                            )'''
if jump_block not in text:
    raise SystemExit('Wizards compatibility could not locate 1.21 generic jump-strength frozen modifier')
# Preserve frozen jump blocking via LivingEntityFrozen; 1.20.1 has no generic living jump-strength attribute.
text = text.replace(',\n' + jump_block, '')
effects.write_text(text)

summons_rel = 'common/src/main/java/net/wizards/entity/WizardSummons.java'
add_import(summons_rel, 'import net.minecraft.registry.Registries;')
summons = file_at(summons_rel)
text = summons.read_text()
text, ids = re.subn(
    r'EntityAttributes\.([A-Z0-9_]+)\.getIdAsString\(\)',
    r'Registries.ATTRIBUTE.getId(EntityAttributes.\1).toString()',
    text,
)
if ids != 5:
    raise SystemExit(f'Wizards compatibility expected 5 summon attribute ID rewrites, found {ids}')
text = text.replace('EntityAttributeModifier.Operation.ADD_VALUE', 'EntityAttributeModifier.Operation.ADDITION')
summons.write_text(text)

# ---------------------------------------------------------------------------
# Identifier and EntityType builder API drift.
# ---------------------------------------------------------------------------
replace_exact(
    'common/src/main/java/net/wizards/content/WizardSpells.java',
    'Identifier.of("spell_power:critical_chance")',
    'new Identifier("spell_power:critical_chance")',
)
replace_exact(
    'common/src/main/java/net/wizards/content/WizardSpells.java',
    'Identifier.of("spell_engine:damage_taken")',
    'new Identifier("spell_engine:damage_taken")',
)
replace_exact(
    'common/src/main/java/net/wizards/item/WizardWeapons.java',
    'var id = Identifier.of(idString);',
    'var id = new Identifier(idString);',
)
entities_rel = 'common/src/main/java/net/wizards/entity/WizardEntities.java'
entities = file_at(entities_rel)
text = entities.read_text()
count = text.count('.dimensions(')
if count != 3:
    raise SystemExit(f'Wizards compatibility expected 3 EntityType dimensions calls, found {count}')
entities.write_text(text.replace('.dimensions(', '.setDimensions('))

# ---------------------------------------------------------------------------
# Status-effect API: direct StatusEffect instead of RegistryEntry<StatusEffect>.
# ---------------------------------------------------------------------------
frozen_rel = 'common/src/main/java/net/wizards/mixin/effect/LivingEntityFrozen.java'
frozen = file_at(frozen_rel)
text = frozen.read_text()
text = text.replace('import net.minecraft.registry.entry.RegistryEntry;\n', '')
old_shadow = '@Shadow public abstract boolean hasStatusEffect(RegistryEntry<StatusEffect> effect);'
if text.count(old_shadow) != 1:
    raise SystemExit('Wizards compatibility could not locate LivingEntityFrozen RegistryEntry shadow')
text = text.replace(old_shadow, '@Shadow public abstract boolean hasStatusEffect(StatusEffect effect);')
if text.count('WizardsEffects.frozen.entry') != 2:
    raise SystemExit('Wizards compatibility expected two frozen RegistryEntry effect uses')
text = text.replace('WizardsEffects.frozen.entry', 'WizardsEffects.frozen.effect')
frozen.write_text(text)

replace_exact(
    'common/src/main/java/net/wizards/mixin/effect/LivingEntityFrostShield.java',
    'WizardsEffects.frostShield.entry',
    'WizardsEffects.frostShield.effect',
)

# ---------------------------------------------------------------------------
# Villager trades. 1.20.1 lacks BuyItemFactory and its multiplier-aware
# SellItemFactory overload accepts ItemStack. Preserve current economics exactly.
# ---------------------------------------------------------------------------
villagers_rel = 'common/src/main/java/net/wizards/villager/WizardVillagers.java'
add_import(villagers_rel, 'import net.minecraft.item.ItemStack;')
add_import(villagers_rel, 'import net.minecraft.village.TradeOffer;')
for rune in ('ARCANE', 'FIRE', 'FROST'):
    replace_exact(
        villagers_rel,
        f'new TradeOffers.SellItemFactory(RuneItems.get(RuneItems.RuneType.{rune}), 2, 8, 128, 3, 0.1f)',
        f'new TradeOffers.SellItemFactory(new ItemStack(RuneItems.get(RuneItems.RuneType.{rune})), 2, 8, 128, 3, 0.1f)',
    )
replace_exact(
    villagers_rel,
    'new TradeOffers.BuyItemFactory(Items.WHITE_WOOL, 10, 12, 5, 6)',
    '(entity, random) -> new TradeOffer(new ItemStack(Items.WHITE_WOOL, 10), new ItemStack(Items.EMERALD, 6), 12, 5, 0.05F)',
)
replace_exact(
    villagers_rel,
    'new TradeOffers.BuyItemFactory(Items.LAPIS_LAZULI, 6, 3, 5, 12)',
    '(entity, random) -> new TradeOffer(new ItemStack(Items.LAPIS_LAZULI, 6), new ItemStack(Items.EMERALD, 12), 3, 5, 0.05F)',
)
for piece in ('head', 'feet'):
    replace_exact(
        villagers_rel,
        f'new TradeOffers.SellItemFactory(WizardArmors.wizardRobeSet.{piece}, 15, 1, 12, 16, 0.1F)',
        f'new TradeOffers.SellItemFactory(new ItemStack(WizardArmors.wizardRobeSet.{piece}), 15, 1, 12, 16, 0.1F)',
    )
for piece in ('chest', 'legs'):
    replace_exact(
        villagers_rel,
        f'new TradeOffers.SellItemFactory(WizardArmors.wizardRobeSet.{piece}, 20, 1, 12, 16, 0.1F)',
        f'new TradeOffers.SellItemFactory(new ItemStack(WizardArmors.wizardRobeSet.{piece}), 20, 1, 12, 16, 0.1F)',
    )

# ---------------------------------------------------------------------------
# ArmorMaterial moved from an interface (1.20.1) to a registered data object.
# Recreate the old proven custom-material contract while retaining all current
# 3.1.1 material values, current custom equip sound, repair ingredients, and IDs.
# ---------------------------------------------------------------------------
armors_rel = 'common/src/main/java/net/wizards/item/WizardArmors.java'
armors = file_at(armors_rel)
text = armors.read_text()
old_material = '''    public static RegistryEntry<ArmorMaterial> material(String name,
                                         int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,
                                         int enchantability, RegistryEntry<SoundEvent> equipSound, Supplier<Ingredient> repairIngredient) {
        var material = new ArmorMaterial(
                Map.of(
                ArmorItem.Type.HELMET, protectionHead,
                ArmorItem.Type.CHESTPLATE, protectionChest,
                ArmorItem.Type.LEGGINGS, protectionLegs,
                ArmorItem.Type.BOOTS, protectionFeet),
                enchantability, equipSound, repairIngredient,
                List.of(new ArmorMaterial.Layer(Identifier.of(WizardsMod.ID, name))),
                0,0
                );
        return Registry.registerReference(Registries.ARMOR_MATERIAL, Identifier.of(WizardsMod.ID, name), material);
    }
'''
new_material = '''    private static final class WizardArmorMaterial implements ArmorMaterial {
        private final String name;
        private final int durabilityMultiplier;
        private final int protectionHead;
        private final int protectionChest;
        private final int protectionLegs;
        private final int protectionFeet;
        private final int enchantability;
        private final RegistryEntry<SoundEvent> equipSound;
        private final Supplier<Ingredient> repairIngredient;

        private WizardArmorMaterial(String name, int durabilityMultiplier,
                                    int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,
                                    int enchantability, RegistryEntry<SoundEvent> equipSound,
                                    Supplier<Ingredient> repairIngredient) {
            this.name = name;
            this.durabilityMultiplier = durabilityMultiplier;
            this.protectionHead = protectionHead;
            this.protectionChest = protectionChest;
            this.protectionLegs = protectionLegs;
            this.protectionFeet = protectionFeet;
            this.enchantability = enchantability;
            this.equipSound = equipSound;
            this.repairIngredient = repairIngredient;
        }

        @Override
        public int getDurability(ArmorItem.Type type) {
            int base = switch (type) {
                case BOOTS -> 13;
                case LEGGINGS -> 15;
                case CHESTPLATE -> 16;
                case HELMET -> 11;
            };
            return base * durabilityMultiplier;
        }

        @Override
        public int getProtection(ArmorItem.Type type) {
            return switch (type) {
                case HELMET -> protectionHead;
                case CHESTPLATE -> protectionChest;
                case LEGGINGS -> protectionLegs;
                case BOOTS -> protectionFeet;
            };
        }

        @Override
        public int getEnchantability() {
            return enchantability;
        }

        @Override
        public SoundEvent getEquipSound() {
            return equipSound.value();
        }

        @Override
        public Ingredient getRepairIngredient() {
            return repairIngredient.get();
        }

        @Override
        public String getName() {
            // Proven 1.20.1 Spell Engine convention: plain material name; Armor.Entry owns namespace/id.
            return name;
        }

        @Override
        public float getToughness() {
            return 0F;
        }

        @Override
        public float getKnockbackResistance() {
            return 0F;
        }
    }

    public static RegistryEntry<ArmorMaterial> material(String name, int durabilityMultiplier,
                                         int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,
                                         int enchantability, RegistryEntry<SoundEvent> equipSound, Supplier<Ingredient> repairIngredient) {
        return new RegistryEntry.Direct<>(new WizardArmorMaterial(
                name, durabilityMultiplier,
                protectionHead, protectionChest, protectionLegs, protectionFeet,
                enchantability, equipSound, repairIngredient));
    }
'''
if text.count(old_material) != 1:
    raise SystemExit('Wizards compatibility could not locate current 1.21 ArmorMaterial factory')
text = text.replace(old_material, new_material)
for name, durability in (
    ('wizard_robe', 10),
    ('arcane_robe', 20),
    ('fire_robe', 20),
    ('frost_robe', 20),
    ('netherite_arcane_robe', 30),
    ('netherite_fire_robe', 30),
    ('netherite_frost_robe', 30),
):
    old = f'            "{name}",\n            1, 3, 2, 1,'
    new = f'            "{name}",\n            {durability},\n            1, 3, 2, 1,'
    if text.count(old) != 1:
        raise SystemExit(f'Wizards compatibility could not locate material durability insertion for {name}')
    text = text.replace(old, new)
armors.write_text(text)

# ---------------------------------------------------------------------------
# Hard guards: these were concrete run-129 failure signatures. A surviving one
# means the compatibility pass is stale/incomplete and must fail before Gradle.
# ---------------------------------------------------------------------------
guards = {
    effects_rel: ['getIdAsString()', 'ADD_MULTIPLIED_BASE', 'GENERIC_JUMP_STRENGTH'],
    summons_rel: ['getIdAsString()', 'ADD_VALUE'],
    villagers_rel: ['BuyItemFactory'],
    entities_rel: ['.dimensions('],
    armors_rel: ['Registries.ARMOR_MATERIAL', 'new ArmorMaterial('],
    frozen_rel: ['RegistryEntry<StatusEffect>', 'WizardsEffects.frozen.entry'],
    'common/src/main/java/net/wizards/mixin/effect/LivingEntityFrostShield.java': ['WizardsEffects.frostShield.entry'],
    'common/src/main/java/net/wizards/content/WizardSpells.java': ['Identifier.of("spell_power:critical_chance")', 'Identifier.of("spell_engine:damage_taken")'],
    'common/src/main/java/net/wizards/item/WizardWeapons.java': ['Identifier.of(idString)'],
}
for rel, banned in guards.items():
    text = file_at(rel).read_text()
    survivors = [token for token in banned if token in text]
    if survivors:
        raise SystemExit(f'Wizards 1.21-only API survived compatibility pass in {rel}: {survivors}')

for rel in (
    'common/src/main/java/net/wizards/client/entity/FrostElementalRenderer.java',
    'common/src/main/java/net/wizards/client/entity/ArcaneEmitterRenderer.java',
):
    if 'float tickDelta, float scale)' in file_at(rel).read_text():
        raise SystemExit(f'Wizards 1.21 six-argument setupTransforms survived in {rel}')

print('Wizards compatibility pass 2 applied: MC 1.20.1 render/entity/trade/attribute/identifier/effect/armor API bridge')
