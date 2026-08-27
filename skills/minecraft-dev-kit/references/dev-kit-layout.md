# Minecraft Dev Kit layout and discovery

## Canonical Drive root

When Google Drive is connected, prefer the user's canonical Dev Kit first:

`https://drive.google.com/drive/folders/1LJ12tL6n__IEf6UdB0iyVx3V3gF1VCeG`

Do not ask the user to re-upload a tool that is already present and readable there. Inspect the relevant subfolder directly rather than broad-scanning all Drive content.

Established high-level layout:

- `00 Documentation & Index`
- `01 Toolchains & SDKs`
- `02 Build & Port Frameworks`
- `03 Launchers & Test Harnesses`
- `04 Reference Runtime Mods`
- `05 Asset & Content Tools`
- `06 Reverse Engineering & Bytecode`
- `07 Profiling, Diagnostics & Rendering`
- `08 Orchestration & Automation`
- `AI Dev Tools`
- `Projects`

Important reusable components include Java toolchains, Gradle distributions, loader MDKs/caches, decompilers, Blockbench/Blender, RenderDoc, profiling tools, FFmpeg/ImageMagick, native Linux headless runtime packages, and project-specific QA toolkits.

## Orchestrator

Prefer the newest verified `mmv-devkit` release in `08 Orchestration & Automation`. Known 2.3.0 capabilities include:

- `validate`
- `check`
- `sync`
- `watch`
- `sources`
- `heritage`
- `port-guard`
- `adopt-tools`
- `cache-doctor`
- `cache-reassemble`
- `archive-split`
- `client-assets`
- `client-natives`
- `world-qa-enable`

Use `--help` on the installed version before assuming newer flags.

## Reuse before download

Before fetching a new JDK, Gradle, loader, mapping, native, model tool, profiler, or dependency cache:

1. Check the canonical Dev Kit folder.
2. Validate the candidate by embedded version/hash/runtime launch, not filename alone.
3. Reuse exact compatible artifacts.
4. Download current official artifacts only when the Dev Kit lacks a compatible copy or the task explicitly requires a newer version.
5. Put genuinely reusable new tools back into the correct Dev Kit subfolder and record provenance/version/checksum.

## Missing-tool behavior

If a missing runtime/tool materially blocks correctness, speed, fidelity, or verification, tell the user exactly what is missing and include a direct official download/use link. Never request credentials, launcher sessions, Microsoft/Minecraft tokens, cookies, or account files for development QA.
