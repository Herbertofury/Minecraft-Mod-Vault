# Future Feature Bridge publication manifest

Frozen: 2026-08-27
Target: Forge 1.20.1 / Forge 47.4.23

## Exact-final runtime identity preserved from prior QA

- Future Feature Bridge final candidate SHA-256 prefix: `b20acda1e2d...`
- Full digest and exact final bridge JAR/source bytes were not recovered in the resumed runtime. Do not fabricate or expand the digest.

## Recovered canonical baseline

- Google Drive project folder: `Ragdoll Forge 1.20.1 Backport`
- Checkpoint: `ragdoll-sable-forge-full-checkpoint-2026-08-27-full-compile-remap-server-green.zip`
- SHA-256: `9be8ed0148daf90c0bd157f34411a473fe6275a8d18c534c3edeb5d413af6d6a`

## Updated Minecraft Dev Kit skill

- Local/download artifact: `minecraft-dev-kit.skill.zip`
- SHA-256: `d417021ea7538ddc6091717dfee17ebad8bed6c8b9aa69b3f5f0eb6148b66c4c`
- ZIP integrity: clean
- Skill validator: passed (`Skill is valid!`)
- Canonical Google Drive path: `/Google Drive/ChatGPT Skills/Minecraft Dev Kit - skill.zip`
- Drive file id: `1q1zpjLUMqb53Gls4soprOntp8F30UBHo`
- Drive round-trip: re-downloaded bytes matched size and SHA-256 exactly.

## Google Drive durability docs

- Release evidence: `/Google Drive/Minecraft Dev Kit/00 Documentation & Index/Future Feature Bridge - Release Evidence - 2026-08-27.md`
  - file id: `1Yw-vrqRVD0yMf_6joXmgrxN8xwMgyCGD`
- Behavior spec: `/Google Drive/Minecraft Dev Kit/00 Documentation & Index/Future Feature Bridge - Spec - 2026-08-27.md`
  - file id: `1jVjHXmR1fqIBX2OSkAJdV1O4Cvce2psZ`
- Checksums: `/Google Drive/Minecraft Dev Kit/00 Documentation & Index/Future Feature Bridge - SHA256SUMS - 2026-08-27.txt`
  - file id: `1UG_GcT5pvOV3TEOsD9FYuW50M-m03buB`
- All three were materialized back from Drive after upload and matched their local bytes/hashes.

## GitHub durability paths

Repository: `Herbertofury/Minecraft-Mod-Vault`

- `verification/ragdoll-forge-1.20.1/future-feature-bridge/2026-08-27/RELEASE-EVIDENCE.md`
- `verification/ragdoll-forge-1.20.1/future-feature-bridge/2026-08-27/SPEC.md`
- `verification/ragdoll-forge-1.20.1/future-feature-bridge/2026-08-27/SHA256SUMS.txt`
- `verification/ragdoll-forge-1.20.1/future-feature-bridge/2026-08-27/minecraft-dev-kit/SKILL.md`
- `verification/ragdoll-forge-1.20.1/future-feature-bridge/2026-08-27/minecraft-dev-kit/future-feature-bridge-qa.md`

## Release lesson promoted into Dev Kit

Never accept mapped compilation, ordinary remap validation, or a dev launch as proof of production JVM symbolic linkage. Resolve and validate production `(owner, name, descriptor)` references and still exercise client-only surfaces such as Creative inventory on the exact release candidate. The bridge QA reference also makes embedded/external provider lanes, same-world provider removal, exactly-once semantics, uninstall cleanup, warning ownership, fixed-position fixtures, deterministic rebuilds, and exact-candidate hashing explicit release gates.
