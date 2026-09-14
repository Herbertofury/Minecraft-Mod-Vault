# Grassier Grass 1.4.5 Render Performance Fix — REPLACEMENT 2 Artifacts

Canonical repair ID: `mc-1.20.1-forge-47.4.20-grassiergrass-1.4.5-sectionoffset-uniform-cache-regression-replacement-2`

## Runnable replacement

- File: `Grassier-Grass-1.4.5-Render-Performance-Fix-REPLACEMENT-2.jar`
- SHA-256: `91006e8220e65572d7e9e1e54b6058526c4adf908faedc750c9f19f13623df8c`
- Size: `2300791` bytes
- Google Drive file ID: `15tK6xmAeAE92S-QQdF8Xd4cBhPoXdlJu`
- Drive folder ID: `1UgypfL65fw105ZpxuNeKjJc5qABFXnhd`
- Install semantics: **REPLACEMENT** — remove the original / REPLACEMENT-1 Grassier Grass JAR and install only this JAR.

## Reproducible source / evidence

- File: `Grassier-Grass-1.4.5-Render-Performance-Fix-REPLACEMENT-2-SOURCE.zip`
- SHA-256: `6f69dfafa6731bed7501e1b09b23cbd50715c465755aaa832a89b99a260dcf34`
- Google Drive file ID: `19ftVafkBcICcBk5pz-uNqgfWjtf5ncqS`
- Drive readback: byte-for-byte verified after final repair-record refresh.

## Verification report

- File: `Grassier-Grass-1.4.5-Render-Performance-Fix-REPLACEMENT-2-VERIFICATION.md`
- SHA-256: `d7acc791e737fec7b4942a9f34a41329061cd683dd09ee6ae7a4f2ae22f6c0a3`
- Google Drive file ID: `16wR8B-ozzo22cT9PAGkGEK9rtHZjG9jW`
- GitHub copy: `VERIFICATION.md` in this directory.

## Repair Mark v2

- File: `Grassier-Grass-1.4.5-Repair-Mark-v2.png`
- SHA-256: `7e3abc159d75fe4ac7f2fc0a0ff3eaa3ff0a34bdf3b9f86d15726fc8d5a4f659`
- Google Drive file ID: `1Ow_xqIQhS4Wif3xLwljV6xjfhxxHWutz`
- Integration: embedded in upstream `logoFile="icon.png"` slot.
- Marker-only diff: PASS — exactly `icon.png` changed from unmarked R2; zero added/removed/metadata-only entries.

## Upstream baseline

- Upstream JAR SHA-256: `0bb02d32a840944feaba7f606823a0bf775f648f134749f988aab4a76a85ca3e`
- Unmarked R2 SHA-256: `f4d8d65d381422522577e6e668f1ebbfc2548e425c207b1b9bb6880a8d920dbd`
- Internal identity remains `grassiergrass` / `1.4.5` / `Grassier Grass`.

## Current certification state

Static/package/bytecode verification is complete and the REPLACEMENT-1 recursive uniform-cache regression is eliminated in the produced bytecode. Runtime certification remains pending until REPLACEMENT-2 is launched in the exact target instance and the same world/render path completes without the prior crash signature.
