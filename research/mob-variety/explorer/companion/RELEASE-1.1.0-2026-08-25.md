# Minecraft Mob Variety Companion 1.1.0

Release date: 2026-08-25

This is an additive release receipt for the private-source-safe Mob Variety Companion. It does not replace or roll back unrelated Minecraft Mod Vault work.

## Catalog continuity

- 293 curated projects
- 19 collection views
- 700 deduplicated real embedded media assets
  - 266 project icons
  - 256 author avatars
  - 224 verified project gallery images
- Canonical Sheet/Doc/PDF remain authoritative and were not modified by this app-only release.
- Source workbook fingerprint remains `82179a97513be2d555f33c41e4d4fad431ad26b6c7cb180757d8a978dd23a909`.
- No generated imagery was used.

## 1.1.0 UX changes

### Browser-style Split Research divider

- Split divider is directly draggable with a 10 px browser-style hit target and a slim visible rail.
- Minimum pane width is enforced so neither catalog nor live website can be accidentally crushed to zero.
- Split ratio persists with the app session and is restored on relaunch.
- Double-click resets to 50/50.
- Right-click swaps catalog and live-site sides.
- Keyboard splitter controls: Left/Right fine adjustment, Shift+Left/Right larger adjustment, Home 25%, End 75%, Enter swaps sides.
- More menu exposes `Swap split sides` and `Reset split 50 / 50`.

### Actual galleries instead of static strips

- Restored the missing packaged `catalog/enhance.js` enhancement layer.
- Card gallery banners, Visual Atlas images, and detail galleries are upgraded into interactive galleries.
- Embedded verified imagery remains the instant/offline fallback.
- On first interaction the app can discover current project screenshots from the exact live project page using an isolated sandboxed Chromium `WebContentsView` and the normal persistent browser session.
- Live discovery favors screenshots/gallery/media/showcase/carousel assets and rejects/penalizes logos, avatars, ads, tracking pixels, navigation art, and footer chrome.
- Galleries support previous/next arrows, thumbnail rail, image counter, keyboard Left/Right, Escape, full lightbox, refresh live gallery, open project here, open project in the user's normal browser, and open the original image.
- Project imagery and author/avatar imagery have hover/focus zoom treatment.
- Discovery is cached for 30 minutes and fails gracefully back to the verified embedded gallery when a provider blocks automated inspection.

## Browser architecture

- Electron 44.0.0
- Chromium 152.0.7977.54
- Live project pages are real sandboxed Chromium tabs, not iframe/embed dependencies.
- Remote sites have Node integration disabled, context isolation enabled, and Chromium sandboxing enabled.
- `Open in your browser` remains first-class and always available for the exact live URL.

## Acceptance evidence

GitHub Actions runtime run: `32907194276`

The release runtime was rehydrated from a SHA-256-verified deterministic source transport, both official Electron 44 Linux and Windows x64 runtimes matched their expected SHA-256 values, and the real Electron/Chromium application was launched under Xvfb.

Self-test result:

```json
{"passed":true,"failures":[],"electron":"44.0.0","chromium":"152.0.7977.54","tabs":2}
```

The runtime self-test exercises normal browser tabs plus the 1.1 split/gallery-discovery path. Final application JavaScript syntax checks also passed for `main.js`, `preload.js`, `catalog-preload.js`, `shell.js`, and `catalog/enhance.js`.

## Final Windows x64 artifact

- Filename: `Minecraft Mob Variety - Companion Windows x64.zip`
- Size: 178,171,452 bytes
- SHA-256: `980f6a67a411f8304e94b1bdc06849166e44e6561c7a635cff4d9dab1acc3ccb`
- ZIP integrity: PASS
- Root folder: `Minecraft Mob Variety Companion`
- Launcher: `Minecraft Mob Variety Companion.exe`

## Rebuildable source

- Filename: `Minecraft Mob Variety - Companion Source.zip`
- Size: 3,354,555 bytes
- SHA-256: `568a6a04cb5a85cd050269dad257fa7c541dc3469a14cc2d3c861a3bae45d8d1`
- ZIP integrity: PASS

## Google Drive canonical release

Folder: `Recommendation Library / Companion App`

- Folder ID: `1vEhnOszAAXGWpC0epxeogo-ufhm5fACJ`
- Source ZIP ID: `1xz_texyF2s0njA8slNceOQUFhNhGevQy`
- Windows part 01 ID: `1X6y0E3E0WfwI4lefxSRXVXb9Xscne0qX`
- Windows part 02 ID: `1YLIbK-IkNhvFsaaOtWrDd5GewUchccR-`
- Reassembler ID: `1rNRrucj4qH6Ex1U8f6ffRyp4E95aUxgH`
- SHA-256 manifest ID: `1xxS6_8bBvif2bdp0xJbGmFpfhbcpdTL-`

Drive multipart hashes:

- part01, 89,128,960 bytes: `dd2616c67f6439883657d4e005c2cc61c0163184763634a5805c44c76b57559e`
- part02, 89,042,492 bytes: `4fa5025e4d0972518a41b1b68ccdfc538f92708cb12128b712c28dcd0cb7147e`

The new Drive files were materialized back out of Google Drive. Their hashes matched the local release exactly; the downloaded parts were concatenated, produced the exact final Windows ZIP SHA-256 above, and passed ZIP integrity testing. The previous 1.0 Companion files were preserved non-destructively in the project History folder before 1.1.0 publication.
