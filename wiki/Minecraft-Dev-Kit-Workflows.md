# Minecraft Dev Kit Workflows

## Live refresh before project work

Before starting a port, repair, or compatibility pass, use the live source manager to check the exact project lane instead of trusting whatever version happened to be downloaded previously:

```text
mmv-devkit check --registry devkit-registry.json --mc 1.20.1 --loader forge
mmv-devkit check --registry devkit-registry.json --mc 1.21.1 --loader neoforge
mmv-devkit check --registry devkit-registry.json --mc 26.2 --loader fabric
```

To refresh managed runtime dependencies, tools, and source mirrors in Google Drive after reviewing the dry-run plan:

```text
mmv-devkit sync --registry devkit-registry.json --mc 1.21.1 --loader neoforge --drive
mmv-devkit sync --registry devkit-registry.json --mc 1.21.1 --loader neoforge --apply --drive
```

For a long-lived workstation or Dev Kit service, `watch --drive --interval 15m` can repeat the same compatibility, dependency, downgrade, and hash checks. It does not blindly pull every newest upload.

## Forge 1.20.1

The core lane is JDK 17, Forge 47.4.23 MDK, Omniporter Gradle cache where useful, Vineflower/Recaf, Prism, spark, Blockbench/GeckoLib for animated entities, and RenderDoc/VisualVM/async-profiler when the problem crosses rendering or JVM performance.

For a mod under active repair, register its exact Modrinth/CurseForge/GitHub identity and let the source manager refresh the compatible runtime and required dependencies before TestGrid executes the build/runtime proof.

## NeoForge 1.21.1

The Dev Kit now has local JDK 21 archives for Windows and Linux, so the 1.21.x lane is offline-ready. Prefer the 1.21.1 ModDevGradle MDK for streamlined work; the NeoGradle MDK remains available for multi-version or toolchain-specific work. GeckoLib and other runtime dependencies should be resolved by provider identity rather than a hard-coded filename when a project is being actively updated.

## Fabric 26.2+

The 26.x Fabric lane targets JDK 25 and the current Fabric example/template workflow. Architectury Loom, Architectury API, Fabric API, Fabric Language Kotlin, Cloth Config, Forge Config API Port, YACL, Cardinal Components, Trinkets, and other common APIs are tracked as provider-backed dependencies where applicable.

## Quilt

The kit preserves the Quilt 1.20.6 template, Quilt Loader, Quilt Config, Quilt mappings, and QFAPI 1.21.1 source for compatibility research. QFAPI is a historical/older-version reference for modern 26.1+ work, where upstream Fabric API is the continuing path.

## Cross-loader / multi-version porting

Preferred order:

1. Refresh the exact source/runtime/dependency lane with `mmv-devkit check`.
2. Architectury Loom / Architectury templates for a clean shared architecture when appropriate.
3. Modstitch and Stonecutter for multi-loader/multi-version source organization patterns.
4. Parchment, Quilt mappings, Enigma, and MDK-provided official mappings for naming/reconciliation.
5. Vineflower for deterministic decompilation; Recaf for interactive bytecode inspection/patching.
6. MixinExtras as the mixin-compatibility reference.
7. DataFixerUpper when serialized/world data migration is part of the port.
8. TestGrid for real build/runtime verification and evidence.

When the source manager finds a required dependency that is not yet in the managed registry, a Drive-backed sync can enroll it into the dependency library lane and mirror matching source in the same transaction.

## Profiling

- **spark**: first choice for in-game/server tick profiling.
- **async-profiler**: native CPU/allocation/lock/JFR profiling.
- **VisualVM**: interactive JVM, heap, thread, and JFR inspection.
- **JFR Converter / VisualVM JFR Streaming**: supporting JFR workflow.
- **RenderDoc**: GPU frame capture and rendering diagnosis.

Profiling tools themselves can be assigned canonical upstream providers so the Dev Kit can detect stale releases instead of treating the current local binary as permanent.

## Models / animation

- **Blockbench**: geometry, pivots, keyframes, texture/model authoring.
- **GeckoLib source + Forge/NeoForge runtime references**: animation/export/runtime behavior.
- **EMF + ETF**: entity model/texture compatibility references.
- **Oculus + Embeddium**: rendering/shader compatibility references for Forge 1.20.1.

The next specialized layer remains automated Blockbench/GeckoLib QA for floating contacts, bad pivots, loop discontinuity, angular jerk, stiff tails, and export validation. Real project assets/outputs are used for QA rather than generative substitutes.

## Worlds / NBT

- **NBT Studio**: direct Java NBT, region, SNBT, and Bedrock NBT inspection/editing.
- **MCA Selector**: chunk/region selection and bulk world surgery.
- **Amulet**: cross-edition world tooling and conversion.
- **DataFixerUpper**: serialized data migration reference.

Back up worlds before destructive edits; the source manager does not auto-run world mutations.

## Bedrock

- **bridge.** for addon/resource/behavior pack development.
- **Amulet** for world-level work.
- Keep Java and Bedrock workflows separate until an explicit conversion step is requested; the Vault records provenance rather than pretending formats are interchangeable.

## Source and Drive rules

Matching source is kept separately from runtime binaries. Existing tracked Drive objects are updated in place by file ID; new objects are created only for newly discovered dependencies/sources. If a runtime is current but its source mirror is absent, the manager performs a source-only refresh rather than replacing working runtime bytes.