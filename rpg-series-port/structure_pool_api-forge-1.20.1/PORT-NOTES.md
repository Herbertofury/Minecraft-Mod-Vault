# Port notes

## Source lineage

- 1.20.1 baseline: FabricExtras/StructurePoolAPI branch `1.20.1`, commit `effa8953409b487a7593ec8ad145e8df5762ca92`.
- Current behavior reference: branch `1.21.1`, commit `c7f9cf0c7ed91bcf3d51f174b078bbf688028ab4`, mod version 1.2.1.
- User-supplied binary: `structure_pool_api-neoforge-1.2.1+1.21.1.jar`.

## Backported behavior

Upstream 1.2.1 fixes structures failing to generate after the first world load. The old 1.20.1 branch attaches injection work directly to Fabric's server-start callback. The 1.2.x architecture instead queues `StructurePoolConfig.Entry` values and processes the queue for each server instance. This Forge port keeps that newer behavior and wires it to Forge `ServerStartingEvent`.

## Intentional 1.20.1 retention

Minecraft 1.21.1 added `StructurePoolAliasLookup` to the jigsaw lookup call. Minecraft 1.20.1 does not have that signature, so the port intentionally keeps the proven 1.20.1 `getPoolKey(StructureBlockInfo)` mixin target while carrying forward the 1.2.1 lifecycle fix.

## Native Forge policy

No Connector and no Forgified Fabric API runtime dependency. Architectury Loom is a build tool only; the output is a Forge JAR.
