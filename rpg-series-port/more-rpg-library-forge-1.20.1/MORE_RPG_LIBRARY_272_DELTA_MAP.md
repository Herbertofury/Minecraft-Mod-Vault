# More RPG Library 2.7.2 -> Forge 1.20.1 forward-port delta map

Exact authorities:
- target-native 1.20.1 substrate: `7da3c766ef5aebd850a0eb2f6a26bde2409f626f`
- live 2.7.2 feature authority: `bfa35e55133bd795676b6beb40c334d2904bf0ea`
- Git comparison: modern line is 268 commits ahead and 34 behind the old substrate, so this is a feature-family forward-port, not a cherry-pick or old-branch rebuild.

## Port wave A - common runtime semantics

Preserve these modern owned systems first because downstream classes depend on them:
- entity ownership/relations: `ControlledOwnerAccess`, `MrpgEntityRelationMatcher`, `ControlOwnerMixin`, `EntityRelationsMixin`
- Control Enemy: `ControlEnemyStatusEffect`
- stealth: `StealthStatusEffect`, `LivingEntityStealth`, `LivingEntityRenderStealth`, `TrackTargetGoalStealth`
- shared entity attributes: `MRPGCEntityAttributes`, `EntityAttributesMixin`
- mob spellcasting AI: `MobSpellCastGoal`, `SpellBehaviorRegistry`, `BackAwayGoal`, `LowHealthFleeGoal`, `ISpellCasterEntity`
- living/projectile behavior: `LivingEntityMixin`, `PersistentProjectileEntityMixin`
- friendly lightning: `FriendlyLightningEntity`

Target proof: native Forge dedicated-server startup plus semantic assertions for owner/friendly/hostile relations, stealth target suppression, functional attributes, channeled spell state and back-away/flee transitions.

## Port wave B - Spell Engine / combat semantics

Modern authority owns substantial behavior not present in the old 1.20.1 line:
- `MrpgLibSpells` (~1411 lines in current authority)
- `CustomSpellImpacts`, `CustomSpellEntityPredicate`
- impact implementations including dash/range, missing-health damage, frozen ticks, knock-up/knockback, lightning, pull-in, rush-to-target, Spellthief, Stop Arrows and Trembling
- `MoreSpellSchools`, `MoreSpellSchoolWeakness`
- optional `CriticalStrikeCompat`
- Better Combat weapon attribute datagen

Spell Engine 1.10.4 frozen/replayed identity is a hard ordering barrier. Critical Strike remains optional and must bind through the graduated Spell Engine compatibility seam rather than become a circular hard dependency.

Target proof: Critical Strike absent/present lanes, custom-impact registration and execution, spell predicates, school weakness behavior, melee/ranged spell tags and representative generated spells.

## Port wave C - effects, items and shared passives

Modern effects include:
- Bleeding, Fatal Poison, Frosted, Frozen Solid, Ignited, Molten Armor, Siren's Tear, Soaked, Stealth, Wither's Curse
- `MRPGCActionImpairing`, `MRPGCEffects`
- Duelist's Focus renderer/model pair and modern passive/status behavior
- shared item groups and upgrade/material items in `MRPGCItemGroups` / `MRPGCItems`

Target proof: registration, save/reload persistence, action-impairing semantics, damage/effect ticks, Duelist's Focus owner/target behavior, no hard Armory dependency for base startup.

## Port wave D - loot/config/data

Modern loot ownership includes:
- `LootConfig`, `WeaknessConfig`, `TweaksConfig`
- `LootInjector`, `LootPoolAdder`, `MRPGCLootTableEntityModifiers`
- `ConditionalItemEntry`, `ConditionalItemLootFunction`, `ItemTagPickerLootFunction`
- `BindSpellFromPoolsLootFunction`, `SpecificSpellScrollPoolLootFunction`

Target proof: predefined and fallback loot tables, conditional entries/functions, tag selection, spell scroll pool binding, config migration/sanitization, clean restart.

## Port wave E - worldgen

Modern worldgen owned systems:
- `ConditionalJigsawStructure`
- `PathAdaptationProcessor`
- `TerrainBlendingProcessor`
- `WaterPillarProcessor`
- `ModStructureTypes`, `ModStructureProcessorTypes`

Use 1.20.1-native registry/codec/worldgen APIs; do not delete behavior because 1.21 APIs differ.

Target proof: deterministic structure registration and representative generated structure/processors in a real Forge world.

## Port wave F - client/render/audio/animation

Modern authority adds/expands:
- status particles and renderers including Duelist's Focus
- heart rendering/types for Fatal Poison
- popup/star/music/rainbow note particles
- friendly lightning and mob beam renderers
- substantially expanded sounds
- player animation catalog including archery, Burstcrack, Decapitate, Puncture, ground/sky channels/releases and other shared combat animations

Target proof: native Forge client/resource bootstrap, particle registration/render smoke, player animation loading and representative status/HUD render paths.

## Data parity floor

Current generated spell/data surface includes representative IDs such as:
`arcane_precision`, `avalanche_melee`, `burstcrack`, `carve`, `cursed_wither_bolt`, `decapitate`, `dragon_breath`, `dragonclaw_melee`, `dragonslayers_fury`, `duelists_focus`, `elder_guardian_shield`, `ender_dragon_shield`, `glacial_shield`, `glacial_splitter`, `lightning_strike_melee`, `lightning_strike_ranged`, `obsidian_shards`, `puncture`, `pyromaniac`, `reef_arrows`, `rimefrost`, `sirens_tears`, `water_flow`, `waterbomb_melee`, `wither_pulse_melee`, `wither_shield`, `zephyrs_speed`.

These are parity inventory, not permission to silently drop IDs that fail an API port.

## Loader/dependency rules

- native Forge 47.4.23 / Minecraft 1.20.1 / Java 17
- no Fabric API, Forgified Fabric API, NeoForge metadata/classes or Connector runtime leakage
- reuse certified Spell Power 1.6.0, Ranged Weapon API 2.3.4, TinyConfig 3.1.0, Structure Pool API 1.2.1 and frozen Spell Engine 1.10.4
- Player Animator / Cloth Config only where actual modern behavior requires them
- Accessories-era implementation must adapt to the chosen Forge 1.20.1 equipment seam without inventing a circular hard dependency
- Armory remains optional for the base library

## Graduation order

1. freeze and independently replay Spell Engine 1.10.4
2. replace `SPELL_ENGINE_1104_EXPECTED_JAR_SHA` fail-closed placeholder
3. exact dual-authority source preparation and machine-readable parity inventory
4. waves A -> F with compile/static gates after each coherent family
5. deterministic two-build release identity
6. dedicated server + semantic self-tests
7. native client and integrated-world gameplay lanes
8. restart/config persistence + optional-provider matrix
9. untouched packaged-server replay
10. first-green identity freeze + independent frozen replay
11. evidence to GitHub + canonical Drive
12. guarded non-force fast-forward to `workbench/rpg-series-forge-1.20.1`
