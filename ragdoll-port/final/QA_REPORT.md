# Final QA Report — Forge 1.20.1 Ragdoll Parity Port

## Release identity

Final candidate JAR SHA-256 values:

- `sable_player_ragdoll-1.20.1-0.7.2.jar` — `737476cf12d110157beabdda66ff1d1cd991c1edcbca5b627db88c56d369004f`
- `ragdoll_reactions-1.20.1-0.7.0.jar` — `f579f37d9808e10f378ddac66d9329f5d3d34218a756138579992416e2256553`
- `sable-forge-1.20.1-2.0.0-all.jar` — `6c3121fd4e5c474bd8799b991ff1845ef5413a35a438787721a9d9b38f8712b2`

The exact hashes above were installed for final native client and dedicated-server QA.

## Build / production namespace gates

- Core full-source Java 17 compile: green, 204 production classes.
- Reactions full-source Java 17 compile: green, 49 production classes.
- Mixin annotation processing/refmap generation: green.
- Production reobfuscation validator: `classes=253 staleMethods=0 staleFields=0 staleInvokeDynamic=0`.
- Eight functional-interface/lambda callsites were explicitly reobfuscated, including renderer provider factories.
- Core mixin config targets Java 17 and contains the generated refmap.
- Sable Forge metadata explicitly selects `META-INF/accesstransformer-forge.cfg` so Forge 47.4.23 applies the correct SRG access transformer.

## Native production client + integrated world

Runtime: Java 17, Forge 47.4.23, Minecraft 1.20.1, real production SRG JARs, Xvfb + llvmpipe software OpenGL.

Observed in the live world:
- Jade discovered `dev.leo.sableplayerragdoll.compat.jade.RagdollJadePlugin`.
- Curios loaded its slot/entity data.
- Sable created Rapier physics pipelines for Overworld, Nether, and End.
- Player Ragdoll and Ragdoll Reactions both reported active.
- Manual player ragdoll created 6 parts / 5 constraints, seated/launched the player, and manual get-up removed the sublevel cleanly.
- Controlled vanilla TNT reaction logged player explosion power 4.00 / radius 10.00 and launched the player ragdoll.
- The same TNT scenario logged zombie explosion tumble and created 7 Sable sublevels / 6 joints for the mob ragdoll.
- Sable's stale UDP movement-snapshot race no longer emits a task-owned ERROR when a plot is already untracked.
- `ragdoll_seat` has an explicit invisible model, removing the missing-model warning.

## Save / restart persistence

- Actual client `Save and Quit to Title` produced normal server shutdown and `ThreadedAnvilChunkStorage: All dimensions are saved`.
- Client then exited through the actual `Quit Game` button.
- A fresh Java process reopened the same saved world using the exact three final hashes.
- `RagdollQA joined the game` again; all three Sable dimension pipelines rebuilt; both ragdoll systems activated.
- Restart task-owned error counters: `non-existent sub-level=0`, `IllegalAccessError=0`, `ModLoadingException=0`, `Mixin apply failed=0`, `ragdoll_seat.json missing model=0`.
- The restarted world was Save-and-Quit again, all dimensions saved, then the client exited normally.

## Exact-hash dedicated server

A fresh dedicated-server copy was populated with the same final three JAR hashes plus the tested Jade and Curios versions.

- Initial QA port 25565 was already occupied by another process; the disposable QA server was safely isolated to 25566. This was an environmental bind conflict, not a mod failure.
- Forge reached `Done (39.283s)!`.
- Jade loaded the ragdoll plugin.
- Curios initialized.
- Sable created Rapier pipelines for Overworld, Nether, and End.
- Both ragdoll systems reported active.
- Task-owned fatal scan remained empty.
- A normal `stop` command produced `Stopping server`, `Saving players`, `Saving worlds`, and `ThreadedAnvilChunkStorage: All dimensions are saved`.

## Parity audit

Compared with the supplied 1.21.1 originals:
- Core: 196 original classes vs 204 final classes.
- Reactions: 49 original classes vs 49 final classes.
- Missing original `assets/` or `data/` files: 0 for both mods.
- Missing Reactions classes: 0.
- The only two original core class files not present are `AccessoriesRenderHelper` nested implementation classes used solely by the Accessories loader bridge. Forge 1.20.1 has no Accessories runtime; Curios is the functional Forge equipment path and was native-tested.

## Proven platform-only exceptions

1. Accessories is not available as a Forge 1.20.1 runtime, so that loader-specific bridge cannot be executed. Curios integration is functional.
2. Vanilla Wind Charge is a newer Minecraft feature and does not exist in 1.20.1; therefore that specific 1.21 reaction has no 1.20.1 vanilla event/entity to react to.

These are platform absences, not unported failures in available 1.20.1 behavior.
