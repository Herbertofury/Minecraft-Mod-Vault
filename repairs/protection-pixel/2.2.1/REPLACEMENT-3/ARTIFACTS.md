# Protection Pixel 2.2.1 — Event-Driven Armor Render Index — REPLACEMENT-3

Target: Minecraft 1.20.1 / Forge 47.x / Java 17

REPLACEMENT-3 replaces the prior per-tick armor target cache with a block-entity lifecycle-maintained sparse index. Steady render discovery performs no chunk scan and no unrelated block-entity scan.

## Final install artifact

- `Protection-Pixel-2.2.1-Event-Driven-Armor-Render-Index-REPLACEMENT-3.jar`
- SHA-256: `2a750dcfda368e2e205587aae26d81a20ba66c73e57f8b4ab705a68956bd3613`
- Size: 3,251,907 bytes
- Drive file ID: `1a5_wTzfQTobbfh0d9-3JvksKH9zGCHyu`
- Drive folder ID: `1xJ-cw4uHGDuLEEXvrHaS-dPQaK1vh2LP`
- Install mode: **REPLACEMENT** — remove stock/R1/R2 and install only R3.

## Architecture

- `ArmorhangerBlockEntity.onLoad()` / `ArmorloadplatformBlockEntity.onLoad()` register actual targets.
- Existing `setRemoved()` paths unregister targets after upstream cleanup.
- Render callback iterates a reusable sparse target view only.
- No render/tick chunk grid discovery, no unrelated block-entity scan, no tick subscriber, no timer, no background thread.
- Actual target rendering remains every frame with upstream radius/visuals/content intact.
- Active world identity uses a weak reference replaced only on world transition; normal render binds allocate no new weak reference.

## Verification

- ZIP integrity PASS, 1,407 entries, 713 Java 17 classes.
- Whole-JAR ASM BasicVerifier: 713 classes / 3,202 methods / 0 failures.
- R2 -> R3: add 5 index classes, remove old `ArmorRenderTargetCache.class`, change exactly 3 existing classes, metadata-only changes 0, 1,399 existing contents unchanged.
- 50,000 randomized lifecycle operations / 17,786 render queries PASS.
- Stress world included 20,000 unrelated block entities without lookup cost scaling to them.
- Repair Mark v2 gate PASS; unmarked -> final changes exactly `logo.png`.
- Neutral/private-label scan PASS.

## Durable artifacts

- Verification report Drive ID: `1XPhoDFZubdC_NBxxRIxtpHXBl2Rx7gDf`
- Repair record Drive ID: `1ZtNB5Lx_-DTnEoSdPypbO7uEyFSTehlm`
- Source/evidence Drive ID: `1AEtAgoLG6zrlbsDV_AeAcmsQDOAglIDs`
- Repair Mark Drive ID: `1tyRabQ4oWvZCe7kic3LwQ4mKofzHn8BG`
- 48px Repair Mark Drive ID: `13OMWM7B2R5kQP1q6BgzfvYqeB-c7aF5E`
- All six Drive deliverables were redownloaded and SHA-256 verified.
- Repair Brain record: `mc-1.20.1-forge-protection-pixel-2.2.1-event-driven-armor-render-index-replacement-3-2026-09-14`
- Repair Brain final SHA-256: `b454f0cfddd51cb47990b1aee443d64d64407a8c991df61ad4f19bc6b26c920d`
- Reusable global policy immediately preceding it: `minecraft-global-hot-path-event-driven-indexing-policy-2026-09-14`.

## Remaining gate

Native target runtime/Spark validation is still required before calling this exact SHA runtime-verified. Run `/sparkc profiler start --timeout 90 --thread *` after exercising existing/new/removed armor hangers/load platforms, equipment changes, chunk/radius/dimension transitions, ALT+U, and Save/Quit/reload.
