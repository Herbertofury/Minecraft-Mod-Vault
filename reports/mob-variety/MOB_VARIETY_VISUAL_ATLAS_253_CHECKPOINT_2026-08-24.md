# Minecraft Mob Variety — 253 Project Visual Atlas Checkpoint

Date: 2026-08-24
Research watermark: **253 curated projects**. Future research remains delta-only from project **#254** onward.

## Completed visual enrichment

- Preserved the ranked/sortable 253-project QOL baseline and all established research decisions.
- Removed the interrupted `_VISUAL_TMP` IMPORTDATA scratch sheet.
- Added fixed-size **real project icons**, **real author avatars**, and **real gallery previews**; no generated imagery.
- Cached visual coverage from the verified repository workflow:
  - 227 / 253 project icons
  - 238 / 253 author avatars
  - 209 / 253 gallery strips
- Master Index now has constrained visual columns and remains filter/sort capable across `A:Z`.
- Added a dedicated **Visual Atlas** sheet with Rank, icon, linked project, author/avatar, gallery preview, score, category, edition, loader, versions, status, project URL and visual-source URL.
- Images in the authoritative XLSX use move/size-with-cell anchors to avoid sorting detachment and row/column blowout.
- Companion recommendation guide gained a 16-page visual appendix using the same cached real imagery. Original 49 pages remain intact; final guide/PDF is 65 pages.

## QA

Authoritative XLSX:
- 25 sheets total, including Visual Atlas
- `_VISUAL_TMP`: absent
- Master Index: 254 rows, 26 columns
- Visual Atlas: 254 rows, 13 columns
- Master project-name hyperlinks: 253 / 253
- Spreadsheet error-token scan: 0
- Embedded image drawings: 674 on Master Index + 674 on Visual Atlas

DOCX/PDF:
- DOCX rendered successfully to 65 pages and visual appendix was visually inspected.
- PDF preflight: 65 pages, openable, unencrypted, non-XFA, not likely scanned.
- PDF spot-check rendering confirmed title/front matter, start of atlas, late atlas pages and final page without clipping/overflow.

Native Google conversion:
- Native Google Doc round-trip preserves the visual atlas and all 16 atlas contact-sheet images.
- Native Google Sheet round-trip preserves all 25 sheets, 253/253 Master project hyperlinks, filters, Visual Atlas, and 660 drawings per visual sheet. Google conversion drops 14 floating drawings per visual sheet; therefore the raw XLSX remains the pixel-complete authoritative spreadsheet artifact while the native Sheet is the convenient editable view.

## Full Drive deliverables

- Canonical raw XLSX (updated in place): https://docs.google.com/spreadsheets/d/1hKjAyLlXckYnIHmwuBOm6uULMGWJUNRf/edit
- Native Google Sheet visual edition: https://docs.google.com/spreadsheets/d/1lQtUmsXYP4OevS4wQg-nvhbdOAatEKIY3QYqeIKbNEw/edit
- Raw DOCX visual edition: https://docs.google.com/document/d/10YPahDPNO1ZDIRwl0_a2q2jgL5i9VNDB/edit
- Native Google Doc visual edition: https://docs.google.com/document/d/14fw6oMPyTbX9C68Xuk9XLtVYrf6o-1WtGVRjZWuVM58/edit
- PDF visual edition: https://drive.google.com/file/d/1jl4svhYmCxmnaGnhAPmUfScjjkZcZlwS/view

## Checksums

- XLSX — 9,034,565 bytes — SHA-256 `612e9c32b603c9ff46a0904c97e12096bcaabad993f0577f5b34c98ac25fba60`
- DOCX — 6,037,877 bytes — SHA-256 `948b09f5314ac5817ecb286ea4c5dec9899b5c4d51f0705130c7f42054ac4574`
- PDF — 3,730,587 bytes — SHA-256 `9e19b24605dd22c784da3a2eef1f91c11b5941d3c32d0b326f9a692221ab3c79`
- GitHub Actions visual cache artifact `mob-visual-assets-253` — SHA-256 `3bf8cea2611fe50bb17b90842a569cc1dcf3e394bc24598360e1575cb8f2632a`

## Continuation rule

Do not restart the 253-project catalog or repeat the first three scours. Preserve the ranked/sortable visual baseline and continue future discovery at **#254+ only**. Any future visual refresh should reuse `tools/visual_metadata/fetch_visuals.py` plus `apply_known_overrides.py`, then re-run the same fixed-thumbnail QA before publication.
