# Competing launchers, managers and authoring workflows

Start with the existing [launcher landscape](research/minecraft-launcher-landscape-2026.md), [world conversion landscape](research/minecraft-world-conversion-landscape-2026.md), [mod conversion stack](research/minecraft-mod-conversion-stack-2026.md) and [WorldForge tooling](research/worldforge-tooling-landscape-2026.md). Their historical review dates remain intact; this page adds a decision-oriented comparison layer rather than another launcher catalog.

| Comparator and primary source | Documented role | Proposed equivalent-work comparison |
|---|---|---|
| [Prism Launcher](https://prismlauncher.org/) | Multiple instances, accounts, mods and modpack management | Install an identical permitted pack; update one mod; recover a failed launch; inspect whether local modifications survive. |
| [Modrinth App](https://modrinth.com/app) | Integrated mod discovery, play and updates | Compare compatible-version selection, provenance, actionable failure messages and complete install-to-launch latency. |
| [packwiz](https://packwiz.infra.link/) | Version-controlled modpack management and distribution workflow | Round-trip a pack manifest with optional content and pinned hashes; compare reviewable diffs, reproducibility and missing-download handling. |
| Existing CurseForge, GDLauncher, ATLauncher, XMCL and Bedrock lanes | Specialist comparisons retained in the launcher ledger | Recheck exact versions and official provider limitations before repeating feature claims. |

## Improvement backlog, not accepted results
- [ ] Mixed-source identity corpus: exact hash match, ambiguous filename, locally patched JAR, duplicate display version and unavailable provider.
- [ ] Failure-injection comparison: interrupted download, incomplete install, dependency conflict and transactional rollback with byte-identical restoration.
- [ ] Java and Bedrock remain distinct acceptance lanes. Verify world/packs/UUID references rather than assuming Java pack semantics transfer.
- [ ] Compare UI responsiveness and peak memory on the same uncapped library while preserving metadata, warnings and off-screen records.
- [ ] Keep TestGrid server evidence separate from real rendered-client evidence.
- [ ] Conversion comparisons must report lost entities, blocks, scripts and behavior explicitly; unsupported conversion is not success.

## Review and evidence rules

Review date: **2026-11-12**. This is a comparison agenda, not a benchmark result or proof of superiority. Official-source capability descriptions and our proposed experiments are separate. Pin actual versions and licenses before reuse; record publication date separately from the date checked. Retain failed comparisons and rejected candidates.

During recurring reviews, inspect upstream releases, breaking changes, security notices and relevant issue reports; add a dated change note only for a meaningful change. Reuse the existing catalogs instead of creating another disconnected list. A missing observation is **unknown**, not unsupported. No new feature is silently authorized merely by appearing here.

For each proposed adoption, record baseline revision, candidate revision, identical workload, required outputs, correctness/security/compatibility checks, latency distribution, memory and reproducible raw evidence. Never make a target-project superiority claim from a synthetic microbenchmark. Keep private assets and traces out of public evidence.

## Publication status

This repository document is the canonical comparison entry point for this addition. A repository-backed wiki page can link to it without duplicating its tables. Native GitHub Wiki publication must be verified separately; a wiki-enabled setting or a source mirror is not publication proof.
