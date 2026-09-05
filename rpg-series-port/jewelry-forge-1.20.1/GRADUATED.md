# Jewelry 2.4.0 -> Native Forge 1.20.1 — Graduated

**Graduated:** 2026-08-26 MDT  
**Repository:** `Herbertofury/Minecraft-Mod-Vault`  
**Branch/head:** `workbench/rpg-series-forge-1.20.1` @ `5716d7a8ed14e660a547f24d87547f841349f45f`  
**PR:** #4  
**Final acceptance run:** #105 / `32955574401`  
**Final CI artifact:** `9601931317`

## Exact source pins
- 1.20.1 substrate: `ZsoltMolnarrr/Jewelry@f20b7d94c4c6cdd5a4ed26e4066374b64654fb96` (1.3.7)
- current target: `ZsoltMolnarrr/Jewelry@572cb8759d13075b97e7a1acd969a6203db594cb` (2.4.0)

## Canonical artifacts
- `jewelry-forge-2.4.0+1.20.1.jar`
  - SHA-256: `0ffe10e010d03db0d698052ddc42d3b138061b7416d8a8cd9307df754f13946d`
- `jewelry-2.4.0-forge-1.20.1-source-ci.zip`
  - SHA-256: `ed796bbb5013b5631f4e89b2e040a7feb71d0aac6ea9b718bc022462588eca34`
- run #105 evidence ZIP
  - SHA-256: `2ae1107e5e7c784e9a4d9fa23ba7f89a74ce8175a1bca89d194666f96f5e0dda`
  - includes the actual release JAR, source ZIP, package logs, client log, and foundation artifacts.

## Acceptance evidence
- Full Jewelry 2.4.0 item/unique catalog and 23 language files retained.
- Current Spell Power / Ranged attributes and recovery/recipe/resource data retained.
- Native Forge 47 registry writes use Forge-owned registry helpers; no late vanilla registry mutation.
- NeoForge biome modifier resource translated to Forge namespace/codec while preserving overworld gem-vein semantics.
- Curios is optional, matching current upstream metadata intent.
- Fresh Forge 47.4.23 packaged-server boot succeeds with Curios absent.
- Fresh Forge 47.4.23 packaged-server boot succeeds with Curios present.
- Forge headless dev-client bootstrap succeeds with current RPG dependencies + Curios.
- Ranged Weapon API's embedded MixinExtras remains owned by Ranged; Jewelry's dev-runtime bridge does not duplicate it into Jewelry's release JAR.
- Canonical source ZIP contains the dev-runtime fix.
- Final workflow artifact was corrected to include `generated/forge/build/libs/*.jar` and run #105 proves the release JAR is present.
- Final evidence ZIP was uploaded to Drive and downloaded back with the exact same SHA-256.

## Known non-blocking warnings
- Forge/Yarn deprecation warnings (`FMLJavaModLoadingContext.get()` and `Identifier(String,String)`) remain compile-time warnings only. They are not runtime failures and are safe to clean up in a later API-polish pass.

## Suite state
Jewelry is now a graduated foundation/content module and is ready to participate in whole-suite integration testing. The active lane can move to the next RPG Series module.
