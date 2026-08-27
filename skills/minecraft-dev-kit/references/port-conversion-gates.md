# Port and conversion gates

## Feature authority

For an old mod with partial newer rewrites, do not choose one version blindly. Treat the newest upstream release/source as behavioral/API authority where appropriate, while using older content-rich releases as feature/content authority when newer versions intentionally or accidentally omit material.

Build an explicit matrix of versions -> content/features -> implementation status -> target mapping.

## Complete source inventory

Before declaring a full port:

- enumerate source code packages/classes;
- registries and content IDs;
- entities, blocks, items, effects, dimensions, worldgen, structures, loot, recipes, tags, advancements;
- models, textures, animations, sounds, particles, shaders, localization;
- data generators and generated data;
- config/network/save formats;
- compatibility hooks/integrations;
- source-pack sidecars and variants;
- archival assets needed for provenance or faithful regeneration.

Classify each item as carried directly, regenerated/derived, intentionally superseded, intentionally excluded by project guardrail, or still missing.

## Conversion acceptance

A conversion is not complete until:

1. Target loader + Minecraft build passes.
2. Full intended content accounting has no unexplained gaps.
3. Deterministic audits/regression tests pass.
4. Dedicated server loads when supported.
5. Native client reaches a real world and exercises representative content.
6. Models/textures/animations are visually checked in the real renderer.
7. Network/state/save behavior is exercised where affected.
8. Restart/persistence is proven for stateful changes.
9. Final artifact and source are reproducible and hashed.
10. Known intentional differences are documented, not hidden.

## Compatibility-first world rule

When project guardrails require a standalone/forever-world-safe port, do not modify vanilla Overworld/Nether/End world generation merely to surface ported content. Prefer dedicated dimensions, structures, portals, recipes, or opt-in mechanics unless the original intended behavior and project requirements explicitly demand vanilla worldgen changes.
