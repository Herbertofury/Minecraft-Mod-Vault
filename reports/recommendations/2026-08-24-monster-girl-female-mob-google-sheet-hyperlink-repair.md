# Monster Girl / Female Mob Google Sheet hyperlink repair

Verified **2026-08-24** against the live native Google Sheet.

## Canonical sheet

- [Minecraft Monster Girl & Female Mob Vault - Expanded Master Tracker (88 Entries)](https://docs.google.com/spreadsheets/d/1VR6VIh_rGyW1i0RljyW3TRt-amYX5qdc6G6PlT3qpEw/edit)

## Repair

- Project and Creator display columns across all current catalog views were made visibly link-styled (blue + underlined) while preserving their official hyperlink targets.
- New rows [Touhou Little Maid: Love & Loathe](https://www.curseforge.com/minecraft/mc-mods/touhou-little-maid-love-loathe) / [JumDa5he](https://www.curseforge.com/minecraft/mc-mods/touhou-little-maid-love-loathe) and [Sable: MaidRagdoll](https://www.curseforge.com/minecraft/mc-mods/sable-maidragdoll) / [gly091020](https://www.curseforge.com/members/gly091020/projects) were converted to explicit Google Sheets `HYPERLINK()` formulas wherever they appear.
- Dashboard leader names were converted to explicit `HYPERLINK()` formulas.
- Master Index leader cards were converted to explicit `HYPERLINK()` formulas.
- Dashboard internal-view navigation was restored as explicit `HYPERLINK()` formulas after detecting and repairing an intermediate right-column regression.
- Current Dashboard view counts were reconciled to: Gameplay & Companions 25; Maid Ecosystem 11; 3D Model Assets 11; Afdian & Creator Hubs 12; Java 1.20.1 39; Author Collection 29; Bedrock 2; Legacy & Remakes 25; Watch & WIP 6.

## Verification

A post-write Google Sheets API round-trip confirmed the affected live cells expose the expected `formulaValue = HYPERLINK(...)`, hyperlink URI, and visible underline formatting. The original Sheet file ID was preserved; no separate replacement Sheet was created.
