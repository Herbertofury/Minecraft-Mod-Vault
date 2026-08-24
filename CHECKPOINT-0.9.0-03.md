# Minecraft Mod Vault 0.9.0 - Checkpoint 03: verified release

Recorded: 2026-08-21T04:35:00Z
Implementation commit: `4191fad`
State: **all implementation, runtime, UI, rollback, restart, source rebuild, and fresh-package gates green; final remote publication pending**

## Verified release identity

- Linux x64 binary: `0e6720b4b1b9a6bb31f90fe2d498a7f4fbfac43df264d7e87bd23062d89ed2ee`
- Windows x64 binary: `6b5967461a34cb8658d93644f9d60196bb3bdbd162519f9dda1aff3ef4fa9598`
- Compatibility Brain seed: `eae01d1588056a1b7a680240752f3bd45add28b956cf115c3c1179a64f18d2d4`

## Acceptance gates

- Complete source convergence checks pass.
- Repair Lab production UI passes 25 assertions/38 authenticated calls.
- Porting Lab production UI passes 29 assertions/23 authenticated calls.
- Source fixture migrates, refuses unacknowledged execution, builds, hashes its artifact, exports evidence, persists, and rolls back.
- Managed JAR forensics, plan, workspace, exact duplicate quarantine, receipt, and restore pass.
- Fresh runtime creates the entire SQLite brain from embedded seeds.
- Fresh package manifests pass, source rebuild is byte-identical, extracted runtime reruns both UI suites, and state survives restart.

Next action is publication only: deterministic final archives, Drive full-byte verification, GitHub source/release publication if supported, and a provider receipt.
