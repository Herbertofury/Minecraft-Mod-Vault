# Protection Pixel 2.2.1 — REPLACEMENT-2

Target: Minecraft 1.20.1 / Forge 47.x / Java 17

This checkpoint records the Protection Pixel armor-render discovery optimization layered on the previously accepted UI/QoL repair.

## Final install artifact

- `Protection-Pixel-2.2.1-UI-and-Armor-Render-Performance-Fix-REPLACEMENT-2.jar`
- SHA-256: `8805c7b504d7f7e68c714770cbfdd4551c492a423e722ce970f71dde89297511`
- Size: 3,247,645 bytes
- Install mode: **REPLACEMENT**
- Drive file ID: `1EUXPA4RZVycjSV6VD3IJ2JcBy7JcXml_`
- Drive folder ID: `1kcaoFsI7P1qQ0XyTHqt1Cq0IX0BxX83X`

## Static / behavioral proof

- ZIP integrity: PASS
- Whole-JAR ASM BasicVerifier: 709 classes / 3,177 methods / 0 failures
- R1 -> R2: add exactly `ArmorRenderTargetCache.class`, change exactly `ArmorrenderProcedure.class`, remove 0, metadata-only change 0
- `ArmorrenderProcedure` descriptor surface unchanged from R1
- Existing 13 R1 repair-layer classes and UI mixin behavior retained
- 10,000 randomized cache-equivalence/invalidation cases: PASS
- Final cache stores target positions only and holds client-level identity through `WeakReference`; no static strong world/BlockEntity retention
- Known targets are re-resolved live every rendered frame; full render-distance discovery is memoized per stable game tick/player chunk/render distance/world identity
- No thread, scheduler, off-thread world access, content/radius/fidelity reduction, entity cap, or preview-render throttling added
- Repair Mark v2 gate: PASS
- Neutral/private-label scan: PASS

## Persistence

- Verification report Drive ID: `15njxmTDqazVuDQgPWShEnvNroKZjATR_`
- Repair record Drive ID: `1hWpxJWTAtY6FZfyED5Fele5xwebGdG74`
- Source/evidence Drive ID: `1X61LkAXtwcawtWKa_whGrHZbOrR1Q7fc`
- Repair Mark Drive ID: `1ttVifQ3GHXM1kmb28cqER966a2vWo0ln`
- 48px QA mark Drive ID: `117fh7fexdinj7MfrZcTALvU8THZUTXmq`
- All six Drive deliverables were redownloaded and SHA-256 verified after upload.
- Repair Brain append preserved the previous 725,564 bytes exactly; new canonical SHA-256: `a07ba6103572efa8fa98ac252659d8f4743a4fee5c0f6d7c993685bc2adc0a06`
- Repair Brain record ID: `mc-1.20.1-forge-protection-pixel-2.2.1-armor-render-target-cache-replacement-2-2026-09-14`

## Runtime gate

The exact final SHA is not yet called runtime-verified. Install only REPLACEMENT-2, exercise armor hangers/load platforms plus ALT+U, Save & Quit/reload, then run `/sparkc profiler start --timeout 90 --thread *` and compare `ArmorrenderProcedure.execute` against the prior profile.
