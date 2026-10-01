# Port notes

- Upstream 1.20.1 baseline commit: 325c09a1d78538ebfeb7e48218c22f9cc74876a9 (0.9.13-era branch).
- Upstream current 1.21.1 commit inspected: 96a89bb8f51f72a9087bbf1b1fb241d59c335d72 (1.3.2, 2026-08-21).
- User binary: runes-neoforge-1.3.2+1.21.1.jar SHA-256 51fee4d0f16ddb705c8cac116050834c74f5dc05f695dcf203689bad72cd6ba6.
- Changelog delta 0.9.13 -> 1.3.2 includes pouches, translations, Architectury migration, FFAPI removal on NeoForge, and recipe integration.
- 1.21 data-component Bundle API cannot be byte/API backported to 1.20.1; pouches are implemented natively with Items NBT, rune-only filtering, capacity bars/tooltips, cursor/slot interaction, and filled-model predicate.
- 1.21 data-pack directory names and recipe result schema are translated back to the 1.20.1 format.

## Verification design

- CI fetches upstream Runes resources at exact commit `96a89bb8f51f72a9087bbf1b1fb241d59c335d72` rather than duplicating third-party texture/audio assets in the public workbench repository.
- `tools/prepare_upstream_resources.py` deterministically converts 1.21 resource/data conventions back to 1.20.1 and removes the Bundle API recipe gate because pouches are native in this port.
- Server smoke includes a runtime pouch self-test: registry IDs, custom recipe load, 4-stack capacity, NBT persistence, extraction, overflow rejection, and non-rune rejection.
- Client smoke boots through resource/model loading under Xvfb and fails on Runes-specific missing-model or loader signatures.
- EMI 1.1.24+1.20.1 integration is compile-time optional; the mod remains runnable without EMI.
