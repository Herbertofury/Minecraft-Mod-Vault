# Minecraft Mod Vault TestGrid

TestGrid is an additive Minecraft runtime driver and evidence system. It does not replace OmniManager, OmniBridge, Creator Vault, Repair Lab, Porting Lab, or WorldForge.

Run it without the UI:

```powershell
Minecraft-Mod-Vault-0.13.0-windows-x64-cli.exe testgrid capabilities
Minecraft-Mod-Vault-0.13.0-windows-x64-cli.exe testgrid validate .\testgrid\examples\java-server.json
Minecraft-Mod-Vault-0.13.0-windows-x64-cli.exe testgrid run .\testgrid\examples\java-server.json
Minecraft-Mod-Vault-0.13.0-windows-x64-cli.exe testgrid inspect <run-id>
```

Run the local API persistently:

```powershell
Minecraft-Mod-Vault-0.13.0-windows-x64-cli.exe serve --port 8765 --token-file .\mmv-token.txt
```

The server binds only to `127.0.0.1`. Supply the token as `X-MMV-Token` or the `token` query parameter.

## Capability boundary

- Java servers, Bedrock Dedicated Server, build tools, protocol probes, RCON, logs, files, hashes, and declared artifacts run headlessly.
- Java client screenshots/video require a real compatible render driver such as a loader-specific test client or HeadlessMC. TestGrid captures its declared output but does not claim every Minecraft client can render without a display.
- Bedrock Dedicated Server is headless. The Windows retail Bedrock client is not; use a real-client automation adapter when graphical client evidence is required.
- Marketplace worlds remain licensed content. TestGrid can test content the user legitimately owns, but it does not extract or redistribute Marketplace packages.

RCON passwords are referenced through `passwordEnv`; they are never accepted as manifest fields and persisted environment values with secret-like names are redacted.
