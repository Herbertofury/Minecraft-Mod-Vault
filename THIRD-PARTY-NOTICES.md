# Third-Party Notices

Minecraft Mod Vault contains an independently implemented CurseForge fingerprint routine whose behavior was verified against the MIT-licensed `packwiz/packwiz` implementation in `curseforge/murmur2/hash.go`.

## packwiz

Copyright (c) 2019 packwiz contributors

MIT License

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

Project: https://github.com/packwiz/packwiz

## Runtime creator-transcription toolchain

Minecraft Mod Vault can optionally download current release artifacts for these tools on demand when a captionless Creator Picks video needs local transcription. They are not embedded in the application binary.

- yt-dlp: https://github.com/yt-dlp/yt-dlp
- whisper.cpp: https://github.com/ggml-org/whisper.cpp
- FFmpeg Windows builds: https://github.com/BtbN/FFmpeg-Builds
- Whisper model files: https://huggingface.co/ggerganov/whisper.cpp

The runtime bootstrap verifies release-asset size and SHA-256 when GitHub exposes a digest, and always verifies the configured Whisper model hash.

## Research and optional execution catalog

The Mod Doctor catalogs official documentation, repositories, APIs, build plugins, mapping projects, decompilers, runtime bridges, compatibility layers, and repair tools by URL. Those referenced projects are not copied into or embedded as executable third-party code merely because they appear in the catalog. A future optional executor must download an exact version into an isolated workspace, retain its source/license/provenance, verify hashes, and comply with the project-specific integration contract before use.

PaperMC DataConverter is cataloged only as data-conversion architecture and test-corpus research for modded worlds. The Vault does not dispatch its Fabric mod against modded saves because it does not execute data fixers registered by other mods.


## 0.9.0 Porting Lab research references

The following projects are referenced as optional research, analysis, build, or verification tools. Their code and binaries are not embedded in Minecraft Mod Vault 0.9.0 merely because they appear in the catalog:

- InterMed — https://github.com/jarettr/intermed
- modcrawl — https://github.com/SirCesarium/modcrawl
- Modpack Inspector — https://github.com/Rearth/Modpack-Inspector
- ModLens MCP — https://github.com/CreeperHost/modlens-mcp
- ModpackResolver — https://github.com/iTrooz/ModpackResolver
- Retromod — https://github.com/Bownlux/Retromod

Any future executor that downloads one of these tools must pin an exact release or revision, preserve its license and provenance, verify the artifact, isolate its workspace, and retain the result as evidence. Experimental or alpha output is never silently promoted to a verified repair.

## 0.9.0 statically linked Compatibility Brain dependencies

Minecraft Mod Vault 0.9.0 statically links the following vendored Go modules for the local pure-Go SQLite Compatibility Brain:

- modernc.org/sqlite v1.57.0
- modernc.org/libc v1.74.4
- modernc.org/mathutil v1.7.1
- modernc.org/memory v1.11.0
- github.com/dustin/go-humanize v1.0.1
- github.com/google/uuid v1.6.0
- github.com/mattn/go-isatty v0.0.24
- github.com/ncruces/go-strftime v1.0.0
- github.com/remyoudompheng/bigfft revision 24d4a6f8daec
- golang.org/x/sys v0.47.0

The release packages include the complete applicable license texts and third-party notices in `THIRD-PARTY-LICENSES/`. SQLite's upstream amalgamation is public domain; the modernc.org translation and supporting modules retain their respective notices. The source archive also includes the exact vendored module graph and license files used for the reproducible cgo-free build.

## OmniBridge 0.11.0 optional external adapters and research references

Minecraft Mod Vault does not bundle the tools below in the standard archive. Users may configure separately obtained executables. Each project retains its own license and distribution terms.

- Chunker — MIT — https://github.com/HiveGamesOSS/Chunker
- Regolith — MIT — https://github.com/Bedrock-OSS/regolith
- Geyser PackConverter — MIT — https://github.com/GeyserMC/PackConverter
- JE2BE Resource Pack Converter — MIT — https://github.com/Seraphic-Studio/JE2BE-Resource-Pack-Converter
- je2be-core — GPL-3.0 — https://github.com/kbinani/je2be-core
- Amulet — project-specific licenses — https://github.com/Amulet-Team/Amulet-Map-Editor
- Vineflower — Apache-2.0 — https://github.com/Vineflower/vineflower
- CFR — MIT — https://github.com/leibnitz27/cfr
- DataFixerUpper — LGPL-2.1 — https://github.com/Mojang/DataFixerUpper

Repository links in the in-app catalog are references, not embedded code or an assertion that every external project can be redistributed by Minecraft Mod Vault.
