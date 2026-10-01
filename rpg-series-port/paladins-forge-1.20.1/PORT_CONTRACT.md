# Paladins 3.1.1 -> Native Forge 1.20.1 Port Contract

Status: **GRADUATED**. Canonical post-promotion deep acceptance run #212 / `33215195254` passed on executable authority commit `257171c71a363285bb5bdbb58083121f3a7456d3`. Rogues is the next RPG Series leaf.

## Graduation authority

- Canonical runtime authority commit: `257171c71a363285bb5bdbb58083121f3a7456d3`
- Pre-promotion exact repair head: `1a9c93399e377b2461632ffedcfa3403ac5338c6`
- Pre-promotion deep acceptance: run #211 / `33214337111` — `FULL_DEEP_BEHAVIOR_PASS`
- Canonical post-promotion deep acceptance: run #212 / `33215195254` — `FULL_DEEP_BEHAVIOR_PASS`
- Certified Paladins release SHA-256: `95e8f9e074dd1c432f9486c922173b76a3b3bc57667b28fe154e47dba5c374ee`
- Certified deterministic source ZIP SHA-256: `fb0e5812857a2fd46de488cd17a80011ef5d18795ff96fa1a3ebed5fd19a4377`
- Run #212 evidence ZIP SHA-256: `152e65f28022d81767e85ee5fd7e905b8d3b73bb3aed53e78fd56ae3c620235c`
- Run #212 evidence ZIP size: `16,710,637` bytes
- Run #212 proved fresh Forge 47.4.23 packaged-server readiness, native LWJGL client join through the real `ConnectScreen` network stack, Divine Protection one/two-charge damage interception, Judgement/STUN input blocking plus positive post-clear controls, Priest Absorption lifecycle, Levitate gravity/Slow Falling semantics, reproducible release identity, and deterministic source identity.
- Canonical Drive evidence file: `paladins-run212-canonical-graduation-evidence.zip`, file ID `1t9ASV5-M_2rVtoHd96zyblGdrHA3zW4i`; Drive round-trip size and SHA-256 verification passed.

## Feature authority

- Upstream: `ZsoltMolnarrr/Paladins`
- Current 1.21.1 commit: `9d5611d3799c56951255fc3e3e61aee4233f3d28`
- Current tree: `be72e53347fa0bbae33515370d374906d192c2d8`
- Current version: `3.1.1`
- Current Minecraft/Yarn: `1.21.1` / `1.21.1+build.3`
- Current loader architecture: Fabric + NeoForge / Architectury
- Java source authority: current 3.1.1. Do not replace current content with historical 1.20.1 content.

## Target mapping substrate

- Historical Paladins 1.20.1 branch commit: `f310c0b12eb3791c6e83f6bda0accdf032aa8a17`
- Historical version: `1.4.0`
- Historical Minecraft/Yarn: `1.20.1` / `1.20.1+build.10`
- Use only for target-era mappings, signatures, assets/data comparisons, and behavioral archaeology where current APIs moved.

## Target runtime

- Minecraft `1.20.1`
- Forge `47.4.23`
- Yarn `1.20.1+build.10`
- Java `17`
- Native Forge release artifact. No Fabric runtime, no NeoForge runtime, no Architectury/Fabric API leakage in the packaged boundary unless an already-graduated shared common ABI requires compile-time annotations only.

## Known current dependency surface

Current 3.1.1 declares:

- Shield API `2.1.0`
- Armor Model API `1.0.0+1.21.1`
- TinyConfig `3.1.0`
- Runes `1.3.1+1.21.1`
- Structure Pool API `1.2.0+1.21.1`
- Spell Power `1.6.0+1.21.1`
- Spell Engine `1.10.0+1.21.1`
- Cloth Config `15.0.130`
- Player Animator `2.0.1+1.21.1`
- Curios `9.5.1+1.21.1` on NeoForge.

Prefer the already-graduated 1.20.1 foundations in this project where their public contract satisfies current Paladins behavior. Port/graduate any genuinely missing foundation before weakening Paladins behavior.

## Shield API source authority + substrate

- Source project: `FabricExtras/ShieldAPI`
- **Exact Paladins-pinned Shield API 2.1.0 tag commit:** `bccbc4fded8956d16cf7bebb24e5cfd3d2f91347`
- Exact 2.1.0 tree: `5d5c158276891582132f5ed5a8ba0d712b8c9bfd`
- Tag: `2.1.0`; release published 2025-09-17; commit message `migration to Architectury`.
- Newer 1.21.1 head `3e1f38fe1be03e21a45075cc9fe39bfff7a41296` is Shield API 2.2.0 and is **not** the source authority for this Paladins pin.
- Exact target-era 1.20.1 branch: `cdcf7ffdcffb31a1dd8c36ba7a27cf312b0e8e71`, Shield API `1.0.1`, Yarn `1.20.1+build.10`.
- Port the exact 2.1.0 public behavior to target-native Forge 1.20.1, using the 1.20.1 branch only as mapping/runtime substrate. Do not substitute historical 1.0.1 wholesale and do not silently upgrade Paladins to Shield API 2.2.0.

## Shield API 2.1.0 -> 1.20.1 translation contract

Source-proven behavior must win over historical implementation details.

### CustomShieldItem attribute surface

Shield API 2.1.0 stores a mutable supplier of `AttributeModifiersComponent`, takes `List<Pair<RegistryEntry<EntityAttribute>, EntityAttributeModifier>>`, and emits each supplied modifier for `AttributeModifierSlot.HAND`.

The 1.20.1 substrate exposes the same conceptual contract through `Multimap<EntityAttribute, EntityAttributeModifier>` and `getAttributeModifiers(EquipmentSlot)`, with shield modifiers returned for `EquipmentSlot.OFFHAND`.

Target rule:

- Preserve the exact caller-supplied modifier list and the public `setAttributeModifiers(...)` mutation contract from 2.1.0.
- Translate registry entries to the raw 1.20.1 `EntityAttribute` surface and build an immutable multimap.
- Return that translated modifier set through the target-native equipment-slot API; do not invent data components on 1.20.1.
- Preserve repair ingredient behavior and `CustomShieldItem.instances` registration.
- Translate `RegistryEntry<SoundEvent>` to target-era raw `SoundEvent` without changing fallback-to-super semantics.

### Shield durability hook

Shield API 2.1.0 injects `PlayerEntity.damageShield` at HEAD and, for `CustomShieldItem`, increments the USED stat server-side, applies `1 + floor(amount)` durability damage when amount >= 3, clears the correct hand on break, and plays the shield break sound.

Target rule:

- Use the 1.20.1 `ItemStack.damage(int, LivingEntity, Consumer<LivingEntity>)` / `sendToolBreakStatus(hand)` callback signature from the historical substrate.
- Preserve every 2.1.0 semantic branch and threshold. The callback/signature is the compatibility seam; behavior is not.

### Shield disable behavior — current semantics are authoritative

This is an explicit anti-regression gate.

- Shield API 2.1.0 injects `disableShield` at HEAD and unconditionally sets a 100-tick cooldown on every registered `CustomShieldItem`.
- Historical Shield API 1.0.1 injects target-era `disableShield(boolean sprinting)` at TAIL and uses Efficiency + sprinting probability, then clears the active item and sends BREAK_SHIELD status.

Target rule:

- Adapt only the method descriptor to 1.20.1 (`disableShield(boolean sprinting)` as required by the target mapping).
- Preserve 2.1.0 behavior: every invocation must apply a 100-tick cooldown to all `CustomShieldItem.instances`.
- Do **not** restore the historical probability calculation, Efficiency dependency, conditional clearActiveItem, or BREAK_SHIELD status side effects merely because they compile on 1.20.1.
- Acceptance must include a deterministic runtime assertion proving this current behavior, including a non-sprinting invocation where the old implementation could have skipped the cooldown.

## Acceptance policy

Mirror the Archers discipline:

1. Freeze exact immutable current + historical pins and source manifests.
2. Materialize every current Java/resource/generated asset/data file.
3. Use an explicit deterministic 1.21.1 -> 1.20.1 compatibility transform pipeline.
4. Common compile against separately built named 1.20.1 ABI JARs; Forge loader compile against separate exact Forge artifacts.
5. No dependency shading that changes runtime ownership.
6. Java 17, metadata, leakage, content-count, and deterministic-byte gates.
7. Real Forge dedicated server + headless/native client acceptance with separate dependency JARs.
8. Semantic Shield API runtime tests before Paladins is allowed to consume it: custom shield construction, mutable attribute replacement, repair ingredient behavior, equip-sound fallback/custom behavior, >=3 damage durability path, hand clearing on break, stat increment, and unconditional 100-tick disable cooldown.
9. Semantic Paladins runtime self-tests for shields, armor, weapons/spells, effects/entities if present, villagers/POIs/trades if present, configs, Curios integration, recipes/tags/data, and current-only features.
10. Fresh official Forge 47.4.23 packaged-server replay with the exact release JAR.
11. Persist source/evidence/release artifacts to both GitHub and canonical Google Drive before graduation.

Acceptance satisfied by canonical run #212. Paladins 3.1.1 is graduated; Rogues is the next active leaf.