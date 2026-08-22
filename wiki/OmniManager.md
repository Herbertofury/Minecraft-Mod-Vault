# 📚 OmniManager

> **Status:** 🟢 Current product identity documented on `main`.

OmniManager treats the **local artifact as the primary object** and storefront/provider records as evidence attached to it.

## Identity resolution

- Modrinth SHA-512 identification.
- CurseForge MurmurHash2 fingerprint identification.
- Fabric, Quilt, Forge, NeoForge, Bukkit, Paper, Velocity and Bungee metadata parsing.
- Embedded JAR artwork plus provider artwork with provenance.
- Parallel provider records instead of forcing one storefront to own truth.
- Separate arbitration for name, author, art, project URL, installed file, installed version, compatible update, loader/environment and dependencies.
- Fuzzy matches remain review suggestions only.
- Patched/custom/local builds are protected from silent overwrite.

## Daily management direction

- instant local scan + background enrichment;
- card and detailed-list workspaces;
- broad search across names/filenames/authors/mod IDs/UUIDs/loaders/versions/providers/content types/status;
- bulk verified updates;
- enable/disable;
- recoverable Vault trash;
- transaction history and undo;
- confidence/source/artwork provenance;
- installed-vs-latest comparison;
- direct handoff to repair/porting flows without losing identity.

## Related pages

- [Cross-Store Identity](Cross-Store-Identity)
- [Native Bedrock Management](Native-Bedrock)
- [Validation & Fidelity](Validation-and-Fidelity)
