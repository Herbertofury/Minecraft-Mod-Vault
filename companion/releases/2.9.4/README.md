# Minecraft Catalog Companion 2.9.4

2.9.4 is the native-runtime follow-up to the failed 2.9.3 Bok's Banging Butterflies acceptance test. The 2.9.3 HTTP/parser fixture was not sufficient: the real Electron path could still finish without ever traversing the exact CurseForge gallery DOM.

## Production failure reproduced from the Windows test

The live Bok's Banging Butterflies CurseForge project exposes `Gallery (11)` and eleven direct ForgeCDN attachment links, but the card could still end at `Live source did not expose this media yet`.

The remaining causes were runtime orchestration and DOM ownership gaps rather than a missing list of image URLs:

- CurseForge's exact Gallery tab is inside navigation while the generic deep-media page-link collector excluded nav links.
- Bounded streaming probes could finish before the real attachment region with no guaranteed exact `/gallery` Chromium DOM rescue.
- Direct ForgeCDN anchors with lazy placeholder child images could score below the ordinary gallery threshold.
- Prime discovery could become terminal from icon/author evidence before a real gallery rescue lane settled.

## 2.9.4 fix

- Adds an exact same-project CurseForge `/gallery` Chromium DOM rescue lane using the Companion's persistent Electron session.
- Exact gallery extraction is bounded after the exact project H1 so global promotions before the project cannot become gallery media.
- Direct `media.forgecdn.net/attachments/...` links are authoritative project gallery candidates on the exact gallery surface even when the nested image is lazy/placeholder markup.
- Preserves the exact same-project Gallery route through nav filtering while leaving unrelated navigation excluded.
- Polls the live gallery DOM for its advertised `Gallery(N)` count and allows a bounded visible-card settle window instead of terminating as soon as icon/author data arrives.
- Empty provider Gallery evidence remains `sourceGalleryAbsent`, so a project Description-media fallback such as DivineRPG remains independent.
- Live-media cache schema is bumped to **13** so stale pre-fix blank/negative states are invalidated.
- The packaged native Electron self-test now contains a Bok-shaped 11-item exact-gallery fixture plus a pre-project promotion image and hard-fails unless all eleven project attachments survive with zero promotion leakage.

## Exact-package QA

All **40 release suites** pass against:

1. the source tree;
2. a fresh extraction of the final Source ZIP; and
3. `resources/app` from a fresh extraction of the final Windows x64 ZIP.

Focused gates include:

- `curseforge-gallery-dom-rescue-qa.js`;
- `curseforge-gallery-anchor-qa.js` — exactly 11 Bok-shaped direct attachment URLs;
- `curseforge-scoped-negative-qa.js`;
- `curseforge-description-link-media-qa.js`;
- all existing identity/role, progressive-media, no-cap, provider, concurrency and stress gates.

The build host is Linux and Wine is unavailable, so the Windows Electron executable is **not** falsely claimed as executed here. `Self-Test.cmd` ships in the Windows package and exercises the new exact-gallery DOM path in native Electron/Chromium.

## Artifacts

Windows x64 ZIP SHA-256:

`e81af7438e2449bb44b5e9f8cef9368a111eb55099e8a02279f1765abe5f00c1`

Source ZIP SHA-256:

`f09f0f9ef7cd5fd3eaae2505f268771d2f5315ea08db64cf054fb8978b82ef6c`

Drive-safe Windows parts:

- part01: `f6137f648fa8d55cc44d031e66e700ca72d65d7f14b5a33dd054a568e290c75a`
- part02: `332c4d66c9c7e26780f53fa0bdf741e64ee5cb4bb49bc6501845e2075fc62d61`

The uploaded Drive parts were materialized back from Google Drive, concatenated, and verified byte-for-byte against the exact final Windows ZIP. `unzip -t` reports no compressed-data errors.

## Native acceptance target

Open Bok's Banging Butterflies in the exact 2.9.4 Windows build. A visible card should enter `Checking exact CurseForge gallery…` while the exact gallery lane is active and then populate from the real project gallery instead of becoming terminal from icon/author-only evidence.

If that still fails on native Windows, run `Self-Test.cmd` and preserve the output. That result will separate packaged Chromium DOM extraction from live provider/network/session behavior without another speculative parser-only rewrite.