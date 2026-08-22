# 🌉 OmniBridge — 0.11.0 Feature Expansion

> [!IMPORTANT]
> **Status: 📋 Roadmap / incremental implementation.** OmniBridge is an additive expansion. It does not replace OmniManager, Repair Lab, Porting Lab or Compatibility Brain.

OmniBridge is the conversion and migration layer that ties artifact identity, repair knowledge, creator tools and runtime verification together.

## Major capability families

- Java ↔ Bedrock conversion
- Minecraft version updates/backports
- Forge / NeoForge / Fabric loader migration
- resource pack and CIT conversion
- standalone entity generation
- furniture generation
- behavior/action translation
- animation retargeting
- Blockbench and Blender round trips
- image/3D model → Minecraft build/entity pipelines
- batch conversion
- fidelity/provenance reporting
- plugin/adapter registry
- automatic upgrade watching
- creator intelligence
- TestGrid and Agent Driver
- WorldForge

## Fidelity classes

Every converted feature should be labeled honestly:

- **Exact**
- **Translated**
- **Reconstructed**
- **Inferred**
- **User Modified**
- **Unsupported**

Low-confidence areas must be reviewable rather than silently emitted as truth.

## Explore

- [Supported Formats](Supported-Formats)
- [Conversion Workflows](Conversion-Workflows)
- [Java ↔ Bedrock](Java-Bedrock)
- [Loaders & Versions](Loaders-and-Versions)
- [Creator Intelligence](Creator-Intelligence)
- [WorldForge](WorldForge)
