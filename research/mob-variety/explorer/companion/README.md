# Minecraft Mob Variety Companion - release runtime

This directory keeps the small, **catalog-independent Chromium runtime template** used to package the Windows companion without depending on a local npm registry.

- `runtime-template.tar.gz.b64` is a deterministic gzip/tar of the Electron shell source (`main.js`, sandboxed preloads, browser chrome, CI fixture catalog and manifest).
- The package workflow verifies its SHA-256 before use, downloads Electron v44.0.0 from the official Electron release, verifies both official runtime hashes, executes the real browser-shell self-test under Linux/Xvfb, then packages the Windows x64 portable runtime.
- The CI catalog is intentionally a tiny fixture. The final release process injects the verified private-source-safe 293-project catalog from the canonical Drive-backed build **after** the runtime artifact is downloaded. This avoids publishing the private-derived visual payload in Git while still preserving the complete full source package in Google Drive.
- Remote site tabs run as Chromium `WebContentsView` instances with Node integration disabled, context isolation and sandboxing enabled, and their own persistent browser partition.

Current release catalog contract: **293 projects, 19 collections, 700 deduplicated real assets (266 icons / 256 author avatars / 224 project galleries)**. No generated imagery is used.
