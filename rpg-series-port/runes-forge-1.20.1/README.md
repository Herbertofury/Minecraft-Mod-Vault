# Runes 1.3.2 — Native Forge 1.20.1 Backport

Target: Minecraft 1.20.1, Forge 47.4.x, Java 17.

This backport combines the proven upstream 1.20.1 crafting implementation with the 1.3.2 content/resources and Forge-native registration. It has no Connector or Forgified Fabric API runtime dependency.

## 1.3.x features carried back
- Rune Crafting Altar and custom crafting recipe type
- all six rune stones and `runes:runes` tag
- modern 20-language resource set
- Small / Medium / Large Rune Pouches with IDs matching 1.3.2
- Visual Workbench tag compatibility
- Spell Engine `spell_quiver` tag data

The 1.21 Bundle API is not available on 1.20.1. Rune pouches therefore use a native NBT-backed 1.20.1 implementation with the same 4x/8x/12x capacities and a stable vanilla-style `Items` NBT list. The RPG Series Unified compat module can consume that format directly for Spell Engine ammo integration.
