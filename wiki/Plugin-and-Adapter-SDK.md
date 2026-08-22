# 🧩 Plugin & Adapter SDK

> **Status:** 📋 OmniBridge/WorldForge roadmap.

The adapter system should make new formats, versions, loaders, world records and validators addable without turning the core into a pile of one-off special cases.

## Adapter families

- source readers
- target writers
- version translators
- loader migrations
- model/animation translators
- world format adapters
- semantic mapping providers
- validators
- repair rules
- editor tools/brushes
- external-tool bridges
- TestGrid workers

## Registry metadata

Every adapter should declare:

- supported formats/versions/loaders;
- maintenance state;
- license;
- required permissions;
- compatibility/fidelity tests;
- version pinning and dependencies;
- security/trust metadata.

## Safety

Third-party plugins should be isolated where practical, with explicit capabilities, safe mode, crash isolation, rollback and diagnostics.
