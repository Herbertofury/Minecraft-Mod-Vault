# SLR Shared Mana 1.1.0 — QA Report

## Result

**GREEN for the release gates available in this environment.**

### Verified

- Java 17 target compile: 13 Java source files / 16 class files; no compile errors.
- Forge 1.20.1 / 47.4.23 production reobfuscation via ForgeAutoRenamingTool 1.0.2.
- Reobf inheritance audit: 0 missing-class errors. One known Forge mapping-table warning for an unrelated vanilla `LivingEntity` lambda.
- Shipping JAR ZIP integrity.
- Production linkage uses SRG names (for example `CommandSourceStack.m_81375_`) rather than development Mojmap method names.
- Packaged coremod JS executed under Nashorn 15.4 with real ASM and returned 44 unique definitions: 43 spend transformers + 1 read-only Spirit Bow affordability transformer.
- Synthetic bytecode assertions prove GETFIELD `PlayerVariables.MP` becomes `SlrBorrowBridge.readMp(Object)` and spend-path PUTFIELD becomes `writeMp(Object,double)`; the read-only target leaves PUTFIELD untouched.
- Forbidden SLR regen/reward/reset/potion/GuildBuff paths are absent from the transformer list.
- No new player/tick event loop exists.
- Default config is `enabled=true` and `separatePoolsBorrowFromIron=false`.
- Classic Iron interception/HUD/regen paths are gated to classic shared mode only.
- Separate mode's Iron deficit debit uses Iron's own `MagicData.setMana` path and its native `UpdateClient.SendManaUpdate`, preserving `ChangeManaEvent` behavior and immediate native HUD sync.
- Shipping `coremods.json`, coremod JS, metadata version, Java 17 class version, and bridge ABI were audited from the final reobfuscated JAR.

## Runtime limitation

A fresh full Minecraft client/server launch with the exact production SLR and Iron's Spells JARs was not rerun in this container because those exact dependency binaries are not persisted here and the external binary-download path is blocked. This report therefore does **not** claim a fresh full-game-stack launch. The release is supported by compile, production-reobf, packaged Nashorn/ASM transformer, linkage, structural, and regression evidence.
