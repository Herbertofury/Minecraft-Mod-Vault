# Minecraft Catalog Companion 2.9.0

2.9.0 is the persistent research-workspace + zero-busywork release.

## It-just-works research QoL

- Per-catalog workspace state now survives restart: query, structured filters, sort, density/view mode, preset, minimum score, and active toggles return exactly where you left them.
- Explicit hash/deep-link state still wins over local restore, so shared/catalog links never get silently overridden by an old local workspace.
- Named saved views can be saved, updated, selected, and deleted; `Ctrl/Cmd+Shift+S` saves the current view.
- The old transient compare selection is now a persistent four-item research tray with Compare, Open selected, Copy links, Export selected, and Clear.
- Creator actions now include Open creator in browser and Copy creator, while project detail can Copy all verified project/source links in one action.
- Clipboard text operations use a protected local IPC bridge rather than depending entirely on page clipboard permission.

## Media QoL without another network tax

- Failed gallery images/videos are quarantined immediately and fall through to the next already-discovered valid media item before any provider refresh is attempted.
- A forced provider refresh happens only when the app has no known local/live-media fallback left.
- Gallery status exposes the active media type, provider, and parser/source provenance where known.
- The lightbox has Copy media URL for the exact active asset.
- No provider polling lane, project cap, source cap, gallery cap, or viewport-culling regression was added.

## Exact-package QA

All 38 release suites pass against:

- the source tree;
- a fresh extraction of the final Source ZIP;
- the staged Windows `resources/app`;
- a fresh extraction of the final Windows ZIP's `resources/app`.

Focused final-package regression evidence:

- shared media single-flight: 32 concurrent consumers, 1 physical request;
- progressive streamed head: 181 bytes read, zero body bytes sent before abort in the controlled fixture;
- startup overlap fixture: 106.62 ms barrier median vs 15.81 ms overlapped median = 6.74x reduction in app-added critical-path delay while preserving cache work;
- stress fixture: 96 cards / 192 flows / 192 physical requests, all 192 simultaneously accepted by the fixture server, p95 166.1 ms;
- Modrinth bulk fixture: 400 projects / 7 chunks / lossless;
- no gallery cap.

Timing figures are deterministic localhost/runtime regression measurements, not claims about third-party WAN latency.

## Native Windows gate

The build host is Linux, so the Windows Electron executable was not falsely claimed as executed here. `Self-Test.cmd` is included beside the app for the native Windows Electron/Chromium smoke gate.

Final Windows ZIP SHA-256:

`4306c9e35b57369aa575aeef746a3a03999c72b046c39b8fca83336366b25ce4`

Source ZIP SHA-256:

`6e37c81bb6fadb792c92435ada76e8a093115e0908c63402d842fe9d523daf20`
