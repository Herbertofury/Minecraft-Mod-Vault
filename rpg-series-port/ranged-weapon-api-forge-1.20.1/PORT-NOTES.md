# Ranged Weapon API 2.3.4 -> native Forge 1.20.1

- Exact Minecraft 1.20.1 substrate: FabricExtras/RangedWeaponAPI `d95ba51c2f5c35bc8d397057092ba6043b00b705` (1.1.4 branch head).
- Exact 2.3.4 release reference: `c834f2699faefbdfcefa84f7f45708cd1a6bc55a`.
- Preserve mod ID `ranged_weapon_api` and namespace `ranged_weapon`.
- The 1.20.1 line is used only for game-version mappings/mixin targets. 2.3.4 config, attribute, scaling, item-slot and resource semantics are the behavior target.
- Minecraft 1.21 data-component attribute plumbing is represented on 1.20.1 through `Item#getAttributeModifiers`, with stable UUIDs derived from the 2.x namespaced modifier IDs. Existing vanilla item modifiers are merged rather than replaced.
- 2.x damage, pull-time, haste and velocity attributes are Forge-registered and attached to all living entities. Vanilla bow/crossbow baselines and custom `RangedConfig` attributes apply in either hand.
- The modern public `RangedConfig` record and modifier list are retained, plus legacy 1.20.1 `CustomRangedWeapon` configuration methods for downstream compatibility.
- Status effects use 1.6/2.x-equivalent base-multiplier operations. Potion helper requests are fulfilled during Forge's potion registry phase.
- Custom repair suppliers, model predicates, FOV and crossbow rendering behavior remain supported.
- 2.3.4 translations/assets are reproduced from the exact modern commit during CI; Fabric metadata and loader initializers are excluded.

## Verification contract

Build/package checks, no Fabric/NeoForge metadata leakage, server runtime invariants for vanilla/custom bows and stacked optional attributes, client resource/model bootstrap, and a final fresh standalone Forge install using the packaged JAR before release.
