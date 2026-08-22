# 🔎 Cross-Store Identity

Minecraft content often exists on multiple hosts—or only locally. OmniManager's identity model avoids reducing a valid custom artifact to an anonymous filename because a preferred storefront does not recognize it.

## Evidence layers

| Evidence | Strength |
|---|---|
| exact provider hash/fingerprint | strongest identity evidence |
| embedded mod metadata / IDs | strong structural evidence |
| source repository / project IDs | strong provenance evidence |
| author + version + filename | contextual evidence |
| fuzzy title similarity | suggestion only |

## Arbitration rule

Different fields may legitimately come from different sources. A Modrinth record might provide the best icon, a local JAR might provide the authoritative installed version, GitHub might provide source/license information, and CurseForge might expose the newest compatible release.

That is intentional: **identity is a graph of evidence, not a single provider row.**

## Update safety

An update must prove it targets the same project/file lineage. Similar names or display versions are not enough to overwrite a patched/local build.
