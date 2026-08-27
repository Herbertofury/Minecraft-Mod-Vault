# Minecraft Catalog Companion 2.5.0

Universal live-media performance release for the Minecraft Mod Vault companion.

## Highlights

- 23 first-class provider families: CurseForge, Modrinth, GitHub, GitLab, Hangar, SpigotMC, Bukkit, BuiltByBit, Nexus Mods, ModDB, Polymart, Planet Minecraft, MCPEDL, ModBay, AFDIAN, Patreon, Minecraft Marketplace, BOOTH, Fourthwall, Ko-fi, itch.io, Gumroad, and alltheysm.
- Provider-adaptive minimum full-response keepers plus short physical probes instead of running every engine as a full download on every provider.
- Non-exclusive public project-icon metadata seeds for SpigotMC/Spiget, Hangar, and GitLab; canonical provider pages still continue for full uncapped gallery enrichment.
- Real Chromium session fetch/DOM, Node progressive streaming, vendored native Rust wreq-js 3.2.0, native Rust Impit 0.14.4 / Chrome 151, HTTP/3 specialty hedging, same-session image cache warming, origin preconnect, connection reuse, and single-flight collapse.
- BuiltByBit defaults to real Chromium browser navigation only when no authenticated official API token is available; direct automated scrape lanes are disabled.
- No project/result/gallery caps and no generated replacement imagery.

## QA

23 release suites pass on the source tree, Windows staging `resources/app`, and a fresh extraction of the final Windows ZIP. Deterministic real local-TCP five-run medians: public metadata icon seed 21.05 ms vs canonical-page first media 83.11 ms (3.94x); provider-adaptive network policy reduced bytes by 66.7% and first-media P95 from 29.67 ms to 3.92 ms in its fixture. These are regression benchmarks, not provider WAN claims.

The Linux build host cannot natively execute the Windows Electron `.exe`; `Self-Test.cmd` is included with the Windows release for that final runtime gate.

Release binaries, multipart Windows package, source archive, portable explorers, QA results, checkpoint, and SHA-256 manifest are published in the Google Drive `Minecraft Mod Vault` folder.
