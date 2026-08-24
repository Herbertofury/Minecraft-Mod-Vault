# Minecraft Mod Vault 0.10.0 — OmniManager

Minecraft Mod Vault 0.10.0 turns the former file manager into **OmniManager**: a local-first universal Minecraft content library for Java Edition, Bedrock Edition and server software. It preserves the Repair Lab, Porting Lab, Compatibility Brain, federated browser, Creator Archive and existing installers while adding a manager designed around exact identity, reversible actions and cross-provider evidence.

## The cross-store identity fix

OmniManager does not reduce a locally installed file to its filename or to the catalog of whichever storefront happens to be open. It builds one identity graph from:

1. exact local cryptographic hashes;
2. Modrinth SHA-512 file lookup;
3. CurseForge-compatible MurmurHash2 fingerprints;
4. embedded Fabric, Quilt, Forge, NeoForge, Bukkit, Paper, Velocity and Bungee metadata;
5. embedded icons and canonical project URLs;
6. provider, author, release and artwork evidence from every configured source;
7. conservative, review-only fuzzy suggestions when exact evidence is unavailable.

Provider identities remain visible rather than being destructively collapsed. Name, author, artwork, installed file, newest compatible file and provenance can therefore come from different verified records without turning a CurseForge-hosted mod into a generic “Uploaded” cube in another catalog.

The release carries a regression fixture based on the reported `cataclysm_dimension-forge1.20.1-1.5.7.jar` failure. Its embedded identity resolves to **Cataclysm Dimensions**, version **1.5.7**, author **P1nero**, Forge, and its real icon; exact provider evidence can then enrich the same record with canonical project art and release identity. Provider file identity is compared even when a creator reuses the same human-facing version label.

## Safe updates instead of blind replacement

- Installed and newest-compatible files are tracked as distinct immutable identities.
- Loader, game version, side, dependencies, release channel and Java requirements participate in compatibility.
- Patched, custom, locally rebuilt and otherwise divergent JARs are protected from unattended replacement.
- Bulk operations create receipts and can be undone.
- Disable, quarantine, trash, update and Bedrock activation are reversible transactions.
- Original files remain recoverable and every meaningful mutation is hash recorded.

## Full Bedrock library support

OmniManager understands `.mcpack`, `.mcaddon`, `.mcworld` and `.mctemplate` packages plus installed behavior packs, resource packs, skin packs, world templates, development packs and worlds. Stable, Preview/Beta and user-defined `com.mojang` roots can coexist.

It reads real Bedrock manifests, localized names/descriptions, UUIDs, semantic versions, minimum engine versions, module types, dependencies, script modules, capabilities, authors, license, project URL, `pack_icon.png` and world art. Native install and per-world behavior/resource activation both carry byte-preserving undo receipts.

Archive intake rejects traversal, absolute paths, symlinks, duplicate/case-colliding entries, extraction escape, excessive file count, unreasonable expansion and suspicious compression ratios.

## Premium everyday workspace

OmniManager provides:

- instant local results followed by bounded background enrichment;
- card and professional dense-list layouts;
- name, filename, author, mod ID, UUID, loader, game version and provider search;
- Java, Bedrock, server, profile, content-type, source, status and release filters;
- cross-provider badges and artwork provenance;
- identity-confidence presentation and conflict evidence;
- installed-to-latest comparison;
- bulk selection and transactional actions;
- details/evidence drawer;
- native Bedrock package installation and world activation;
- direct Repair Lab and Porting Lab handoffs;
- keyboard search, select-all, escape/close, reduced-motion and forced-color support;
- no viewport virtualization, hidden card caps or deferred off-screen installed content.

## Verification performed

- verified suite named Go tests passed in the captured machine-readable suite.
- `go vet ./...` passed.
- Go race detection passed.
- Every JavaScript file passed `node --check`.
- JSON assets passed strict parsing.
- Linux amd64 and Windows amd64 binaries were built with Go 1.27.0 and `CGO_ENABLED=0`.
- A fresh isolated runtime rendered the production OmniManager route with 0 enabled controls and exercised 0 safe interactive actions through headless Chromium/CDP.
- The production browser run required visible search, filters, Bedrock and cross-provider surfaces, the premium stylesheet, and zero page exceptions or console errors.
- Capability, static-control, route, quality, screenshot and regression-name audits passed before release packaging.

See `FINAL-RELEASE-VERIFICATION-0.10.0.json` and the verification evidence archive for hashes and observed results.
