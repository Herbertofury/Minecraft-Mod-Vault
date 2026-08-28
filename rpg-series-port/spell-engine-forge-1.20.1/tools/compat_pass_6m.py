#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import subprocess
import sys

work = pathlib.Path(sys.argv[1]).resolve()
java = work / "common/src/main/java"
if not java.is_dir():
    raise SystemExit(f"missing Spell Engine common Java root: {java}")

bridge = java / "net/spell_engine/compat/registry/RegistrationBridge.java"
bridge.parent.mkdir(parents=True, exist_ok=True)
bridge.write_text('''package net.spell_engine.compat.registry;

import net.minecraft.registry.Registry;
import net.minecraft.registry.RegistryKey;
import net.minecraft.registry.entry.RegistryEntry;
import net.minecraft.util.Identifier;

/**
 * Loader-neutral registration seam used by the Forge 1.20.1 RPG Series ports.
 *
 * Common code keeps its upstream construction/configuration flow. A loader may temporarily install
 * a registrar for one exact registry while its native registration event is firing. Calls outside an
 * installed scope retain vanilla Registry.register behavior, so this compatibility seam is inert for
 * ordinary common-code callers.
 */
public final class RegistrationBridge {
    @FunctionalInterface
    public interface Registrar {
        void register(Registry<?> registry, Identifier id, Object value);
    }

    private record State(Registry<?> expectedRegistry, Registrar registrar) { }
    private static final ThreadLocal<State> ACTIVE = new ThreadLocal<>();

    private RegistrationBridge() { }

    public static void withRegistrar(Registry<?> expectedRegistry, Registrar registrar, Runnable action) {
        if (ACTIVE.get() != null) {
            throw new IllegalStateException("Nested native registry bridge scopes are not supported");
        }
        ACTIVE.set(new State(expectedRegistry, registrar));
        try {
            action.run();
        } finally {
            ACTIVE.remove();
        }
    }

    public static <T> T register(Registry<T> registry, Identifier id, T value) {
        var state = ACTIVE.get();
        if (state == null) {
            return Registry.register(registry, id, value);
        }
        requireExpected(state, registry, id);
        state.registrar().register(registry, id, value);
        return value;
    }

    public static <T> T register(Registry<T> registry, RegistryKey<T> key, T value) {
        return register(registry, key.getValue(), value);
    }

    public static <T> RegistryEntry<T> registerReference(Registry<T> registry, Identifier id, T value) {
        var state = ACTIVE.get();
        if (state == null) {
            return Registry.registerReference(registry, id, value);
        }
        requireExpected(state, registry, id);
        state.registrar().register(registry, id, value);
        // 1.20.1 consumers use RegistryEntry as a value holder here. Forge owns the actual registry
        // insertion; a direct entry avoids illegally touching the locked vanilla wrapper a second time.
        return RegistryEntry.of(value);
    }

    private static void requireExpected(State state, Registry<?> actual, Identifier id) {
        if (state.expectedRegistry() != actual) {
            throw new IllegalStateException("Cross-registry registration during native Forge phase: " + id);
        }
    }
}
''', encoding="utf-8")

targets = {
    "net/spell_engine/rpg_series/item/Weapon.java": ("Registry.register(", "RegistrationBridge.register("),
    "net/spell_engine/rpg_series/item/RangedWeapon.java": ("Registry.register(", "RegistrationBridge.register("),
    "net/spell_engine/rpg_series/item/Armor.java": ("Registry.register(", "RegistrationBridge.register("),
    "net/spell_engine/api/effect/Effects.java": ("Registry.registerReference(", "RegistrationBridge.registerReference("),
}

for rel, (old, new) in targets.items():
    path = java / rel
    if not path.is_file():
        raise SystemExit(f"registration bridge target missing: {path}")
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count < 1:
        raise SystemExit(f"registration bridge target no longer contains {old}: {path}")
    text = text.replace(old, new)
    if "import net.spell_engine.compat.registry.RegistrationBridge;" not in text:
        marker = "package " + ".".join(rel[:-5].split("/")[:-1]) + ";\n"
        if marker not in text:
            raise SystemExit(f"could not find package marker in {path}")
        text = text.replace(marker, marker + "\nimport net.spell_engine.compat.registry.RegistrationBridge;\n", 1)
    path.write_text(text, encoding="utf-8")
    print(f"[Spell Engine compat 6m] bridged {count} registry call(s): {rel}")

print("[Spell Engine compat 6m] loader-neutral native registration bridge installed")

# Keep the build driver's established final-pass call stable while extending the migration with a
# separately reviewable client-lifecycle pass. This avoids silently burying loader-client behavior in
# the registry transform and lets 6n stay fail-closed on its own generated output.
next_pass = pathlib.Path(__file__).with_name("compat_pass_6n.py")
if not next_pass.is_file():
    raise SystemExit(f"missing chained Spell Engine client lifecycle pass: {next_pass}")
baseline = pathlib.Path(sys.argv[2]).resolve() if len(sys.argv) > 2 else pathlib.Path('.')
subprocess.run([sys.executable, str(next_pass), str(work), str(baseline)], check=True)
