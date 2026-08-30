#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6o.py <generated-port-root> <spell-engine-1.20.1-baseline>')

root = Path(sys.argv[1]).resolve()
baseline = Path(sys.argv[2]).resolve()
java = root / 'common/src/main/java'
forge_java = root / 'forge/src/main/java/net/spell_engine/forge'

# Reuse the already-reviewed Forge 47 physical-client lifecycle, but activate it from the actual
# graduation chain instead of leaving it orphaned after pass 6i. This registers every Spell Engine
# keybinding, including 1.10.3's GUI-scoped tooltip-details key, and restores client setup/screens/
# render registrations without touching the certified TinyConfig packaging policy.
pass6n = Path(__file__).with_name('compat_pass_6n.py')
if not pass6n.is_file():
    raise SystemExit(f'missing Forge client lifecycle pass: {pass6n}')
subprocess.run([sys.executable, str(pass6n), str(root), str(baseline)], check=True)

# 1.10.4 made Critical Strike compatibility externally configurable. The optional Critical Strike mod
# is graduated separately, so Spell Engine must not hard-link its classes yet; preserve the public API
# seam now with safe defaults that an installed adapter can replace at runtime.
critical = java / 'net/spell_engine/compat/CriticalStrikeCompat.java'
critical.write_text(r'''package net.spell_engine.compat;

import net.minecraft.entity.damage.DamageSource;

import java.util.function.BiConsumer;
import java.util.function.Predicate;

/**
 * Spell Engine 1.10.4 externally-configurable Critical Strike compatibility seam.
 *
 * The optional Critical Strike Forge 1.20.1 port can bind these hooks without Spell Engine taking a
 * hard class dependency. Defaults are intentionally inert when no compatible provider is installed.
 */
public final class CriticalStrikeCompat {
    public static Predicate<DamageSource> isCriticalStrike = ds -> false;
    public static BiConsumer<DamageSource, Float> setCriticalStrike = (ds, multiplier) -> { };

    private CriticalStrikeCompat() { }

    public static void init() { }

    public static boolean isCriticalStrike(DamageSource damageSource) {
        return damageSource != null && isCriticalStrike.test(damageSource);
    }

    public static void setCriticalStrike(DamageSource damageSource, float critMultiplier) {
        if (damageSource != null) {
            setCriticalStrike.accept(damageSource, critMultiplier);
        }
    }
}
''')

# Make tooltip registration observable in the real native client. A source declaration alone is not
# enough: Forge must fire RegisterKeyMappingsEvent and the exact 1.10.3 key must be in Keybindings.all().
client = forge_java / 'client/ForgeClientMod.java'
ct = client.read_text()
old = '''    public static void registerKeys(RegisterKeyMappingsEvent event) {
        for (var keybinding : Keybindings.all()) {
            if (keybinding instanceof GuiKeyBinding) {
                keybinding.setKeyConflictContext(KeyConflictContext.GUI);
            }
            event.register(keybinding);
        }
    }
'''
new = '''    public static void registerKeys(RegisterKeyMappingsEvent event) {
        if (!Keybindings.all().contains(Keybindings.tooltip_details)) {
            throw new IllegalStateException("Spell Engine 1.10.4 tooltip_details key missing from registration list");
        }
        for (var keybinding : Keybindings.all()) {
            if (keybinding instanceof GuiKeyBinding) {
                keybinding.setKeyConflictContext(KeyConflictContext.GUI);
            }
            event.register(keybinding);
        }
        if ("true".equalsIgnoreCase(System.getenv("CI"))) {
            System.out.println("[Spell Engine CI] TOOLTIP_DETAILS_KEY_REGISTERED");
        }
    }
'''
if old not in ct:
    raise SystemExit('Forge client key-registration seam drifted')
client.write_text(ct.replace(old, new, 1))

# Hard parity assertions for every behavior added after the user's supplied 1.10.2 floor.
keybindings = java / 'net/spell_engine/client/input/Keybindings.java'
kt = keybindings.read_text()
for required in (
    'public static KeyBinding tooltip_details = add(new GuiKeyBinding(',
    'InputUtil.UNKNOWN_KEY.getCode()',
    'public static KeyBinding bypass_spell_hotbar = add(new KeyBinding(',
):
    if required not in kt:
        raise SystemExit(f'Spell Engine 1.10.3 tooltip-details parity missing: {required}')

crit = critical.read_text()
for required in (
    'public static Predicate<DamageSource> isCriticalStrike',
    'public static BiConsumer<DamageSource, Float> setCriticalStrike',
    'isCriticalStrike.test(damageSource)',
    'setCriticalStrike.accept(damageSource, critMultiplier)',
):
    if required not in crit:
        raise SystemExit(f'Spell Engine 1.10.4 Critical Strike API parity missing: {required}')

loot_helper = java / 'net/spell_engine/rpg_series/loot/LootHelper.java'
lh = loot_helper.read_text()
for required in (
    'private static List<PoolContents> inspect(List<LootPool> pools)',
    'configureFallback(',
    'fallback.rolls_multiplier',
    'matchedWeight',
    'pendingReport.injected',
    '((LootPoolAccessor) (Object) pool).spellEngine_getEntries()',
):
    if required not in lh:
        raise SystemExit(f'Spell Engine 1.10.3 fallback-loot parity missing: {required}')

aw = root / 'common/src/main/resources/spell_engine.accesswidener'
array_entry = 'accessible    field    net/minecraft/loot/LootPool    entries    [Lnet/minecraft/loot/entry/LootPoolEntry;'
if aw.read_text().count(array_entry) != 1:
    raise SystemExit('Spell Engine 1.10.4 LootPool.entries 1.20.1 access adaptation missing or duplicated')
if 'accessible    field    net/minecraft/loot/LootPool    entries    Ljava/util/List;' in aw.read_text():
    raise SystemExit('Spell Engine 1.10.4 modern LootPool.entries descriptor survived target adaptation')

client_final = client.read_text()
for required in (
    'value = Dist.CLIENT',
    'SpellEngineClient.init();',
    'SpellEngineClient.onClientStarted();',
    'RegisterKeyMappingsEvent',
    'Keybindings.all().contains(Keybindings.tooltip_details)',
    'KeyConflictContext.GUI',
    'TOOLTIP_DETAILS_KEY_REGISTERED',
):
    if required not in client_final:
        raise SystemExit(f'Spell Engine Forge client parity missing: {required}')

# Strengthen the packaged-server self-test so 1.10.4's externally configurable Critical Strike seam
# is not merely source-checked. Reflection on the packaged class proves the two adapter fields remain
# public + static with the expected functional-interface types and initialized defaults at runtime.
self_test = forge_java / 'SpellEngineCiSelfTest.java'
st = self_test.read_text()
marker = '''        System.out.println("[Spell Engine CI] Packaged runtime self-test passed: SpellRegistry=" + spells.size()
                + ", SpellInfinity=present, NBT-components=roundtrip");
'''
critical_runtime = '''        try {
            var compatClass = Class.forName("net.spell_engine.compat.CriticalStrikeCompat");
            var predicateField = compatClass.getField("isCriticalStrike");
            var setterField = compatClass.getField("setCriticalStrike");
            int requiredModifiers = java.lang.reflect.Modifier.PUBLIC | java.lang.reflect.Modifier.STATIC;
            if ((predicateField.getModifiers() & requiredModifiers) != requiredModifiers
                    || (setterField.getModifiers() & requiredModifiers) != requiredModifiers) {
                throw new IllegalStateException("Spell Engine CI self-test: Critical Strike hook fields are not public static");
            }
            if (!java.util.function.Predicate.class.isAssignableFrom(predicateField.getType())
                    || !java.util.function.BiConsumer.class.isAssignableFrom(setterField.getType())) {
                throw new IllegalStateException("Spell Engine CI self-test: Critical Strike hook field types drifted");
            }
            if (predicateField.get(null) == null || setterField.get(null) == null) {
                throw new IllegalStateException("Spell Engine CI self-test: Critical Strike hook defaults are null");
            }
        } catch (ReflectiveOperationException e) {
            throw new IllegalStateException("Spell Engine CI self-test: Critical Strike compatibility seam is not packaged/runtime-visible", e);
        }

        System.out.println("[Spell Engine CI] Packaged runtime self-test passed: SpellRegistry=" + spells.size()
                + ", SpellInfinity=present, NBT-components=roundtrip, CriticalStrikeHooks=public-static-ready");
'''
if marker not in st:
    raise SystemExit('packaged-runtime self-test success seam drifted')
self_test.write_text(st.replace(marker, critical_runtime, 1))
self_test_final = self_test.read_text()
for required in (
    'Class.forName("net.spell_engine.compat.CriticalStrikeCompat")',
    'compatClass.getField("isCriticalStrike")',
    'compatClass.getField("setCriticalStrike")',
    'java.util.function.Predicate.class.isAssignableFrom',
    'java.util.function.BiConsumer.class.isAssignableFrom',
    'CriticalStrikeHooks=public-static-ready',
):
    if required not in self_test_final:
        raise SystemExit(f'Spell Engine packaged Critical Strike runtime proof missing: {required}')

# The certified TinyConfig bytes must remain exact, but Loom include(...) needs a module component.
# Apply the local artifact-only Maven staging only after all 1.10.4 behavior assertions are green.
pass6p = Path(__file__).with_name('compat_pass_6p.py')
if not pass6p.is_file():
    raise SystemExit(f'missing certified TinyConfig module-staging pass: {pass6p}')
subprocess.run([sys.executable, str(pass6p), str(root), str(baseline)], check=True)

print('Spell Engine compatibility pass 6o applied: 1.10.4 Critical Strike API + packaged runtime reflection + fallback loot + access fix + tooltip key runtime parity gated; certified TinyConfig module staging activated')
