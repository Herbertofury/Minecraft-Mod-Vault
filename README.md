# Minecraft-Mod-Vault

## Recovered C2ME OpenCL + Radium Forge 1.20.1 build

A previously interrupted temporary build task has been recovered and independently rechecked from the durable ProjectDump recovery branch.

Verified deliverable:

- ZIP: https://github.com/Herbertofury/ProjectDump/raw/tmp-c2me-opencl-radium-build/tmp-c2me-build/c2me-opencl-radiumcompat-forge1201.zip
- ZIP size: `5,403,128` bytes
- ZIP SHA-256: `35d48f7f392b268819c653a57032bdf3a8f62f665d9d03d80b6716ce83d0565f`
- JAR: `c2meF-0.2.0+alpha.12.1-opencl-radiumcompat-all.jar`
- JAR size: `5,398,697` bytes
- JAR SHA-256: `0f6f9e8d349bc092197e25c99e3a992eda1dab3a6dd1b62d15f75dbbd16cd393`
- Base c2meF nested entries preserved: `25`
- Added runtime entries: `4`
- Final nested entries: `31`

Runtime proof was performed on Minecraft `1.20.1`, Forge `47.4.0`, Java `17.0.20`, and exact Radium `0.12.4+git.26c9d8e`. The real headless Forge server loaded C2ME and Radium, applied both targeted Radium compatibility overrides, reached `Done (33.024s)!`, and cleanly fell back when no OpenCL context was available on the GPU-less runner. The build verifier also confirmed the archive was structurally valid and that the required LWJGL, Caffeine, and Commons Compress runtimes were visible to Forge.

### Install

1. Remove `c2meF-0.2.0+alpha.12-all.jar` from the instance `mods` folder.
2. Install `c2meF-0.2.0+alpha.12.1-opencl-radiumcompat-all.jar` from the recovered bundle.
3. Keep `radium-mc1.20.1-0.12.4+git.26c9d8e.jar` installed.
4. Copy `config/lithium.properties` from the bundle into the instance config folder.
5. If Twilight Forest is installed, keep `tfthreadsafetyaddon-forge-1.20.1-1.1.0.jar` as a separate mod.
6. Do not leave a second c2meF all-jar installed.

The temporary build PRs are execution scaffolding only and are not intended to merge into this repository.