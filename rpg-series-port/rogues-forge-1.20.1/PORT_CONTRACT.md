# Rogues 3.1.1 -> native Forge 1.20.1 port contract

Status: **GRADUATED — FULL NATIVE RUNTIME ACCEPTANCE PASSED**

## Graduation authority
- Repository: `Herbertofury/Minecraft-Mod-Vault`
- Canonical branch: `workbench/rpg-series-forge-1.20.1`
- Product/source identity authority: `77af5b416ab0889e770028ad3b93fbf2034d311c`
- Canonical QA-only replay authority: `1bbeb46d32fbbe21f5df99b6f7b593b2e4936611`
- Independent locked-identity replay: run #245 / `33269680789` / job `99145812360` — SUCCESS.
- Final canonical replay: run #247 / `33270906726` / job `99149089245` — SUCCESS.
- Exact release JAR SHA-256: `8845d27451fac83666b4ed4dd0ea5bac96c6a38b39c5d3585b8c9492274f3f23`.
- Deterministic source ZIP SHA-256: `44f1ee45b7f50bfd5ee4c370ae03a8eaad51d0dd1aa2dfa11acb8b09c2f7c8e8`.
- Final canonical evidence ZIP SHA-256: `2cc50d0b399012ebfb5eba4a53fa1f67ffd343486ef22a06f58a25a4e630a503`.
- Final canonical evidence size: `17,204,109` bytes.
- Final canonical evidence GitHub artifact: `9720186150`.
- Final canonical evidence Drive file: `rogues-run247-canonical-graduation-evidence.zip`, ID `1itlRnN-LCldj4M-uZkS8B7AOa3Xwn3uF`.
- Drive round-trip: PASS; re-downloaded bytes retained exact size/SHA-256 and ZIP integrity.

The documentation-only graduation commit after this receipt does not replace the runtime/product authority above.

## Authorities
- Current feature/content authority: `ZsoltMolnarrr/Rogues` `1.21.1` commit `d4a7af565559dcff4384eabb2481f63eb5f97d55`, tree `c6fac5d7c80807b41668274843c9732762d7cb03`, version 3.1.1.
- Historical mapping/API substrate only: `1.20.1` commit `bdfe6447b90758129e12430b497d97c181222b12`, tree `0a4f7f94e77031732843f8a40a8460184bd3577a`, version 1.2.0.
- Target: Minecraft 1.20.1, Forge 47.4.23, Java 17.

## Non-negotiable preservation
Current 3.1.1 owns features and content. The historical branch may prove target-era names/signatures only and may never replace current implementations wholesale. Preserve Bear Trap, current warrior/rogue spell suite, current equipment/armor definitions, effects, villages/trades, generated recipes/tags/lang/sounds, and optional current equipment integrations.

## Dependency ownership
Armor Model API, Structure Pool API, Spell Power, Spell Engine, TinyConfig, Cloth Config, Player Animator, and Curios remain separate real dependencies. No foundation source injection and no dependency shading into the Rogues release.

## Acceptance result
1. Exact upstream pins and deterministic source materialization — PASS.
2. Current 3.1.1 common source/resources survive the 1.20.1 translation — PASS; 427 JSON resources validated in the final replay.
3. Native Forge registration owns registry mutation/phase ordering — PASS.
4. Java 17 common + Forge compile and remapped package — PASS; 39 Rogues-owned classes passed the Java 17 package gate.
5. Packaged dedicated Forge server reaches ready state with the exact release candidate — PASS.
6. Native LWJGL Forge client reaches post-bootstrap and a real client joins the packaged server — PASS; `ROGUES_REAL_PLAYER_JOIN_PASS`.
7. Semantic gameplay acceptance — PASS: Charge lifecycle, Stealth visibility/break cleanup, Bear Trap registration/behavior, ROOT movement+jump blocking with attacks allowed, SHOCK full-stun attack blocking, representative weapon/spell/equipment data, and positive post-clear controls.
8. Deterministic release/source identity is certified — PASS; exact JAR/source hashes remained locked across independent and canonical replays.
9. Pre-promotion acceptance then exact promoted canonical end-to-end replay — PASS; #244 first deep, #245 independent replay, canonical promotion, #247 final canonical replay.
10. GitHub + Google Drive evidence persisted and byte/hash verified — PASS.

## Important production-runtime repairs retained
- Final JAR manifest explicitly activates `rogues.mixins.json`; dev-only `--mixin.config` is not relied upon.
- `LivingEntityStealth.updatePotionVisibility` uses a standard remappable Sponge `@Redirect`; packaged refmap resolves the production method to `m_8034_()V`.
- Stealth/Shadow Step target suppression remains enforced at both universal `MobEntity#setTarget` and `MobEntity#getTarget` boundaries outside configured reveal range.
- The final canonical target-acquisition QA polls the live Minecraft `execute on target` relationship across a bounded window rather than relying on a single timing-sensitive sample.

Final deep marker: `FULL_DEEP_BEHAVIOR_PASS`.
