# 🧯 Troubleshooting

## A mod is not identified correctly

- inspect embedded mod ID/version/loader metadata;
- compare exact hashes/provider fingerprints;
- review provider candidate evidence;
- do **not** authorize replacement from a fuzzy name match alone;
- preserve the local artifact as primary evidence.

## A repaired/ported project still crashes

- capture the exact crash/log and target runtime versions;
- confirm the intended build is actually installed/loaded;
- verify dependencies and loader version;
- return to Repair/Porting Lab with the reproducer;
- add a regression fixture after root cause is fixed.

## Bedrock pack activation is wrong

- verify UUID/version and world behavior/resource pack records;
- compare against the pre-transaction bytes;
- restore the exact previous activation state if needed.

## World conversion/edit concerns

WorldForge remains roadmap. Its acceptance rules explicitly require transactional snapshots, item/container preservation, orientation/connection correctness, unknown-data preservation and real Minecraft verification.

## Wiki status confusion

If a feature page says 📋 Planned, it is intentionally not being presented as shipped. Use [Release Status](Release-Status) as the authoritative status page.
