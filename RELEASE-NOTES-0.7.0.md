# Minecraft Mod Vault 0.7.0 release notes

Version 0.7.0 is the federated-browser and updater-hardening release. It preserves the compiled desktop architecture from 0.6.0 while replacing narrow provider paths with a broad, capability-aware source layer.

## Universal browser

- Expanded the integrated provider registry to 28 source lanes.
- Search all enabled sources or any exact subset from the Mods workspace.
- Added stable federated pagination so provider windows are expanded first, merged/ranked second and sliced last. This eliminates double-offset skips and unstable Load More behavior.
- Added Grid/List modes, source filtering, provider presets, **Select all relevant**, Reset browser, provider health and per-source result facets.
- Provider chips that cannot publish the chosen content type remain visibly irrelevant and are skipped by the backend rather than wasting network calls.
- Added type-aware browsing for Mods, Modpacks, Resource Packs, Shaders, Data Packs, Plugins, Bedrock Add-ons, Worlds/Maps, Skins and Tools.
- Added live Modrinth and CurseForge taxonomy merging alongside the custom Vault category system.
- Improved duplicate merging so same-title projects from clearly different authors are not incorrectly collapsed while strong cross-provider identity matches keep their source variants.
- Preserved provider artwork, project icons, author avatars, galleries, compatibility, source variants and richer statistics through cross-provider merging.

## Source expansion

- Added Smithed v2 native search/details/dependencies and verified data/resource-pack installation.
- Added Sponge Ore native search/details/version metadata and verified plugin installation.
- Added Polymart in-app discovery/details.
- Added official Vanilla Tweaks picker discovery/details.
- Added MinecraftMaps.com, ResourcePack.net and Texture-Packs.com integrated search/details plus verified detected-package installation.
- Added MCreator Community, ShaderPacks.com, Shaderpacks.net, MinecraftShader.com and The Skindex specialist lanes.
- Added MinecraftHub as a curated cross-provider discovery lane. Its results stay in-app, project pages retain creator/media/version clues, and install resolution delegates back to the original provider rather than using the directory as a mirror.
- Expanded CurseForge Java and Bedrock routes, including Java data packs, Bukkit plugins, add-ons/customization and Bedrock add-ons, scripts, maps, textures and skins.
- Expanded Planet Minecraft and GitHub search/install semantics by content type.

## Verified installation

- Added bounded in-app detected-download resolution for integrated web providers.
- Rejects HTML/fake downloads before installation.
- Validates JAR/ZIP/MRPACK/MCPACK/MCADDON/MCWORLD containers before placing content.
- Supports bounded two-step intermediate download pages for providers that use an internal download handoff.
- Added safe Java-world extraction with `level.dat` verification, traversal rejection, extraction limits and unique save naming.
- Bedrock packages keep their own handoff path.
- GitHub release installs support mods, plugins, resource packs, shaders and data packs with SHA-256 verification when the release publishes a digest.

## Universal Updater

- Retains exact Modrinth SHA-512 identity and CurseForge MurmurHash2 fingerprint matching.
- Adds exact GitHub repository identity from embedded JAR metadata. Automatic GitHub replacement requires the release/asset metadata to explicitly prove the target Minecraft version and loader.
- Adds developer-declared canonical Modrinth/CurseForge project identity. This recovers old or repacked JARs that no longer match provider fingerprints.
- Modrinth canonical metadata uses provider-declared target-version/loader releases and recognizes an already-current declared version.
- CurseForge canonical metadata resolves exact slugs with the official Search Mods API and target-compatible files when an API key is configured.
- CurseForge candidates that introduce required dependencies remain review-only.
- Automatic apply still stages downloads, verifies provider hashes/size, parses downloaded JAR metadata, rejects wrong mod IDs and wrong loaders, backs up the original, atomically replaces it and rolls back on failure.
- Port/fork/continuation discovery now uses every enabled relevant provider rather than a three-source hard-code.

## Recommendations and Creator Picks

- Broadened recommendation queries across the full live taxonomy, configured interests, current game/loader compatibility and all relevant enabled providers.
- Provider cache includes fresh and stale-fallback behavior, keeping useful results during partial provider outages.
- Creator Picks resolution uses the full relevant provider set.
- Source variants in project details are actionable in-app search routes rather than passive labels.

## Reliability

- Provider calls use bounded timeouts, retries for transient 429/5xx responses and cache pruning.
- Provider settings automatically migrate when the provider schema grows.
- Added regression coverage for source paging, content-type routing, verified package chains, world extraction, provider cache fallback, duplicate merging, plugin facets, MinecraftHub canonical-source installation, updater metadata identity and rollback/preflight behavior.

## Creator Archive: exhaustive channel recommendation intelligence

- Added permanent tracked archives for `@AsianHalfSquat` and `@EnderVerseMC`.
- Full-history channel enumeration uses the YouTube uploads playlist when an API key is configured and a yt-dlp Videos/Shorts/Streams sweep as the fallback path. Historical indexing has no 200-video cap.
- Every indexed upload is queued resumably for description parsing, direct source-link resolution, public/automatic captions, persisted timed transcripts, and local whisper.cpp speech recognition when captions are unavailable.
- Local transcription now supports Large v3 Turbo Q5, Large v3 Turbo, and Base models, Minecraft terminology prompting, saved timed segments, retry cooldowns, and configurable parallel analysis workers.
- Recommendation evidence combines description links, description names/context, video titles, transcript mentions/context, provider metadata, project summaries, timestamps, confidence, and unresolved mention review.
- Creator Archive UI adds channel/kind/search filters, Mods and Videos views, newest/oldest recommendation ordering, per-video ordering, mod-name/channel/confidence sorting, transcript search, exact timestamp navigation, progress counters, and per-channel/full rescans.
- Creator following is now user-extensible: paste an `@handle`, channel URL, legacy `/user/` or `/c/` URL, or `UC...` channel ID to add a creator with immediate resumable full-history sync.
- Added a curated high-signal creator catalog with current showcase, deep-modded, historical, and utility tiers plus one-click Follow + Sync and Add Top Current Picks actions.
- Followed creators support pause/resume of automatic refresh; manual per-channel rescans still work while paused. Non-core creators can be unfollowed without deleting their indexed videos, transcripts, or recommendations.
- Archive channel filters now resolve by channel ID or handle, so custom and legacy channel URLs remain stable after canonical resolution.
