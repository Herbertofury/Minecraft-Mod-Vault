# Minecraft Mob Variety Explorer - Update Contract

## Canonical sources

- Google Sheet: `1Jh94vNazaxtZbHJ-BWRjXLklbpoSsexIz9_OXKiZDIU`
- Google Doc: `1n7t_NCbe31fxYt0XPluakcgtYwCBFvayLFCkXspFOj4`
- PDF: `1jl4svhYmCxmnaGnhAPmUfScjjkZcZlwS`
- Project Drive folder: `1Zwtn7OPyp3qVeicr6BZQ10JISeH3vRXC`
- History folder: `1Sww9cb_fuJBPYU8fVkKH7106a0doUGKA`
- Current Explorer HTML: `1y0QUdAdqaBnfvxdY4mSpGm7yLWHg0JUM`
- Current source ZIP: `1U4MPaqtXp7QALzRvvvFjQsDl-mOEP96b`

## Source of truth

Use the exact current XLSX export of the native Google Sheet. Never substitute a title-matched copy when the canonical file ID is available. The companion Doc and PDF remain linked from the Explorer but do not override row-level Sheet data.

## Change detection

1. Export the canonical Sheet as XLSX.
2. Compute SHA-256 of the exact export.
3. Compare against `MOB_VARIETY_BUILD.sourceWorkbookSha256` embedded in the current Explorer.
4. If identical, make no artifact changes and do not send a change notification.
5. If changed, rebuild from that exact export and run the complete verification gate.

## Rebuild invariants

- Clean filename stays `Minecraft Mob Variety - Explorer.html`; counts/scours/build detail live inside it.
- Preserve every current master row and all direct authoritative project links.
- Recompute scour membership from the Second/Third/Fourth addition tabs; all remaining entries are Foundation / First Pass.
- Recompute collection membership from current category/collection tabs.
- Embed only real imagery from the workbook; do not generate replacement images.
- Keep collision-proof stable internal IDs derived from project name + short SHA-1 digest.
- Keep the private Sheet private; do not enable anonymous/public Sheet access as a sync shortcut.
- Browser storage is optional. Storage denial must never prevent startup.

## Verification gate

Before publication, require builder integrity, unique project IDs, complete primary links, correct scour coverage, JavaScript syntax pass, Chromium UI QA with zero page/console errors, working search/filter/table/details/favorite/compare/help/CSV/theme/direct-link flows, hidden empty compare tray, and usable desktop/mobile layouts.

## Publication

Replace Drive file `1y0QUdAdqaBnfvxdY4mSpGm7yLWHg0JUM` in place after a verified rebuild so the public identity inside the user's Drive stays stable. Refresh source ZIP `1U4MPaqtXp7QALzRvvvFjQsDl-mOEP96b` in place as well. If in-place replacement is impossible, preserve the superseded copy under History before creating the cleanly named new current object.

Checkpoint Explorer deltas additively under `research/mob-variety/explorer/` without reverting unrelated concurrent work.

## Research continuation

The fourth scour closed at 293 unique projects. New research begins at **#294** and remains delta-only against the current master.
