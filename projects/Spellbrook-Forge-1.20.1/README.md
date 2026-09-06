# Spellbrook — Forge 1.20.1

Standalone broom mod plus optional native Hexerei integration.

## Standalone
- One persistent Broomstick item with all captured Spellbrook finishes represented as NBT-backed variants.
- All captured core-crystal visuals represented as swappable core variants.
- Server-authoritative mount flight: WASD steering, jump ascend, crouch descend, sprint boost.
- 27 visible storage slots plus three reserved compatibility/module slots, matching the 30-slot Hexerei storage footprint.
- Broom inventory, finish, core, owner and persistent broom UUID survive break/place cycles.
- Exact captured models are rendered as normal baked Minecraft item models; the entity renderer renders that same stack in-world.
- No player/world scan and no background polling; flight work runs only for existing broom entities.

## Hexerei installed
Spellbrook conditionally registers a `BroomEntity` subclass and delegates riding/storage/module behavior to Hexerei. Spellbrook finishes and cores remain Spellbrook data/rendering, while Hexerei provides its native flight controller, 30-slot handler, brush/module slots, mounting, synchronization and persistence conventions. A default Hexerei broom brush is inserted when a Spellbrook broom enters the Hexerei path so native flight has the expected module.

Hexerei is optional and is never bundled.

## Credits
Original Spellbrook broom art/audio/design: **Levah, Crocwise, Team Spellbrook / Spellbrook Ltd.**, used with permission as described in `NOTICE-SPELLBROOK.md`.

Optional Hexerei integration: **JoeFoxe / Hexerei**.

## Status
Source implementation checkpoint exists. Final build/native QA is blocked until the ChatGPT container runner becomes available; the captured binary assets are intentionally kept outside the public repo until their permission boundary is confirmed.
