# Minecraft Mod Vault 0.11.0

## Minecraft Mod Vault 0.11.0 — OmniBridge

**One immutable source, every truthful Java and Bedrock target.** OmniBridge imports mods, source projects, data/resource packs, behavior/resource packs, scripts, worlds, templates, modpacks and structures; builds a Universal Minecraft Content Graph; emits deterministic target artifacts; generates target-native implementation projects for code; runs reviewed specialist adapters; and preserves every unresolved semantic gap as a source-backed contract instead of silently dropping it.

Release highlights:

- Java ↔ Bedrock content conversion with exact, translated, generated, tool-assisted, review and blocked states per graph node.
- Bedrock `.mcpack`, `.mcaddon`, `.mcworld`, `.mctemplate`, editable BP/RP/script projects and complete world products.
- Java data packs, resource packs, paired vanilla add-on families, worlds, Fabric/NeoForge/Forge/multi-loader projects and standalone world-template mod projects.
- Target-native Java feature matrices for Bedrock blocks, items, entities, scripts, cameras, dialogue, trades, volumes, animation/rendering, worldgen, dimensions, UI and every other detected add-on surface.
- Version Atlas-backed pack formats, Java versions, mappings, loaders and build-tool scaffolds.
- Allowlisted Chunker, JE2BE Resource Pack Converter, Geyser PackConverter and Regolith execution with isolated workspaces, logs, hashes and source re-verification.
- Deterministic proof bundles, source/tree hashes, structural package validation and no silent feature loss.

Read [the 0.11.0 release notes](RELEASE-NOTES-0.11.0.md), [architecture](OMNIBRIDGE-ARCHITECTURE.md), [capability matrix](CONVERSION-CAPABILITY-MATRIX.md), and [adapter catalog](OMNIBRIDGE-TOOL-ADAPTERS.md).

## Minecraft Mod Vault 0.10.0 — OmniManager

The universal Java, Bedrock and server mod manager is now a first-class part of Minecraft Mod Vault. It resolves installed content across storefront boundaries, recovers embedded names and art when catalogs fail, compares exact installed/provider file identities, protects custom builds, manages Bedrock packages and worlds natively, and makes every mutation recoverable.

Read [the 0.10.0 release notes](RELEASE-NOTES-0.10.0.md), [identity contract](PROVIDER-IDENTITY-CONTRACT.md), [Bedrock contract](BEDROCK-MANAGEMENT-CONTRACT.md), and [architecture](OMNIMANAGER-ARCHITECTURE.md).


Minecraft Mod Vault is a compiled desktop Minecraft discovery, installation, management, update, recommendation, creator-analysis, Java, Bedrock, CIT, furniture, visual-pack, plugin, map, and modpack application.

Version 0.9.0 keeps every federated browser, updater, installer, Creator Archive, and Mod Doctor capability from 0.8.0, then unifies three major foundations: **Porting Lab** for installed-JAR forensics and isolated migration workspaces, **Repair Lab** for immutable source-project intake and controlled migration/build/rollback, and an offline **SQLite Compatibility Brain** that searches official version/toolchain seeds, current repair research, and durable repair history. The **Mods** workspace remains a federated browser instead of a collection of provider hyperlinks. Search **all enabled sources or any exact subset** from one query box, filter by Minecraft version, loader/platform, content type, category, source, and sort mode, then inspect rich project details without leaving the Vault.

## Universal Mods browser

The browser currently has **28 integrated source lanes**. Every listed lane participates through an in-app search/details implementation rather than a giant external-link button:

- [Modrinth](https://modrinth.com/) and [CurseForge](https://www.curseforge.com/minecraft) for broad Java/Bedrock projects, packs, plugins, data packs and visuals.
- [GitHub](https://github.com/), [Smithed](https://smithed.dev/), [Planet Minecraft](https://www.planetminecraft.com/), [MCPEDL](https://mcpedl.com/), and [Minecraft Marketplace](https://www.minecraft.net/marketplace) for source releases, community content, Bedrock content and official Marketplace discovery.
- [Hangar](https://hangar.papermc.io/), [SpigotMC](https://www.spigotmc.org/resources/), [BukkitDev](https://dev.bukkit.org/bukkit-plugins), [Sponge Ore](https://ore.spongepowered.org/), [BuiltByBit](https://builtbybit.com/resources/), and [Polymart](https://polymart.org/resources) for server/plugin ecosystems.
- [ATLauncher](https://atlauncher.com/packs/all), [Technic](https://www.technicpack.net/modpacks), and [Feed The Beast](https://www.feed-the-beast.com/modpacks) for modpacks.
- [Minecraft Maps](https://www.minecraftmaps.com/), [ResourcePack.net](https://resourcepack.net/), [Texture-Packs.com](https://texture-packs.com/), [Vanilla Tweaks](https://vanillatweaks.net/), [ShaderPacks.com](https://shaderpacks.com/), [Shaderpacks.net](https://shaderpacks.net/), and [MinecraftShader.com](https://minecraftshader.com/) for maps, packs and visual content.
- [MCreator Community](https://mcreator.net/modifications), [The Skindex](https://www.minecraftskins.com/), [MinecraftHub](https://minecrafthub.io/resources), [Mod DB](https://www.moddb.com/games/minecraft/mods), and [Nexus Mods](https://www.nexusmods.com/minecraft) for specialist/community discovery.

Provider chips are real source selectors. Presets such as **Every site**, **Curated discovery**, **Java client**, **Data + visual packs**, **Shaders**, **Skins**, **Maps + worlds**, **Plugins**, **Modpacks**, **Bedrock**, **Community**, and **Marketplaces** simply select meaningful source sets. A source text filter and **Select all relevant** make large source sets manageable.

Search is federated with stable source windows. Providers are queried concurrently, source results are normalized, then cross-provider variants are merged before the requested result page is sliced. **Load More** expands the stable provider windows and appends only unique projects. The browser never virtualizes away off-screen cards.

Cards and details preserve project art, icons, creator avatars, galleries, downloads/usage, categories, Minecraft versions, loaders/platforms, update dates and source variants when providers expose them. Clicking a source variant runs that source inside the Vault rather than sending the user to a generic provider home page.

## Content types and taxonomy

The same browser handles Mods, Modpacks, Resource/Texture Packs, Shaders, Data Packs, Server Plugins, Bedrock Add-ons, Maps/Worlds, Skins and Tools. The taxonomy combines provider-native categories with Vault-specific discovery groups, including:

- railroads, trains, vehicles, ships and transport;
- technology, automation, redstone, computers and storage;
- magic, spellcraft, RPG, combat, mobs and bosses;
- furniture, CIT, building, architecture, farming, food and cute/cozy content;
- living foliage, waving grass/leaves, animation, particles, water splashes/wakes and visual effects;
- card games, collectibles, tameable pets, companions and creature-collecting/Pokémon-like gameplay;
- exploration, structures, world generation, horror, space, performance and quality-of-life.

Modrinth categories and current public CurseForge filters are refreshed at runtime and merged without duplicating equivalent permanent categories.

## Provider-aware installation

Installation is provider-specific rather than pretending every site behaves the same:

- Modrinth uses native version resolution, hashes and required-dependency installation.
- CurseForge uses its official API when configured and verified package detection for appropriate public/Bedrock lanes.
- GitHub selects real release assets and verifies published digest/size when available.
- Hangar, Spigot, BuiltByBit, Smithed and Sponge Ore use their provider-native installation paths where available.
- Planet Minecraft, MCPEDL, BukkitDev, Mod DB, Minecraft Maps, ResourcePack.net, Texture-Packs.com, MCreator and supported shader directories use bounded in-app download-page resolution, reject HTML/fake downloads, validate the final container, then install to the correct target.
- The Skindex saves only validated Minecraft skin PNG dimensions.
- MinecraftHub is a curated cross-provider index: Vault resolves its original source and delegates installation back through the corresponding verified provider integration instead of treating MinecraftHub as a download mirror.

World ZIP installation verifies a real `level.dat`, rejects traversal, extracts safely and creates a unique save destination. Bedrock `.mcpack`, `.mcaddon` and `.mcworld` files remain in the Bedrock handoff flow rather than being mixed into Java folders.

## Universal Updater

The updater scans actual installed JARs and records SHA-1, SHA-512, CurseForge-compatible MurmurHash2 fingerprints and embedded Fabric/Quilt/Forge/NeoForge metadata.

Identity resolution proceeds from strongest evidence to weakest:

1. exact Modrinth SHA-512 identity;
2. exact CurseForge fingerprint identity when an API key is configured;
3. developer-declared canonical Modrinth or CurseForge project URLs embedded in the JAR metadata;
4. developer-declared GitHub repository metadata with explicit release compatibility evidence;
5. multi-provider continuation/port/fork discovery for review.

Canonical Modrinth/CurseForge metadata is particularly useful for old or repacked JARs whose bytes no longer match a provider fingerprint. Target releases still have to match the requested Minecraft version and loader. Before an automatic replacement reaches the real mods directory, Vault downloads it to staging, verifies provider hashes/size where available, parses the downloaded JAR, checks mod identity and loader, stages required Modrinth dependencies, creates timestamped backups, then performs an atomic replacement. Any failure rolls the operation back. Disabled state is preserved.

Ambiguous ports, forks, substitutions and dependency-changing CurseForge candidates remain review-only.

## Mod Doctor: repair, compatibility, upgrade, and downgrade intelligence

Mod Doctor scans the real installed JARs and constructs an evidence-backed compatibility graph instead of guessing from filenames. It understands Fabric, Quilt, Forge, and NeoForge metadata; required, optional, incompatible, and conflict relationships; Java class-file levels; namespaces; Mixin/refmap/plugin contracts; access wideners and transformers; coremods and transformation services; nested JARs; signatures; native libraries; source provenance; and packaged data/assets.

The graph detects duplicate IDs, duplicate artifacts, exact and conflicting class collisions, missing required dependencies, declared conflicts, and required-dependency cycles. It also emits a deterministic dependency-first transaction order so staged upgrades and ports can be built and deployed in the correct sequence.

For upgrades, downgrades, and loader changes, Mod Doctor chooses among source rebuild, namespace remap, source transformation, narrow binary repair, compatibility adapter, runtime bridge, data/schema migration, or review-only substitution. A machine-readable 158-source catalog, 96 detailed execution records, and 167 unique runtime tool cards connect every recommendation to current official documentation, repositories, and tools. Runtime bridges such as Connector or Kilt remain controlled candidates and require real target-runtime verification.

Legacy migrations are routed through era-correct toolchains instead of the modern default. Mod Doctor distinguishes RetroMCP/RetroFuturaGradle, Cleanroom/Fugue/MixinBooter, UniMixins/LWJGL3ify, Legacy Fabric, Ornithe/Feather/Ploceus, StationAPI, and LegacyFix from modern NeoForm/NeoForm Runtime, Fabric Loom/Fabric Meta, ForgeGradle, and ModDevGradle. Historical source reconstruction, actual source/data porting, and optional runtime compatibility layers are separate verified tracks.

Paste a crash report, `latest.log`, `debug.log`, or build failure into the same workspace. The classifier prioritizes the earliest causal dependency, linkage, Mixin, registry/data, side, native, or resource failure and connects it to reusable Repair Brain patterns plus the exact tools needed to prove and repair it.

Exact byte-identical duplicate JARs are the first repair class the Vault will execute automatically. Candidates are re-hashed immediately before mutation, moved into a Vault quarantine rather than deleted, written to an atomic receipt, and restorable through a verified reverse transaction. Different bytes, ambiguous identity, symbolic links, drifted files, or anything outside managed Minecraft directories remain blocked or review-only.

See [`ULTIMATE-TOOLS-AND-PORTING-KNOWLEDGE.md`](ULTIMATE-TOOLS-AND-PORTING-KNOWLEDGE.md) for the complete catalog.


## Porting Lab: official version intelligence and isolated migration workspaces

Porting Lab turns “make this mod work on another version or loader” into a bounded engineering workflow instead of a metadata edit. Its embedded **Version Atlas** contains 907 official Minecraft manifests with immutable SHA-256 source evidence, Java requirements, client/server availability, mappings, protocol, world/data, resource-pack, and loader/build-tool coverage. It also carries exact reviewed Forge, NeoForge, Fabric, Quilt, Gradle-plugin, and mapping-tool coordinates where the upstream evidence provides them.

A plan records the source and target version/loader, upgrade or downgrade direction, source-versus-binary starting point, computed risk, exact pins, semantic boundaries, warnings, selected tools, eight ordered phases, a client/server verification matrix, and a definition of completion. Binary inputs from managed directories are analyzed live for SHA-256/SHA-512/CurseForge fingerprint identity, loader metadata, mod IDs, dependencies, Java class level, namespace clues, Mixins, access rules, coremods, transformation services, nested JARs, native libraries, signatures, assets/data, reflection/Unsafe/MethodHandles, Kotlin/Scala, side references, and risk signals.

Generating a workspace copies—not moves—the chosen JAR into an isolated, hash-recorded directory. The workspace includes the immutable plan, human-readable instructions, exact source/target coordinates, generated build scaffolding, evidence and decompilation areas, verification scripts, and a manifest with every produced file hash. The original input is re-identified before workspace creation, and a changed input path or changed digest aborts the operation.

The toolchain radar probes the local Java, compiler, Git, Gradle, Node, Python, archive, and platform environment. Tool recommendations deliberately combine official native loader toolchains with independent analyzers and cross-checks such as InterMed, modcrawl, Vineflower, CFR, Tiny Remapper, japicmp, Retromod, RetroFuturaGradle, Unimined, Modstitch, Stonecutter, packwiz, and Ferium. Experimental tools are labeled and gated; no transformed JAR is called “ported” until the complete runtime matrix passes.

## Repair Lab: secure source migration, controlled builds, and proof bundles

Repair Lab is the source-project counterpart to Porting Lab. Importing a ZIP creates an immutable, SHA-256-identified original and a disposable working tree. Intake rejects path traversal, symbolic links, duplicate archive paths, entry-count and expanded-size abuse, suspicious compression ratios, and files outside the bounded extraction root. The original archive and extracted source tree are re-verified after every controlled run.

The profiler detects the project root, loader, Minecraft version, mappings, Java target, build system, wrappers, metadata files, and pack formats. A requested upgrade, downgrade, or loader target resolves through the embedded Version Atlas and Compatibility Brain. Automatic edits are deliberately conservative: recognized Gradle properties, plugin/tool coordinates, loader metadata, Java levels, mappings, Minecraft constraints, and data/resource-pack formats are staged as a reviewable per-file change list. Semantic API rewrites remain explicit source-repair work rather than being disguised as a successful metadata port.

Build, test, and clean actions are restricted to detected project wrappers and fixed commands. Execution requires the exact acknowledgement that build scripts are code, runs with a sanitized environment and dedicated caches, records complete logs, supports timeout/cancellation, hashes every output artifact, and rechecks the immutable source. A completed session can export a deterministic prepared-source ZIP plus a proof bundle containing the plan, edits, logs, hashes, receipts, and artifacts. One-click rollback restores the working copy from the immutable source while preserving the evidence trail. This is application-level containment, not a claim of operating-system or virtual-machine sandboxing.

## Offline Compatibility Brain

The first run creates a local pure-Go SQLite database in WAL mode from reviewed embedded seeds. The current corpus contains 919 Minecraft version records from the earliest pre-classic identifiers through the newest captured snapshot, 949 loader/game relationships, 550 loader releases, more than 15,000 toolchain releases, 342 searchable knowledge documents, and the durable Repair Brain history. Full-text search connects crash signatures, loader/build eras, mapping systems, source converters, decompilers, bytecode tools, compatibility bridges, launchers, world/NBT repair tools, publishing systems, and primary documentation.

The database is local-first and rebuildable from provenance-recorded embedded sources. Each seed and generated atlas carries source identity, retrieval date, record counts, and hashes. Remembered repair records are evidence to search, not commands to apply blindly: the exact artifact, source, version, loader, Java runtime, dependencies, and current upstream behavior are re-evaluated for every repair.

## Living recommendations

Primary recommendations are not frozen into the release. While the application is running it repeatedly refreshes live provider searches, popular/recent feeds, rotating taxonomy categories, configured interests, Minecraft-version/loader compatibility and analyzed creator signals. The last useful ranked cache is persisted so startup can show results immediately and refresh stale data in the background.

The default recommendation refresh is 15 minutes and creator discovery refresh is hourly. Provider calls have bounded retries, timeouts, fresh caching and stale-cache fallback so one source outage does not blank the whole browser.

## Creator Archive and Creator Picks

A fresh install follows 13 high-signal Minecraft channels automatically: AsianHalfSquat, EnderVerseMC, Noxus, ChosenArchitect, direwolf20, Gaming On Caffeine, SystemCollapse, Lashmak, PwrDown, Mischief of Mice, PopularMMOs, DanTDM, and The Breakdown. Existing installations are migrated in place without losing pause state or indexed history. Automatic channel enumeration shares the configured archive concurrency budget so first-run backfills drain steadily instead of launching every channel crawl at once.

You can still paste any additional YouTube @handle, channel URL, legacy /user/ or /c/ URL, or UC channel ID. Custom creators use the same full-history, Shorts, streams, transcript, evidence, project-resolution, retry, and sorting pipeline.

Creator Picks turns Minecraft recommendation videos into browseable mod lists. Completed analyses preserve video thumbnail/creator identity, timestamps, deep-links, resolved projects, provider/source, project icon/author metadata, evidence and confidence.

Analysis tries, in order:

1. direct links and timestamp lists in the video description;
2. public YouTube/TikTok captions with timecodes;
3. local speech-to-text for captionless videos using current [yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://github.com/BtbN/FFmpeg-Builds), and [whisper.cpp](https://github.com/ggml-org/whisper.cpp) artifacts prepared on demand.

Long recommendation videos are not intentionally truncated to a tiny fixed mod count. Unresolved mentions remain visible in **Needs review** with a one-click search back into the Mods browser.

## CIT, Furniture and Bedrock

CIT Packs, Furniture and Bedrock remain first-class workspaces. They preserve the curated cozy/cute/furniture collections already built for the Vault while their live-search actions feed into the same current federated browser. Bedrock browsing includes add-ons, scripts, maps/worlds, texture packs and skins where providers expose those lanes.

## Desktop architecture

The compiled executable embeds the frontend and starts a loopback-only HTTP backend on `127.0.0.1` using a random port and random per-run authorization token. The native backend owns filesystem access, provider requests, installs, update transactions, persistent caches and creator analysis.

The app itself does not require Node.js or Electron.

## Build and verification

```sh
gofmt -w *.go
go test -mod=vendor ./...
go vet -mod=vendor ./...
node --check web/app.js
node --check web/catalog.js
node --check web/repair-lab.js
git diff --check
./scripts/build.sh
```

Windows PowerShell can use `scripts/build.ps1`.

Release verification also launches the final Linux executable from a freshly extracted release package and exercises the real authenticated loopback backend. See `Minecraft-Mod-Vault-0.9.0-BUILD-VERIFICATION.txt` in the packaged release for the exact pass.

## Release

Version: **0.9.0**

See `RELEASE-NOTES-0.9.0.md`, `ULTIMATE-TOOLS-AND-PORTING-KNOWLEDGE.md`, `RESEARCH-SOURCES.md`, `THIRD-PARTY-NOTICES.md`, the build verification report, and the SHA-256 manifest shipped with the release.

### Exhaustive creator recommendation archive

Creator Archive ships with 20 protected creator archives spanning YouTube and TikTok. The release defaults and hot-drop creator catalogs currently include AsianHalfSquat, EnderVerseMC, Kizamiringo, Katsumi (`@its_katsumi`), SpeedyChunks (`@speedychunks`), NoxusMinecraft (`@noxusminecraft`), UnyxYT (`@unyxyt`), CurseForge, HendyVideos, Noxus, ChosenArchitect, direwolf20, Gaming On Caffeine, SystemCollapse, Lashmak, PwrDown, Mischief of Mice, PopularMMOs, DanTDM, and The Breakdown. They are followed automatically on both fresh installs and upgraded installations. Any built-in creator can pause automatic refresh without losing data. Paste an additional YouTube or TikTok creator handle/profile URL (plus legacy YouTube `/user/`, `/c/`, or `UC...` channel ID forms) and the Vault immediately saves the custom watch, starts a resumable full-history backfill, includes uploads/Shorts/TikToks/streams as applicable, and continues checking incrementally for new posts. User-added creators can be unfollowed while preserving everything already indexed.

For every followed creator, the Vault keeps a resumable local corpus of video metadata, descriptions, timed transcripts, resolved Minecraft projects, evidence, confidence, timestamps, and unresolved mentions. Browse recommendations newest-to-oldest or oldest-to-newest, by channel, video, mod name, confidence, or video mod count, and search across the archive. Public/automatic captions are preferred; yt-dlp caption recovery and local whisper.cpp Large v3 Turbo speech-to-text provide layered fallback paths.
