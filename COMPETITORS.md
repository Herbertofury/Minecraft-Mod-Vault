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


## Potential foundations, reuse and differentiation

| Candidate foundation | Specific component/pattern to investigate | Intended improvement | Compatibility and empirical gate |
|---|---|---|---|
| packwiz | Inspect manifest/hash and distribution semantics for a possible import/export adapter | Reviewable pack changes without losing local overrides or optional content | Round-trip identical permitted packs; preserve filenames, hashes, provider identities, side constraints and optional choices. Do not equate loader migration with semantic mod porting. |
| Prism Launcher | Study instance boundaries, launch configuration and diagnostic UX; reuse code only if licensing permits | Make failed launches explainable without burying logs or corrupting working instances | Compare clean install, custom Java arguments, dependency failure and rollback on the same pack and hardware. |
| Modrinth App / official provider interfaces | Investigate provider identity and compatible-version metadata through permitted interfaces | Show exact installed-to-candidate changes and provenance rather than filename guesses | Test API unavailability, ambiguous matches, removed versions and custom patches. Provider access restrictions remain binding. |
| Vault's current identity and transaction contracts | Reuse existing artifact-first identifiers, receipt/undo semantics and TestGrid evidence | Avoid replacing working foundations with duplicate managers | Future Enderloom/other-tool interoperability should exchange versioned manifests and sanitized receipts first, not share writable profile directories. |
| Existing world/conversion catalog | Reuse knowledge of format and behavior boundaries | Make conversion loss visible and reversible | Clone a permitted test world, compare entities/blocks/scripts and verify restoration. No unsupported conversion may be advertised as complete. |

**License and interoperability gate:** inspect exact upstream code licenses, dependency licenses, notices, API terms and redistribution permissions before importing code. Pack metadata does not confer mod redistribution rights. Record edition, loader, game, Java/runtime and provider API versions. Exported manifests must exclude account tokens and personal paths.

**Empirical evidence required:** cold/warm install-to-play, update and rollback journeys; peak RAM; UI latency at full library size; complete identity/dependency output; byte-level recovery; real-client behavior where relevant. An import/export adapter remains a candidate until tested against Vault's current baseline with its complete requirements preserved. No new integration was implemented or benchmarked by this documentation expansion.

## Review and evidence rules

Review date: **2026-11-12**. This is a comparison agenda, not a benchmark result or proof of superiority. Official-source capability descriptions and our proposed experiments are separate. Pin actual versions and licenses before reuse; record publication date separately from the date checked. Retain failed comparisons and rejected candidates.

During recurring reviews, inspect upstream releases, breaking changes, security notices and relevant issue reports; add a dated change note only for a meaningful change. Reuse the existing catalogs instead of creating another disconnected list. A missing observation is **unknown**, not unsupported. No new feature is silently authorized merely by appearing here.

For each proposed adoption, record baseline revision, candidate revision, identical workload, required outputs, correctness/security/compatibility checks, latency distribution, memory and reproducible raw evidence. Never make a target-project superiority claim from a synthetic microbenchmark. Keep private assets and traces out of public evidence.

## Publication status

This repository document is the canonical comparison entry point for this addition. A repository-backed wiki page can link to it without duplicating its tables. Native GitHub Wiki publication must be verified separately; a wiki-enabled setting or a source mirror is not publication proof.
