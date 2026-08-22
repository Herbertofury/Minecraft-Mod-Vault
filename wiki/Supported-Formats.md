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
| Resource packs | Java/Bedrock packs | cross-edition/version transformation |
| CIT/custom item packs | CIT/CIT Resewn-style assets | functional furniture/item/block mods/add-ons |
| Models | Blockbench, glTF/GLB, Blender workflows | Java/Bedrock models/entities/builds |
| Animation | Bedrock controllers, Blockbench, skeletal formats | retargeted Java/Bedrock systems |
| Worlds | Java Anvil, Bedrock LevelDB, `.mcworld`, `.mctemplate` | WorldForge conversion/edit/repair |
| Schematics/structures | `.schem`, `.schematic`, `.litematic`, NBT, `.mcstructure` | WorldForge import/export/placement |

> [!WARNING]
> A listed roadmap target is not automatically a claim of current production support. Check [Release Status](Release-Status) and the relevant feature page.
