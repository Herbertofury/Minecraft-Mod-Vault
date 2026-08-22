# 🧬 Loaders & Minecraft Versions

## Current repair/porting foundation

Porting Lab + Compatibility Brain already establish the evidence model: detect the actual source loader/version/toolchain before proposing changes.

## OmniBridge roadmap

- Forge
- NeoForge
- Fabric
- Quilt-era inputs where practical
- Bukkit/Paper/server ecosystems where relevant to artifact identity or adapters

### Migration domains

| Domain | Examples |
|---|---|
| registration/lifecycle | deferred registers, entry points, event buses |
| mappings/names | obfuscated/mappings-era → modern unobfuscated Java releases |
| build/toolchain | Gradle plugins, Java runtime requirements, packaging |
| networking | packet APIs, codecs, directions, login/play phases |
| rendering | model/render layer APIs, client setup, shader hooks |
| data/component systems | capabilities/components/attachments/data components |
| worldgen | registries, codecs, datapack-driven definitions |
| testing | GameTest, loader JUnit, real client/server smoke tests |

## Non-negotiable

Loader conversion is never just import/package renaming. The generated project must compile **and** behave correctly in the target runtime.
