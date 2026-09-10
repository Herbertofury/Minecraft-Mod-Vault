# Backpacks 2.0.2.4 Semantic Hardening — Build 1

Status: **build-green; expanded native regression pending**.

- Minecraft: 1.20.1
- Forge: 47.4.23
- Java: Eclipse Temurin 17.0.20.1+1
- Gradle: 8.8
- Offline build: `gradle --offline --no-daemon clean build` — PASS including reobf
- Generic Bedrock semantic coverage: PASS — 320 obligations, 0 uncovered, 0 unresolved, 0 proof failures
- Semantic visual audit: PASS
- JAR SHA-256: `bd919aa49f7d019c046e7d1b5752f02a73188890d1cc038a46d3a383f8ac307c`
- Source checkpoint SHA-256: `d9446dbb53180576d384c96157ecc4cf68157708aebb03db3e09acdf51cc4721`

Drive copies:
- JAR: https://drive.google.com/file/d/14r-i4q7ZkTAv1VXIBAGtcUucwBk5Ve5I/view
- Source: https://drive.google.com/file/d/1bsLrOSgehhvDT1Yy3U88tq9FlHfp5oMO/view
- Checkpoint: https://drive.google.com/file/d/1ds53IxHkxO9DM23dg9-IqR7IDEiIcG-Z/view
- SHA256SUMS: https://drive.google.com/file/d/1l9C0C4B_szMZLaFsUHhFCpzG9DXWmBeZ/view

## Repairs represented by this checkpoint

- Exact Bedrock X-basis geometry conversion and 527-state emissive-alpha routing retained.
- Authored 21-cuboid backpack-crafter geometry is now live through a Java block-entity renderer with cardinal-facing and source material/mining mappings.
- Guide source semantics restored: stack 64, version 2.0.2 provisioning marker, authored lore, join provisioning, and page-turn audio.
- Locator source semantics restored: click-to-toggle, session-only activation, five-tick cadence, main-hand deactivation, and source-style cross-dimension direction indication.
- Third-person main-hand attachable evaluates `animation.sqst_bkpk.swing` from the actual LivingEntity movement state rather than a static held transform.
- Bedrock offhand worn-placement gesture is adapted to the Curios back slot.

## Continuing after Build 1

The working tree has already advanced past this checkpoint to harden Creative anti-duplication/storage identity and default backpack naming semantics. Build 1 remains intentionally immutable as a recovery point.