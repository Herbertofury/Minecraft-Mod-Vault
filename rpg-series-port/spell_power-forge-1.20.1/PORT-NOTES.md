# Spell Power 1.6.0 -> native Forge 1.20.1

- Exact 1.20.1 compatibility substrate: upstream commit `681993d5f823aa96b1b24e21b145e89f46147f2d` (0.12.0 line).
- Exact 1.6.0 behavior/resource reference: upstream commit `6fed879e796cbe82c43684d914a8fa99a99e8b12` from the 1.21.1 line. CI is pinned to this immutable commit.
- Architecture: deterministic baseline assembly + explicit Forge-native overlays; no Connector or Forgified Fabric API runtime dependency.
- 1.21 data-driven enchantments are not copied blindly. Their gameplay semantics are emulated on the 1.20.1 class-based enchantment API.

## Backported 1.6 behavior

- Generic Spell Power multiplier model, including main-hand Spell Power enchantment semantics.
- Specialized Sunfire/Soulfrost/Energize armor enchantment attribute behavior and modern exclusivity (specialized enchants exclude each other, not generic Spell Power).
- Spell Volatility and Amplify Spell are main-hand/weapon semantics, mutually exclusive, max level 3, with modern +4% / +10% per-level tuning; Spell Haste is max level 3 at +4% per level.
- Modern enchantment cost curves where the 1.20.1 API can express them; legacy config keys remain readable for downstream compatibility.
- Configurable matching-attribute requirement for specialized spell-power enchantments.
- `c:is_magic` Magic Protection compatibility instead of only private Spell Power damage tags.
- 1.6-style `Result.Value` / `isCritical` API.
- Sub-1 spell haste support with a 10% floor.
- Hyperbolic/linear/quadratic resistance config with 90% default cap.
- Base spell-power/crit config, innate crit attribute modifiers, all-LivingEntity attribute attachment, and login-time base-value migration.
- Optional potion registration (`register_potions`) using the 1.6 potion ID convention.
- Current translations/effect descriptions/visual assets, `c:is_magic` data, and mob-panic damage-type tagging.
- Status-effect attribute modifiers use 1.6-equivalent base-multiplier semantics rather than the older total-multiplier behavior.

## Verification contract

CI must build the remapped Forge JAR, reject Fabric/NeoForge metadata leakage, boot a dedicated Forge server, exercise the 1.6 enchantment/multiplier invariants, boot a Forge client through resource initialization, and preserve both the verified JAR and an exact tracked source-overlay ZIP.
