# ✅ Validation, Fidelity & Provenance

Minecraft Mod Vault's quality bar is intentionally stricter than “the file exists.”

## Conversion fidelity labels

| Label | Meaning |
|---|---|
| **Exact** | target representation preserves source semantics directly |
| **Translated** | equivalent target mechanism is used |
| **Reconstructed** | behavior recreated from evidence/templates |
| **Preserved Opaque** | unknown/custom bytes retained without interpretation |
| **Inferred** | best-supported guess; requires visible confidence/review |
| **User Modified** | explicit user override |
| **Unsupported** | no safe representation; must be surfaced, not silently deleted |

## WorldForge hard fidelity gates

A move/paste/schematic import fails the quality gate if it silently causes:

- missing/duplicated chest items;
- nested shulker/bundle data loss;
- broken loot-table state;
- double-chest split/merge mistakes;
- wrong stairs/rails/doors/beds/pistons/observers/sign orientation;
- fences/walls/panes/redstone connection corruption;
- lost block-entity NBT;
- dangling/misplaced entities or UUID relationships.

The planned transform order is **semantic state transform first, targeted neighbor-state recalculation after the whole structure is placed**.

## Proof hierarchy

Static checks < build success < isolated tests < GameTest/runtime tests < real client/player-facing verification.

The highest level required depends on the claim being made.
