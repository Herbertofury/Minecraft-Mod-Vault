# ⚡ OmniBridge TestGrid

> **Status:** 📋 0.11.0 roadmap.

TestGrid is designed to make Minecraft testing **fast without substituting mocks for reality**.

## Tiered execution

| Tier | Runtime | Best for |
|---|---|---|
| **0** | no Minecraft process | schemas, manifests, dependencies, bytecode/symbol checks, deterministic transforms |
| **1** | loader/JVM-aware harness | codecs, registries, data generation, packet logic, suitable loader tests |
| **2** | real headless/GameTest server | block/entity interaction, AI, redstone, persistence, server behavior |
| **3** | real Minecraft client, virtual display when possible | rendering, models, animation, client events, screens |
| **4** | full Agent Test Driver | player input, mod GUIs, vehicles, complex interaction, visual/manual semantics |

## Disposable runtime farm

- immutable cached base profiles;
- content-addressed official runtime cache;
- copy-on-write/reflink/hardlink overlays where safe;
- tiny deterministic test worlds;
- auto-clean successful instances;
- retain failed instances for reproduction;
- Truth Profile vs Accelerated Profile;
- Rust orchestration where benchmarks justify it.

## A/B testing

Compare:

- original vs converted;
- old version vs port;
- loader A vs loader B;
- dependency implementation A vs B;
- before vs after patch;
- accelerated vs truth profile.

The report should normalize gameplay semantics instead of treating unstable internal IDs as the truth.
