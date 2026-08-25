# 📦 Supported Formats & Coverage

This page distinguishes **current management support** from **OmniBridge conversion targets**.

## Current/documented management

| Format/content | Current role |
|---|---|
| Java mod JARs | identity, management, repair/porting handoff |
| `.mcpack` | Bedrock management |
| `.mcaddon` | Bedrock management |
| `.mcworld` | Bedrock management |
| `.mctemplate` | Bedrock management |
| behavior/resource/skin packs | Bedrock management |

## OmniBridge conversion targets — roadmap

| Source family | Examples | Target direction |
|---|---|---|
| Java mods | Forge, NeoForge, Fabric, Quilt-era inputs where practical | Java versions/loaders; Bedrock where semantics can be recreated |
| Bedrock add-ons | BP/RP, `.mcpack`, `.mcaddon` | Java mod or newer Bedrock format |
| Resource packs | Java/Bedrock packs, LabPBR/PBR material data | cross-edition/version transformation including Bedrock RTX/MER where representable |
| CIT/custom item packs | CIT/CIT Resewn-style assets | functional furniture/item/block mods/add-ons |
| Models | Blockbench, glTF/GLB, Blender workflows | Java/Bedrock models/entities/builds |
| Animation | Bedrock controllers, Blockbench, skeletal formats | retargeted Java/Bedrock systems |
| Java worlds | Anvil `.mca`/legacy `.mcr`, `level.dat`, `playerdata`, `data`, dimension folders | Java version upgrades/downgrades; Java ↔ Bedrock world conversion |
| Bedrock worlds | LevelDB worlds, `.mcworld`, `.mctemplate`, older StorageVersion-era stores | Bedrock version migration; Bedrock ↔ Java conversion |
| Early Pocket worlds | pre-LevelDB `chunks.dat` where detectable | WorldForge legacy Bedrock/Pocket import and preservation research |
| Legacy Console worlds | Xbox 360, PS3, Wii U plus platform/save-container variants where verified | Java/Bedrock migration through platform-aware adapters |
| Server worlds | Java UUID player records; Bedrock XUID / `player_server_*` records | server → singleplayer selected-player identity migration |
| Schematics/structures | `.schem`, `.schematic`, `.litematic`, Java structure `.nbt`, `.mcstructure` | common structure IR → import/export/placement with target block mappings |
| Raw NBT / region data | big-endian Java NBT, little-endian Bedrock structure NBT, MCA/MCR | inspection, repair, exceptional translation and provenance |

## World/version conversion is not a single format toggle

The canonical conversion path is documented in **[WorldForge — World & Version Conversion Matrix](WorldForge-Conversion-Matrix)**. In particular:

- Java forward upgrades use the real target Minecraft DataFixerUpper/WorldUpgrader path before semantic repair;
- Java and Bedrock downgrades are target-aware translations, **not** version-number edits;
- Bedrock adapters must recognize storage/chunk eras rather than assuming all LevelDB worlds use one layout;
- Legacy Console support is recorded by exact platform/save-container/title-update era;
- structures/builds use their own common IR and never stand in for complete world conversion.

> [!WARNING]
> A listed roadmap target is not automatically a claim of current production support. Check [Release Status](Release-Status), the relevant feature page, and the per-lane capability verdict before converting an irreplaceable world.