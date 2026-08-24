# Minecraft Ecosystem Category Scout — 2026-08-25

**Material run:** yes  
**Coverage:** Java + Bedrock, all materially active versions  
**Permanent categories:** 34  
**Canonical Recommendations Vault:** 100 retained serious projects/finds  

## Material delta

- **#34 Monster Girl / Female-Coded Mob — leader update:** [Monster Girl 2.7.0](https://www.curseforge.com/minecraft/mc-mods/monster-girl) is now the practical Forge 1.20.1 leader. The 2.7.0 line also exists for NeoForge 1.21.1; 26.1.2 remains on 2.6.0. New content includes Succubus, Death Lash, Little Devil taming and expanded companion control. **Critical safety note:** the exact Forge 1.20.1 file warns that deleting the mod after a save has loaded it can make that save unplayable. **TEST BRANCH / BACKUP REQUIRED.**
- **#4 Performance / #11 Create specialist — new runner-up:** [StellarCreateOptimization 1.0.2](https://www.curseforge.com/minecraft/mc-mods/stellarcreateoptimization) has direct Forge 1.20.1 and NeoForge 1.21.1 builds for [Create](https://modrinth.com/mod/create) 6.0.x. It targets chutes, block entities, particles, shaders, Chain Conveyors and large-factory rendering/server cost, with explicit [Embeddium](https://modrinth.com/mod/embeddium) and [Oculus](https://www.curseforge.com/minecraft/mc-mods/oculus)/[Iris](https://modrinth.com/mod/iris) compatibility work. Young project: benchmark locally before promotion.
- **#9 Combat/Animation compatibility — new advanced runner-up:** [Epic Fight Indestructible - Unofficially Enhanced!! 20.14.17](https://www.curseforge.com/minecraft/mc-mods/epic-fight-indestructible-unofficially-enhanced) targets Forge 1.20.1 + [Epic Fight](https://www.curseforge.com/minecraft/mc-mods/epic-fight-mod) 20.14.17, adding dynamic mob patches, live refresh, [CustomNPCs](https://www.curseforge.com/minecraft/mc-mods/custom-npcs) coverage, optional [Timeless and Classics Zero](https://www.curseforge.com/minecraft/mc-mods/timeless-and-classics-zero) behavior and animation/crash fixes. Keep as **TEST BRANCH**; the simpler guard-fix path remains the baseline leader.
- **#34 Maid/social — new specialist:** [Touhou Little Maid: Love & Loathe 2.0.6](https://www.curseforge.com/minecraft/mc-mods/touhou-little-maid-love-loathe) has Forge 1.20.1 and NeoForge 1.21.1 builds and adds emotion, hunger, broadcast commands and trust/fear-driven AI behavior to [Touhou Little Maid](https://www.curseforge.com/minecraft/mc-mods/touhou-little-maid). It is a stateful maid addon, not a standalone mob.
- **#34 Experimental maid physics — new watch:** [Sable: MaidRagdoll 0.11-beta](https://www.curseforge.com/minecraft/mc-mods/sable-maidragdoll) is a 1.21.1 NeoForge maid-physics layer with partial model support, ragdoll behavior, Cake Box throwing and a Death Cheat Charm recovery path. **WATCH / FUTURE VERSION.**

## Afdian / creator-source status

- [Feather_aya](https://afdian.com/a/FliegeSA) now records that public Afdian model sales stopped in May 2026. Prior-buyer update access remains, and some later variants may be free/reduced. Preserve as **MANUAL-AUDIT / SALES DISCONTINUED ON AFDIAN** rather than inventing product rows.
- [你个人机cc](https://afdian.com/a/ccnie) continues to expose an active [Yes Steve Model](https://www.curseforge.com/minecraft/mc-mods/yes-steve-model) catalog, with many entries requiring 2.5.1+.
- [映白](https://afdian.com/a/ehaku) continues to expose an active [Yes Steve Model](https://www.curseforge.com/minecraft/mc-mods/yes-steve-model) / [Touhou Little Maid](https://www.curseforge.com/minecraft/mc-mods/touhou-little-maid) creator catalog with current engine requirements and explicit usage/licensing constraints. Preserve official provenance per model.

## Specialist rechecks

- **#31 Animated:** [Actions & Stuff 1.11](https://www.minecraft.net/en-us/marketplace/pdp?id=53c7af69-2425-490e-8e9a-8ad2c2e7cbfe) overall / [Fresh Animations](https://modrinth.com/resourcepack/fresh-animations) Java — **NO MATERIAL LEADER CHANGE**.
- **#32 CIT:** [Kaydicraft CIT 1.45](https://modrinth.com/resourcepack/kaydicraft) / [FAYE](https://www.planetminecraft.com/texture-pack/the-faye-wallpaper-amp-flooring-set-cit/) — **NO MATERIAL LEADER CHANGE**.
- **#33 3D:** [3D Default 1.15.0](https://modrinth.com/resourcepack/3d-default) — **NO MATERIAL LEADER CHANGE**.
- **#16–#18 Bedrock:** **NO MATERIAL LEADER CHANGE**.
- **#19–#20 maps/artistic:** **NO MATERIAL CHANGE** after the current Planet Minecraft sweep.
- **All other categories:** independently rechecked; **NO MATERIAL LEADER CHANGE**. The complete 34-category ranking remains in the canonical Google Sheet and curated master guide.

## Safety snapshot

| Project | Client only? | Server required? | Save/world touched? | Safe to remove? | Update friendliness |
|---|---|---|---|---|---|
| [StellarCreateOptimization](https://www.curseforge.com/minecraft/mc-mods/stellarcreateoptimization) | No | Client+server environment | No dedicated world-content registration surfaced | High from save-integrity perspective; young-project caution | Good / test first |
| [Epic Fight Indestructible - Unofficially Enhanced!!](https://www.curseforge.com/minecraft/mc-mods/epic-fight-indestructible-unofficially-enhanced) | No | Yes for integrated behavior stack | Behavior/config mob patching; no new worldgen surfaced | Medium-high with backup/config rollback | Good with strict dependency matching |
| [Monster Girl](https://www.curseforge.com/minecraft/mc-mods/monster-girl) | No | Yes in multiplayer | Persistent entities/items/gameplay | **Low**; exact Forge 1.20.1 file has no-delete warning | Caution |
| [Touhou Little Maid: Love & Loathe](https://www.curseforge.com/minecraft/mc-mods/touhou-little-maid-love-loathe) | No | Yes | Stateful maid emotion/trust/fear/hunger/AI | Medium-low | Caution |
| [Sable: MaidRagdoll](https://www.curseforge.com/minecraft/mc-mods/sable-maidragdoll) | No | Yes | Beta maid/ragdoll runtime state; persistence not fully documented | Unknown-medium | Caution / beta |

## Persistence

The canonical Google Sheet and curated Google Doc were updated in place. The separate Monster Girl/Female Mob companion Sheet + Doc were expanded to **88 curated projects/source hubs**, including **39 explicit 1.20.1 routes** and **11 maid-ecosystem entries**. This file is the durable GitHub checkpoint for the 2026-08-25 material run.
