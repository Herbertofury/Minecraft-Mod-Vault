# Provider identity contract

1. Local file bytes are the anchor.
2. Exact Modrinth SHA-512 and CurseForge fingerprint matches are exact provider evidence.
3. Embedded loader metadata and artwork are usable even with every provider offline.
4. Canonical URLs and IDs are preserved, not rewritten into one storefront.
5. Conflicting names, authors, versions and artwork remain explainable evidence.
6. Fuzzy matching may suggest; it may not auto-update, auto-rename or auto-delete.
7. Latest means the newest file compatible with the selected game, loader/platform, side, Java and release policy—not merely the largest version string.
8. Custom/patched artifacts require deliberate user review before replacement.
9. Artwork is cached with provenance and falls back to embedded art before generic placeholders.
10. Provider failure never removes local content from the library.
