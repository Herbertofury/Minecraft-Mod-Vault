# Minecraft Catalog Companion 2.9.1

2.9.1 fixes two distinct CurseForge media-discovery failures found during native catalog testing without adding provider polling, gallery caps, or serialized work.

## Bok's Banging Butterflies — full gallery recovery

- CurseForge currently exposes a real `Gallery (11)` for Bok's Banging Butterflies.
- Exact gallery attachment links are now first-class owned-media candidates inside the bounded project gallery region.
- The authoritative ForgeCDN attachment URL can be retained even when the nested `<img>` is only a placeholder/thumbnail or its markup changes.
- Avatar, navigation, advertising, sponsor, and unrelated sibling media remain excluded by the existing ownership gates.

## DivineRPG — post media survives an empty gallery tab

- DivineRPG's CurseForge `/gallery` surface legitimately reports no gallery items while the canonical project Description contains the useful project imagery.
- An empty provider gallery tab is now represented as `sourceGalleryAbsent`, not global `galleryAbsent`.
- The canonical Description/post-media pass therefore continues independently instead of being suppressed by the earlier `/gallery` negative.
- Image-bearing GitHub/GitLab links inside the bounded project Description are normalized to their raw media targets after ownership validation, allowing the real DivineRPG dimensions/bosses/NPC/armory media to populate the project gallery.
- The UI now reports `Gallery tab empty — checking project post…` rather than prematurely declaring that the whole project has no media.

## Cache and correctness

- Live-media cache schema is now v11 so bad pre-fix negative states cannot keep either project stuck after upgrading.
- A discovered real gallery clears an earlier global negative if one somehow arrives from another lane.
- Provider-gallery absence remains useful evidence for scheduling, but it can no longer become a terminal project-media result by itself.

## Exact-package QA

All 39 release suites pass against the source tree, fresh Source ZIP, staged Windows `resources/app`, and a fresh extraction of the final Windows ZIP.

Dedicated final-package fixture coverage includes:

- Bok-style link-wrapped CurseForge attachment gallery extraction;
- DivineRPG-style empty `/gallery` plus image-rich canonical Description fallback;
- GitHub raw-media canonicalization;
- scoped negative-state semantics and cache schema 11;
- all existing author/media identity, progressive streaming, startup overlap, native HTTP race, no-cap, and 192-flow stress suites.

Focused controlled-runtime evidence remains healthy: 32 concurrent media consumers single-flight to one physical request; the 96-card/192-flow stress fixture accepted all 192 requests concurrently; and the startup-overlap fixture preserved all cache work while reducing the measured app-added barrier from 106.92 ms median to 15.67 ms median. These are deterministic localhost/runtime regression measurements, not claims about third-party WAN latency.

## Native Windows gate

The build host is Linux, so the Windows Electron executable was not falsely claimed as executed here. `Self-Test.cmd` ships with the release for native Windows Electron/Chromium acceptance.

Final Windows ZIP SHA-256:

`c424a7d58f93719166f05f636edf6ccdb7eb0e49bd9ae40060a8e3505a91087e`

Source ZIP SHA-256:

`8321df8477ec3714bba2358e3e2bd8ab21b0235461392a9c4dcb4c05c53f4947`
