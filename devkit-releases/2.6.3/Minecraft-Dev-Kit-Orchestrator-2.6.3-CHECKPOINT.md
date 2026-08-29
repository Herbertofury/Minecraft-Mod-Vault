# Minecraft Dev Kit Orchestrator 2.6.3 — Verified Checkpoint

Generated: 2026-08-29 UTC

## Requested behavior

The Dev Kit now enforces: **full target-native port/conversion first, future-vanilla backports second as an always-offered explicit opt-in layer**. The backport offer must state exactly what vanilla code/data/resources/assets/providers/QA must be brought over.

## Implemented

- Skill workflow changed to base-port-first.
- `port-conversion-gates.md` now defines Phase A base-port certification and Phase B optional future-vanilla parity.
- `vanilla-feature-atlas.md` now treats atlas results as requirements discovery during Phase A and implementation authorization only after the base gate.
- `plan-backport` now emits `stage`, `offerRequired`, `optIn`, `baselinePortGate`, `implementationReady`, and `parityRole`.
- `plan-backport --opt-in` requires `--port-report` whose `passed=true` and `errors=0`.
- The base artifact stays independently runnable; opt-out/disable recovery is part of the optional-layer QA matrix.
- Provider detection remains fail-closed: provider presence without per-feature ownership proof remains `UNRESOLVED`.

## Verification observed

- `go test ./...`: PASS.
- Linux amd64 build: PASS; reports `Minecraft Dev Kit Orchestrator 2.6.3`.
- Windows amd64 cross-build: PASS.
- Offer-only plan before base certification: PASS (`optIn=false`, `baselinePortGate=NOT_PROVIDED`, `implementationReady=false`).
- Explicit opt-in without base report: correctly REJECTED.
- Explicit opt-in with failing base report: correctly REJECTED.
- Explicit opt-in with passing base report: PASS and implementation authorization becomes true for a resolved provider-absent future feature.
- Updated Minecraft Dev Kit skill validation/package: PASS.
- Source ZIP integrity: PASS; generated Python cache files excluded.

## Artifact SHA-256

- `Minecraft-Dev-Kit-Orchestrator-2.6.3-SOURCE.zip`: `990f9feda867f39e97bd3270ec9adc768281d950c68fa26a1199b6d95b34c7b8`
- `mmv-devkit-linux-amd64-v2.6.3`: `fa043902a7fa42abf59abecf6ea93c9c09fdd0aced2fe3b7f3641e65469fc024`
- `mmv-devkit-windows-amd64-v2.6.3.exe`: `2650e0ef8032682f0ff7fe5f8b1141a5574bc69f4f8043c2344408956d15ea0f`
- `skill.zip`: `20e99c37969acbf40ee060650e47b55991b223928df0f271cf503b3fa4d684a8`

## Canonical atlas

The existing Vanilla Feature Atlas remains the discovery/provenance source. This change modifies **sequencing and authorization**, not its historical data.

## Publication

- Drive source ZIP: `1h4dHhyBuijZVCcM2Bjrip0p196YHsB40`
- Drive Linux binary: `1eqZSzqXfqrNofmyiI0qH1m8BPNIx2vlo`
- Drive Windows binary: `1opmCGsYbUfd1Ae6buLjaHLCkE3ZqYpnR`
- Drive versioned skill: `1XbZNF9BrimkOJfSUqLHPfpLiWMNdCPOS`
- Drive stable skill alias (updated in place): `1q1zpjLUMqb53Gls4soprOntp8F30UBHo`
- The public repository's live `devkit-orchestrator/` tree is older than this verified package; this checkpoint deliberately does not overwrite that tree piecemeal. The verified 2.6.3 source package on Drive is the canonical source checkpoint for this release until a coherent full-tree promotion is made.
