# Minecraft Mod Vault 0.9.0 — Ultimate Repair and Porting Release

This release turns Minecraft Mod Vault into a two-lane repair system:

- **Porting Lab** performs installed-JAR forensics, evidence-backed upgrade/downgrade planning, isolated migration workspace generation, and cryptographic duplicate quarantine/restore.
- **Repair Lab** accepts hostile/untrusted source ZIPs with traversal, symlink, duplicate-path, archive-bomb, and extraction-escape defenses; preserves immutable originals; detects the real loader/build contract; stages only recognized migration edits; executes wrapper-only build/test/clean actions behind exact code acknowledgement; hashes artifacts; exports proof; and rolls back.
- **Compatibility Brain** is an offline SQLite 3.53.3 WAL/FTS5 evidence store seeded with official Minecraft and loader/toolchain metadata, current porting research, and durable repair history.

## Verified release artifacts

The connector used for this publication can write repository content but cannot create GitHub Release binary assets. The complete byte-verified packages are therefore preserved in the canonical Google Drive project folder, with their exact provider IDs, sizes, and SHA-256 values recorded in `DRIVE-REMOTE-VERIFICATION.json`.

| Artifact | Size | SHA-256 | Google Drive |
|---|---:|---|---|
| Windows x64 ZIP | 14,071,813 | `d3d80652ba83163c7262cc79535b83812a4920b1a6fb0cf38fd4b4602dad7b20` | https://drive.google.com/file/d/1eUQw4rJmMeQkuZHEdYfg9GfXK5gAORXd/view |
| Linux x64 tar.gz | 13,986,202 | `2ed67c583c57d15d54f71827e3881a768ae43f3ccec9c88b46deef9c6b7e95d1` | https://drive.google.com/file/d/1fiJl4eHWAXIClBw1IfVujRaakYaRT0vM/view |
| Vendor-complete source ZIP | 36,994,941 | `806cdbe4f219e636ec32f7856b1b7ae8ec58040c2346c856cd388e30707d2fb3` | https://drive.google.com/file/d/1zL_Ma3MF3eNQK4lk9HcuBjhd81_YMmtZ/view |
| Verification evidence ZIP | 1,899,873 | `07ee17974107c6823230d81d7b3b2ad5602c1ee7258afb25f4c6978060fc4fba` | https://drive.google.com/file/d/17aqTwWuKBYNd__GH4EjHrLguvp6fLBO-/view |

Each Drive object was downloaded again after upload, hashed locally, compared with the immutable release manifest, and archive-tested. See `DRIVE-REMOTE-VERIFICATION.json`.

## Verification highlights

- Go tests and vet passed with Go 1.27.0 and vendored dependencies.
- Linux and Windows binaries were rebuilt from the extracted source archive and matched the packaged executables byte for byte.
- Repair Lab passed 25 production-UI assertions and 38 authenticated API calls with zero page, console, or request errors.
- Porting Lab passed 29 production-UI assertions and 23 authenticated API calls with zero page or console errors.
- Fresh initialization loaded SQLite 3.53.3 in WAL mode with 919 Minecraft versions, 15,133 toolchain releases, 342 searchable knowledge documents, and all 8 durable Repair Brain records.
- Process-restart persistence passed for repair sessions and the Compatibility Brain.
- Final runtime packages contain 53 internally hashed files; the source package contains 2,161 internally hashed files.

See `FINAL-RELEASE-VERIFICATION.json`, `RELEASE-NOTES-0.9.0.md`, and the evidence archive for the complete observed proof.
