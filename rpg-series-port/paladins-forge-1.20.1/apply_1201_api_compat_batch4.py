#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f"missing generated Paladins Java root: {root}")

path = root / "net/paladins/effect/PaladinEffects.java"
text = path.read_text(encoding="utf-8")

# Forge native registration owns the real STATUS_EFFECT registry insertion. Spell Engine's
# 1.20.1 RegistrationBridge therefore returns a direct value holder during the RegisterEvent,
# which intentionally has no RegistryKey. Current Paladins immediately feeds that holder into
# Protection.register(holder, pop), whose 1.20.1 implementation unconditionally unwraps the
# holder key. Preserve the current protection semantics by selecting Spell Engine's explicit
# RegistryKey overload and carrying the same holder/pop payload into Protection.Entry.
import_anchor = "import net.minecraft.registry.Registries;\n"
import_insert = "import net.minecraft.registry.Registries;\nimport net.minecraft.registry.RegistryKey;\nimport net.minecraft.registry.RegistryKeys;\n"
if text.count(import_anchor) != 1:
    raise SystemExit(f"[Divine Protection key bridge] expected exactly one Registries import, found {text.count(import_anchor)}")
if "import net.minecraft.registry.RegistryKey;" in text or "import net.minecraft.registry.RegistryKeys;" in text:
    raise SystemExit("[Divine Protection key bridge] registry-key imports unexpectedly already present; source assumption changed")
text = text.replace(import_anchor, import_insert, 1)

old = '''        Protection.register(DIVINE_PROTECTION.entry, new Protection.Pop(\n                List.of(DivineProtectionStatusEffect.particles),\n                PaladinSounds.divine_protection_impact.soundEvent()\n        ));'''
new = '''        var divineProtectionPop = new Protection.Pop(\n                List.of(DivineProtectionStatusEffect.particles),\n                PaladinSounds.divine_protection_impact.soundEvent()\n        );\n        Protection.register(\n                RegistryKey.of(RegistryKeys.STATUS_EFFECT, DIVINE_PROTECTION.id),\n                new Protection.Entry(DIVINE_PROTECTION.entry, null, 1, divineProtectionPop, divineProtectionPop)\n        );'''
count = text.count(old)
if count != 1:
    raise SystemExit(f"[Divine Protection key bridge] expected exactly one current holder-only registration shape, found {count}")
text = text.replace(old, new, 1)

# Fail closed against the exact runtime regression found by Paladins acceptance run #198.
forbidden = "Protection.register(DIVINE_PROTECTION.entry,"
if forbidden in text:
    raise SystemExit("[Divine Protection key bridge] holder-only Protection.register survived transform")
required = (
    "RegistryKey.of(RegistryKeys.STATUS_EFFECT, DIVINE_PROTECTION.id)",
    "new Protection.Entry(DIVINE_PROTECTION.entry, null, 1, divineProtectionPop, divineProtectionPop)",
)
for marker in required:
    if text.count(marker) != 1:
        raise SystemExit(f"[Divine Protection key bridge] required marker count changed: {marker!r} -> {text.count(marker)}")

path.write_text(text, encoding="utf-8")
print("[Paladins 1.20.1 API batch4] Divine Protection uses explicit status-effect RegistryKey while retaining current holder/pop semantics")
