# Minecraft Catalog Companion 2.9.5

2.9.5 is the corrective release candidate produced after the real Windows 2.9.4 acceptance test still showed `Live source did not expose this media yet` for Bok's Banging Butterflies. The prior parser-only and hidden-DOM fixes were not sufficient, so this pass traced the complete renderer/cache/prime pipeline and removed the concrete terminal-state failure instead of adding another narrow parser exception.

## Root bug found in the real app path

The renderer's prime scheduler treated an icon + author/avatar cache result as if project media discovery were complete:

```js
if ((s.gallery.length || s.icon || s.galleryAbsent) && s.quickLoaded && !s.cacheStale && (s.author || !s.authorName)) continue;
```

That meant the exact state visible in the failed Windows screenshots — project icon present, author present, gallery absent — could permanently suppress another gallery prime. The card could therefore become terminal even though CurseForge still exposed a real `Gallery (11)`.

2.9.5 removes `s.icon` from that completion condition. Only an actual gallery result or an authoritative project-level gallery-negative can satisfy gallery completion. Cache schema is bumped to **14** so old 2.9.4 icon/author-only blank states are invalidated on upgrade.

## Exact CurseForge gallery recovery is now redundant by design

The exact same-project `/gallery` route now has three independent recovery paths rather than depending on a single hidden-DOM rescue:

1. **Chromium-session full HTML** — retains and parses the complete exact Gallery response using the application's persistent Electron session.
2. **Node full HTML fallback** — independently retains and parses the complete Gallery body.
3. **Chromium DOM rescue** — remains as the dynamic/lazy-markup fallback.

For visible and near-visible cards, the prime path now keeps full exact Gallery responses and delivers late full-gallery results even after the initial icon/author result has painted. The old cheap first-media probe remains for responsiveness, but it is no longer allowed to make gallery discovery terminal.

## Live CurseForge topology accounted for

The current Bok page exposes the tab order `Description -> Comments -> Files -> Gallery (11) -> Relations -> Issues`, followed by eleven project media items. Exact `/gallery` parsing now treats those tab labels as navigation, not content/end boundaries. The Gallery region is identity-bounded after the exact project H1 and ends only at real project/global boundaries, preventing both early truncation and unrelated promotion leakage.

The native packaged self-test fixture mirrors this topology, contains all eleven Bok-shaped direct ForgeCDN attachment links, includes a promotion before the project H1, and fails unless all 11 project attachments survive with zero promotion leakage.

## QA

All **40 release suites** pass against:

- the 2.9.5 source tree;
- a fresh extraction of the final Source ZIP; and
- `resources/app` from a fresh extraction of the final Windows x64 ZIP.

Focused gates cover the exact 11-item Bok topology, full-HTML gallery recovery, DOM rescue, renderer cache terminal-state prevention, cache schema 14, source-scoped negatives, Description-media fallback, project/author/media role isolation, progressive delivery, no gallery caps, provider concurrency and stress behavior.

The build host is Linux and Wine is unavailable, so the Windows Electron executable is **not** falsely claimed as executed here. `Self-Test.cmd` is included in the Windows package and now directly exercises the same packaged Electron/Chromium Gallery paths that matter for this regression.

## Artifacts

Windows x64 ZIP:

- size: `196792972` bytes
- SHA-256: `4fcb507f058813cd15a88e73ed210f14b594deccd7c72b138b25205ad2993377`

Source ZIP:

- size: `22000623` bytes
- SHA-256: `c736caecdb74fe72f683b543e99405e4ed848acce29eeb407497fcf75207547e`

Drive-safe Windows parts:

- part01: `47c8a5b9f8ee132478613d8bc029c1408e16619ee444443a2563f61b827f5b12`
- part02: `a60067c905fef8e9095f716ff7426a435197e78b9326a838e08338a241b46c0e`
- part03: `737ed9feb61c57d196a865c7195069f8db0b720edab4a6cd57eeb7c59bfd30aa`

The uploaded Google Drive parts were materialized back, concatenated, and verified byte-for-byte against the final Windows ZIP SHA-256. The Drive Source ZIP was also round-tripped and matched its final source hash. `unzip -t` reports no compressed-data errors.

## Acceptance target

The real Windows acceptance target remains Bok's Banging Butterflies. 2.9.5 specifically removes the icon/author cache short-circuit that could suppress its Gallery discovery and adds retained full-response recovery ahead of the DOM fallback. This release is intentionally described as a corrective release candidate until the real Windows card shows the 11-image Gallery.