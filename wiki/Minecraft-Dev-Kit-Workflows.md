# Minecraft Dev Kit Workflows

## Forge 1.20.1

```text
mmv-devkit plan build --loader forge --mc 1.20.1
mmv-devkit plan port --loader forge --mc 1.20.1
```

Expected core: JDK 17, Forge 47.4.23 MDK, Omniporter Gradle cache where useful, Vineflower/Recaf, Prism, spark, Blockbench/GeckoLib for animated entities, and RenderDoc/VisualVM/async-profiler when the problem crosses rendering or JVM performance.

## NeoForge 1.21.1

```text
mmv-devkit plan build --loader neoforge --mc 1.21.1
```

The resolver targets JDK 21, then prefers the 1.21.1 ModDevGradle MDK. The NeoGradle MDK stays available for multi-version or toolchain-specific work. GeckoLib 4.9.2 and spark NeoForge are runtime references.

## Fabric 26.2

```text
mmv-devkit plan build --loader fabric --mc 26.2
```

The resolver targets JDK 25 and the Fabric 26.2 example project. Fabric's 26.2 guidance uses Loom 1.17; the collected Architectury Loom 1.17 source is also available for cross-loader work.

## Quilt

The kit preserves the Quilt 1.20.6 template, Quilt Loader, Quilt Config, Quilt mappings, and QFAPI 1.21.1 source for compatibility research. QFAPI is a historical/older-version reference for modern 26.1+ work, where upstream Fabric API is the continuing path.

## Cross-loader / multi-version porting

Preferred order:

1. Architectury Loom / Architectury templates for a clean shared architecture when appropriate.
2. Modstitch and Stonecutter for multi-loader/multi-version source organization patterns.
3. Parchment, Quilt mappings, Enigma, and MDK-provided official mappings for naming/reconciliation.
4. Vineflower for deterministic decompilation; Recaf for interactive bytecode inspection/patching.
5. MixinExtras as the mixin-compatibility reference.
6. DataFixerUpper when serialized/world data migration is part of the port.
7. TestGrid for real build/runtime verification and evidence.

## Profiling

- **spark**: first choice for in-game/server tick profiling.
- **async-profiler**: native CPU/allocation/lock/JFR profiling.
- **VisualVM**: interactive JVM, heap, thread, and JFR inspection.
- **JFR Converter / VisualVM JFR Streaming**: supporting JFR workflow.
- **RenderDoc**: GPU frame capture and rendering diagnosis.

## Models / animation

- **Blockbench**: geometry, pivots, keyframes, texture/model authoring.
- **GeckoLib source + Forge/NeoForge runtime references**: animation/export/runtime behavior.
- **EMF + ETF**: entity model/texture compatibility references.
- **Oculus + Embeddium**: rendering/shader compatibility references for Forge 1.20.1.

The planned next layer is automated Blockbench/GeckoLib QA for floating contacts, bad pivots, loop discontinuity, angular jerk, stiff tails, and export validation.

## Worlds / NBT

- **NBT Studio**: direct Java NBT, region, SNBT, and Bedrock NBT inspection/editing.
- **MCA Selector**: chunk/region selection and bulk world surgery.
- **Amulet**: cross-edition world tooling and conversion.
- **DataFixerUpper**: serialized data migration reference.

Back up worlds before destructive edits; the orchestrator does not auto-run world mutations.

## Bedrock

- **bridge.** for addon/resource/behavior pack development.
- **Amulet** for world-level work.
- Keep Java and Bedrock workflows separate until an explicit conversion step is requested; the Vault should record provenance rather than pretending formats are interchangeable.
