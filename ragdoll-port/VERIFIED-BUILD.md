# Forge 1.20.1 Ragdoll Backport - Verified Build

Target: Minecraft 1.20.1 / Forge 47.x / Java 17

Artifacts:
- `sable_player_ragdoll-1.20.1-forge-0.7.2-premium.1.jar`
  - SHA-256: `5fe65079a1b51a118e30a4281af95fcea20ca54391d9a4150c5ade530596c6ba`
- `ragdoll_reactions-1.20.1-forge-0.7.0-premium.1.jar`
  - SHA-256: `4bb9b7db37d6f803d49499b70f4030e0157ab3286f6be2e94d61a8c2a237aa6e`

Implementation notes:
- Both original NeoForge 1.21.1 Java 21 binaries were rebuilt as Forge 1.20.1 Java 17 compatibility implementations while preserving the original mod IDs.
- Key `dev.leo.sableplayerragdoll.api` launch/options/session signatures used by addons are preserved.
- Direct Sable 1.21.1 physics coupling is replaced behind a runtime-discovered ragdoll backend bridge so the Forge build can bind to a compatible installed physics mod without hardcoding one internal class name.
- Ragdoll Reactions implements hurt, fall, and death triggers with bounded directional launch velocity and an editable `config/ragdoll_reactions-forge.toml`.
- Original icons/language/sounds and safe resources are retained.

Verification performed:
- Valid ZIP/JAR integrity for both artifacts.
- All output classes are classfile major 61 (Java 17).
- Forge `META-INF/mods.toml` parses correctly.
- `@Mod("sable_player_ragdoll")` and `@Mod("ragdoll_reactions")` verified via `javap -v`.
- No bundled fake `net.minecraft`/`net.minecraftforge` stubs.
- No NeoForge metadata/classes in output.
- Compatibility API descriptors verified with `javap`.
- Simulated Forge event bus registration succeeded.
- Dynamic backend discovery test located a test ragdoll backend JAR and invoked spawn/release successfully.
- Ragdoll Reactions `LivingHurtEvent` runtime harness successfully triggered the core API and opened an active ragdoll session.

Google Drive canonical folder: `Minecraft Mod Vault/Ragdoll Forge 1.20.1 Backport`.

Remaining real-machine gate: launch these builds in the user's exact Forge 1.20.1 modpack with the chosen ragdoll physics backend and inspect the resulting `latest.log` plus in-game ragdoll behavior. Do not mark user-confirmed until that test is complete.
