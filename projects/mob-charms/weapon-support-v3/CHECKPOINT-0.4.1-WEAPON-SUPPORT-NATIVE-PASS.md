# Mob Charms 0.4.1 V3 - Weapon Support Native QA PASS

Date: 2026-09-06

## Result
PASS. The current Broom Dangle V3 release was exercised in the real packaged Forge 1.20.1 client across every distinct weapon presentation lane represented by the current support matrix. No production code change was required by this pass.

## Exact build under test
- Mob Charms V3 JAR: `mobcharms-0.4.1-dev-BROOM-DANGLE-NATIVE-PASS.jar`
- SHA-256: `f03c62f0cf644a36940be5846c57d814141eae999d209931183b078aed9eeb11`
- Punchy: 2.7e
- Punchy SHA-256: `459e6889c11ecfbd1faf13749dafa8eea8df39e4d569e8b2bb0cb80aae94dcf2`
- Minecraft: 1.20.1
- Forge: 47.4.23
- Java: Eclipse Temurin 17.0.20.1+1
- External provider used for current-provider coverage: Simply Swords 1.70.2-1.20.1

## Native matrix
One Bee charm was used for every stage so fit, chain, orientation and pendulum behavior could be compared directly.

1. `minecraft:diamond_sword` - original Charmsy full-item baseline
2. `minecraft:diamond_axe` - authored axe family
3. `minecraft:diamond_pickaxe` - authored pickaxe family
4. `minecraft:diamond_shovel` - model-preserving overlay
5. `minecraft:diamond_hoe` - model-preserving overlay
6. `minecraft:trident` - spear-style overlay
7. `minecraft:bow` - canonical sword-presentation overlay
8. `minecraft:crossbow` - canonical sword-presentation overlay
9. `minecraft:fishing_rod` - canonical sword-presentation overlay
10. `simplyswords:diamond_longsword` - model-preserving sword overlay
11. `simplyswords:diamond_spear` - provider spear overlay
12. `simplyswords:diamond_greataxe` - provider axe overlay
13. `simplyswords:diamond_greathammer` - provider mace overlay
14. `simplyswords:diamond_warglaive` - hard wrap + dangle
15. `simplyswords:diamond_twinblade` - hard wrap + dangle

Authoritative native result: 15/15 stage markers followed by `COMPLETE stages=15`.

## Visual review
- Original vanilla-sword Charmsy presentation remains intact.
- Weapon models remain native/preserved on overlay/provider lanes rather than being replaced by synthetic weapon geometry.
- Bee remains attached at the grip/attachment region across the tested vanilla and Simply Swords shapes.
- Chain/tether and mascot remain visible; no double weapon, giant charm, missing mascot, disappearing overlay or obvious grip inversion was found.
- Warglaive and twinblade special hard-wrap lanes remain readable and preserve the intended dangle.
- Motion sampling during the run shows the attachment moving with held-item motion rather than behaving as a static screen decal.

## Runtime regression scan
Within the second authoritative 15-stage run:
- Stage markers: 15/15
- Complete marker: PASS
- Mob Charms / linkage / mod-loading fatal matches: 0
- Task-related missing-model / missing-texture / bake-error matches: 0
- Integrated world saved overworld, Nether and End and returned to the title screen before clean client exit.

The reused disposable BroomFitQA save emitted pre-showcase noise because Hexerei was intentionally omitted from this weapon-only runtime while the save still contains old Hexerei registry/datapack state. Those warnings occurred before the authoritative weapon-QA start marker and did not affect the weapon matrix. They are harness/save noise, not a Mob Charms regression.

## Findings / next action
No release-blocking fix is justified by this pass. The only remaining quality opportunity is optional polish: bow, crossbow and fishing rod currently use the deliberate canonical sword-presentation overlay instead of bespoke per-item authored anchors. They are visually functional in this pass; changing them would be a quality enhancement, not a bug repair.

## Drive evidence
- Native 720p video Drive file ID: `1uk1ZAN6r3n-7-FZK1KRlvdIgwdlUceEP`
- Video size: 45,630,867 bytes
- Video SHA-256: `f2e6cc9a9b60e4974f6facde38685bf88b0e5beab908eda60636da481a90e554`
- Contact sheet Drive file ID: `1_NGWp6nCZKBexFAYCKKlncTl4F2YH3nu`
- Contact sheet SHA-256: `8580423b060aba71a7bfde7bbea933fc7dc82324d4d1f20c3a23650f091844d8`
- Both Drive objects were downloaded back and matched the local SHA-256 exactly.
