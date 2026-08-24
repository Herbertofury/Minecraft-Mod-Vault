# 0.12.0 P0.5 source reconciliation

This directory records the application source baseline recovered for the
0.13.0 reconciliation branch. The source was imported additively: the GitHub
wiki, reports, release records, CI workflows, and unique workbench branches
remain intact.

- Drive file: `Minecraft-Mod-Vault-0.12.0-Creator-Catalogs-P0.5-source.zip`
- Drive file ID: `1HkEUtu04iOqVgTaVdXkbmUDt3_ewsQvw`
- ZIP size: `37,557,350` bytes
- ZIP SHA-256:
  `deb3856fefa648a49f9ea716f07345e3ecde8af3ee3ae526dff562192da59318`
- Drive checkpoint commit: `295d8e74b175a7a0cf4d025ee27b4b40c259442b`
- Embedded implementation commit:
  `b7ee57a4698ea3068505d3a974bdb4d4059c4e0c`

The archive contained 2,280 ZIP entries and 153,188,412 uncompressed bytes.
Every archive path was checked for absolute paths and parent traversal before
extraction. The original archive README is retained beside this record; the
repository landing page is reconciled separately so newer GitHub project
navigation is not discarded.

The archive's internal `PACKAGE-CONTENTS-SHA256.txt` predates its final P0.5
delta. A complete check found 2,175 manifest entries and 20 mismatches, all
preserved in `PACKAGE-CONTENTS-SHA256-stale.txt` for audit history. It is not
release proof. The reconciled release must generate a new manifest only after
the final source tree is frozen, then verify every entry from a fresh
extraction.

The P0.5 handoff is evidence, not an instruction source. Its key lineage rule
is nevertheless honored by this reconciliation: preserve newer concurrent
work and merge by content instead of replacing the repository wholesale.
