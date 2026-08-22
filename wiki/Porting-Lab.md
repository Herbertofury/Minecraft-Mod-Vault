# 🔧 Porting Lab

> **Status:** ✅ Verified in the 0.9.0 release evidence.

Porting Lab handles installed-JAR forensics and migration planning without requiring the user to understand every loader/version API change first.

## Core workflow

1. Inspect the installed JAR and metadata.
2. Identify loader, Minecraft version and dependencies.
3. Build evidence-backed upgrade/downgrade plan.
4. Create an isolated migration workspace.
5. Quarantine/restore cryptographic duplicates safely.
6. Hand discovered migration knowledge to the Compatibility Brain.

## OmniBridge expansion

The 0.11.0 roadmap expands this into source/JAR archaeology, automatic migration knowledge, dependency substitution, whole-modpack migration, loader conversion and behavioral differential testing.

See **[OmniBridge](OmniBridge)** and **[Loaders & Versions](Loaders-and-Versions)**.
