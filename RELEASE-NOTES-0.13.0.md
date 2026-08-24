# Minecraft Mod Vault 0.13.0

## TestGrid + the complete Vault

Version 0.13.0 adds TestGrid without replacing or narrowing Minecraft Mod Vault. OmniManager, OmniBridge, Mod Doctor, Compatibility Brain, Repair Lab, Porting Lab, WorldForge, Creator Vault, Java management, and native Bedrock management remain part of the same desktop application.

### Added

- Headless, manifest-driven execution for Java servers, Bedrock Dedicated Server, build tools, render adapters, and other explicit executables.
- TCP, Java server-list, Bedrock RakNet, authenticated RCON, log, file, SHA-256, command, process, and artifact assertions.
- Combined logs plus JSON, JUnit XML, and standalone HTML reports for each run.
- Hash-addressed artifact capture and persistent run history.
- A loopback-only, token-protected TestGrid API and persistent service mode.
- A dedicated TestGrid desktop studio alongside the existing Vault workspaces.
- Windows GUI, Windows CLI, and Linux x64 release binaries.
- Real Windows and Linux CI proofs for the CLI self-test manifests.

### Control and safety boundaries

- Commands are arrays of arguments and are launched directly; manifests do not pass untrusted strings through a shell.
- RCON secrets are read from environment variables, not stored in manifests, and secret-like environment values are redacted from reports.
- Paths are resolved beneath explicit workspace roots and captured artifacts are size bounded.
- Dedicated servers can run headlessly. Rendered Java-client evidence still requires a compatible render driver, while the retail Bedrock client requires a real-client adapter.
- TestGrid may test licensed content the user owns, but it does not extract or redistribute Marketplace packages.

### Quick start

```powershell
Minecraft-Mod-Vault-0.13.0-windows-x64-cli.exe testgrid capabilities
Minecraft-Mod-Vault-0.13.0-windows-x64-cli.exe testgrid validate .\testgrid\examples\windows-self-test.json
Minecraft-Mod-Vault-0.13.0-windows-x64-cli.exe testgrid run .\testgrid\examples\windows-self-test.json
```

See [`testgrid/README.md`](testgrid/README.md) and [`wiki/TestGrid.md`](wiki/TestGrid.md) for manifests, reports, API routes, and runtime boundaries.
