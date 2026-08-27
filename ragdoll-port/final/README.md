# Sable Ragdolls + Ragdoll Reactions — Forge 1.20.1 parity build

Install all three JARs in the same Forge 1.20.1 `mods` folder:

1. `sable-forge-1.20.1-2.0.0-all.jar`
2. `sable_player_ragdoll-1.20.1-0.7.2.jar`
3. `ragdoll_reactions-1.20.1-0.7.0.jar`

Runtime used for final QA: Minecraft 1.20.1, Forge 47.4.23, Java 17.

Optional integrations tested in the final native client/server runs:
- Jade 11.13.3
- Curios 5.14.1+1.20.1

The goal of this port is behavioral parity with the supplied 1.21.1 originals wherever the 1.20.1 platform provides the underlying feature.

Platform-only exceptions:
- Accessories has no Forge 1.20.1 runtime. Curios is the working Forge equipment integration; retained snapshot/storage fields preserve the model where possible.
- Vanilla Wind Charge does not exist in Minecraft 1.20.1, so its 1.21-only reaction cannot occur.

See `QA_REPORT.md` for native client/world, restart, dedicated-server, remap, Jade, Curios, player ragdoll, mob ragdoll, and reaction evidence.
