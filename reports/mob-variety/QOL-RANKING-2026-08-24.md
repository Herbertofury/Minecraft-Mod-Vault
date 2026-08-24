# Minecraft Mob Variety - Ranked/Sortable QOL Checkpoint

**Date:** 2026-08-24  
**Catalog baseline:** 253 unique projects  
**Canonical native Google Sheet:** https://docs.google.com/spreadsheets/d/16wfCuXWLPmXSyc8qQ0QvdDIbXqPWuPQlvDyhH3dy5Zg/edit

## What changed

The 253-project tracker was upgraded in place for fast browsing and ranking without changing the curated project set.

- All **22 data tabs** now have complete native sort/filter controls across every used column.
- The header row and **Name** column are frozen on every data tab.
- Existing `Overall` was standardized to **Overall Score**.
- Every data tab now has **Rank (1 = Best)**.
- Rank is formula-driven from the existing Overall Score and recalculates automatically.
- Overall Score and Rank receive red/yellow/green heatmaps for fast visual scanning.
- Every project name remains directly hyperlinked to its canonical/primary project page.
- The Read Me tab documents how to sort worst->best or best->worst.

## Ranking model

Overall Score remains:

- 35% Variety
- 30% Depth
- 20% Polish
- 15% Freshness

Rank formula pattern:

`1 + COUNTIF(score_range, ">" & current_score)`

This keeps ties honest: identical Overall Scores share the same rank.

## Verification

Local XLSX audit:

- 22 data tabs detected.
- 1,036 populated project-name cells checked; **0 missing hyperlinks**.
- 22/22 tabs have `Overall Score` + `Rank (1 = Best)`.
- 22/22 tabs have complete AutoFilter ranges.
- 22/22 tabs freeze at `B2` (top row + Name column).
- Rank formulas and cached rank values agree on every row.
- **0 spreadsheet error tokens** (`#REF!`, `#DIV/0!`, `#VALUE!`, `#N/A`, `#NAME?`, etc.).
- Dashboard master count still evaluates to **253**.

Native Google Sheets read-back:

- Title: `Minecraft Mob Variety - Master Tracker (253 Projects - Ranked QOL)`.
- All 22 data tabs report `frozenRowCount=1` and `frozenColumnCount=1`.
- Master Index rank formulas evaluate (for example 9.2 -> rank 36).
- Cows-Bovine evaluates 9.7 -> rank 1.
- Creature-Collections evaluates Cobblemon 9.9 -> rank 1.
- Bedrock and Third-Scour-Additions rank formulas evaluate correctly.
- Representative project-name hyperlinks remain live after the write.

## Raw XLSX artifact

Drive file ID: `1hKjAyLlXckYnIHmwuBOm6uULMGWJUNRf`  
File name: `Minecraft_Mob_Variety_Master_Tracker_2026-08-24_QOL_RANKED_253_PROJECTS.xlsx`  
Size: **384,147 bytes**  
SHA-256: `31bbc41c46b554afb1f4d89af3152f569b79014a503551cd33155add5da7031e`

The raw XLSX was replaced **in place** under the same Drive file ID and its remote metadata read back at 384,147 bytes.

## Continuity

The research baseline remains **253 projects**. Future research passes still start from 253 and search delta-only; this checkpoint changes presentation/QOL only, not the curated count.
