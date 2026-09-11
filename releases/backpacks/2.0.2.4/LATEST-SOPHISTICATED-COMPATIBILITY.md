# Backpacks 2.0.2.4 - Latest Sophisticated Stack Compatibility Audit

**Status:** NATIVE CLIENT + DEDICATED SERVER COMPATIBILITY CERTIFIED  
**Audit timestamp (UTC):** 2026-09-11T04:54:53Z  
**Target:** Minecraft 1.20.1 / Forge 47.4.23 / Java 17

## Frozen release identity

The certified Backpacks 2.0.2.4 artifact was not modified during this audit.

- File: `Backpacks-2.0.2.4-FINAL-Forge-1.20.1.jar`
- Size: 2,906,589 bytes
- SHA-256: `252e83b9f7fbe20fd15a4de047d4b02f1abcf87852dbbe7944beb386a85b1342`
- Result: hash remained unchanged after all compatibility testing.

## Latest dependency matrix tested

| Component | Version / identity | SHA-256 |
|---|---|---|
| Backpacks Java port | 2.0.2.4 FINAL | `252e83b9f7fbe20fd15a4de047d4b02f1abcf87852dbbe7944beb386a85b1342` |
| Sophisticated Backpacks | 3.26.3.2157 | `6ab4a0cb739a6c11442a73d1b5f42c40c2ae93b31162fa63d81e1b66706e9e31` |
| Sophisticated Core | 1.5.1.2335 | `f4038a2e0d0e3504258fbe315cb27a37f1a769d71985bfa20afb983a3bdfab64` |
| Curios compatibility build | 5.14.1 + 1.20.1 Performance-Leak-Fix REPLACEMENT-3 | `d7ef9c4ada0ed0415b138b84ca0ddd01d0a295abab3ebe2a7f699e7cef81550d` |
| VanillaBackport | 1.1.7.10 | `157f95106c7f00f92b05c72f25e2d87e2ffbef04137e9ea609fc03838bb39e2d` |
| Platform | 1.3.4 | `e35c013fb8cbd62900ccdf0872fc6c14b9858a711f6b2279c479f4b4f668d870` |
| Forge packaged server payload | 1.20.1-47.4.23 | `dc628857ae45ac198417781c6712b0a973cf026488cf080ed54e03b23fb8d161` |
| Java runtime | Temurin 17.0.20.1+1 | runtime identity verified |

## Upstream provenance

### Sophisticated Backpacks

- Baseline used by the original certification: 3.26.0, commit `6e8d930ac25d019621a267ae63d4c4cc44a1ff5b`.
- Latest tested 1.20.x head: commit `af44319ac81d062ac0600daae240daa344da0076`.
- Upstream GitHub Actions run: `34392261051` (successful Mod Build).
- Build-libs artifact ID: `10120256831`.
- The actual tested JAR reports version `3.26.3.2157` in `META-INF/mods.toml`.
- The baseline-to-current Git delta changes linked-storage/wrapper/context and translation/test files. None of the direct classes consumed by Backpacks 2.0.2.4 changed: `BackpackItem`, `BackpackBlock`, `BackpackBlockEntity`, `IBackpackWrapper`, `ModBlocks`, or `BackpackBlockEntityRenderer`.

### Sophisticated Core

- Baseline used by the original certification: 1.5.0, commit `e5c526bdb5a46b9c1cf8cff05fd67497b174d459`.
- Latest tested 1.20.x head: commit `d15c18345b4815cef3a0974578b89f0da0a08b3e`.
- Upstream GitHub Actions run: `34253827802`; its `Build Mod` job and CurseForge/Modrinth/GitHub Packages publication jobs all succeeded. The workflow's overall red conclusion came from the separate Code Quality job, not compilation, packaging, or publication.
- Build-libs artifact ID: `10070889644`.
- The actual tested JAR reports version `1.5.1.2335` in `META-INF/mods.toml`.
- The baseline-to-current Core delta is limited to linked-storage/Create integration, associated mixin/interfaces, `gradle.properties`, and language data. The direct `IStorageWrapper` and `InventoryHandler` ABI paths used by the port remain unchanged.

## Declared dependency-range check

The frozen 2.0.2.4 `META-INF/mods.toml` permits:

- Sophisticated Backpacks `[3.26.0,)`
- Sophisticated Core `[1.3.84,)`
- Curios `[5.14.1,)`
- VanillaBackport `[1.1.7,)`
- Forge `[47.4.0,)`
- Minecraft `[1.20.1,1.20.2)`

The latest tested Sophisticated Backpacks/Core versions satisfy those ranges.

## Native dedicated-server proof

The exact Forge 47.4.23 packaged server was run with the frozen Backpacks JAR and the latest dependency matrix above.

- Fresh latest-stack server reached `Done (23.740s)!`.
- Clean follow-up boot reached `Done (5.447s)!`.
- RCON interaction boot reached `Done (5.653s)!`.
- QAPlayer completed the Forge mod handshake and joined the real server.
- The final server shutdown was issued through RCON and logged `Stopping server`, `Saving players`, `Saving worlds`, and `All dimensions are saved`.

## Packaged native-client proof

The production Forge client package was reconstructed from the previously verified Forge-client export. All 3,575 unique Mojang asset objects were revalidated before launch. The client used Temurin 17.0.20.1+1 under Xvfb/Mesa software rendering.

### Pass 1 - world join and Guide UI

- The client discovered the exact latest-stack JAR filenames in its real `work/mods` directory.
- It connected to `127.0.0.1:25565` and the server logged QAPlayer joining.
- The port emitted its normal activation chat message.
- The auto-provisioned Guide was used and the real Backpacks 2.0.2 Guide screen rendered successfully.
- The player inventory/Curios-capable screen rendered without linkage failure.
- Client exited through X11 `WM_DELETE_WINDOW` and logged `Stopping!`.

### Pass 2 - live Curios Warden render

The client was restarted against the same latest stack. Through localhost-only disposable RCON, the server:

1. switched QAPlayer to Creative;
2. created a deterministic stone QA platform;
3. teleported QAPlayer to it;
4. set midnight/clear weather; and
5. executed `curios replace back 0 QAPlayer with sqst_bkpk:backpack_23`.

Curios returned: `Replaced slot back for QAPlayer with [Warden]`.

The client was switched to third-person rear view. The actual Warden backpack rendered on QAPlayer's back at midnight, including its bright cyan/emissive geometry. This exercises the port's registered `ScaiCurioRenderer` in a real packaged Forge client against Sophisticated Backpacks 3.26.3.2157 and Sophisticated Core 1.5.1.2335.

## Strict failure scan

Across the preserved client/server logs from both client passes and the corresponding server passes:

- `NoSuchMethodError`: 0
- `NoSuchFieldError`: 0
- `AbstractMethodError`: 0
- `VerifyError`: 0
- `LinkageError`: 0
- `ClassCastException`: 0
- `Unknown recipe category: sqst_bkpk:backpack_crafting`: 0
- candidate-owned `sqst_bkpk.*ERROR` / `ERROR.*sqst_bkpk` signatures: 0

Observed non-candidate noise was limited to expected no-network Mojang/Yggdrasil lookups, optional integration/mixin class probes for mods not installed in the disposable matrix, generated-config correction notices, and an upstream dedicated-server client-class probe warning. None produced a Backpacks crash, missing mandatory dependency, binary linkage failure, or failed gameplay interaction.

## Evidence

Preserved evidence includes:

- first client/server native logs;
- second client/server native logs;
- RCON interaction transcript;
- real world-join screenshot;
- real Guide UI screenshot;
- real inventory screenshot;
- real Warden-on-Curios-back midnight screenshot;
- actual JAR `mods.toml` snapshots for Backpacks/Sophisticated Backpacks/Sophisticated Core;
- machine-readable strict failure counts and hashes.

Full evidence ZIP (Google Drive): https://drive.google.com/file/d/1iCEgDhxdYyyOjUPVADrj-q5yKrO_6vM1/view

Compatibility checksum manifest (Google Drive): https://drive.google.com/file/d/1zuino0t7e-Iq1xHwhNcsqSpH7kUlJe1y/view

## Conclusion

**No compatibility patch is required.** The exact frozen Backpacks 2.0.2.4 JAR remains the release artifact, and it is now additionally certified against the tested latest Minecraft 1.20.1 Sophisticated stack: **Sophisticated Backpacks 3.26.3.2157 + Sophisticated Core 1.5.1.2335**.

This audit extends compatibility evidence; it does not replace or weaken the original 2.0.2.4 release certification.