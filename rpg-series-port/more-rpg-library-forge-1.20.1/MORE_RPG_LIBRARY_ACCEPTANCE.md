# More RPG Library 2.7.x -> Forge 1.20.1 Acceptance Contract

## Source authorities

- Target-version substrate: `ProfessorFichte/More-RPG-Library@7da3c766ef5aebd850a0eb2f6a26bde2409f626f` (`1.2.22-1.20.1`, Fabric, Java 17).
- User-supplied released feature reference: `2.7.1+1.21.1`.
- Live bleeding-edge feature authority: `ProfessorFichte/More-RPG-Library@bfa35e55133bd795676b6beb40c334d2904bf0ea`, whose `gradle.properties` reports `2.7.2` development.
- Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17.

## Dependency strategy

Reuse certified Forge 1.20.1 foundations where already graduated:

- Spell Engine base 1.10.2 (with a final upstream-delta audit before ecosystem freeze)
- Spell Power Attributes 1.6.0
- Ranged Weapon API 2.3.4
- TinyConfig 3.1.0
- Structure Pool API 1.2.1 when still referenced by target-version behavior

Do not create artificial dependency cycles:

- Critical Strike is compile/compat integration in current common source; it must remain optional unless runtime code proves a hard requirement.
- Armory integration is conditional and must not block the base library port.
- Modern Accessories wiring is an implementation detail to adapt to a Forge 1.20.1-compatible equipment/accessory seam; do not silently declare a new hard public dependency unless behavior truly requires it.
- Historical Trinkets/AzureLib Armor dependencies are target-era implementation references, not automatically final Forge dependencies.

## Feature-parity floor

The port is not complete if it merely rebuilds the old 1.2.22 feature set. Audit and preserve every representable mod-owned 2.7.x behavior, including at minimum:

- 2.7.x Spell Engine 1.10 integration behavior
- improved controlled-entity relations and controlled-enemy behavior
- ControlEnemy status-effect functionality
- generalized stealth status effect
- current entity-protection/relation checks
- modern mob spell-caster goal behavior, including channeled-spell tracking/back-away behavior
- current custom spell impacts and predicates
- current functional entity attributes available to all relevant living entities
- current elemental weakness configuration
- current loot config, custom loot entries/functions and spell-scroll behavior
- current spell schools and weakness handling
- current particles, sounds and player animations
- current weapon passives/skills moved into the library from LNE/Berserker/Forcemaster integrations
- Duelist's Focus and other modern shared status/passive behavior
- current structures/processors, including ConditionalJigsawStructure, PathAdaptionProcessor, WaterPillarProcessor and TerrainBlendingProcessor where target-version APIs permit a faithful adaptation
- modern shared item groups / content exposure without creating a hard Armory/Arsenal dependency
- current config migration/sanitization behavior
- current fixes for hostile/friendly targeting, armor piercing, lightning entities, render crashes, status effects and cooldown behavior
- 2.7.2 removal of Forgified Fabric API as a required dependency must be reflected in the native Forge target: no FFAPI runtime requirement.

## Porting rules

1. Start from the 1.20.1 branch's mappings and version-native implementations for registries, entities, loot, rendering and worldgen where they are behaviorally valid.
2. Forward-port modern 2.7.x mod-owned features by feature family, not by blindly copying 1.21.1 APIs.
3. Replace Fabric loader/events/registry/network hooks with native Forge equivalents.
4. Preserve mod ID `more_rpg_classes` unless an upstream compatibility reason proves otherwise.
5. Preserve data/resource IDs and save/config compatibility where representable.
6. Never solve a missing 1.21 vanilla API by deleting behavior; record cross-version vanilla dependencies and use target-native adaptations or the explicit opt-in future-vanilla parity process.
7. No Fabric/NeoForge/Connector metadata or hard runtime symbols may leak into the shipped Forge JAR.

## Acceptance gates

- exact source pins and reproducible source assembly
- modern-vs-1.20.1 source/content inventory with explicit parity ledger
- deterministic clean Forge 1.20.1 release JAR build
- Java 17 bytecode gate
- Forge metadata/dependency gate
- dedicated-server registration/config/loot/worldgen/entity startup
- targeted runtime assertions for functional attributes, status effects, controlled-entity relations, mob spellcaster behavior, custom impacts/predicates and loot functions
- native client for particles/renderers/player animations/UI-facing behavior
- integrated-world gameplay lane for shared passives/skills and entity relation behavior
- restart/config persistence gate
- optional-integration lanes: Critical Strike absent/present; Armory absent/present when its port exists; accessory/equipment provider behavior appropriate to final Forge design
- clean fresh packaged Forge server using untouched release JAR
- frozen JAR/source identity replay
- evidence/checksums to GitHub + canonical Drive folder
- guarded non-force promotion to `workbench/rpg-series-forge-1.20.1`

## Downstream unlocks

Graduating this foundation should materially unblock the mandatory expansion wave: Artificers, Bards, Berserker, Elemental Wizards, Forcemaster, More Relics / MRPGC Skill Tree integrations, LNE content and other More RPG Classes/Content projects. Dependency order must still be verified per project rather than assumed.
