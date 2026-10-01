# Structure Pool API 1.2.1 — Native Forge 1.20.1 Backport

Target: Minecraft 1.20.1, Forge 47.4.x, Java 17.

This project preserves the public package/mod ID and the 1.20.1 worldgen implementation while backporting the upstream 1.2.1 server-lifecycle queue fix. It builds a direct Forge JAR and does not require Sinytra Connector or Forgified Fabric API.

Upstream source: https://github.com/FabricExtras/StructurePoolAPI
Upstream license: MIT.

## Compatibility goals

- `modId=structure_pool_api`
- Existing 1.20.1 callers using `StructurePoolAPI` and `StructurePoolConfig` remain source/binary-shape compatible where Minecraft mappings allow.
- Injection registrations survive integrated-server world changes, matching upstream 1.2.1 behavior.
- Structure spawn limits retain the proven 1.20.1 generation hook rather than forcing the 1.21.1 alias-lookup signature onto 1.20.1.
