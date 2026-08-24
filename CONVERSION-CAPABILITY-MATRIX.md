# OmniBridge 0.11.0 conversion capability matrix

| Source capability | Bedrock target | Java vanilla pack target | Java loader project target | Verification boundary |
|---|---|---|---|---|
| Textures and icons | Copy, normalize paths, rebuild texture indexes | Copy and rebuild Java asset paths | Copy into `assets/<namespace>` | Visual parity in target client |
| Languages | Convert key/value representation | Convert to Java locale JSON | Convert and preserve original | Formatting/fallback review |
| Common recipes | Translate shaped, shapeless, and furnace-family schemas | Translate to target Java recipe path | Translate into project data resources | Registry IDs and custom serializers |
| Models | Copy simple assets; specialist pack adapters for richer Java models | Preserve/normalize Java models | Preserve plus renderer/model contract | Predicates, transforms, geometry review |
| Sounds | Copy and preserve sound files | Copy Java sound assets | Copy plus sound-event surface | Sound index/event parity |
| Particles/animations/controllers | Preserve Bedrock-native files; generate target manifests | Review contracts | Client particle/animation implementation surfaces | Real client rendering test |
| Blocks/items/entities | Generate Bedrock definitions or preserve native definitions | Related recipes/assets only | Registry, component, attributes, goals and renderer contracts | Target-native gameplay implementation |
| Functions and commands | Preserve native functions or retain migration source | Command migration review source | Server command/function contracts | Every command tested in target edition |
| Loot/tags/advancements/predicates | Preserve or review target schema | Preserve native Java; Bedrock source retained for translation | Data and serializer/trigger contracts | Schema and runtime semantics |
| Bedrock Script API | Preserve for Bedrock and pin Script API manifest | Cannot execute in vanilla Java pack | Event/tick/command/network/persistence contracts and original JS | Implement and test logical sides |
| Java source | Script/data contracts for Bedrock | Extract data/resource layers | Loader/version migration workspace | Compile and runtime test |
| JVM bytecode/Mixins/ASM | Decompile/remap/semantic reconstruction route; never text-translated | Extract deterministic data/assets only | Dual-decompiler/remap/loader migration route | Reproduce source build and behavior |
| Camera/dialogue/trades/volumes | Preserve Bedrock definitions | Review source | Concrete camera/menu/trade/region surfaces | Client/server behavior tests |
| Biomes/features/worldgen/dimensions | Preserve Bedrock data or generate review contracts | Related Java data lanes where representable | Datagen/registry/biome-modifier/dimension contracts | Seed and placement verification |
| Worlds | Chunker/je2be/Amulet lane | Source world retained | World-template mod/exporter lane | Chunks, entities, inventories, dimensions, packs |
| World templates | `.mctemplate` and companion add-on product | Optional source world in pack family | Standalone world-template mod project | Fresh world, restart and upgrade tests |
| Structures | SchemConvert/world-adapter route | Java structure data when native | Structure processor/placement contract | Palette/block-state placement test |
| Java version/loader changes | Not applicable to Bedrock runtime | Target pack format from Version Atlas | Target Java/loader/mappings/build-tool scaffold | Compile, client, server, persistence |
| Bedrock engine changes | Target minimum engine version and manifest regeneration | Java target is separately versioned | Java target is separately versioned | Content-log and GameTest |

## Conversion levels

- **Exact:** source bytes or same-schema data preserved, with target indexes/metadata regenerated when necessary.
- **Translated:** a deterministic schema/path converter exists and is tested.
- **Generated:** a valid target artifact or source implementation surface is generated, but semantic parity still requires target tests.
- **Tool-assisted:** a reviewed specialist converter must run and produce a validated artifact.
- **Review:** related target concepts exist, but automatic equivalence cannot be proven.
- **Blocked:** no direct target representation exists; the original and a reimplementation contract are preserved.
