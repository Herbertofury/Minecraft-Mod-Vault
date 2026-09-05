#!/usr/bin/env python3
from __future__ import annotations
import pathlib, sys

root = pathlib.Path(sys.argv[1]).resolve()
path = root / "net/rogues/mixin/LivingEntityStealth.java"
text = path.read_text(encoding="utf-8")
replacements = (
    ("import net.minecraft.registry.entry.RegistryEntry;\n", ""),
    ("target = \"Lnet/minecraft/entity/LivingEntity;hasStatusEffect(Lnet/minecraft/registry/entry/RegistryEntry;)Z\"", "target = \"Lnet/minecraft/entity/LivingEntity;hasStatusEffect(Lnet/minecraft/entity/effect/StatusEffect;)Z\""),
    ("LivingEntity instance, RegistryEntry<StatusEffect> effect, Operation<Boolean> original", "LivingEntity instance, StatusEffect effect, Operation<Boolean> original"),
)
for old, new in replacements:
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[Rogues 1.20.1 compat batch3] expected one historical mixin ABI seam, found {count}: {old}")
    text = text.replace(old, new, 1)
path.write_text(text, encoding="utf-8")
if "registry/entry/RegistryEntry" in text or "RegistryEntry<StatusEffect>" in text:
    raise SystemExit("[Rogues 1.20.1 compat batch3] holder-form status effect target survived")
print("[Rogues 1.20.1 compat batch3] LivingEntityStealth updatePotionVisibility target restored to historical 1.20.1 raw StatusEffect ABI")
