# Minecraft Mob Variety Companion 1.0.0

Release date: 2026-08-25

## Product

The Minecraft Mob Variety Explorer is now also a standalone Windows x64 Chromium companion application.

- Canonical catalog: 293 projects
- Collections: 19
- Real deduplicated media assets: 700
  - project icons: 266
  - author avatars: 256
  - project gallery images: 224
- No generated imagery is used.
- Canonical private source remains Google Sheet `1Jh94vNazaxtZbHJ-BWRjXLklbpoSsexIz9_OXKiZDIU`.

## Desktop behavior

- Electron 44.0.0 / Chromium 152.0.7977.54.
- Real live sites run in sandboxed Chromium `WebContentsView` tabs inside the app; no iframe/embed dependency.
- Always-visible `Open in your browser` and Copy URL actions.
- Back/forward/reload/address navigation, target-blank tab handling, session restore, downloads, permissions, find, zoom, and persistent live-site session.
- Split Research mode keeps the catalog/visual atlas beside the live project site.
- Catalog keeps search, numeric queries, filters, ranking, card/table/visual-atlas browsing, hover/focus image enlargement, lightbox galleries, author media, favorites, notes, compare, recent projects, and export.

## Runtime acceptance

GitHub Actions run: https://github.com/Herbertofury/Minecraft-Mod-Vault/actions/runs/32904335099

All release gates passed:

1. deterministic runtime bundle reassembly and SHA-256 validation;
2. JavaScript syntax checks;
3. official Electron 44 Linux/Windows runtime SHA-256 validation;
4. real Electron/Chromium shell launch under Xvfb;
5. browser-shell self-test PASS with `MOB_VARIETY_SELF_TEST {"passed":true,"failures":[],"electron":"44.0.0","chromium":"152.0.7977.54","tabs":2}`;
6. Windows x64 packaging;
7. final private catalog injection;
8. final ZIP integrity test;
9. exact source-file hash comparison inside the Windows ZIP;
10. Drive re-download/hash/reassembly verification.

The isolated build-infrastructure branch is `mob-variety-companion-runtime`; it was intentionally not used to overwrite unrelated concurrent `main` work.

## Final artifacts

Final Windows x64 ZIP:
- size: 160,233,261 bytes
- SHA-256: `6660d4fd0a43385e58fbc2d41d6447c645959ec18e002b44087241e50b3e0424`

Drive multipart delivery in `Recommendation Library/Companion App`:
- `Minecraft Mob Variety - Companion Windows x64.zip.part01`
  - Drive ID `1DYgP7qNAPc8iVavc38m492tar_-AekeC`
  - size 89,128,960 bytes
  - SHA-256 `e284148c52580c668e80914a1eba04a3970e206a32d5366a064c84485f9c82df`
- `Minecraft Mob Variety - Companion Windows x64.zip.part02`
  - Drive ID `1Asxd6X8BSAlFeHQItcJqBHwlqqFrtdNI`
  - size 71,104,301 bytes
  - SHA-256 `8a2c973aa8ba7bd0add51152366a150f8ec527ad0f718ef4f82b60da012996fd`
- `Minecraft Mob Variety - Companion - Reassemble.cmd`
  - Drive ID `1pG35l_08ziGwt-_mXkmv7YhPF07Axi-Y`
- `Minecraft Mob Variety - Companion - SHA256.txt`
  - Drive ID `16YgSuheXet34GOa4nc1nZc_GMwCtxe6E`
- `Minecraft Mob Variety - Companion Source.zip`
  - Drive ID `1KUpl15cpIlYRW3Xn_6NvDEwL1VTbFGg-`
  - size 3,346,013 bytes
  - SHA-256 `afc538652ade2c7a7177f772bdca5fda2fdb8fcf82bf8079292625b0a9830876`

Drive round-trip verification re-downloaded all five current companion files. Both multipart files matched their local SHA-256 values, their concatenation matched the final Windows ZIP SHA-256 exactly, and the reassembled ZIP passed archive integrity testing.

Superseded companion artifacts were moved to the existing `History` folder rather than deleted.
