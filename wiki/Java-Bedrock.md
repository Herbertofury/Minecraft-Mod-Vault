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

See **[WorldForge](WorldForge)** for terrain, players, maps, entities, POIs, structures, inventories and world metadata.
