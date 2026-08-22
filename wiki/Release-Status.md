# 🚦 Release & Capability Status

This page is the wiki's truth anchor. It deliberately separates **repository evidence**, **current product identity**, and **future roadmap**.

## Current evidence snapshot — 2026-08-21

| Area | Status | Evidence / note |
|---|---|---|
| **0.9.0 — Repair & Porting release** | ✅ Verified evidence in `main` | `releases/v0.9.0` contains checksums, Drive verification, final verification and release README. |
| **Repair Lab** | ✅ Verified | 0.9.0 release evidence documents production-UI/API assertions, hardened archive handling and reproducible outputs. |
| **Porting Lab** | ✅ Verified | 0.9.0 release evidence documents production-UI/API assertions and installed-JAR forensics/planning. |
| **Compatibility Brain** | ✅ Verified | 0.9.0 release evidence documents SQLite WAL/FTS5 state, seeded version/toolchain knowledge and persistence. |
| **0.10.0 — OmniManager identity** | 🟢 Current `main` README | `main` describes cross-store identity, native Bedrock management and premium daily-management flows. |
| **0.10.0 release evidence path** | ⚠️ Reconciliation needed | `main` references `releases/v0.10.0` and `release/v0.10.0`, but those were not visible in the repository API during this wiki refresh. Do not fabricate them. |
| **0.11.0 — OmniBridge expansion** | 📋 Roadmap / 🚧 as implemented later | The advanced TODO is additive-only and must be promoted page-by-page only after real implementation + verification. |
| **WorldForge** | 📋 Roadmap | Universal world editor/converter/repair system; not claimed shipped here. |
| **TestGrid / Agent Driver** | 📋 Roadmap | QA architecture and player-like automation; not claimed shipped here. |

## Promotion rule

A roadmap item moves to **Verified** only when all of the following are true:

- real implementation exists;
- user-facing controls are wired end to end;
- representative real workflows were executed;
- discovered failures were fixed and retested;
- existing features did not regress;
- a usable build/checkpoint was preserved;
- the wiki is updated **after** that proof exists.

> [!WARNING]
> A polished UI, generated file, passing build, unit test, or status document alone is not enough to change a roadmap feature to ✅ Verified.

## Source anchors

- Current project README: https://github.com/Herbertofury/Minecraft-Mod-Vault/blob/main/README.md
- 0.9.0 evidence index: https://github.com/Herbertofury/Minecraft-Mod-Vault/tree/main/releases/v0.9.0
- Roadmap overview: [Roadmap](Roadmap)
