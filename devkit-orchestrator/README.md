# Minecraft Dev Kit Orchestrator

`mmv-devkit` turns the Minecraft Dev Kit folder into a capability-based toolchain instead of a pile of downloads. It is standalone, dependency-free, and designed to complement Minecraft Mod Vault TestGrid.

## Responsibilities

- Discover the synced Dev Kit root.
- Scan the real tool inventory against stable IDs.
- Resolve the best installed JDK, Gradle/build framework, MDK/template, loader reference, mappings, decompiler/bytecode workbench, profiler, renderer debugger, world/NBT tool, model/animation tool, Bedrock tool, or launcher for a requested task.
- Plan `build`, `port`, `repair`, `decompile`, `profile`, `render`, `world`, `model`, `bedrock`, `launch`, and `mappings` workflows.
- Select Java by Minecraft generation: Java 17 through 1.20.x, Java 21 for 1.21.x, Java 25 for 26.x.
- Safely prepare ZIP and `.tar.gz` archives into a local cache without modifying the Drive originals.
- Reject archive path traversal and refuse installer execution unless explicitly allowed.
- Emit JSON plans for TestGrid/agent consumption.

## TestGrid split

The orchestrator owns **tool selection and preparation**. TestGrid remains the **execution and evidence** layer for builds, launches, log assertions, RCON/network probes, hashes, runtime reports, and other acceptance evidence. This prevents duplicated automation logic.

## Current release

The complete 1.0.0 release, Windows x64 binary, Linux x64 binary, source, tests, generated registry catalog, and verification evidence are stored in the canonical Minecraft Dev Kit Google Drive under `08 Orchestration & Automation`.

The source-controlled documentation lives in:

- [`wiki/Dev-Kit-Orchestrator.md`](../wiki/Dev-Kit-Orchestrator.md)
- [`wiki/Minecraft-Dev-Kit-Tool-Catalog.md`](../wiki/Minecraft-Dev-Kit-Tool-Catalog.md)
- [`wiki/Minecraft-Dev-Kit-Workflows.md`](../wiki/Minecraft-Dev-Kit-Workflows.md)

## Verified 1.0.0 checks

- race-enabled Go unit tests
- `go vet`
- Linux amd64 build
- Windows amd64 cross-build
- registry parse/uniqueness
- synthetic Dev Kit scan
- Forge 1.20.1 -> Java 17 selection
- NeoForge 1.21.1 -> exact Java 21 gap detection
- interspersed CLI flags
- ZIP/path traversal rejection
- platform-aware artifact preference
- Wiki catalog generation from the machine registry

## One remaining offline inventory gap

The collected Drive kit currently has JDK 17, 25, and 26, but not the exact JDK 21 archive required by the resolver for fully offline Minecraft/NeoForge 1.21.x work. Online Gradle toolchain resolution can cover compatible projects; `doctor` reports the gap rather than silently substituting an incompatible runtime.
