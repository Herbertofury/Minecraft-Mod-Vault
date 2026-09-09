# Spellbrook Forge 1.20.1 — All 29 Native Spell GIFs (PRETTY FINAL v4)

Status: **PASS — complete 29-spell native runtime showcase**

This checkpoint records the complete Spellbrook spell showcase captured from the real Forge 1.20.1 client/integrated server using PRETTY FINAL v4. No image generation was used.

## Coverage

All 29 Spellbrook spells were captured and validated:

### Fire
1. Fireball
2. Flame Wall
3. Flame Thrower
4. Controlled Fireball
5. Fire Bolt
6. Fire Tornado
7. Phoenix Fireball
8. Solar Strike

### Earth
9. Boulder
10. Earth Barrier
11. Earth Pillar
12. Earthquake
13. Rock Crusher
14. Rolling Boulder

### Water
15. Icicle
16. Water Whip
17. Wave
18. Water Geyser
19. Water Orb
20. Water Tornado
21. Permafrost Lance
22. Freeze

### Nature
23. Bubble
24. Healing Dew
25. Healing Orb
26. Poison Thorn
27. Restraining Vines
28. Thorns
29. Floral Stairway

## Runtime acceptance

- Total spells: **29**
- Passed: **29**
- Failed: **0**
- Offensive/targeted spells: **25**
- Offensive spells with server-observed real `hurtTime`: **25/25**
- Heal/barrier/mobility spells: validated through their actual player/self effects, health, position, manifestations, and status state.

Offensive spells were fired at a real 500-HP Iron Golem. The fixture sampled authoritative server health, `hurtTime`, fire ticks, effects, projectile counts, manifestation counts, and movement state. GIFs therefore show real Minecraft damage reactions rather than visual-only VFX.

Representative Fireball result:
- target health `500 -> 494` on first impact sample
- `hurtTime=10`
- fire ticks `60`
- mana `140 -> 122`
- projectile `0 -> 1`
- later target health `492`

The visual audit also confirmed persistent wall/tornado/freeze effects, healing presentation, barrier presentation, and Floral Stairway mobility rather than 29 near-identical target clips.

## Canonical media

Drive folder ID: `1xFoGI00owzravjW9mgPScXNlUQLUipQp`

Folder path:
`Minecraft Mod Vault/Spellbrook Forge 1.20.1/Spellbrook-All-29-Spells-Native-GIFs-v4`

Bundle:
- `Spellbrook-All-29-Spells-Native-GIF-Pack-v4.zip`
- Size: **16,927,057 bytes**
- SHA-256: `2427d6411642022106d2912d288452c7d572d10f6ae7dcfd8a1872fdad4a5570`
- Drive file ID: `1G4rtNdWWhVF30Z7RlB_18eWjj8-L9iHD`

Contact sheet:
- Drive file ID: `1QDuIYy59h5pMwMre_TTPWOVFW_vHGqYg`

Animated HTML gallery:
- Drive file ID: `1oOcLoFHZHFggz32Q7vfTmHKE1A00j-i4`

Runtime proof table:
- Drive file ID: `1lHPTg6Vjd28c-Y8vOH8hwjfHJbFOLAGm`

GIF integrity table:
- Drive file ID: `1z18ppACtn6IVyN-nbZrQhA95vGTRJXun`

SHA manifest:
- Drive file ID: `1QRUlhjd4q0JrlZCWlDFvvVlBZGM_-8aR`

Full results JSON:
- Drive file ID: `1dIkNZHOCPzzVIfxG4CDZBd1fjmarbTxy`

This media checkpoint does not add generated imagery. Every GIF/contact frame originates from the actual Forge client runtime.
