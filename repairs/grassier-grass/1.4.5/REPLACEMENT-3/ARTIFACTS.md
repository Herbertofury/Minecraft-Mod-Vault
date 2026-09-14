# Grassier Grass 1.4.5 Render Performance Fix — REPLACEMENT 3 Artifacts

REPLACEMENT-3 supersedes both REPLACEMENT-1 and REPLACEMENT-2.

## Runnable replacement

- File: `Grassier-Grass-1.4.5-Render-Performance-Fix-REPLACEMENT-3.jar`
- SHA-256: `a2122e4402cc583087cc2262f91dde458c05be5d9fc4eabb06df2ef7a0ed0dc5`
- Size: `2299859` bytes
- Google Drive file ID: `1rLqrCQ427gKNdPdLsuM_PPnGyv44O1iw`
- Drive folder ID: `16aIBwFn9HZQJ8s19B8534zzmLFJqx0tJ`
- Install mode: **REPLACEMENT** — remove original / REPLACEMENT-1 / REPLACEMENT-2 and install only this JAR.

## Verification report

- File: `Grassier-Grass-1.4.5-Render-Performance-Fix-REPLACEMENT-3-VERIFICATION.md`
- SHA-256: `6464b73a075f956768e5c7b6a20baaa692556fcefcc8d0355d8ca2769ad5bd3e`
- Google Drive file ID: `17Zrf1gdrVHuQXqRS2Wb_RtVeh5jXnrnJ`

## Reproducible source/evidence

- File: `Grassier-Grass-1.4.5-Render-Performance-Fix-REPLACEMENT-3-SOURCE.zip`
- SHA-256: `63d1f844c59d506f38d16dc3b4d277dfbceb6b3e4e2bfd4333955aac299116c8`
- Google Drive file ID: `1Vi3IPTUJppsH4zbGSfcAmOkUhZFN0Hnf`

## Repair Mark v2

- File: `Grassier-Grass-1.4.5-Repair-Mark-v2.png`
- SHA-256: `7e3abc159d75fe4ac7f2fc0a0ff3eaa3ff0a34bdf3b9f86d15726fc8d5a4f659`
- Google Drive file ID: `13YnJsdPN5R6pnoXO5-rXUeqSXf2qLPZx`
- Embedded into upstream `logoFile="icon.png"` slot.

All four Drive objects were re-downloaded and compared byte-for-byte against the local artifacts after upload.

## Safety hardening relative to R2

REPLACEMENT-3 removes the added cross-frame `SectionOffset` uniform cache entirely. The final class contains no `sectionOffsetUniform` helper and no cached shader/uniform fields. It restores the same two direct `ShaderInstance.getUniform("SectionOffset")` lookups as upstream while retaining the larger queue/culling optimizations.

## Certification state

Static/package/bytecode/frame/queue-state verification is complete. Exact full Noxviola runtime certification remains pending until this exact SHA is launched through the same world/render path.