# Entity Render Guard 1.0.0

Client-side Forge 1.20.1 compatibility shield for Accelerated Rendering.

## Purpose

Noxviola's Dream crashed while rendering Ars Nouveau's `ars_nouveau:whirlisprig` through Accelerated Rendering. Entity Render Guard keeps Accelerated Rendering enabled globally but routes only matched entity types through its vanilla entity/item/text pipelines for that entity's render scope, then immediately restores acceleration.

## Features

- Ships with `ars_nouveau:whirlisprig` protected.
- Supports exact IDs, globs such as `some_mod:*`, and `regex:` patterns.
- Can learn future qualifying Accelerated Rendering entity crashes from recent client crash reports once at startup.
- No watchdog, polling loop, timer, file watcher, entity deletion, world edits, or global performance disable.
- First launch creates `config/entity-render-guard.cfg`.

## Target

- Minecraft 1.20.1
- Forge 47.4.20+
- Java 17+
- Designed against Accelerated Rendering 1.0.14-1.20.1-alpha

## Release artifacts

Canonical release JAR (Google Drive): https://drive.google.com/file/d/1IvBdvl0b9cG_Z_tA4Etdxo-zKk6SFFos/view

Verified source archive (Google Drive): https://drive.google.com/file/d/1FgqMy2OCr0vbB45gYXhHLw1bMlglI8ei/view

Native QA log (Google Drive): https://drive.google.com/file/d/1A3AfocEGEOhyrE3G9Exhk3q8iOgNrZVi/view

Verification receipt (Google Drive): https://drive.google.com/file/d/1jctRZ_24SXCVv6W3va07b5samS_Llfmj/view

Release SHA-256 file (Google Drive): https://drive.google.com/file/d/1Yssv-M-LXkKjKpmxU1bRqGjOR8R75Vx5/view

See `VERIFICATION.md` for the native production-JAR proof.