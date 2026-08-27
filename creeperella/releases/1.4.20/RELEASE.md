# Creeperella: Bloom & Boom 1.4.20 — Forge 1.20.1

Release gate: Minecraft 1.20.1 / Forge 47.4.23 / Java 17.

## Why 1.4.20 supersedes 1.4.19

Real Forge client rendering exposed a Cat rest-state transform accumulator that deterministic/software animation QA could not see. The custom Cat `waist` parent is outside vanilla `CreeperModel`'s animation reset set; rest translations/rolls accumulated every rendered frame until the complete Cat rig moved out of view while its shadow/particles remained. 1.4.20 restores the authored bind pose with `waist.resetPose()` and `body.resetPose()` before applying each frame's source/rest layers. Approved 1.4.19 geometry, IK, grounding, tail motion, and personality curves are preserved.

The same native gate also exposed Cherry/Blossom item/block atlas warnings. 1.4.20 replaces invalid entity-texture atlas references with deterministic source-derived item/block textures.

## Verified gates

- `gradle --offline --no-daemon clean build`: PASS, including `reobfJar`.
- Final JAR: 244,958 bytes; SHA-256 `a29acf3296a10f1fdd15c35484ec2ce53d519b17f9f4fe3e7692cb7aa1285b62`.
- Fresh verified-source rebuild: 206 packaged entries, 0 content differences versus release JAR.
- Post-build dedicated Forge server: `Done (2.805s)!`.
- Native Forge/Xvfb/Mesa client reel: PASS.
- Native Cat grids: `NONE`, `TAIL_WAG`, `KNEAD`, `CURIOUS_LOOK`, `PAW_LIFT` each logged `spawned=14`; final clear removed 14.
- Exact Cat gait/secondary-motion gates: 17,328 locomotion samples each.
- Cherry animation parity: 235 sampled evaluations.
- Pie animation parity: 712 sampled evaluations.
- Supplied source-pack accounting: 45/45.
- No Creeperella-specific missing-texture/model/atlas warning remained in final native client log.

## Canonical Drive artifacts

- JAR: https://drive.google.com/file/d/1d88CBKo6_WNPoZTpFnDEOt7WQsuWBF09/view
- Verified source ZIP: https://drive.google.com/file/d/1GnFYV6OsdWTGI8WjWF1DfRHBX6SHrYEF/view
- Native Forge client QA evidence: https://drive.google.com/file/d/1EiCdy9crgPhvS4n5xDrgQfXKLwFZNxBe/view
- Final verification: https://drive.google.com/file/d/1ivhjO3VrgBX9XSPREQCY5bemACf7qVUq/view
- SHA-256 manifest: https://drive.google.com/file/d/1JHxxFnG7lONhWhE_8JC6v1rxZL5JNjx-/view
- QA Toolkit 1.2.0 in Minecraft Dev Kit: https://drive.google.com/file/d/1p0aUgOCHJDJCtGPPVfiWsLK7jfmwELhy/view

## Native QA contract

The hardened 1.2.0 client-QA toolkit seeds a fresh deterministic save, keeps MONSTER-category QA Cats alive using Normal difficulty plus `doMobSpawning=false`, drives phases from an in-world scheduled datapack rather than chat typing, and refuses to pass unless all five Cat phases contain a complete 14-entity grid. X11 input is not the correctness source of truth.

Full external Mojang asset objects were unavailable in the network-isolated runner. Visual-only QA explicitly proves vanilla texture resources from the cached 1.20.1 client JAR; missing external sound objects are reported as a warning rather than silently treated as complete. Set `VISUAL_ONLY_OK=0` when the full Mojang object cache is required.
