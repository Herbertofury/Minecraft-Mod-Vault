# OmniBridge specialist tool and reference catalog

OmniBridge distinguishes **executable adapters** from **research/integration references**. A repository being listed does not grant it permission to run.

## Reviewed executable contracts

| Tool | Role | OmniBridge command contract | Output handling |
|---|---|---|---|
| Chunker / ChunkerCLI | Java ↔ Bedrock world and version conversion | `-i <world> -f <target-format> -o <isolated-output>` plus validated optional mapping/settings files | Packages `.mcworld`, `.mctemplate`, or Java world ZIP and validates it |
| JE2BE Resource Pack Converter | Java resource pack → Bedrock resource pack with PBR/RTX support | `convert <input.zip> <output.mcpack>` with reviewed pack-name/description/PBR options | Hashes and validates `.mcpack` |
| Geyser PackConverter / Thunder | Java resource pack → Bedrock resource pack | `nogui --input <isolated-input.zip>` | Detects produced `.mcpack`/ZIP, hashes and validates |
| Regolith | Compile generated Bedrock BP/RP/script projects | `run default` in a copied project using an exact export profile | Packages exported BP/RP as `.mcaddon`, hashes and validates |

## World and structure references

- Chunker: https://github.com/HiveGamesOSS/Chunker
- je2be-core: https://github.com/kbinani/je2be-core
- Amulet: https://github.com/Amulet-Team/Amulet-Map-Editor
- SchemConvert: https://github.com/SchemConvert/SchemConvert

## Resource/model conversion references

- Geyser PackConverter: https://github.com/GeyserMC/PackConverter
- Geyser Rainbow: https://github.com/GeyserMC/Rainbow
- JE2BE Resource Pack Converter: https://github.com/Seraphic-Studio/JE2BE-Resource-Pack-Converter
- EgoConverter++: https://github.com/ego-smp-labs/ego-converter-plus
- convert-pack: https://github.com/3vorp/convert-pack
- ResourcePackConverter: https://github.com/agentdid127/ResourcePackConverter

## Java source, mappings and multi-version references

- Fabric Loom: https://github.com/FabricMC/fabric-loom
- Tiny Remapper: https://github.com/FabricMC/tiny-remapper
- Modstitch: https://github.com/isXander/modstitch
- Stonecutter: https://github.com/Kikugie/Stonecutter
- Architectury Loom/API: https://github.com/architectury/architectury-loom
- Mojang DataFixerUpper: https://github.com/Mojang/DataFixerUpper
- Vineflower: https://github.com/Vineflower/vineflower
- CFR: https://github.com/leibnitz27/cfr

## Bedrock project references

- Mojang Bedrock samples: https://github.com/Mojang/bedrock-samples
- Regolith: https://github.com/Bedrock-OSS/regolith
- bridge.: https://github.com/bridge-core/editor

## Java data/resource-pack references

- beet + mecha: https://github.com/mcbeet/beet
- mcmeta version data: https://github.com/misode/mcmeta
- pack-format: https://github.com/Nixinova/Minecraft-Pack-Format

## Experimental Java-mod → Bedrock research

- PortKit: https://github.com/anchapin/portkit
- ModMorpher: https://github.com/Indozilla1234/Modmorpher

These experimental projects may inform analysis or scaffolding, but OmniBridge does not promote their output to semantic parity without source review, target-native implementation, builds, logs, and gameplay verification.
