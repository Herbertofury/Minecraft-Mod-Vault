# Bundle API 1.1.0 -> native Forge 1.20.1

Authority: `TheRedBrain/bundle-api@10d750662dc58ce3a95b73eae14e110b1e1cc638`.
1.20.1 mapping/storage reference: `ThePotatoArchivist/BundleBackportish@ccaf6fdc17c460e2c2f8b99e023aa7ec6c73f737`.
Primary current consumer: `ZsoltMolnarrr/Archers@82a5329474548f5bcc7bddabd3cbabc2ab2cbda0`.

Target: Minecraft 1.20.1, Forge 47.4.23, Java 17, Yarn 1.20.1+build.10.

The 1.21 data-component implementation is translated to owning-ItemStack NBT rather than emulated globally. Canonical storage is `BundleAPI.Items` plus positive `BundleAPI.SizeMultiplier`; root vanilla `Items` is accepted as a migration source. Unrelated NBT is preserved. Occupancy uses exact Apache Commons `Fraction`: ordinary item = `1/(maxStackCount * sizeMultiplier)`, nested custom bundle = `1/16 + nested occupancy`, occupied beehive = 1, total capacity = 1.

Archers quiver acceptance remains current behavior: `ItemTags.ARROWS` only; size multipliers 4/8/12 => 256/512/768 ordinary arrows; current slot/cursor interactions, drop-all behavior, ordering, merge semantics, fullness bar and tooltip must be preserved before graduation.

This prep lane is intentionally isolated from the active Wizards acceptance head. It is not graduated until native Forge compile/package, deterministic rebuild, Java 17, server/client/package runtime, NBT round-trip/migration, mutation/capacity tests, and real Archers 3.1.1 consumer integration all pass.
