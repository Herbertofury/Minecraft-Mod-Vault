# Future Feature Bridge - Durable Behavior Spec

Date: 2026-08-27
Target: Minecraft Forge 1.20.1 / Forge 47.4.23

## Purpose

The Future Feature Bridge gives the Forge 1.20.1 Ragdoll/Reactions stack a stable wind-charge semantic layer that works in two modes:

- EMBEDDED fallback when no compatible external future-content provider is installed.
- EXTERNAL provider mode when a compatible provider such as Just Trial Spawners (JTS) is present.

The bridge must feel like one feature, not two competing implementations.

## Provider selection contract

1. Resolve provider capability at runtime from the currently installed mod set.
2. Prefer a compatible external provider when present.
3. Fall back automatically to the embedded implementation when no compatible external provider exists.
4. Never persist a stale provider lock across restarts or across same-world provider removal.
5. Provider selection must be observable in diagnostics.

Known accepted resolver examples from the final QA checkpoint:

- `minecraft:wind_charge -> ragdoll_reactions_future [EMBEDDED]`
- `justtrialspawners [EXTERNAL] / PROJECTILE_IMPACT_OR_DISCARD`

## Semantic event contract

The bridge normalizes a wind-charge use into one semantic impact/discard event for Reactions.

For one physical projectile/use:

- consume the real item exactly once when Survival consumption applies;
- emit exactly one Reactions trigger;
- produce at most one intended ragdoll transition for the impacted actor;
- do not double-fire because both provider hooks and fallback hooks observed the same lifecycle;
- discard/impact observation must be idempotent for that projectile lifecycle.

The final JTS lane proved the intended acceptance shape with item count 2 -> 1, one Reactions trigger, and one zombie ragdoll.

## Embedded gameplay acceptance contract

The embedded path must reproduce the intended wind-charge interaction semantics without forcing an external provider dependency.

Project acceptance fixture from the final QA checkpoint:

- oak trapdoor closed before burst;
- iron trapdoor closed before burst;
- oak trapdoor becomes `open=true` after the burst;
- iron trapdoor remains `open=false`.

Use authoritative block state, not only visual observation.

## Cleanup and uninstall contract

`prepare-uninstall` must remove bridge-owned fallback content that could otherwise remain in a save or inventory after the companion is removed.

Cleanup must cover at least:

- loaded fallback projectile entities;
- dropped bridge charge items/entities;
- player-held or player-inventory bridge charges;
- ender-chest bridge charges where supported by the implementation.

The exact-final checkpoint observed one successful cleanup run removing:

- 3 loaded fallback projectiles;
- 1 dropped charge;
- 4 player/ender-chest charges.

Cleanup must be safe to run deliberately before uninstall and must not delete unrelated provider-native content.

## Same-world provider removal contract

The same save must remain loadable after removing an optional external provider stack.

Expected behavior:

1. Forge may emit missing-registry warnings for content belonging to the removed provider/dependencies.
2. Those warnings are not bridge failures by themselves.
3. The original Ragdoll/Reactions systems must still activate.
4. Player join must complete if the save is otherwise valid.
5. Resolver must return to EMBEDDED mode.
6. No bridge-owned stale registry/provider state may prevent startup or gameplay.

## Production linkage contract

Mapped-development compilation, remap validation, and dev launch are necessary but insufficient.

Before release, inspect every bridge call that crosses into production Forge/Minecraft/provider APIs and validate the exact production JVM symbolic reference:

- owner class;
- method or field name;
- descriptor/signature;
- static/interface/invoke shape where relevant.

A release-blocking example already proven by this project was the mapped-development `BuildCreativeModeTabContentsEvent.accept(ItemLike)` convenience call, which linked differently in the production Forge 47.4.23 runtime and caused a Creative-inventory `NoSuchMethodError`.

The Dev Kit must retain an offline production-linkage gate capable of reproducing such descriptor failures before native launch.

## Native-client contract

After offline linkage checks, the exact release candidate still must be tested in real/native Forge.

Required client surfaces include:

- clean boot;
- world join;
- Creative inventory or any other client-only registration surface touched by the changed code;
- embedded gameplay assertion;
- external-provider gameplay assertion when applicable;
- save/quit with `All dimensions are saved`.

Do not substitute a dev launch for native production-client evidence.

## Dedicated-server contract

Test the exact release candidate in both applicable server profiles:

- base Ragdoll/Reactions + bridge, no optional provider;
- base + bridge + external provider + provider dependencies.

Required evidence:

- server reaches `Done (...)!`;
- original ragdoll systems activate;
- no bridge-owned mod-loading/linkage failure;
- clean stop/save reaches `All dimensions are saved`.

Warnings must be classified by ownership. Legacy upstream client-class mixin warnings can be documented as tolerated only after proving they are not introduced or worsened by the bridge.

## QA fixture contract

A harness failure is not a product failure.

When proximity or collision determines expected state:

- platform or pin the actor to a fixed test position;
- record the precondition;
- trigger the action;
- read authoritative post-action block/entity state;
- reject a fixture in which the actor fell, drifted, unloaded, or moved away from the target.

## Exact-candidate and reproducibility contract

All final gates must name the same candidate digest.

- Rebuild at least twice when practical.
- Compare produced artifacts byte-for-byte or by SHA-256.
- If a candidate changes, prior runtime evidence does not automatically transfer.
- Never expand a truncated digest from memory. The preserved final checkpoint currently identifies the bridge candidate only as `b20acda1e2d...`.

## Evidence ownership

Release evidence must explicitly distinguish:

- recovered bytes and hashes verified in the current session;
- preserved runtime evidence from an earlier exact-hash run;
- upstream/legacy warnings;
- bridge-owned errors;
- missing/unrecovered artifacts.

This prevents a clean server gate from hiding a native-client failure and prevents missing bytes from being silently reconstructed as if they were the original release.
