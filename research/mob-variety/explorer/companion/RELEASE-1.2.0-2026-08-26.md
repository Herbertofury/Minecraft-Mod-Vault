# Minecraft Catalog Companion 1.2.0 — 2026-08-26

This is the universal multi-catalog successor to the Mob Variety Companion. It preserves the existing Mob Variety rich catalog and upgrades the desktop app into a reusable catalog research/browser workspace.

## Bundled catalogs

- Minecraft Mob Variety: 293 projects, 19 collections, 700 real embedded media assets (266 project icons, 256 author images, 224 gallery images).
- Minecraft Monster Girl & Female Mob Vault: **312 current entries** from canonical Google Sheet `1MNY5dXvCzf6km7uojySMaHBjJJPEKJSuHnjeRCgdYTQ`.
- Mob Girls companion sources: Doc `1tvUipQCA1Ah7coX1MHhrfvol1PPSjqrC-tR991yMf9s`, PDF `10kQcQMPGX1bK1MVYmGxEytAquTbtiZE7`.

## Fixes

- Research Split now uses an **18 px uncovered native hit lane** between the two Electron `WebContentsView` panes, with real pointer drag, persisted ratio, keyboard resize, 50/50 reset, and side swap.
- Gallery overlay hit-testing bug fixed: hover/focus zoom now resolves the underlying media even when the gallery open-button overlays the image.
- Interactive galleries retain prev/next, thumbnails, lightbox, keyboard navigation, original/source actions, and live project-page media discovery.

## Universal catalog workspace

- Catalog button opens a switcher and source controls.
- Hot-drop/import support: XLSX/XLSM, CSV/TSV, JSON, DOCX, PDF, HTML, Markdown, text.
- Zero-dependency XLSX parser supports worksheet discovery, shared strings, hyperlink formulas and worksheet hyperlink relationships.
- Structured Sheet/table source wins catalog identity; Doc/PDF/reference sources enrich matching entries without inflating the structured row count.
- Duplicate structured rows collapse by authoritative URL and stable name/creator identity.
- Local sources are watched and refresh after file changes.
- Saved private Google sources refresh through the app's persistent authenticated Chromium session.
- Structured Google sources auto-check every five minutes; manual current/all refresh remains available.
- Favorites, notes, browser session and catalog choice persist independently of source updates.

## Acceptance evidence

- Real Chromium smoke test: PASS.
- Split divider real mouse drag emitted 11 resize commands and changed ratio: PASS.
- Mob Girls catalog switch/render: exactly 312 cards: PASS.
- Gallery fixture: live image injection, nav controls, hover zoom, lightbox, 2 thumbnails, keyboard next-image: PASS.
- Existing Mob Variety embedded gallery hover regression: PASS.
- Current Mob Girls XLSX independently parsed to exactly 312 entries: PASS.
- CSV literal source URL mapping + duplicate-row collapse: PASS.
- Structured identity protected from unmatched guide/reference URLs: PASS.
- JavaScript syntax checks for main, shell, generic catalog, gallery enhancer, ingestion, and preload modules: PASS.
- Windows ZIP integrity: PASS.
- Source ZIP integrity and exact key-file comparison: PASS.
- Multipart reassembly reproduces the exact Windows ZIP: PASS.

## Release artifact

Windows x64 ZIP size: **178,144,561 bytes**, which is **26,891 bytes smaller** than the 1.1 release despite the universal catalog engine and second bundled catalog.

SHA-256:

`8093148164dc66c0c583f3c547f7402d1f0f0ab10cda793a2ac0edee397f753f`

Source ZIP SHA-256:

`5577da47c2a5ec14bdd860aa406a7054446e07498c79a45afa9d0b8de8e9d7b7`

The Windows release reuses the previously verified Electron 44 / Chromium 152 runtime; 1.2.0 changes only the application payload under `resources/app`.

## Drive

Canonical Companion App folder: `1vEhnOszAAXGWpC0epxeogo-ufhm5fACJ`.

Current release is stored there as cleanly named multipart Windows package + source + reassembler + SHA-256 manifest. The superseded 1.1 files were moved non-destructively to project History `1Sww9cb_fuJBPYU8fVkKH7106a0doUGKA`.