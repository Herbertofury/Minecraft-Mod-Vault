# Dimensional Lifeguard 0.1.0-safe-preview.1 - Verification Receipt

Date: 2026-09-04 (America/Denver)  
Target: Minecraft 1.20.1, Forge 47.4.20+, Java 17  
Native QA runtime: Forge 47.4.23, Eclipse Temurin 17.0.20.1  
Mod ID: `dimensionallifeguard`

## Release artifacts

- `Dimensional-Lifeguard-0.1.0-safe-preview.1-Forge-1.20.1.jar`
  - size: 26225 bytes
  - SHA-256: `43149d5e4277407d654e679cfa25b9a9d64739683afeb698c69f24af2e38ca95`
- `Dimensional-Lifeguard-0.1.0-safe-preview.1-source.zip`
  - size: 78637 bytes
  - SHA-256: `8f31f57279d157b563d9b4f8cb23b0adbf6e00f788fe8d1eeb8c40825b7ecd83`

## Acceptance intent

Dimensional Lifeguard keeps modded dimensions registered while deferring compatible `ServerLevel` construction until a real server access requires that dimension. Safety outranks maximum dormancy: vanilla dimensions, configured eager dimensions/namespaces, persisted forced-chunk/ticket dimensions, and compatibility-recovery dimensions load eagerly.

The first preview intentionally does **not** live-unload a world after it has been woken. That avoids stale `ServerLevel` references in large mods while still eliminating the startup/save flood from never-used dimensions.

## Production build proof

- ForgeGradle 6.0.54 production build: PASS.
- Java compilation: PASS.
- Mixin annotation processing and production SRG refmap: PASS.
- `reobfJar`: PASS.
- JAR ZIP integrity: PASS.
- Required release entries present: `META-INF/mods.toml`, mod classes, mixin config, production refmap.
- Manifest advertises `MixinConfigs: dimensionallifeguard.mixins.json`.
- `mods.toml` requires Minecraft 1.20.1 and Forge `[47.4.20,)`.

## Clean-source reproducibility

The canonical source ZIP was extracted to a fresh directory and rebuilt independently with the verified offline Forge toolchain.

- fresh build: PASS
- fresh JAR SHA-256: `43149d5e4277407d654e679cfa25b9a9d64739683afeb698c69f24af2e38ca95`
- canonical release JAR SHA-256: `43149d5e4277407d654e679cfa25b9a9d64739683afeb698c69f24af2e38ca95`
- byte-for-byte equality: PASS

## Native lifecycle QA - 24 synthetic mod dimensions

Harness: 24 datapack dimensions `dltest:dim_01` through `dltest:dim_24`, plus the three vanilla dimensions.

### Startup laziness

Observed on Forge 47.4.23:

- registered LevelStems: 27
- eager/live at startup: 3
- dormant at startup: 24
- server reached ready state successfully

### Wake-on-access

Vanilla `execute in dltest:dim_03` exercised the normal `MinecraftServer#getLevel` access path.

Observed in the production-reobfuscated release JAR on the real `forgeserver` launch target:

- production server ready: `Done (2.790s)!`
- FastWake `dltest:dim_03`: **3 ms**
- reason: `API_ACCESS`
- production caller: `net.minecraft.commands.arguments.DimensionArgument#m_88808_`
- after wake: registered=27, live=4, dormant=23
- shutdown save set: Overworld, Nether, End, and `dltest:dim_03` only

A separate production-JAR baseline boot reached `Done (2.251s)!`, reported registered=27/live=3/dormant=24, and saved only the three vanilla dimensions on shutdown.

Earlier named-userdev lifecycle runs measured synthetic FastWake construction at 1-4 ms across repeated dimensions. These numbers measure Lifeguard's world-construction path in this synthetic harness; real destination chunk generation and mod-specific load hooks can add latency.

## Persisted forced-chunk safety

Test sequence:

1. force-load a chunk in `dltest:dim_02`;
2. stop and restart;
3. verify Lifeguard detects persisted forced data;
4. remove force-load;
5. stop and restart again.

Results:

- with persisted force-load: `dim_02` auto-promoted eager; startup became 4 live / 23 dormant - PASS
- after removal and restart: returned to 3 live / 24 dormant - PASS

This protects remote machinery and Forge/vanilla chunk-ticket persistence from being silently put to sleep.

## Transactional wake rollback and compatibility recovery

Late activation is transactional. If construction or a load hook throws after partial creation, Lifeguard removes the partial world from Forge's live map, marks the world snapshot dirty, closes the partial level, keeps the registered LevelStem dormant, and records the dimension for eager compatibility recovery next restart.

Recovery-ledger runtime test:

- seeded prior-failure entry `dltest:dim_04`
- next startup auto-promoted `dim_04` eager - PASS
- `/dimensionallifeguard recovery list` exposed it - PASS
- `recovery clear dltest:dim_04` removed the override - PASS

## Wake flood diagnostics

Eight API wake requests within the configured five-second window were exercised. Lifeguard emitted its dimension-flood warning with the external caller and retained per-dimension wake timings.

Observed burst timings: 4, 3, 2, 2, 1, 2, 1, 1 ms - PASS.

This is specifically intended to identify integrations that eagerly enumerate/get every world and defeat lazy startup.

## Operator surface verified

- `/dimensionallifeguard status`
- `/dimensionallifeguard list`
- `/dimensionallifeguard report`
- `/dimensionallifeguard wake namespace:dimension`
- `/dimensionallifeguard recovery list`
- `/dimensionallifeguard recovery clear namespace:dimension`
- `/dimensionallifeguard recovery clear-all`

The namespaced command parser uses Minecraft's resource-location argument type, including IDs containing `:`, after a QA-discovered parser defect was corrected.

## Safety design frozen for Preview 1

- all `minecraft:*` dimensions are hard-eager
- static dimension registry entries remain registered
- persisted vanilla/Forge forced chunks fail open to eager load
- off-thread wake is marshalled to the server thread
- Forge live-world map is updated through `forgeGetWorldMap()` and `markWorldsDirty()`
- Forge level-load lifecycle is replayed on late activation
- FastWake starts with reduced view/simulation distance and ramps toward server targets
- no automatic live unload in Preview 1
- `startup.enabled=false` restores vanilla eager startup behavior without deleting world data

## Remaining real-pack gate

This preview has strong synthetic and production-Forge runtime proof, but it has **not yet been declared fully compatible with the user's 600+ mod Noxviola's Dream instance**. That pack is the next acceptance environment.

The user's current pack uses Forge 47.4.20. The release metadata deliberately supports 47.4.20+, but native runtime QA in this session used 47.4.23. The exact 47.4.20 Forge installer could not be fetched inside the code runner because external DNS is unavailable there, so no claim of an exact 47.4.20 boot is made here.

First Noxviola gate:

1. back up the world;
2. install only this JAR with default Lifeguard settings;
3. boot the same benchmark save;
4. collect `latest.log` and `launcher_log.txt`;
5. compare live/dormant dimensions, server startup/world-entry time, save enumeration, wake callers, wake timings, and any compatibility-recovery entries;
6. add narrow BCLib/WorldsTogether, Valkyrien Skies, Ad Astra/Planets+, or other adapters only when the pack logs prove they are needed.

No claim of literal zero-lag dimension entry is made: the synthetic Lifeguard construction path is already in the low-single-digit millisecond range, but real worldgen, chunk IO, structures, entities, and mod-specific level-load hooks can dominate first entry. The FastWake ramp is designed to hide/spread that cost without skipping correctness work.
