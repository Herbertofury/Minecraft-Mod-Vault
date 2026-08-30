# Minecraft Dev Kit Orchestrator 2.7.0 verification

Target verification date: 2026-08-28

## Compatibility Atlas acceptance

- Built-in plugin pack: TaCZ, Better Combat, Punchy, Epic Fight, Ragdoll/Reactions, Create, Valkyrien Skies, Clockwork, Eureka, Yes Steve Model.
- `compat init`: non-destructive custom plugin handling plus `--refresh` upstream sidecar behavior.
- `compat ingest`: runtime bytecode descriptors, source evidence, SHA-256, API fingerprint, semantic-anchor resolution, Minecraft/mod-version metadata and multi-loader source inference.
- `compat diff`: exact lane lookup, ambiguity rejection, complete symbol additions/removals, required-anchor move/signature/missing drift and production-linkage guidance.
- `compat matrix`: exact power set through 16 plugins, retaining dependency-invalid cases with reasons.
- `compat plan`: target lane selection plus optional hard-link audit.
- `compat scaffold`: generated Class.forName adapters with provider probe classes and no optional provider imports.
- `compat verify`: fails closed while requested lanes are missing or hard links remain.

## Automated verification

- `go test ./...`: PASS after Compatibility Atlas implementation.
- Same-arity signature-drift regression: PASS (`compat diff` identifies `(I)V -> (F)V` as `SIGNATURE_DRIFT`).
- Ambiguous version-only lane reference rejection: PASS.
- Hard-link scanner ignores reflection strings/comments but catches direct imports: PASS.
- Built-in plugin parse/probe validation: PASS.
- Linux amd64 build: required for release package.
- Windows amd64 cross-build: required for release package.

## Real corpus stress evidence

### TaCZ source evidence

- Input: `TACZ-1.20.1.zip`
- SHA-256: `8fe2eff59c9549a63c718ec392f5224df671c2b90ce82e8fb9a429d74c7bd7bc`
- Learned lane: `1.20.1/forge/source@8fe2eff59c95`
- Classes: 645
- Methods: 2,903
- Symbols: 3,548
- Required semantic anchors missing: 0
- Evidence level: `SOURCE_EVIDENCE` (not a substitute for the exact runtime JAR).

### Better Combat 26.2 source evidence

- Input: `BetterCombat-26.2.zip`
- SHA-256: `fa07b2bd49fc27703936b30403cf070197d82748b1b04d906b53ece27a3d5ff1`
- Learned lane: `26.2/fabric+neoforge/3.2.2`
- Classes: 114
- Methods: 379
- Symbols: 493
- Required semantic anchors missing: 0
- Multi-loader inference: PASS (Fabric + NeoForge retained in one source lane).
- Evidence level: `SOURCE_EVIDENCE`.

## Preserved 2.3/2.4 capabilities

Archive splitting/reassembly, cache doctor, client assets, client natives, world QA enablement, two-version heritage/port-guard, exact provider fetch, source management and Drive-aware tool management remain part of the same orchestrator. Compatibility Atlas is additive and does not weaken those gates.

## Vanilla Feature Atlas acceptance

- Sound identity split into SoundEvent registration / `sounds.json` definition / external OGG object provenance.
- Cross-version planner fails closed rather than stubbing an unresolved vanilla identifier.
- Complete dependency-closure checklist covers behavior, registry/bootstrap, data/resources/assets, translations, particles/game events, networking/codecs, persistence/data fixers, acquisition, worldgen, rendering/UI, dedicated-server behavior, and compatibility ownership.
- Built-in Vanilla Backport provider uses optional detection and requires per-feature ownership before suppressing the canonical Future Vanilla Backport fallback.
- Provider-present, provider-absent, same-world provider-removal, production-linkage, native-client, and dedicated-server lanes are part of generated plans when applicable.
- Reproducible atlas builder pins Mojang/mcmeta/minecraft-data/provider revisions and emits SHA-256 manifests.
- `go test ./...`: PASS after Vanilla Feature Atlas implementation.


## 2.6.3 base-port-first / opt-in backport gate

- `plan-backport` remains usable before completion as an offer/planning tool.
- The plan is `POST_PORT_OPTIONAL_VANILLA_PARITY`, `offerRequired=true`, `optIn=false`, and `implementationReady=false` by default.
- `--opt-in` without a passing `--port-report` is rejected.
- A passing `port-guard` report plus `--opt-in` can authorize implementation when the feature/provider decision is resolved.
- The generated dependency closure and QA lanes explicitly preserve the base target-native port and include opt-out recovery.


## 2.6.4 canonical ownership / Stonecutter matrix gates

- `go test ./...` must cover catalog construction, canonical fallback, proven-provider surrender, external-provider conflict detection, version/loader matrix parsing, Java-era selection, and matrix validation.
- Real 26.2 -> 1.20.1 atlas catalog generation must succeed against the complete atlas.
- `resolve-owner` must select Vanilla Backport only for proven claimed surfaces and must select `futurevanillabackport` when no proven external provider is installed.
- Individual port ownership of cataloged future-vanilla surfaces is forbidden by policy.
- `multiversion scaffold` pins Stonecutter 0.9.7 and emits 1.20.1/1.21.1/26.2-compatible Java lanes without overwriting existing files unless explicitly forced.


## 2.7.0 provider-native resilient download gates

- Exact Modrinth resolution retains canonical project/version provenance and hashes, then supplies primary URL + canonical CDN + Maven fallback routes.
- Artifact transfer streams to disk, retries bounded transient failures, resumes interrupted responses using HTTP Range, verifies size + SHA-512/SHA-256/SHA-1 as available, and promotes only verified bytes.
- CurseForge metadata and CDN requests use `x-api-key`; redirect requests preserve the key. `CURSEFORGE_API_KEY`, `CF_API_KEY`, and `CURSEFORGE_KEY` are accepted.
- CurseForge exact-file resolution requests `/download-url` when `downloadUrl` is omitted and reconstructs the official CDN route only from exact file ID + filename.
- Regression coverage includes the AoA 3.6.11 Modrinth identity (`9qn2AQBc` / `VCCCalGp`), primary-URL fallback, interrupted-transfer resume, hash-failure non-replacement, CurseForge redirect authentication, actionable missing-key failure, and download-url fallback.
- `go test ./...`, `go vet ./...`, Linux amd64 build, and Windows amd64 cross-build are release gates.
