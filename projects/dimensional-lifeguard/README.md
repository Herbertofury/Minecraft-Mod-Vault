# Dimensional Lifeguard

**Lazy worlds. Instant portals.**

Safety-first Forge 1.20.1 dimension lifecycle optimizer for very large modpacks. Compatible registered dimensions stay dormant until a real server access needs them; vanilla dimensions, persisted forced-chunk/ticket worlds, configured eager worlds, and compatibility-recovery worlds remain eager.

## Safe Preview 1

Version: `0.1.0-safe-preview.1`

Target: Minecraft 1.20.1, Forge 47.4.20+, Java 17.

Production runtime QA was completed on Forge 47.4.23 with Eclipse Temurin 17.0.20.1. The user's real Noxviola's Dream Forge 47.4.20 pack is the next integration gate.

### Verified production behavior

- 27 registered LevelStems -> 3 live / 24 dormant at startup.
- Real production `forgeserver` boot reached `Done (2.251s)!` in the synthetic baseline.
- Production `execute in dltest:dim_03` wake created the dormant world in 3 ms.
- After that wake: 4 live / 23 dormant.
- Shutdown saved only Overworld, Nether, End, and the one woken test dimension.
- Persisted vanilla/Forge forced chunks promote their dimension to eager on restart.
- Failed late activation is transactional and writes a per-dimension eager recovery override for the next restart.
- Eight API wakes within five seconds trigger the dimension-flood detector and caller attribution.
- No automatic live unloading is enabled in Preview 1; once woken, a dimension remains resident for the session to avoid stale `ServerLevel` references in large mod stacks.

## Exact verified artifacts

JAR SHA-256:
`43149d5e4277407d654e679cfa25b9a9d64739683afeb698c69f24af2e38ca95`

Source ZIP SHA-256:
`8f31f57279d157b563d9b4f8cb23b0adbf6e00f788fe8d1eeb8c40825b7ecd83`

Verification receipt SHA-256:
`3b715f53aa73187888511c9c498054561e48979b418a5274faf416d213bf2dfb`

### Google Drive project artifacts

- Release JAR: https://drive.google.com/file/d/1b2zuAaX-C9IvgBccAefpsmll1HWpLL6J/view
- Reproducible source ZIP: https://drive.google.com/file/d/1BojKtNc2aDVDDeT-UzeErPGdSi9SJmty/view
- Verification receipt: https://drive.google.com/file/d/1aQivPlmSRdY-HBZCZqoB1QpulVC4Iuvq/view
- SHA-256 manifest: https://drive.google.com/file/d/1XjLE9jCUiGqhPyLMT0g9mRrYQbuKPx4o/view
- Production baseline log: https://drive.google.com/file/d/1EzUpZPKWQyqakIhqxrL5Vducr4DAYidd/view
- Production wake log: https://drive.google.com/file/d/1V7zAIDF3QXnQj8SG6YzqZlLJ__QspSxL/view

The source ZIP was extracted into a clean directory and rebuilt; the rebuilt JAR was byte-for-byte identical to the release JAR.

## First Noxviola gate

1. Back up the world.
2. Install only the verified JAR with default Lifeguard settings.
3. Boot the same benchmark save.
4. Run `/dimensionallifeguard status`.
5. Travel through representative mod dimensions normally.
6. Run `/dimensionallifeguard report`.
7. Preserve `latest.log` and `launcher_log.txt` for the compatibility/performance pass.

Do not eager-whitelist whole mod namespaces unless real pack evidence requires it. Use wake-caller diagnostics to build narrow BCLib/WorldsTogether, Valkyrien Skies, Ad Astra/Planets+, or other compatibility adapters instead.
