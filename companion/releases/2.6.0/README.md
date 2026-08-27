# Minecraft Catalog Companion 2.6.0

Identity-safe live media, real parallel parsing, and integrated Translate Web Pages (TWP) release.

## Wrong-image fix

The first streamed CurseForge/ForgeCDN-looking URL is no longer treated as project media. The fast gallery path now waits for the exact project H1, enters only the project-owned gallery region, rejects promotion/campaign/ad artwork, and preserves the real project image/icon/gallery relationship. Live-media cache schema is bumped to v8 so polluted 2.5 identities are rediscovered.

A release regression fixture specifically places `PUBG BATTLEGROUNDS UGC CONTEST` artwork before the `Maid Useful Tasks` H1 and verifies that the banner cannot become project media.

## Real parallel speed work

- Prewarmed `worker_threads` provider parser pool moves expensive full-page enrichment across CPU cores while first-media prefix parsing remains hot-path.
- Final fresh-extracted benchmark: 32 x ~405 KiB provider fixtures, 4 workers, 125.65 ms serial vs 49.75 ms parallel = 2.53x parser throughput.
- Visible/near-visible source images start Chromium `Image.decode()` work in parallel with normal paint.
- Same-session preview-byte warming is widened, and six Chromium media `WebContentsView`s are preconstructed for cold visible-card discovery.
- Existing Node, Chromium, wreq Rust/BoringSSL, Impit Rust/HTTP3, preconnect, single-flight and uncapped gallery enrichment remain intact. No projects, providers, sources, records, or gallery results are hidden/capped to manufacture the speedup.

## Integrated TWP translator

The protected browser chrome now contains an integrated TWP-compatible translator based on `FilipePS/Traduzir-paginas-web` architecture:

- Bing, Google, Yandex, DeepL;
- whole-page translation;
- selected-text translation;
- Original / Translated toggles;
- dynamic-page MutationObserver translation;
- per-site automatic translation;
- target-language selection;
- persistent translation cache and request single-flight;
- isolated Chromium world for page text manipulation so remote pages get no Node/Electron privileges.

### Auto-update

Like the existing uBlock integration, the translator checks the official TWP GitHub releases every six hours. It downloads the official tagged source archive, verifies TWP manifest identity/version, MPL-2.0 licensing and required translation-core files, stores the relevant upstream source plus SHA-256 receipt under user data, and hot-reloads only strict allow-listed translation-service endpoint recipes into the audited Electron adapter. Downloaded upstream JavaScript is not blindly executed in the privileged main process.

## QA

26 release suites pass on the source tree, the staged Windows `resources/app`, and a fresh extraction of the final Windows ZIP. Key fresh-extracted measurements:

- parser pool: 2.53x throughput on 32 large provider fixtures;
- 32 identical media consumers -> 1 physical GET;
- 128 real localhost media streams -> 128 simultaneous physical requests;
- provider API seed: 21.11 ms vs canonical first media 83.84 ms in its deterministic fixture;
- TWP translation batching: 100 text segments -> 4 parallel requests with cache reuse verified;
- TWP updater fixture verifies source retention, SHA-256 receipt, manifest/license checks and endpoint allow-list enforcement.

The Linux build host cannot execute the final Windows `.exe`; the release includes `Self-Test.cmd` for the native Windows Electron/Chromium runtime gate.

Windows ZIP SHA-256: `da41e00ebce8d14cf18c7716d8f90fcc16283081b36e55574a394de29020d95b`.
