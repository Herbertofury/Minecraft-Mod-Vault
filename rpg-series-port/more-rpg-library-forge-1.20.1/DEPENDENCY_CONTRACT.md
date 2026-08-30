# More RPG Library 2.7.x -> Forge 1.20.1 dependency contract

Authorities:
- target-native substrate: `ProfessorFichte/More-RPG-Library@7da3c766ef5aebd850a0eb2f6a26bde2409f626f` (1.2.22 / Minecraft 1.20.1)
- mandatory released feature floor: 2.7.1 / Minecraft 1.21.1
- live feature-delta authority: `bfa35e55133bd795676b6beb40c334d2904bf0ea` (2.7.2 development)

## Ordering barriers

Heavy More RPG graduation MUST NOT start until Spell Engine 1.10.2 has a frozen Forge 1.20.1 release identity. Spell Engine and Spell Power are the public hard ecosystem requirements of the modern library, so using an unfinished Spell Engine port would make every downstream class/content result untrustworthy.

## Already-certified foundations available to this port

- Spell Power 1.6.0 Forge 1.20.1
- Ranged Weapon API 2.3.4 Forge 1.20.1
- TinyConfig 3.1.0 Forge 1.20.1
- Structure Pool API 1.2.1 Forge 1.20.1
- Gazebos 2.2.0 Forge 1.20.1 once its frozen replay/promotion completes

## Modern implementation integrations

Treat these according to actual source ownership rather than Gradle configuration labels:

- Spell Engine: hard ordering/runtime foundation; freeze identity before More RPG.
- Spell Power: hard runtime foundation.
- TinyConfig: library/config implementation dependency; use the certified 3.1.0 API/Forge bytes.
- Ranged Weapon API: real implementation integration; use certified 2.3.4 rather than source injection.
- Player Animator: external runtime integration where exercised by animations.
- Cloth Config: configuration/UI integration; preserve target-native Forge dependency rather than Fabric metadata.
- Critical Strike: compile/optional integration in modern common code; absence must not prevent base library startup unless source semantics prove otherwise.
- Armory: optional compatibility/content-group integration; do not make Armory an ordering prerequisite for the base More RPG Library.
- Accessories/owo on modern NeoForge: adapt to the target Forge 1.20.1 equipment ecosystem without forcing a circular dependency. Prefer the established Forge/Curios-compatible route where source behavior can be preserved; prove optional-provider presence/absence separately.
- AzureLib Armor / older 1.20.1 substrate integrations: retain only where still owned by modern 2.7.x behavior. Do not reintroduce dead legacy dependencies simply because the 1.20.1 branch used them.

## Parity families that must survive the version port

The Forge 1.20.1 result is not accepted if it merely resembles old 1.2.x. At minimum audit and preserve the live 2.7.x families for controlled-entity relations, stealth, weapon passives/skills, smarter mob spellcasting/channel behavior, custom spell impacts/predicates, elemental weaknesses, loot functions, particles/renderers, shared More Armory/More Arsenal item-group behavior where target-native, ConditionalJigsawStructure/path/terrain processors, and reusable mechanics moved from LNE/Berserker/Forcemaster into the library.

## Acceptance shape

1. exact-source dual-line preparation and deterministic source inventory;
2. no Fabric/NeoForge runtime leakage into native Forge output;
3. certified foundation identity checks before build;
4. deterministic two-build release identity;
5. optional-integration absence/presence gates where behavior changes;
6. real Forge 1.20.1 dedicated-server startup + library-owned semantic self-tests;
7. native Forge client/resource/particle/animation bootstrap for client-owned surfaces;
8. fresh packaged-server replay using untouched release JAR;
9. first-green identity freeze and independent frozen replay before canonical promotion.
