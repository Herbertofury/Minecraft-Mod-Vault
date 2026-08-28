# Ragdoll Forge 1.20.1 — Final Release Checkpoint

Status: **RELEASE READY**  
Date: 2026-08-28  
Target: **Minecraft 1.20.1 / Forge 47.4.23 / Java 17**

## Final project JAR identities

- `sable_player_ragdoll-1.20.1-0.7.2.jar` — `0e9ee19b3b4b3efe4da07d2f7c47c29f98eeb805f96f95698163975fb71e1b47`
- `ragdoll_reactions-1.20.1-0.7.0.jar` — `ceb726e002bd0a3446012170b639c869605b2ae5f417546d428d543f171ac967`
- `ragdoll_reactions_future-1.20.1-1.0.0.jar` — `ad044549697c09a95800a933782fd62ab82e8d9e462e44e64996cdd3ae36b453`

Pinned dependencies: Sable `3e1b2a1e…`, Ragdollified `7d386138…`, Jade `a632c779…`, Curios `1e817919…`.

Optional JTS provider stack: Just Trial Spawners `ed3db6b8…`, FTB Library `f0933594…`, Architectury `218b471d…`.

## Static production gates

- Core compile: **125 Java sources -> 201 classes**.
- Production remap: **201 classes / 1751 method remaps / 603 field remaps**.
- Stale remap scan: **0 stale methods / 0 stale fields**.
- Production symbolic linkage: core **201/0**, reactions **49/0**, future bridge **12/0**.
- Core packaged classes are byte-identical to the newest remap output.
- Reactions and bridge class bytes are byte-identical to the gameplay-proven Candidate C classes; final hashes changed for release resource metadata only.

## Production fixes captured

- Direct Mixin targets use exact production SRG members where required.
- Reflection resolves mapped + SRG aliases for affected client model/renderer members.
- LambdaMetafactory `invokedynamic` SAM names are remapped by functional-interface owner + bootstrap SAM descriptor.
- `pack.mcmeta` is present for all three project mods.
- Invisible `ragdoll_seat` now has a valid zero-geometry blockstate/model, removing the last task-owned model-bake warning without changing appearance.

See `../final-2026-08-28/FINAL-SOURCE-DELTA.md` for the reproducible source delta.

## Runtime matrix

### Exact-hash dedicated server

The exact final project JARs and pinned dependencies reached Forge 47.4.23 dedicated-server `Done`, resolved the embedded provider, activated both ragdoll systems, and completed Minecraft's clean save/shutdown path through `All dimensions are saved`.

### Packaged production client — embedded provider

The real production-named `forgeclient` target (not userdev) reached an integrated world using the exact final JARs and official 1.20.1 assets.

Verified: embedded resolver active; both ragdoll systems active; `RagdollQA` joined; real Creative mode and Creative inventory opened; no task-owned `NoSuchMethodError`, `AbstractMethodError`, `NoClassDefFoundError`, Mixin application failure, missing pack metadata, missing bridge model/sound, or `ragdoll_seat` warning survived.

### Uninstall sanitizer

Seeded two embedded wind charges in normal inventory and two in ender storage, then ran `/ragdollfuture prepare-uninstall`.

Exact result: `projectiles=0 dropped=0 inventory_ender=4 total=4`.

### Packaged production client — Just Trial Spawners

Resolver selected `justtrialspawners [EXTERNAL] / PROJECTILE_IMPACT_OR_DISCARD item=justtrialspawners:wind_charge entity=justtrialspawners:breeze_wind_charge`.

A real JTS wind charge was used in Survival: item count **2 -> 1**, exactly **1** `external_pending`, exactly **1** `semantic_trigger`, origin `EXTERNAL_FALLBACK_PROJECTILE_DISCARD`, source `com.breakinblocks.justtrialspawners.common.entity.WindChargeProjectile`. World saved cleanly.

### Same-world provider removal

JTS / FTB Library / Architectury were removed from that same save. Expected historical registry/data-pack notices appeared, then the world loaded, the bridge automatically resolved back to `[EMBEDDED]`, both ragdoll systems activated, the player joined, and the world saved cleanly again.

## Benign warning classification

Retained environment/upstream noise: offline Mojang auth/network failures, headless OpenAL failure, software-renderer `Can't keep up`, pre-existing upstream dedicated dist-cleaner warnings, and expected Forge registry/data-pack history when the optional provider is added/removed. None are task-owned release failures.

## Durable release artifacts

Google Drive project folder ID: `1T7YP1nKcJtWJjsNXvDmRfMpA4asfJIQt`.

- Full release `Ragdoll-Forge-1.20.1-Release-2026-08-28.zip` — SHA-256 `373baea9096c2e8f1126e634b8e1183620c364d0dff45d6357105a020f870524`, Drive ID `1aTlCPROYPrEzv5dTVl1nkvWPIVcRDDV7`.
- Project-only bundle `Ragdoll-Forge-1.20.1-Project-Jars-2026-08-28.zip` — SHA-256 `56f4cda9197d016e45b406c44ccefe1e2cbe0845795ae4ce5775f11cf9d91dab`, Drive ID `1tyoHo9a0ytWGP2RFo2t5CrP-TgLpNdj_`.
- Final manifest — SHA-256 `c34c82dd3f1ead6c55c40f6dc848ef4ff71b2b7545d937b68e495ef014f583ff`, Drive ID `1Zf2rnavLVPLQvkY79GqHccdQhYdhNZgA`.

All Drive artifacts were fetched back and matched byte-for-byte by SHA-256; ZIP integrity tests were clean.

## Dev Kit update

Canonical `Minecraft Dev Kit - skill.zip` was updated in place, validated with the official skill validator/packager, fetched back, and hash-verified.

SHA-256 `c65d417d69a144f95cfc2beb247969d3b1ffa14b45532acf0f50a96a26775559`, Drive file ID `1q1zpjLUMqb53Gls4soprOntp8F30UBHo`.
