# Spell Power 1.6.0 -> native Forge 1.20.1

- Exact 1.20.1 compatibility substrate: upstream branch commit `681993d5f823aa96b1b24e21b145e89f46147f2d` (0.12.0).
- Exact modern behavior/source line: upstream `1.21.1`, version 1.6.0.
- Architecture: deterministic baseline assembly + explicit Forge-native overlays; no Connector or Forgified Fabric API runtime dependency.
- 1.21 data-driven enchantments are not copied blindly. The 1.20.1 class-based enchantment implementation is retained and upgraded with the modern behavior/API changes.
- Current backports in this checkpoint: generic spell-power multiplier school, 1.6-style Result.Value/isCritical API, sub-1 haste support (10% floor), hyperbolic/linear/quadratic resistance config, 90% default resistance cap, modern base spell power/crit config, innate crit attribute modifiers, current translation/visual assets, `c:is_magic` compatibility data, Forge-native all-living-entity attribute attachment.
- Native Forge registry lifecycle is used for attributes, status effects and enchantments; intrusive registry objects are not registered through Fabric static-init hooks.
