# 🧱 WorldForge vs Minecraft Bedrock Editor

> **Status:** 📋 Parity/surpass contract. Benchmark baseline documented for Bedrock Editor 26.40-era capabilities; re-check before implementation/release.

WorldForge treats the current Microsoft Bedrock Editor as a **living benchmark**, not a tutorial to copy once and forget.

## Required parity families

- Tool/Crosshair editing modes
- Selection, Magic Select, brush selection
- Clipboard + Paste Preview
- Brush + smart fill
- Line, Ruler, Extrude, Repeater
- Terrain + Flood
- Biome assignment
- chunk delete/regenerate
- Block Inspector + Workbench
- Entity Inspector + Summon
- structures/prefabs/Jigsaw/layouts
- Custom Mesh / STL / glTF / GLB
- procedural generators
- camera/cinematic/timeline
- Vibrant Visuals/environment authoring
- World Options + Navigation/minimap
- Test World
- import/export/project packaging
- Editor extensions/API
- transactions, long operations and undo/redo
- collaborative/BDS editing where supported

## Where WorldForge goes further

| Area | Planned advantage |
|---|---|
| editions | Java **and** Bedrock in one world model |
| versions | old/mixed-version repair and migration |
| packages | direct `.mcworld` / `.mctemplate` workspace |
| modded worlds | custom registry/data awareness |
| fidelity | release-blocking inventory/orientation tests |
| pruning | simple + build-aware + target-version retrogen |
| scale | streamed huge selections/operations rather than copying arbitrary caps |
| repair | corruption, maps, identity, POI/structure metadata |
| QA | automated TestGrid + Agent Driver |
| workspace | professional dockable/multi-monitor layouts |

## Arc de Triomphe acceptance fixture

WorldForge should reproduce the official tutorial workflow—selection/delete/fill, repeated geometry, stair orientation, Paste Preview, brush, extrude/carving and repeater—and then compare clicks/actions, completion time, undo behavior, correctness and resulting world data.

## Primary reference

https://learn.microsoft.com/en-us/minecraft/creator/documents/bedrockeditor/
