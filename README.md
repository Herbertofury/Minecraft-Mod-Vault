# Minecraft Mod Vault

**The evidence-first Minecraft mod updater, repair laboratory, compatibility brain, and version-porting workbench.**

Minecraft Mod Vault is being built to end the usual cycle of guessing at broken dependencies, blindly replacing JARs, losing working builds, and calling a metadata edit a successful port. It combines live installed-mod forensics, source-project migration, exact version/toolchain intelligence, reversible repair transactions, and permanent repair knowledge in one local-first application.

## Minecraft Mod Vault 0.9.0

Version 0.9.0 unifies three major systems without removing the existing federated mod browser, updater, installers, Creator Archive, Mod Doctor, CIT, furniture, Bedrock, manager, or provider workflows.

| System | What it does |
|---|---|
| **Porting Lab** | Opens an exact managed JAR, records hashes and loader metadata, inspects classes/Mixins/access rules/runtime risks, resolves an upgrade/downgrade route, probes the toolchain, and generates an isolated hash-locked migration workspace. |
| **Repair Lab** | Imports an untrusted source ZIP through hardened extraction, preserves immutable originals, detects the real project/build/loader/version contract, stages reviewable known-field migrations, performs acknowledged wrapper-only builds, hashes outputs, exports proof, and rolls back. |
| **Compatibility Brain** | Builds a local SQLite WAL/FTS5 evidence database from reviewed Minecraft, loader, mappings, build-tool, repair-research, and durable repair-history seeds. |

### Verified release state

- **Go 1.27.0**, vendored dependencies, cgo-free Linux and Windows builds.
- **SQLite 3.53.3** in WAL mode.
- **919** Minecraft versions, **15,133** toolchain releases, **342** searchable knowledge documents, and **8** durable repair records in the fresh embedded brain.
- Repair Lab passed **25 production-UI assertions** and **38 authenticated backend calls**.
- Porting Lab passed **29 production-UI assertions** and **23 authenticated backend calls**.
- Zero page exceptions, console errors, or meaningful request failures in the exercised release flows.
- The extracted source package rebuilt both executables **byte-for-byte identically**.
- Internal package manifests, fresh-runtime initialization, artifact/proof downloads, quarantine/restore, rollback, and process-restart persistence all passed.

## Download and proof

The complete release index, exact Drive artifacts, checksums, and verification receipts are published under [`releases/v0.9.0`](releases/v0.9.0/README.md).

| Package | SHA-256 |
|---|---|
| Windows x64 ZIP | `d3d80652ba83163c7262cc79535b83812a4920b1a6fb0cf38fd4b4602dad7b20` |
| Linux x64 tar.gz | `2ed67c583c57d15d54f71827e3881a768ae43f3ccec9c88b46deef9c6b7e95d1` |
| Vendor-complete source ZIP | `806cdbe4f219e636ec32f7856b1b7ae8ec58040c2346c856cd388e30707d2fb3` |
| Verification evidence ZIP | `07ee17974107c6823230d81d7b3b2ad5602c1ee7258afb25f4c6978060fc4fba` |

Every Drive object was downloaded again after upload, compared by size and SHA-256, and archive-tested. The repository carries both the final verification receipt and the separate remote-byte verification receipt.

## Repair philosophy

Minecraft Mod Vault does not confuse these operations:

- discovering a possible continuation;
- resolving exact dependencies;
- editing metadata;
- remapping namespaces;
- decompiling or reconstructing source;
- migrating loader APIs and registries;
- compiling successfully;
- starting a client;
- passing a dedicated server, persistence, networking, and gameplay matrix.

Each is a different gate. Originals remain immutable or recoverable, mutations occur in disposable workspaces or receipt-backed quarantine transactions, and completion requires evidence appropriate to the actual target.

Build scripts are code. Repair Lab requires an explicit acknowledgement and restricts execution to detected project wrappers and fixed actions with dedicated caches, sanitized environment, timeout/cancellation, complete logs, artifact hashes, and post-run source verification. This is honest application-level containment—not a false claim of operating-system or virtual-machine sandboxing.

---

<details>
<summary><strong>Recovered C2ME OpenCL + Radium Forge 1.20.1 compatibility build</strong></summary>

A previously interrupted compatibility task was recovered and independently rechecked from the durable ProjectDump branch.

- ZIP size: `5,403,128` bytes
- ZIP SHA-256: `35d48f7f392b268819c653a57032bdf3a8f62f665d9d03d80b6716ce83d0565f`
- JAR: `c2meF-0.2.0+alpha.12.1-opencl-radiumcompat-all.jar`
- JAR SHA-256: `0f6f9e8d349bc092197e25c99e3a992eda1dab3a6dd1b62d15f75dbbd16cd393`

Runtime proof used Minecraft `1.20.1`, Forge `47.4.0`, Java `17.0.20`, and Radium `0.12.4+git.26c9d8e`. The headless Forge server loaded both mods, applied the targeted Radium compatibility overrides, reached `Done`, and cleanly fell back on the GPU-less verifier when no OpenCL context was available.

</details>
