# ☕ ↔ 🧱 Java ↔ Bedrock

> **Status:** 📋 OmniBridge roadmap for conversion; Bedrock management itself is documented current functionality.

## Bedrock → Java targets

- entity geometry/bones/textures;
- animation and animation-controller semantics;
- Molang expressions where translatable;
- items/blocks/recipes/loot;
- spawn rules and behavior;
- particles/sounds;
- scripts/events translated into appropriate Java systems;
- explicit handling for features with no direct Java equivalent.

## Java → Bedrock targets

- behavior/resource packs;
- entities, blocks, items, recipes and loot;
- geometry/animations/particles/sounds;
- spawn/gameplay behavior;
- Bedrock Script API when declarative components are insufficient.

## Key rule

Cross-edition conversion is semantic. Java and Bedrock often encode the same player-facing idea through completely different runtime systems, so OmniBridge must recreate behavior rather than perform superficial file-format substitution.

## World conversion

World conversion is also semantic and must be scored by layer rather than reported as one pass/fail result. See **[WorldForge — World & Version Conversion Matrix](WorldForge-Conversion-Matrix)** for the current tool stack, version-upgrade/downgrade methods, Legacy Console lanes and torture-test corpus.

### Current practical lanes

| Direction | Strong current baseline | Important boundary |
|---|---|---|
| Java → Bedrock | [Chunker](https://www.chunker.app/), [je2be](https://je2be.app/), [Universal Minecraft Tool](https://www.universalminecrafttool.com/) | Chunker's current official conversion scope does not include cross-edition entities or player inventories; those require another capable lane or explicit post-fix/verification |
| Bedrock → Java | Chunker, je2be, UMT | preserve Bedrock actor IDs/XUID provenance and map deliberately to Java UUID/player records |
| multi-version editing / inspection | [Amulet](https://www.amuletmc.com/) | valuable editor/translation architecture, but current entity/item limitations mean it is not treated as a complete semantic converter |

WorldForge records separate verdicts for:

- terrain/chunks/block states and biomes;
- block entities and containers;
- entities;
- players/inventory/ender chest;
- items/components;
- maps;
- dimensions/portals;
- villages/trades/reputation/POIs;
- structures/references/jigsaws;
- command/data systems;
- lighting/heightmaps/ticks;
- resource/behavior-pack dependencies and custom/modded identifiers.

A successful terrain conversion therefore **cannot** silently imply player, entity, inventory, map or invisible-metadata fidelity.

### Identity translation

Java UUIDs, Bedrock actor IDs and Bedrock XUID/player records are different identity systems. Conversion must retain provenance and explicitly choose which server player becomes the local singleplayer identity when that workflow is requested. The [Server-to-Singleplayer World Converter](https://github.com/imSirr/world-converter) is tracked as a useful specialist reference for this exact case.

### Packs alongside worlds

A world can depend on assets/behavior that the other edition cannot represent natively. WorldForge therefore inventories attached datapacks/resource packs/behavior packs/resource packs before conversion and reports required companion conversion separately rather than silently dropping them.

See **[WorldForge](WorldForge)** for terrain, players, maps, entities, POIs, structures, inventories and world metadata, and **[Conversion Workflows](Conversion-Workflows)** for the broader OmniBridge content-conversion pipeline.