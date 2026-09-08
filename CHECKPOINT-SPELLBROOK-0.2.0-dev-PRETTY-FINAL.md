# Spellbrook Forge 1.20.1 — 0.2.0-dev PRETTY FINAL

Status: **PASS — release-polish checkpoint**

This checkpoint records the final polished Spellbrook candidate for Minecraft 1.20.1 / Forge 47.4.23 / Java 17. The canonical binary/source/evidence artifacts are stored in the Spellbrook Google Drive project folder; GitHub keeps the durable release identity, hashes, and QA summary without committing large generated JAR/ZIP artifacts into the vault history.

## Final artifact

- JAR: `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL.jar`
- Size: **6,930,801 bytes**
- SHA-256: `124688c04f05cde1472092a5a0e36836ad29c35914b7641a3eba8f6a8211bb53`
- Drive: https://drive.google.com/file/d/1t5WdhWrBw_BNZO9rDu9Y_D3hus-4f-GQ/view?usp=drivesdk

## Source and evidence

- Source ZIP: `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL-SOURCE.zip`
  - SHA-256: `82bc4cbaea4f246b859de2675b0077d1c4131ec9e777d64226c00f2c6b81af18`
  - Drive: https://drive.google.com/file/d/1spi0nxpOsWh3Q04RQQKKW3Nm37AZEPSI/view?usp=drivesdk
- Native QA evidence ZIP: `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL-EVIDENCE.zip`
  - SHA-256: `8dd53fd56c702c01f4d8755f9124c26a22ad653e1336dab8f1069a8e3e668c33`
  - Drive: https://drive.google.com/file/d/1Qqf41f5If1RaplLzlNhnMoA-T9s1e-cF/view?usp=drivesdk
- SHA-256 manifest: https://drive.google.com/file/d/1-AUEVXKoOtg1WTybZ14dJSjV0CIzi10E/view?usp=drivesdk
- Final verification report: https://drive.google.com/file/d/1GynyxdlUzsQ2W2xATAh0mlkqVM1X0Jm7/view?usp=drivesdk
- Dedicated-server log: https://drive.google.com/file/d/1X2sOxu6MO6-6dR0kB-6mUYRvXSN8fAnQ/view?usp=drivesdk

Canonical Drive folder: https://drive.google.com/drive/folders/1xA5xcdiMVD3kDo0UD5l4ouI_goWMsZFg

## Polish pass

- Centralized player-facing content naming, category presentation, and rarity styling while preserving internal content IDs.
- Added variant-specific broom/core presentation with flight/handling and ownership/mastery information.
- Cleaned armor, hats, tools, furniture, vessels, compass, and whistle tooltips; removed broken control characters and developer-facing placeholder copy.
- Reorganized the creative tab around player progression rather than raw asset order.
- Polished Glimmer Store presentation with the captured Spellbrook store artwork, shipped Spellbrook font, item grid, category/page rail, status panel, Glimmer balance, and a real **Keys & Mysteries** filtered section backed by existing key/shard content.
- Polished Spellbrook Studies with captured school art, live stats, and a fifth Earth page.
- Native visual QA caught an Earth flavor-text overlap; the text was wrapped/repositioned and the exact screen was re-tested successfully.

## Verification

- `compileJava`: PASS.
- Final normal offline Forge build: PASS, including `reobfJar`.
- Final JAR resource audit: 1,382 Spellbrook JSON resources parsed successfully; required polished client classes/font/store assets present.
- Native Forge client reached a real integrated world (`Dev joined the game`).
- Native screenshots verified Featured Store, Keys & Mysteries, Fire Studies, and corrected Earth Studies layouts.
- Clean Save & Quit completed with all dimensions saved.
- Client task-owned scan: 0 Spellbrook runtime errors, 0 missing Spellbrook texture failures, 0 unable-to-load Spellbrook model failures, 0 DecoderException, 0 EncoderException, 0 wrong-side/invalid-message failures.
- Dedicated Forge server reached `Done (35.641s)!`, then on intentional shutdown saved players/worlds and reported `All dimensions are saved`.
- QA-only offline dependency alias used to restore the visual harness was not part of the shipped build.

## Known harness-only limitations

The restored offline visual harness cannot claim full external Mojang sound-object completeness. Sandbox Mojang/Yggdrasil/update-check DNS failures and the visual-harness external-sound warnings are environment limitations, not Spellbrook release failures.

## Lineage

This PRETTY FINAL is a presentation/UX polish continuation of the previously hardened Spellbrook candidate. Previously proven broom/provider/persistence behavior was preserved rather than redundantly reopened; only gates invalidated by the polish changes were rerun.
