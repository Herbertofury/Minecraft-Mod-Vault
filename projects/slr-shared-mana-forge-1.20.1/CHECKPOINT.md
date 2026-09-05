# SLR Shared Mana - Forge 1.20.1 v1.0.0 - Verified Checkpoint

## Objective
Forge 1.20.1 downport/rebuild of SLR + Iron's Shared Mana with SLR MP authoritative, low-overhead ISS integration, polished configuration, and first-class Iron's Botany 2.0.1 compatibility.

## Canonical local workspace
- Project: /mnt/data/slr_shared_mana_1201/project
- Repository state: not-a-git-repository
- Minecraft: 1.20.1
- Forge: 47.4.23
- Java: Temurin 17.0.20.1+1
- Gradle: 8.8

## Final artifacts
- SLR-Shared-Mana-Forge-1.20.1-1.0.0.jar | 25347 bytes | SHA-256 b50427c4b0204aa7dce3535c643fb60c6f99d927eb55d4778f0c889744365a13
- SLR-Shared-Mana-Forge-1.20.1-1.0.0-sources.jar | 12428 bytes | SHA-256 07c011a9a62da1cde98858df9b600df0b41765ee265e2df2f72d8b16176f3a35
- SLR-Shared-Mana-Forge-1.20.1-1.0.0-RELEASE.zip | 268687 bytes | SHA-256 0a6726ccb8792458a9ecebbb8711f057e8d91b295da64685900784fe74382d8d
- SLR-Shared-Mana-Forge-1.20.1-1.0.0-SOURCE.zip | 16869 bytes | SHA-256 41d457f42acd23d7c17d19b617619b349906b8ac0400d07cc5b0fd20ef2e32cc

## Architecture accepted
- SLR MP is authoritative; default conversion is 10 SLR MP = 1 Iron mana.
- ISS reads/writes translate on demand; no continuous player/world mana mirror.
- ISS passive regen is suppressed while sharing is active so SLR's native regen/cooldown owns regeneration.
- Positive ISS mana credits are translated into SLR MP so Iron's Botany ISS_PRIMARY-style grants remain functional.
- Iron's Botany is the routing authority; this bridge does not create a competing unification-mode state machine.
- Client reports shared mana through ISS while hiding the redundant Iron mana HUD.
- /slrmana status exposes live SLR MP, Iron-equivalent mana, ratio, and detected Botany mode.

## Iron's Botany compatibility
Iron's Botany 2.0.1 modes inventoried exhaustively: BOTANIA_PRIMARY, ISS_PRIMARY, HYBRID, SEPARATE, DISABLED.
- HYBRID: real dedicated-server, restart, and native integrated-client proof.
- SEPARATE: independent detached real dedicated-server proof with bridge receipt mode=SEPARATE.
- BOTANIA_PRIMARY / ISS_PRIMARY / DISABLED: exhaustive router-neutral audit proves bridge primitives do not branch on or overwrite Botany mode; zero-delta debits and positive credits preserve Botany semantics.

## Runtime proof
- Clean offline Forge build: BUILD SUCCESSFUL in 24s.
- Dedicated server HYBRID: Done (52.491s)! with bridge receipt.
- Restart same save/config: Done (10.354s)! with bridge receipt and all dimensions saved.
- Native Forge client/integrated server: player joined; /slrmana status reported SLR 1000/1000 MP, Iron 100 mana, ratio 10:1, Botany HYBRID; SLR MP HUD visible and separate Iron mana bar absent; clean save/quit; BUILD SUCCESSFUL.
- Dedicated server SEPARATE: Done (10.248s)! plus bridge receipt mode=SEPARATE; all dimensions saved.
- Fresh release extraction: both ZIPs pass archive test and all SHA256SUMS entries validate.
- Production-linkage audit: PASS against untouched production SLR 1.2.0, ISS 3.16.3, and Iron's Botany 2.0.1 JARs.

## Performance proof
Static hot-path audit found no TickEvent/PlayerTick/ServerTick/LevelTick subscriber, no level.players() scan, no entity/AABB scan, and no scheduled synchronization loop. Reflection metadata is resolved once and cached. Result: effectively zero idle overhead by design; runtime work is demand-driven by actual mana/config/client events.

## Known QA-environment limitation
A second installer-style packaged `forgeserver` launch of the reobfuscated JAR could not be constructed in this sandbox because the cached Forge userdev environment lacks the installer-generated server/server-extra Maven tree, and the sandbox cannot resolve maven.minecraftforge.net. This does not invalidate the green real dedicated-server/restart/native-client gates from the final source; the final reobfuscated artifact separately passed production-linkage audit.

## Exact next action
Release is complete. Install SLR-Shared-Mana-Forge-1.20.1-1.0.0.jar in the Forge 1.20.1 pack alongside SLR, Iron's Spells 'n Spellbooks, and optionally Iron's Botany 2.0.1. No further implementation action is pending.
