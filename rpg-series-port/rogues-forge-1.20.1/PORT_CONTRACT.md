# Rogues 3.1.1 -> native Forge 1.20.1 port contract

Status: **ACTIVE / NOT GRADUATED** until an exact promoted canonical head passes deep native runtime acceptance.

## Authorities
- Current feature/content authority: `ZsoltMolnarrr/Rogues` `1.21.1` commit `d4a7af565559dcff4384eabb2481f63eb5f97d55`, tree `c6fac5d7c80807b41668274843c9732762d7cb03`, version 3.1.1.
- Historical mapping/API substrate only: `1.20.1` commit `bdfe6447b90758129e12430b497d97c181222b12`, tree `0a4f7f94e77031732843f8a40a8460184bd3577a`, version 1.2.0.
- Target: Minecraft 1.20.1, Forge 47.4.23, Java 17.

## Non-negotiable preservation
Current 3.1.1 owns features and content. The historical branch may prove target-era names/signatures only and may never replace current implementations wholesale. Preserve Bear Trap, current warrior/rogue spell suite, current equipment/armor definitions, effects, villages/trades, generated recipes/tags/lang/sounds, and optional current equipment integrations.

## Dependency ownership
Armor Model API, Structure Pool API, Spell Power, Spell Engine, TinyConfig, Cloth Config, Player Animator, and Curios remain separate real dependencies. No foundation source injection and no dependency shading into the Rogues release.

## Acceptance gates
1. Exact upstream pins and deterministic source materialization.
2. Current 3.1.1 common source/resources survive the 1.20.1 translation.
3. Native Forge registration owns registry mutation/phase ordering.
4. Java 17 common + Forge compile and remapped package.
5. Packaged dedicated Forge server reaches ready state with the exact release candidate.
6. Native LWJGL Forge client reaches post-bootstrap and a real client joins the packaged server.
7. Semantic gameplay acceptance covers at minimum: root versus stun action semantics; Stealth break/removal; Bear Trap entity/effect behavior; representative rogue/warrior spell behavior; current weapon/armor registration and attributes.
8. Deterministic release/source identity is certified.
9. Pre-promotion acceptance passes, then the exact promoted canonical workbench head is replayed end-to-end.
10. GitHub evidence and Google Drive evidence are persisted and byte/hash verified.

No compile-only or startup-only result can graduate Rogues.
