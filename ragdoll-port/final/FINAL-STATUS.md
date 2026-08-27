# Final status

Forge 1.20.1 parity port is native-runtime green.

- Core: 204 classes
- Reactions: 49 classes
- Strict production validator: `classes=253 staleMethods=0 staleFields=0 staleInvokeDynamic=0`
- Native manual player ragdoll: green (6 parts / 5 constraints, seat/launch/get-up)
- Native Reactions player TNT path: green
- Native mob TNT path: green (7 Sable sublevels / 6 joints)
- Same-world save/restart: green
- Exact-hash dedicated server start/save/stop: green
- Jade 11.13.3: green
- Curios 5.14.1+1.20.1: green

Final exact tested JAR hashes are in `JAR-SHA256SUMS.txt`; source lineage and platform-only exceptions are in `PROVENANCE.md` and `QA_REPORT.md`.
