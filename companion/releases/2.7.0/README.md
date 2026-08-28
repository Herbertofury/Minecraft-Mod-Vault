# Minecraft Catalog Companion 2.7.0

2.7.0 is the strict-media-ownership + instant-frontier release.

## Correct media roles

- Project icon, creator avatar, and gallery media are independent semantic roles through provider parsing, cache sanitization, main-process merging, IPC, and renderer merging.
- The same URL cannot survive in multiple roles. Ambiguous cross-role collisions are quarantined rather than painted.
- CurseForge project imagery must be bound to the exact project H1/project entry. Creator avatars must be bound to the exact member/profile identity.
- Global promotions, sponsor art, tier frames/badges, and `More from` sibling-project imagery are rejected.
- Exact CurseForge creator Projects pages can provide both the creator avatar and the exact logo for each matching project. Cards sharing the creator single-flight the author-page request.
- ForgeCDN 256 px project-logo thumbnails remain the fast paint URL while the full original URL is retained.
- A provider response explicitly saying a project has no gallery is now a successful `galleryAbsent` state instead of an endless Refreshing loop.
- Live-media cache schema is v9, invalidating polluted earlier media records.

## Faster without dropping work

- Persistent-cache hydration and live network discovery now start concurrently. 2.6 waited for the whole catalog cache IPC batch before first live prime I/O.
- Initial live-media prime starts on the next microtask; the old 90 ms startup timer is gone.
- Media-prime execution scales to 3x logical CPUs with a 128-job ceiling.
- Eight Chromium media views are prewarmed.
- Existing Node progressive streaming, shared Chromium session/cache, worker-thread full-page parsing, wreq Rust/BoringSSL, Impit Rust/HTTP3, provider API seeds, preconnect, speculative image-byte warming, uncapped enrichment, uBlock Origin, and the auto-updating TWP translator remain intact.
- No project, source, provider, gallery, or result cap was added.

## Exact-package QA

All 30 release suites pass against source, staged Windows `resources/app`, and a fresh extraction of the final Windows ZIP.

Focused final-package deterministic network tests:

- Startup overlap fixture: same 90 ms cache-hydration workload preserved; median 105.88 ms barrier model vs 16.50 ms overlapped = 6.42x faster app-added critical path.
- CurseForge identity fixture: exact owned media median 29.88 ms vs 188.30 ms complete response = 6.30x, while an earlier PUBG/BATTLEGROUNDS promotion is prohibited from opening the media gate.
- Exact no-gallery state: ~23.60 ms in the same streamed fixture.
- Exact author ownership: ~27.48 ms.
- Creator fanout: 12 project cards, 1 physical author-page request, 12 unique exact project logos.
- Stress: 96 cards / 192 real localhost HTTP streams, 192 physical requests, 192 simultaneously active at the fixture server in the captured final benchmark.

These are deterministic real-socket regression measurements, not claims about CurseForge/Patreon/other WAN response time. Remaining cold latency after the app's own barriers are removed is provider RTT/server time and real image-byte transfer.

## Native Windows gate

The build host is Linux, so the Windows Electron `.exe` was not falsely claimed as executed here. The exact final Windows archive was fresh-extracted and its `resources/app` passed all 30 suites. `Self-Test.cmd` ships beside the executable for the native Windows Electron/Chromium gate.

Final Windows ZIP SHA-256:

`aa9fe510258a53fdff4ec0eebd4ab287ed6314b95a317875188b00094cd2c386`
