# Asset import boundary

The captured Spellbrook client resource pack is **not committed to this public repository** because the user's permission to use/download the assets should not be silently broadened into a public redistribution license.

Final private/release build procedure:

1. Start from the verified `Spellbrook-Broom-Research-Kit.zip` captured from the authorized Spellbrook session.
2. Copy the original `assets/broom/models/broomstick/**`, `assets/broom/models/core_crystal/**`, their referenced textures, and the captured broom sounds into `src/main/resources/assets/spellbrook/**`.
3. Rewrite namespace references inside copied model JSON from `broom:` to `spellbrook:`.
4. Preserve every captured model/texture dependency; do not substitute generated art.
5. Validate every JSON, PNG and OGG before build.
6. Confirm all 28 broom finish model routes and all 20 core model routes resolve from the item override JSONs.
7. Copy the exact uploaded `Hexerei-0.4.2.3-Performance-Overhaul-v4-Forge-1.20.1.jar` into `libs/` only for the compatibility compile/QA gate; do not bundle it in Spellbrook.

The verified capture report recorded 29 broom model JSONs (including parent/debug), 21 core model JSONs (including parent), 29 broom textures, 18 core textures, and 26 broom flight/mount OGG files.
