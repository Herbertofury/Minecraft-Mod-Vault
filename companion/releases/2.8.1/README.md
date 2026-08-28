# Minecraft Catalog Companion 2.8.1

2.8.1 is the zero-extra-request semantic-media hardening release built directly on 2.8.0.

## What changed

- Preserves the 23-provider exact project/author adapter registry, independent creator-avatar lane, role quarantine, uncapped galleries, native HTTP races, Chromium fallback, uBlock Origin, and TWP translator.
- Adds social video discovery from exact project/post HTML using `og:video`, `og:video:url`, `og:video:secure_url`, and `twitter:player:stream`, retaining the exact social image as video poster/preview.
- Adds semantic lazy media extraction from `data-bg`, `data-background`, `data-background-image`, `data-image`, and `data-media-src` when the element itself has project/post/gallery/content semantics.
- Adds typed direct MP4/WebM/OGV/MOV/M4V/HLS post links when the anchor is semantically a video/post/media control.
- Mirrors those same capabilities in the protected Chromium DOM fallback for JavaScript/session-heavy providers.
- Tightens structured creator URLs: JSON-LD creator images may still be used, but only provider-recognized creator/profile routes are allowed to launch independent author enrichment. A project URL can no longer masquerade as `author.url` and create a bad extra request.
- Adds fast negative gates so ordinary pages that do not advertise social video/data media/direct video links skip the richer parsing work.
- Adds `social-post-media-qa.js` and raises the exact release matrix to 37 deterministic suites.

## Exact role/sample coverage

- Planet Minecraft: exact member binding remains verified against the Enderwoman / `redstonae` case; creator avatar, project hero/showcase, comments, and sibling projects stay isolated by semantic role.
- AFDIAN: `vm-pic` / `img-pre` transformed previews retain the clean original CDN asset for full-resolution hover/lightbox, while image/GIF/video remain typed gallery media.
- Generic provider path: Schema.org/JSON-LD image, screenshot, associated media, VideoObject, structured creator media, semantic data attributes, social video, and direct video links share the same identity/role gates.

## Performance / zero-loss checks

- No new provider request lane was added by 2.8.1.
- Project/gallery delivery still does not wait for creator-profile enrichment.
- No project/source/provider/gallery/result cap was introduced.
- The final fresh-extracted Windows `resources/app` passes all 37 suites.
- Real-socket stress remains 96 cards / 192 simultaneous HTTP streams, and startup-overlap / progressive-media / parser-worker gates all remain in the release matrix.
- Same-host 32 x ~405 KiB parser regression sampling remains inside the existing multi-core release gate; the new rich-media paths are fast-negative-gated on ordinary pages.

## Native Windows gate

The build/QA host is Linux and Wine is unavailable, so the Windows `.exe` is not falsely claimed as executed here. `Self-Test.cmd` ships beside the executable for the native Electron/Chromium acceptance pass.

Final Windows ZIP SHA-256:

`592d184d50ed9a87ec319a829a1ee21586291bbc357c11e9ff2867a6a4d2dba4`

Source ZIP SHA-256:

`6d8f35cb9e2620c577bd251489e9d8005be32209bd7147584f5d6258e7c36bee`

Split parts:

- part01: `c9fe0d1055b29fc39175087531adfe65db92fdaf72e7dedc834de10ab8fdf0f2`
- part02: `1862a1da9e6a6c74dacaadb82a4b9af4a3637e9d3e5c5ec1d67abf79f49d36b3`
- Reassemble.cmd: `b8e9ebf3c465076af249504e1a8794c7634d4a472fac167bd3195a0e8ec2c492`
