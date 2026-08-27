---
name: minecraft-dev-kit
description: Build, port, convert, repair, benchmark, and release Minecraft mods using the user's Minecraft Dev Kit, with real Minecraft runtime verification rather than compile-only confidence. Use for Forge/NeoForge/Fabric/Quilt mod development, version ports, Bedrock-to-Java conversions, entity/model/animation work, worldgen/gameplay/network changes, dependency/cache problems, release QA, or any request to test a mod in actual Minecraft. Prefer the canonical Dev Kit on Google Drive when connected, reuse its toolchains/caches/QA harnesses, and require native client/server evidence whenever the changed behavior can only be proven in-game.
---

# Minecraft Dev Kit

Treat every Minecraft task as an acceptance contract. Compilation is intermediate evidence; prove the exact changed behavior in the strongest available real runtime before declaring it complete.

## Core workflow

1. Resolve the canonical project, Minecraft version, loader/version, Java version, source lineage, and requested artifact.
2. Resolve the canonical Dev Kit before downloading or rebuilding toolchains. Read `references/dev-kit-layout.md` when tool discovery, cache reuse, or Drive access matters.
3. Establish a baseline from embedded metadata, source, build files, current artifacts, and prior QA/repair receipts.
4. Build and run deterministic static/regression gates first. For optional-provider/future-feature bridges or code that crosses mapped-dev and production Forge APIs, read `references/future-feature-bridge-qa.md` and run its production-linkage gate before native launch.
5. Run the real dedicated server when server/common code, registries, world data, networking, or datapacks are affected.
6. Run the real client when rendering, models, textures, animation, UI, sound, input, client-only code, integrated-server behavior, or gameplay presentation can be affected. Read `references/native-runtime-qa.md` before native client work.
7. For ports/conversions, inventory upstream content and prove feature/content parity before polishing. Read `references/port-conversion-gates.md`.
8. Inspect fresh runtime logs and captured evidence. A command returning success is not proof that the intended entity/state survived or rendered.
9. Fix every task-related failure through diagnose -> repair -> retest. Do not weaken gates to obtain green output.
10. Build a fresh runnable artifact, hash it, package source/evidence/checksums, persist to the project Drive folder, and mirror repository-owned source/checkpoints to GitHub when applicable.

## Runtime verification policy

Choose the strongest applicable gate:

- **Compile/build only**: acceptable only for a change with no meaningful runtime behavior and no available runtime.
- **Dedicated server**: required for common/server logic when feasible. Wait for the real `Done (...)!` readiness marker, then exercise the changed workflow and inspect the fresh log.
- **Native client**: required for visual/entity/animation/client-integrated behavior when feasible. Use the actual loader + Minecraft client, not generated images or a software preview as a substitute.
- **Client + integrated server**: prefer this for entity state, gameplay, commands, datapacks, rendering tied to server sync, and conversion QA.
- **Real-world restart/persistence**: required for save/config/state changes.

A deterministic renderer, unit test, model audit, or server startup can complement native-client proof; none substitutes for it when the defect can exist only in the real client renderer.

## Native visual QA minimum

For model/entity/animation work:

1. Use a deterministic test save: fixed seed/flat world where practical, noon, clear weather, fixed camera, fixed coordinates.
2. Disable uncontrolled natural spawning rather than using Peaceful when QA entities are `MobCategory.MONSTER`; Peaceful can delete valid test entities immediately.
3. Prefer an in-world scheduled datapack or purpose-built QA command to drive phases. Do not make X11/chat typing the source of truth.
4. Assert the expected live entity/state count from the integrated server after each phase.
5. Capture the actual Minecraft window with FFmpeg/X11 or the platform-native capture path.
6. Preserve complete client logs, phase markers, stills, motion strips/clips, and checksums.
7. Visually inspect non-obvious frames, not only frame zero.
8. Fail if the model disappears, floats, clips, accumulates transforms, uses missing textures, emits loader/model/atlas warnings tied to the changed content, or silently changes state.

For articulated custom models, explicitly verify that every custom parent/root `ModelPart` not covered by vanilla animation reset logic returns to bind pose each frame before applying additive transforms. Per-frame transform accumulation is a release blocker.

## Cache and offline-first policy

Prefer the Dev Kit's verified JDK/Gradle/loader/cache assets over network resolution. Use `mmv-devkit` for cache integrity, client assets, natives, split archives, and world QA preparation when available.

For Forge 1.20.1 native client work, the established reusable path is:

```text
mmv-devkit cache-doctor
mmv-devkit client-assets
mmv-devkit client-natives
world QA normalization / command enable
Gradle runClient
native capture + authoritative log assertions
```

Full Mojang external asset objects are required when sound/resource completeness is under test. For visual-only model/texture QA, a verified mapped client JAR can be sufficient only when the harness explicitly proves the needed vanilla texture resources and reports missing external sound objects as a warning rather than pretending the cache is complete.

On Windows, use the bundled `scripts/Prewarm-Minecraft-Client-Assets.ps1` beside `mmv-devkit-windows-amd64-v2.3.0.exe` to resolve Mojang 1.20.1 metadata, download and SHA-1 verify the full external asset cache, re-verify it offline, and optionally split a ZIP into 85 MiB Drive-safe parts.

## Conversion and port discipline

Do not equate "builds on the target version" with "ported." For a full conversion:

- identify newest upstream feature authority plus older content-rich releases;
- inventory every relevant source file/content family;
- preserve intentional gameplay, assets, localization, sidecars, alternate states, and dependencies;
- distinguish direct runtime, derived runtime, and archival/provenance assets;
- prove mapped/converted content counts and hashes where exact preservation matters;
- test actual target-loader runtime paths;
- keep vanilla/Nether/End/worldgen untouched when the project guardrails require standalone compatibility.

## False-pass protections learned from real Forge QA

Treat these as known hazards:

- Peaceful deletes MONSTER-category QA entities even after a successful summon/QA command.
- A custom model parent omitted from vanilla reset sets can accumulate pose transforms frame-over-frame until the model vanishes.
- Hard-killing `runClient` can leave generated Forge runtime config partially written; validate/repair generated config before the next launch.
- A shell/container foreground timeout is not proof Minecraft hung. Detach long native runs and poll milestones/log markers.
- Xvfb may have no window manager; window-title discovery/focus helpers can fail even while GLFW/Minecraft is healthy.
- FFmpeg X11 capture should avoid mouse-pointer queries when bare Xvfb cannot service them.
- Post-processing tools must not consume the marker loop's stdin.
- A successful command is not proof of a live rendered entity. Assert server state and inspect the captured frame.
- Windows-generated ForgeGradle metadata can point at Windows LWJGL classifiers even when Linux native JARs are cached; extract and point LWJGL at verified Linux natives explicitly.
- A mapped-dev compile, ordinary remap check, or dev launch is not proof of production JVM symbolic linkage. Resolve and validate the exact production owner/name/descriptor for invoked symbols; a convenience overload visible to mapped development can still produce `NoSuchMethodError` in the shipped Forge runtime.
- Native-client-only entry points must be exercised on the exact release candidate. Registry/creative-tab population is a proven example: a production descriptor mismatch can stay invisible until the Creative inventory actually opens.
- A moving or falling QA actor can create a false gameplay failure. When a state assertion depends on proximity, use a platformed/fixed-position fixture and read authoritative post-action block/entity state rather than trusting visual position.

Read `references/native-runtime-qa.md` for the exact generalized native-client pattern. For optional-provider bridges, production-linkage hazards, provider removal/fallback, exactly-once event semantics, and uninstall cleanup, also read `references/future-feature-bridge-qa.md`.

## Evidence and release contract

For a substantive release, preserve:

- exact build command/result;
- final JAR/modpack hash and size;
- source archive/checkpoint;
- static/regression gate results;
- dedicated-server readiness + workflow evidence when applicable;
- native-client log + visual evidence when applicable;
- warnings classified as task-related vs benign;
- known limitations stated literally;
- Drive/GitHub publication identities.

Do not call a release verified if the strongest applicable runtime gate was skipped without stating why.
