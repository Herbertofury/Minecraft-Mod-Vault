# Minecraft Mod Vault

**The evidence-first Minecraft mod manager, updater, repair laboratory, compatibility brain, and version-porting workbench for both Java and Bedrock.**

## Minecraft Mod Vault 0.10.0 — OmniManager

OmniManager fixes the failure shown by competing launchers when a perfectly valid local mod came from another storefront and is reduced to an anonymous filename, generic cube, or “Uploaded” label. Minecraft Mod Vault now treats the **local artifact as the primary object** and every storefront as evidence attached to it.

### One identity across every source

- Exact Modrinth SHA-512 identification.
- Exact CurseForge MurmurHash2 fingerprint identification.
- Fabric, Quilt, Forge, NeoForge, Bukkit, Paper, Velocity, and Bungee metadata parsing.
- Embedded JAR artwork plus provider artwork with visible provenance.
- Parallel Modrinth, CurseForge, GitHub, GitLab, community-site, server-repository, generic-page, and Bedrock catalog records.
- Independent arbitration for name, art, author, description, project URL, installed file, installed version, latest compatible file, game version, loader, environment, dependencies, and update source.
- Provider file identity and hashes are compared even when two files reuse the same display version.
- Fuzzy matches are review suggestions only and can never authorize replacement.
- Patched, custom, and locally built artifacts are protected from silent overwrite.

### Native Bedrock management

OmniManager reads, installs, scans, activates, disables, restores, and records transactions for `.mcpack`, `.mcaddon`, `.mcworld`, and `.mctemplate` content. It supports behavior packs, resource packs, skin packs, world templates, development packs, installed worlds, Bedrock Stable, Preview/Beta, and custom `com.mojang` roots.

Localized names/descriptions, UUIDs, semantic versions, minimum engine versions, modules, scripts, capabilities, dependencies, authors, license, project links, icons, and world artwork are preserved. World activation edits the correct behavior/resource pack JSON and can restore the exact previous bytes.

### Premium daily management

- Instant local scan with background cross-store enrichment.
- Card and detailed-list workspaces without an installed-content result cap.
- Search across names, filenames, authors, mod IDs, UUIDs, loaders, versions, providers, profiles, content types, and status.
- Bulk verified updates, enable/disable, recoverable Vault trash, transaction history, and receipt-driven undo.
- Identity confidence, source badges, artwork provenance, installed-to-latest comparison, compatibility warnings, dependency evidence, and exact project/file destinations.
- Direct handoff from an installed JAR into Mod Doctor, Porting Lab, or Repair Lab without losing identity or recovery context.
- Keyboard search and selection, persistent view/density preferences, responsive details, reduced-motion support, forced-color support, and accessible live feedback.

### Release verification

The exact packaged release was checked through Go tests and vet, JavaScript syntax validation, static control-to-handler auditing, fresh source reconstruction, byte-identical Linux/Windows rebuild comparison, internal package manifests, archive-path safety, a real Chromium production flow, process restart, and an uncapped 256-artifact library fixture that verified the final off-screen item remained in the DOM.

The canonical Google Drive project folder contains the Windows, Linux, vendor-complete source, verification evidence, checksums, and final release receipts:

**https://drive.google.com/drive/folders/1nkX40V3f0psEQldm0WjAZH9o-gnAO-Ln**

Release documentation and verification index: [`releases/v0.10.0`](releases/v0.10.0/README.md)

Source branch: [`release/v0.10.0`](../../tree/release/v0.10.0)

See also:

- `RELEASE-NOTES-0.10.0.md`
- `CROSS-STORE-IDENTITY-CONTRACT.md`
- `OMNIMANAGER-ARCHITECTURE.md`
- `PROVIDER-CAPABILITY-MATRIX.md`
- `PRIVACY-AND-NETWORK.md`
- `SBOM-SPDX-0.10.0.json`

---

<details>
<summary><strong>Recovered C2ME OpenCL + Radium Forge 1.20.1 compatibility build</strong></summary>

The earlier recovered build remains preserved with ZIP SHA-256 `35d48f7f392b268819c653a57032bdf3a8f62f665d9d03d80b6716ce83d0565f` and JAR SHA-256 `0f6f9e8d349bc092197e25c99e3a992eda1dab3a6dd1b62d15f75dbbd16cd393`.

</details>
