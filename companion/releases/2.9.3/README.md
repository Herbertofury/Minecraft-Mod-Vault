# Minecraft Catalog Companion 2.9.3

2.9.3 closes the remaining production CurseForge gallery failure exposed by Bok's Banging Butterflies while restoring the source-scoped negative and Description-media behavior documented in 2.9.1.

## Bok's Banging Butterflies — all 11 live gallery items

- The live CurseForge project currently exposes `Gallery (11)`.
- Direct `media.forgecdn.net/attachments/...` links are accepted as authoritative project gallery media after the exact project H1 and project-owned Gallery boundary, even when the nested `<img>` is only a lazy placeholder.
- The parser now starts at the **first** project-owned Gallery marker after the exact H1 rather than the last repeated `/gallery` link in SSR markup. This prevents later navigation/gallery links from moving the extraction window past earlier real images.
- The production-shaped regression fixture contains all 11 current Bok attachment URLs and verifies exact count, canonical originals, progressive stream seed, repeated-gallery-link handling, and global promo isolation.

## Empty gallery is source-scoped, not project-wide

- An empty CurseForge `/gallery` route is represented as `sourceGalleryAbsent`, not global `galleryAbsent`.
- The canonical project Description/post-media lane therefore keeps running. This preserves the DivineRPG-style case where the Gallery tab is empty but the project Description contains useful media.
- A real project-owned gallery result clears older source-scoped or terminal negative state.
- The renderer reports `Gallery tab empty — checking project post…` while canonical fallback discovery is in flight.

## Description-linked repository media

- Image-bearing GitHub links inside the exact, identity-bounded CurseForge Description are canonicalized from `/blob/` pages to `raw.githubusercontent.com` media.
- Image-bearing GitLab `/-/blob/` links are canonicalized to `/-/raw/` media.
- Repository navigation, issues, non-image documents, team/sibling regions, promotions, avatars, ads, and sponsor media stay excluded.

## Cache and identity safety

- Live-media cache schema is v12 so stale pre-fix negative/gallery states cannot pin upgraded installs.
- Existing project/icon/author role quarantine remains intact.
- No gallery cap, project cap, viewport record culling, static/fabricated image substitution, or provider polling was added.

## Exact-package QA

All **39 release suites** pass against the source tree, a fresh extraction of the final Source ZIP, and `resources/app` from a fresh extraction of the final Windows x64 ZIP.

Dedicated 2.9.3 coverage includes:

- Bok's 11-item direct-attachment CurseForge gallery topology;
- first-owned-gallery boundary vs repeated `/gallery` links;
- CurseForge global promotion isolation;
- `sourceGalleryAbsent` + canonical Description fallback;
- GitHub/GitLab Description image-link raw canonicalization;
- cache-v12 negative-state invalidation;
- all existing provider identity, author/media role, progressive streaming, startup overlap, native-network race, no-cap, and stress gates.

The Linux build host does not have Wine, so the Windows Electron executable is **not** falsely claimed as executed here. `Self-Test.cmd` is included in the Windows package for native Windows Electron/Chromium acceptance.

## Final artifact hashes

Windows x64 ZIP SHA-256:

`a150fe894f753fa1a7cb63189a9357e4121dc31e9f5af35b4cf6ed391b122025`

Source ZIP SHA-256:

`1cbe4c969205f528e373466273596256db324ef64604495d25decef943443edb`

Split Windows package:

- part01: `ed84a9fa17867e35605d4b7d0e27de428ba0262fbabb8326479ede2487b1b4f1`
- part02: `cf72e92e7734bb772197e9f85f88cd31cd811d5f38fb2ea996679b1cccad658e`

The two uploaded Drive parts were materialized back from Google Drive and reassembled byte-for-byte to the exact Windows ZIP SHA-256 above.