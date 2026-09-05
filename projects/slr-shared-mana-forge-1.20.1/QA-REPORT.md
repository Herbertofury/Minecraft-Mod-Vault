# SLR Shared Mana 1.0.0 - Verification Report

## Build identity

- Minecraft: 1.20.1
- Forge: 47.4.23
- Java: Temurin 17.0.20.1+1
- Gradle: 8.8
- Final clean offline build: **PASS** (`BUILD SUCCESSFUL in 24s`)

## Real dedicated-server proof

Full runtime stack included SLR 1.2.0, Iron's Spells 3.16.3, Botania 455, Iron's Botany 2.0.1, Curios, Patchouli, GeckoLib, Iron's Lib, and Player Animator.

- HYBRID first full server: **PASS** - `Done (52.491s)!`
- HYBRID restart/persistence: **PASS** - `Done (10.354s)!`
- Shared-mana config hash remained byte-identical across first run + restart.
- SEPARATE challenge with bridge definitely loaded: **PASS** - `Done (10.248s)!` and bridge receipt `Iron's Botany mode=SEPARATE`.
- SEPARATE shutdown: **PASS** - `ThreadedAnvilChunkStorage: All dimensions are saved`.

## Native client / integrated-server proof

Real Forge client was launched under Xvfb with Linux LWJGL natives and entered a copied QA world.

Observed:

- Iron's Botany client/common setup loaded.
- Bridge logged that Iron's mana HUD was hidden.
- Integrated server started and loaded the bridge in HYBRID mode.
- Test player joined a real world.
- `/slrmana status` returned: `SLR 1000.0/1000.0 MP | Iron 100.0 mana | ratio 10.00:1 | Botany HYBRID`.
- Rendered HUD showed SLR `MP 1000/1000` while the separate Iron mana bar was absent.
- Save-and-Quit saved all loaded dimensions; client exited with `BUILD SUCCESSFUL`.

A native runtime screenshot is included at `evidence/native-client-status.png`.

## Iron's Botany mode coverage

Supplied Iron's Botany 2.0.1 production JAR exposes exactly:

`BOTANIA_PRIMARY`, `ISS_PRIMARY`, `HYBRID`, `SEPARATE`, `DISABLED`.

The bridge's mana mutation/regen paths do not branch on or rewrite that mode. Iron's Botany remains the routing authority. This is deliberate: all five modes share the same SLR-backed ISS primitives rather than maintaining a second competing mode state. See `evidence/botany-five-mode-static-audit.txt`.

Runtime challenge coverage:

- `HYBRID`: dedicated server + restart + native integrated client.
- `SEPARATE`: independent dedicated-server start with bridge's own mode receipt.
- `BOTANIA_PRIMARY`, `ISS_PRIMARY`, `DISABLED`: exhaustive production-config/API and router-neutral code-path audit; no duplicate mode implementation exists in this bridge.

## Performance / server cost

Static audit found no TickEvent subscriber, PlayerTickEvent, ServerTickEvent, LevelTickEvent, `level.players()` scan, entity/AABB scan, or scheduled mana mirror. Reflection metadata is resolved once and cached; subsequent work is demand-driven by real mana API activity.

This is intentionally described as **effectively zero idle overhead by design**, not as a mathematically measured zero-nanosecond claim.

## Production-linkage audit

Final reobfuscated release JAR was checked against untouched production SLR 1.2.0, Iron's Spells 3.16.3, and Iron's Botany 2.0.1 JARs:

- ISS `getMana`, `setMana`, `addMana`, `regenPlayerMana` symbols present.
- SLR `MP`, `Mana`, `manaregen`, sync, and `CooldownManager.set` symbols present.
- Iron's Botany `MANA_UNIFICATION_MODE`, ratio, and all five enum constants present.
- No dependency JARs embedded in the release.
- No QA/remapper/dev-runtime path leakage in the release.
- Result: **PASS**.

The sandbox could not independently launch a second production `forgeserver` from the packaged JAR because the Dev Kit cache lacks Forge installer's generated `server`/`server-extra` library tree and sandbox DNS cannot fetch the official installer. This is a harness limitation, not a runtime error from the mod. Real userdev dedicated-server, restart, and native client/integrated-server gates above all passed from the final source, and the final production JAR's linkage was audited separately.

## Non-blocking upstream/environment warnings observed

- Offline Mojang/Yggdrasil/version-check/Patreon/GitHub requests fail in this sandbox.
- SLR emits its existing client-only KeyMapping mixin warning on dedicated server.
- ISS/Iron's Botany include data entries referencing newer vanilla loot-function IDs; they log warnings in 1.20.1 but did not prevent server readiness.

No task-related bridge crash or server-start blocker remained.
