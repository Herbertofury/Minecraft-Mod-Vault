# Spellbrook Forge 1.20.1 — implementation checkpoint

## Objective
Create a standalone Forge 1.20.1 Spellbrook broom mod from the authorized captured client assets, while becoming a first-class Hexerei broom addon when Hexerei 0.4.2.3 is installed.

## Source evidence locked
- Captured pack reconstructed successfully: 26,358/26,358 entries, zero reconstruction failures.
- `assets/broom`: 132 files.
- 29 broom model JSONs, 21 core-crystal model JSONs, 29 broom textures, 18 core textures, 26 broom flight/mount OGGs.
- Exact original custom-model-data routes retained in code/resources.
- Captured client pack does not contain Spellbrook's server plugin/config implementation for authoritative server flight physics/progression/quests; standalone flight here is an independent implementation and is not falsely represented as server-source parity.

## Implemented in this branch
- Forge 1.20.1 / Forge 47.4.23 / Java 17 project.
- 28 broom finish variants plus hidden debug route.
- 20 core routes plus hidden debug route.
- NBT persistence for finish/core/inventory/owner/broom UUID.
- Standalone server-authoritative flight and input packet.
- 30-slot handler with 27-slot visible vanilla chest storage surface.
- Break/place persistence.
- Core swapping.
- Custom renderer using captured baked JSON model routes.
- Optional Hexerei bridge isolated in its own source package.
- Hexerei path subclasses `BroomEntity`, registers `BroomType`s, preserves Hexerei item handler/module behavior and native flight, inserts its normal broom brush, and renders the Spellbrook model instead of Hexerei's stock stick model.
- Explicit credits to Levah, Crocwise, Team Spellbrook / Spellbrook Ltd.; Hexerei credit to JoeFoxe.

## Current blocker
ChatGPT container execution currently fails before executing even trivial commands with an internal `KeyError`. This prevents binary ZIP extraction into the source tree, `javap` against the user's exact performance-overhaul Hexerei JAR, Gradle compilation, dedicated-server launch, native-client visual QA, and final JAR packaging. Repeated unchanged retries are intentionally stopped.

## Exact next action
When container execution is healthy: import the verified Broom Research Kit into `assets/spellbrook`, rewrite `broom:` model namespace references to `spellbrook:`, copy the exact uploaded Hexerei performance-overhaul JAR into project `libs/`, run a compile-only gate, fix any API linkage differences, then run standalone dedicated server + native client and repeat with Hexerei installed before packaging the final JAR/source/evidence bundle.
