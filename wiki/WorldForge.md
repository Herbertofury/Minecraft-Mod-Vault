# 🌍 WorldForge

![WorldForge](https://img.shields.io/badge/WorldForge-universal%20world%20studio-0ea5e9?style=for-the-badge)
![Status](https://img.shields.io/badge/status-roadmap-f59e0b?style=for-the-badge)

> [!IMPORTANT]
> **Status: 📋 0.11.0 roadmap.** WorldForge is the planned universal world converter, editor, repair, migration, pruning and retrogen studio inside OmniBridge.

## The ambition

Combine and surpass the strongest practical ideas from:

- Universal Minecraft Tool
- Chunker
- Amulet
- je2be
- Axiom
- WorldEdit / FAWE
- MCA Selector
- WorldPainter
- Minecraft Bedrock Editor

—but prove superiority through fixtures and benchmarks, not feature-count marketing.

## Core systems

| System | Goal |
|---|---|
| **WorldGraph IR** | version-independent semantic representation + opaque passthrough for unknown data |
| **Java + Bedrock** | Anvil and LevelDB through versioned adapters |
| **Direct packages** | `.mcworld`, `.mctemplate`, `.mcproject` workflows |
| **2D + GPU 3D editor** | enormous worlds, LOD, slicers, overlays and direct semantic inspection |
| **World conversion** | terrain, biomes, items, block entities, entities, players, maps, POIs, structures, metadata |
| **Mod/add-on awareness** | migrate custom content alongside the world rather than deleting unknown IDs |
| **Repair & forensics** | corruption scanning, salvage, broken references, mixed-version diagnosis |
| **Transactional editing** | previews, snapshots, persistent undo, crash recovery, dry runs |
| **Pruning & retrogen** | manual chunk pruning plus player-build-aware regeneration and seam repair |
| **TestGrid integration** | launch source/target worlds and verify behavior in real Minecraft |

```mermaid
flowchart TB
  S[Java / Bedrock / Template / Schematic] --> G[WorldGraph]
  G --> E[2D + 3D Editor]
  G --> C[Converter]
  G --> R[Repair / Forensics]
  G --> P[Prune / Retrogen]
  E --> TX[Transactional Journal]
  C --> TX
  R --> TX
  P --> TX
  TX --> V[Validation]
  V --> T[TestGrid + Agent Driver]
```

## Deep dives

- [Pruning & Retrogen](WorldForge-Pruning-and-Retrogen)
- [Bedrock Editor Parity](WorldForge-Bedrock-Editor)
- [Universal Minecraft Tool Parity](WorldForge-UMT-Parity)
- [Validation & Fidelity](Validation-and-Fidelity)
