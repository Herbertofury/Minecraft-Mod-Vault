# Creeperella: Bloom & Boom 1.4.21 — Forge 1.20.1

Release gate: Minecraft 1.20.1 / Forge 47.4.23 / Java 17.

## Release decision

**PASS / canonical 1.4.21.** This release brings Blossom Creeper to Cherry-quality standalone depth while preserving Cherry's protected source-authored baseline and re-running the mod's deterministic, dedicated-server, and real Forge client gates.

## Blossom parity work

- Separate deterministic Blossom model layer rather than a texture-only Cherry alternate.
- Seven Blossom-only sakura silhouette accents: crown center/left/right, back bloom, left/right shoulder blooms, hip bloom.
- Subtle Blossom secondary motion with per-frame bind-pose reset protection.
- Blossom Powder + Forge gunpowder tag.
- Dedicated Blossom loot table with vanilla skeleton music-disc behavior.
- Blossom TNT, recipe, primed entity and renderer.
- Real QA path calls `BlossomTntBlock.prime`; it does not fake the primed entity directly.
- Blossom Creeper Head block/wall block/block entity/item renderer and charged-Creeper head drop.
- Blossom Creebet painting variant/item.
- English/Russian localization and creative-tab entries.
- New `/creeperella qa family`, `cherry_family`, `charged_family`, and `blossom_showcase` gates.

Protected Cherry model/runtime assets remain hash-identical to the approved 1.4.20 baseline.

## Verification

- Final JAR: 281,868 bytes; SHA-256 `1cfb1a9eb938bf038709ea851ea112d075f6d427a38ff65031d75fc02f4d05c5`.
- Verified source ZIP: 15,265,886 bytes; SHA-256 `9f0f0dfe69e2bb5b1821ce5a9a2d0c7b74031573bba205a065e08809e76b9072`.
- Native/regression evidence ZIP: 70,881,621 bytes; SHA-256 `991401725b7821ca4de6a38674c5beb1e7675c04ebc8c3277cfc7aca35944d3a`.
- Offline Java 17 / Gradle 8.8 / ForgeGradle 6.0.54 clean build + `reobfJar`: PASS.
- Exact final source ZIP fresh rebuild: 239 vs 239 JAR entries; identical entry names; 0 packaged-content differences.
- Final packaged-source dedicated Forge server: `Done (3.491s)!`.
- All 16 blocking deterministic/static gates pass.
- Cat native reel: 5/5 phases, 14 live entities per phase, final clear 14.
- Native family reel: `expected=6 spawned=6`.
- Normal Cherry/Blossom native comparison: `spawned=2`.
- Charged Cherry/Blossom native comparison: `powered=2 spawned=2`.
- Blossom sidecar/TNT gate: `tnt=1 head=1 primed_via_block=1 items=3`.
- No Creeperella-specific missing texture/model/atlas warning in the final family client log.
- Core JAR, source ZIP and QA evidence were downloaded back from Google Drive and compared byte-for-byte: PASS.

## Canonical Drive artifacts

- JAR: https://drive.google.com/file/d/1YaBGM-v25MB7_CqlAoX3c1gDuN6V_3Dg/view
- verified source: https://drive.google.com/file/d/183jOS_rJQCXVad43AaVLaQY4akG6Vzby/view
- native/regression evidence: https://drive.google.com/file/d/1PJDTmq23EomFkYXEcxN8qnTL0BjvWePx/view
- final verification: https://drive.google.com/file/d/1Les-kVIQRt2D6yGjecA2MiQGTUD7vHIU/view
- SHA-256 manifest: https://drive.google.com/file/d/1zNqyQBRrUWqoVKsUOKH8uGjdUpJyMtLE/view
- normal Cherry-vs-Blossom native still: https://drive.google.com/file/d/1iJVc7b_cUpDKmcfzZe9nPtgXINDC1Zbv/view
- charged Cherry-vs-Blossom native still: https://drive.google.com/file/d/1b_62Qpzl74Yt72PNutDHIhmVc0NNsemm/view
- Blossom sidecar native still: https://drive.google.com/file/d/1X2os5xKLaufTwj9EfZbXVPc2eOTNGi-5/view

## Explicit limitation

The Dev Kit Client Asset Prewarm folder still contains only the prewarmer/checksum, not the full Mojang 1.20.1 external asset-object cache. The isolated runner therefore cannot certify vanilla external sound-object completeness. This does not affect the proven entity/model/texture gate: the real Forge client loaded the visual atlases, joined the integrated server, rendered the mod, and completed all native phases. The supplied source packs themselves contain zero sound assets.

## Anti-stall execution rule

This release also hardened the project workflow: preserve a complete Drive source checkpoint before long native runs, detach clients, poll authoritative milestones, and resume from the last passed gate instead of restarting broad scans after a runner timeout.
