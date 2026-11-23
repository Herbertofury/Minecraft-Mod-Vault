# Spell Engine 1.10.2 -> Native Forge 1.20.1

Immutable source boundaries:

- Minecraft 1.20.1 substrate: `ZsoltMolnarrr/SpellEngine@8721120169ddefd230fc73fc7c332318a92f6c7c` (`0.15.12+1.20.1`).
- 1.10.2 target state: `ZsoltMolnarrr/SpellEngine@bc02f7a49da950503010020da491f6bdc5871df7`.

Port policy:

1. Preserve the 1.10.2 public/data behavior wherever Minecraft 1.20.1 can express it.
2. Use explicit 1.20.1 compatibility layers for post-1.20 APIs (notably data components, custom payload networking, registry/event differences).
3. Depend on the already verified native Forge 1.20.1 Spell Power 1.6.0 and Ranged Weapon API 2.3.4 ports.
4. No Fabric or NeoForge runtime leakage in the final Forge JAR.
5. Acceptance is build/package + client resource/bootstrap + clean standalone Forge 47.4.23 packaged-JAR server runtime invariants.
6. Keep immutable upstream source references and deterministic reconstruction rather than vendoring unrelated upstream history.
