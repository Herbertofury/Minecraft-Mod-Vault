# Minecraft Mod Vault 0.12.0 P0.5 Creator Catalog hot-drop ingestion handoff

**FIRST NEXT TASK:** Port this isolated P0.5 delta into the newest concurrent Minecraft Mod Vault source. Do not replace newer files wholesale.

P0.5 makes creator databases real hot-droppable application data. Schema-v1 JSON bundles from the release and `<config>/creator-catalogs/` merge into the existing Creator Channels/Videos/Recommendations state, reload while the app runs, preserve stronger live/provider evidence, and retain last-known-good data when a replacement is malformed.

The first real data pack is AsianHalfSquat: channel target 349 videos, 11 exact seeded videos, 93 evidence-backed recommendations, explicitly incomplete until a network-enabled full-history sync fills the rest. SpeedyChunks, NoxusMinecraft, and UnyxYT are protected TikTok follows; NoxusMinecraft/UnyxYT are supplied through the catalog path to prove ordinary creator additions need no creator-specific Go code.

Runtime proof: fresh P0.5 Linux binary exposed 20 protected creators and AsianHalfSquat 349/11/93; a fifth local catalog was hot-dropped and became visible without restart; a malformed replacement returned HTTP 207 without deleting last-known-good data; restart preserved the hot-dropped record.

**Truth boundary:** this container cannot DNS-resolve YouTube/TikTok, so it does not claim all 349 AsianHalfSquat videos have been live-crawled here. The catalog records only evidence actually obtained, and the normal full-history sync continues the same database when network access is available.

See `CREATOR-CATALOGS.md` and `CHECKPOINT-0.12.0-06-CREATOR-CATALOG-HOTDROP-INGESTION.md`.

---

# Minecraft Mod Vault 0.12.0 P0.3 Creator Modpack Library integration handoff

**FIRST NEXT TASK:** Port this isolated P0.3 delta into the newest concurrent Minecraft Mod Vault build without replacing newer files wholesale. It extends P0.2 Creator Link-Hub Intelligence.

P0.3 adds an evidence-backed multi-pack Creator Modpack Library. Followed creators can have zero, one, or many verified CurseForge/Modrinth packs. Provider profiles discovered from profiles/bios/link hubs/upload descriptions enumerate all current modpacks; direct `My Modpack` links can safely teach a differently named provider identity only when provider data proves a unique owner/author. Owner/member/profile/direct relationships are preserved instead of flattened, failed provider scans retain last-known-good packs, and no evidence produces an explicit empty state rather than a guess.

AsianHalfSquat and EnderVerse now seed their verified CurseForge creator profiles, not fixed pack lists. Current web evidence on 2026-08-23 shows AsianHalfSquat with 11 CurseForge projects containing both Owner and non-owner/member-associated entries, and EnderVerse with 7 projects whose visible entries are owner modpacks.

Implementation commit: `9dd0d90ff04042d0128b5a27a1accd2b2e597cb2`. Full Go tests/vet + JS syntax/diff checks pass under the documented host-toolchain compatibility procedure. Fresh Linux/Windows checkpoint builds were produced. The Linux executable freshly extracted from the deliverable tarball, byte-identical to the raw build, enumerated 4/4 AsianHalfSquat and 3/3 EnderVerse provider-profile fixture packs, then restored both exact libraries after restart with all 17 built-ins intact.

**Remaining live gates:** network-enabled real provider refresh/comparison; a real Linktree/Lnk.Bio direct-pack path in the packaged app; production-browser click/screenshot verification; merge against newest concurrent source; official Go 1.27 plus existing TikTok/OCR live gates.

See `CREATOR-MODPACK-LIBRARY.md`, `CHECKPOINT-0.12.0-04-CREATOR-MODPACK-LIBRARY.md`, and `P0-TIKTOK-CREATOR-VAULT-TODO.md`.

---

# Minecraft Mod Vault 0.12.0 P0.2 Creator Link-Hub integration handoff

**FIRST NEXT TASK:** Port this isolated P0.2 delta into the newest concurrent Minecraft Mod Vault build without replacing newer files wholesale. It extends the existing TikTok Creator Vault checkpoint.

P0.2 adds creator-controlled link intelligence: Katsumi's verified Lnk.Bio + Linktree hubs are seeded; followed YouTube/TikTok profiles automatically discover supported link-in-bio hubs from profile HTML, provider bio text, and recent archived upload descriptions; nested hubs are bounded; outbound modpacks/mods/packs/socials/downloads/support/wishlists are typed and stored with evidence plus first/last-seen timestamps; platform/hub redirect wrappers resolve target metadata without crawling arbitrary destinations; successful refreshes remove stale auto-discovered links; blocked refreshes keep last-known-good cache.

The Creator Archive UI now has a functional Creator Links panel with modpacks first, exact-destination controls, live/seeded/cached status, provenance tooltips, and a lightweight **Refresh links** action that does not force a video rescan. Automatic link refresh also rides the normal creator sync and is attempted independently when video enumeration fails.

Katsumi research snapshot (2026-08-23): `https://lnk.bio/itskatsumii` and `https://linktr.ee/Itskatsumii` are current indexed hubs. The Linktree index exposes Discord, Tips/Donos, Wishlist, Twitch, YouTube and Instagram. The public search index did not expose a trustworthy exact current modpack child URL, so this checkpoint deliberately does not fabricate one; the live crawler is designed to record the current child destination from the creator-controlled hub.

Local verification: full Go tests and `go vet` pass under the documented host-Go compatibility procedure; JS syntax and diff whitespace checks pass; fresh Linux/Windows checkpoint binaries build; a fresh Linux runtime loaded Katsumi with both seeded hubs; a real blocked-network Refresh links call retained the last-known-good cache; that state survived restart. Chromium production click/screenshot proof is externally blocked by an organization policy that blocks loopback/file navigation. See `CREATOR-LINK-HUB-INTELLIGENCE.md` and `CHECKPOINT-0.12.0-03-CREATOR-LINK-HUB-INTELLIGENCE.md`. Live hub-child discovery, merged-source Go 1.27 verification, packaging and remote publication remain.

---

# Minecraft Mod Vault 0.12.0 P0 Creator Vault integration handoff

**FIRST NEXT TASK:** Integrate the isolated TikTok Creator Vault parity checkpoint into the newest concurrent Minecraft Mod Vault build. Do not overwrite newer work from the active build chat.

The checkpoint adds persistent TikTok follows, platform-scoped identities, full-first/incremental-after automatic refresh, retry backoff, Kizamiringo/Katsumi/CurseForge/HendyVideos seeds, visual OCR for text-only recommendations, speech+visual evidence merging, provider-verified mod resolution, and cross-platform Creator Archive UX.

Local verification already covers the full Go test suite, `go vet`, JavaScript syntax, a real Tesseract OCR fixture, Linux/Windows compatibility builds, and a packaged Linux restart/persistence flow for a custom TikTok subscription. Remaining P0 gates are live TikTok profile/video verification on a networked environment, an actual Kizamiringo text-only completeness comparison, Windows.Media.Ocr runtime proof, then merge/retest against the other chat's newest source with the official Go 1.27 toolchain.

See `P0-TIKTOK-CREATOR-VAULT-TODO.md`, `TIKTOK-CREATOR-SOURCE-CATALOG.md`, and the 0.12.0 TikTok checkpoint verification receipt.

**2026-08-23 Katsumi delta:** `https://www.tiktok.com/@its_katsumi` is now a required `curated-core` built-in TikTok follow. Fresh-install and upgrade migration tests expect 17 protected creators, and a fresh packaged runtime exposed `@its_katsumi` as active/required after startup.
**2026-08-23 SpeedyChunks delta:** `https://www.tiktok.com/@speedychunks?lang=en` is now normalized to and seeded as the required `curated-core` TikTok creator `@speedychunks`. Fresh-install/upgrade expectations move from 17 to 18 protected creators. No unverified Linktree/CurseForge/Modrinth identity is seeded; Creator Link-Hub Intelligence and Creator Modpack Library discover those from live public evidence and preserve an explicit empty state when nothing trustworthy is found.


---

# Minecraft Mod Vault 0.9.0 verified release handoff

Recorded: 2026-08-21T04:35:00Z
Verified implementation commit: `4191fad`
State: **verified release ready for final publication**

## What shipped

Minecraft Mod Vault 0.9.0 is one preserved product with two truthful repair workflows and one shared evidence brain:

- **Porting Lab** handles installed managed JARs: live hashes/metadata/class/Mixin/access-rule/runtime-risk forensics, the official Version Atlas, exact upgrade/downgrade/loader plans, eight gated phases, isolated hash-locked workspaces, toolchain probes, and cryptographically reversible exact-duplicate quarantine.
- **Repair Lab** handles source-project ZIPs: hostile-safe immutable intake, source-tree fingerprinting, project/loader/version/mapping/Java/build detection, conservative recognized-field migrations, explicit build-code acknowledgement, wrapper-only fixed actions, dedicated caches, timeouts/cancel, logs, artifact hashes, deterministic prepared-source/proof exports, and rollback.
- **Compatibility Brain** is an embedded, rebuildable pure-Go SQLite evidence store with WAL/FTS5, reviewed Mojang/mcmeta/Fabric/Quilt/Forge/NeoForge/Modrinth/toolchain seeds, the current research catalog, repair patterns, and eight durable Repair Brain records.

## Release proof

- Final Linux binary SHA-256: `0e6720b4b1b9a6bb31f90fe2d498a7f4fbfac43df264d7e87bd23062d89ed2ee`
- Final Windows binary SHA-256: `6b5967461a34cb8658d93644f9d60196bb3bdbd162519f9dda1aff3ef4fa9598`
- Repair Lab: 25 UI assertions, 38 authenticated API calls, zero browser diagnostics.
- Porting Lab: 29 UI assertions, 23 authenticated API calls, zero browser diagnostics.
- Fresh database: SQLite 3.53.3/WAL, 919 Minecraft versions, 15,133 toolchain releases, 342 knowledge documents, 8 repair records.
- Fresh package: all internal manifests passed; extracted source rebuilt both binaries byte-identically; extracted Linux runtime reran both full UI workflows; session/database state survived process restart.

Full evidence is in `Minecraft-Mod-Vault-0.9.0-BUILD-VERIFICATION.txt` and `verification/final-package-verification.json`.

## Important boundaries

Build scripts are project code. The Vault requires explicit acknowledgement and applies bounded commands, sanitized environment, dedicated caches, logs, timeout/cancel, and immutable-source checks, but does not claim OS/VM sandboxing. Known-field edits are migration preparation, not proof of semantic compatibility. A final mod port still requires exact dependency resolution, compilation, client/dedicated-server exercise where supported, log inspection, persistence, artifact hashing, and rollback.

## Remaining publication-only work

Generate the final deterministic archives containing this handoff/evidence, publish all final artifacts to the canonical Drive folder with full redownload/hash verification, publish the canonical GitHub source/release if supported, and record provider IDs/hashes in a separate publication receipt. Do not rebuild or alter the verified binary payload merely to insert its own outer archive hash into itself.
