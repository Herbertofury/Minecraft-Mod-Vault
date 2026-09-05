# SLR Shared Mana - Forge 1.20.1

A clean Forge 1.20.1 shared-mana bridge for **Solo Leveling: Reawakening (SLR)** and **Iron's Spells 'n Spellbooks**, built to coexist with **Iron's Botany 2.0.1** without adding a competing mana router.

## Install

Drop `SLR-Shared-Mana-Forge-1.20.1-1.0.0.jar` into the same `mods` folder as:

- Solo Leveling: Reawakening (SLR) 1.20.1
- Iron's Spells 'n Spellbooks 1.20.1 (tested with 3.16.3)
- Iron's Botany 2.0.1 when you want Botania/ISS routing
- Iron's Botany's normal dependencies (Botania, Curios, Patchouli, Iron's Lib, etc.)

Forge target: **1.20.1-47.4.23**. Java target: **17**.

## What it does

- **SLR MP is the authoritative pool.**
- Iron mana reads resolve from SLR MP on demand at the configured ratio (default `10 SLR MP = 1 Iron mana`).
- Iron mana spends debit SLR MP exactly once and start SLR's native `mana_refresh` cooldown.
- Intentional positive Iron mana grants can credit SLR MP, which lets Iron's Botany `ISS_PRIMARY` feed the shared pool.
- Native ISS passive regen is suppressed while sharing is active so SLR regen/cooldown rules stay authoritative.
- Iron's mana HUD is hidden by default; SLR's MP HUD remains the one visible source of truth.
- `/slrmana status` reports SLR MP, Iron-equivalent mana, ratio, and detected Iron's Botany mode.

## Iron's Botany modes

Iron's Botany remains the routing authority; use its own `manaUnificationMode` config. This bridge supports the full mode set without creating a duplicate selector:

| Mode | Shared Mana behavior |
|---|---|
| `HYBRID` | Botany's combined behavior; the ISS side is backed by SLR MP. |
| `BOTANIA_PRIMARY` | Botany can pay the cast from Botania; a zeroed ISS debit means SLR is not double-charged. |
| `ISS_PRIMARY` | Botany's positive ISS credits become SLR MP. |
| `SEPARATE` | Botania remains separate; the ISS side is still backed by SLR MP. |
| `DISABLED` | Botany stops routing; SLR <-> ISS sharing itself remains active. |

## Config

Forge generates `config/slr-shared-mana.toml`. A fully commented example is included as `slr-shared-mana-example.toml`.

The default config intentionally has **no second Botany mode switch**. Duplicating Iron's Botany's router would create conflicting state; the comments instead document exactly how each Botany mode behaves with the SLR-backed ISS pool.

## Performance design

There is no player-tick mana mirror, no world/entity scan, no nearby-block search, and no scheduled synchronization loop in this mod. The bridge runs only on actual ISS mana reads/writes/regen attempts plus normal one-time lifecycle setup. See `evidence/final-performance-audit.txt`.
