# ☕ ↔ 🧱 Java ↔ Bedrock

> **Status:** 📋 OmniBridge roadmap for conversion; Bedrock management itself is documented current functionality.

This page covers **mods/add-ons and their content**, not only saved worlds. See **[Mod Conversion Stack](Mod-Conversion-Stack)** for the full mappings/Mixins/loaders/bytecode/toolchain architecture.

## Bedrock → Java mod targets

- entity geometry/bones/textures;
- animation and animation-controller semantics;
- Molang expressions where translatable;
- items/blocks/recipes/loot;
- spawn rules and behavior;
- particles/sounds;
- component groups/events translated into Java lifecycle/event/state logic;
- Script API translated into target-native Java APIs, networking, persistence and/or Mixins;
- explicit handling for features with no direct Java equivalent.

### Preferred evidence / tooling

- **[Geyser mappings](https://github.com/GeyserMC/mappings)** and **[mappings-generator](https://github.com/GeyserMC/mappings-generator)** for living Java↔Bedrock block/item/effect/particle/sound/game-semantic mapping evidence.
- **[Blockbench](https://www.blockbench.net/)** for model/rig interchange.
- **[GeckoLib](https://github.com/bernie-g/geckolib)** when Bedrock-style animated models/animations can be reused faithfully in Java.
- **[bridge.](https://github.com/bridge-core/editor)** plus **[Minecraft Creator docs](https://learn.microsoft.com/minecraft/creator/)** / **[bedrock-samples](https://github.com/Mojang/bedrock-samples)** to parse/validate source Bedrock BP/RP/Script API semantics.

## Java → Bedrock add-on targets

- behavior/resource packs;
- entities, blocks, items, recipes and loot;
- geometry/animations/particles/sounds;
- spawn/gameplay behavior;
- Bedrock Script API when declarative components are insufficient;
- custom block/item/resource-pack translation where representable.

### Preferred evidence / tooling

- **[Geyser mappings](https://github.com/GeyserMC/mappings)** for cross-edition IDs/semantics.
- **[Rainbow](https://github.com/GeyserMC/Rainbow)** as a bleeding-edge Java custom blocks/items/resource-pack → Bedrock/Geyser generation reference.
- **[Geyser PackConverter](https://github.com/GeyserMC/PackConverter)** and specialist pack converters for narrower resource-pack lanes.
- **Minecraft Creator schemas / bedrock-samples** as the canonical target contract.

## Behavior IR

Cross-edition behavior must pass through a canonical semantic representation rather than code-to-code text translation. Record:

- lifecycle/event trigger;
- conditions;
- state/variables;
- actions;
- timing/tick semantics;
- authoritative logical side;
- networking/synchronization;
- persistence;
- referenced blocks/items/entities/registries;
- rendering/animation/particle/sound hooks;
- source component/script/code provenance;
- exact / translated / generated / review / blocked verdict.

Then emit target-native Java APIs/Mixins or Bedrock components/controllers/Script API code.

## Key rule

Cross-edition conversion is semantic. Java bytecode, loader events and Mixins do not have direct Bedrock JSON/Script API equivalents, and Bedrock component/event/controller graphs do not have direct Java-class equivalents. OmniBridge must recreate behavior rather than perform superficial file-format substitution.

## World conversion

World conversion is also semantic and must be scored by layer rather than reported as one pass/fail result. See **[WorldForge — World & Version Conversion Matrix](WorldForge-Conversion-Matrix)** for the current tool stack, version-upgrade/downgrade methods, Legacy Console lanes and torture-test corpus.

### Current practical world lanes

| Direction | Strong current baseline | Important boundary |
|---|---|---|
| Java → Bedrock | [Chunker](https://www.chunker.app/), [je2be](https://je2be.app/), [Universal Minecraft Tool](https://www.universalminecrafttool.com/) | Chunker's current official conversion scope does not include cross-edition entities or player inventories; those require another capable lane or explicit post-fix/verification |
| Bedrock → Java | Chunker, je2be, UMT | preserve Bedrock actor IDs/XUID provenance and map deliberately to Java UUID/player records |
| multi-version editing / inspection | [Amulet](https://www.amuletmc.com/) | valuable editor/translation architecture, but current entity/item limitations mean it is not treated as a complete semantic converter |

WorldForge records separate verdicts for terrain/chunks, block entities, entities, players/inventory, items/components, maps, dimensions/portals, villages/trades/POIs, structures, command/data systems, lighting/heightmaps/ticks and companion packs/custom identifiers.

A successful terrain conversion therefore **cannot** silently imply player, entity, inventory, map or invisible-metadata fidelity.

### Identity translation

Java UUIDs, Bedrock actor IDs and Bedrock XUID/player records are different identity systems. Conversion must retain provenance and explicitly choose which server player becomes the local singleplayer identity when that workflow is requested. The [Server-to-Singleplayer World Converter](https://github.com/imSirr/world-converter) is tracked as a useful specialist reference for this exact case.

### Packs alongside worlds

A world can depend on assets/behavior that the other edition cannot represent natively. WorldForge therefore inventories attached datapacks/resource packs/behavior packs/resource packs before conversion and reports required companion conversion separately rather than silently dropping them.

See **[WorldForge](WorldForge)** for saved-world migration and **[Conversion Workflows](Conversion-Workflows)** for the broader OmniBridge pipeline.
