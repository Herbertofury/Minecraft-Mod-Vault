# Minecraft Dev Kit lessons captured from the final Ragdoll Forge release

Canonical validated skill SHA-256: `c65d417d69a144f95cfc2beb247969d3b1ffa14b45532acf0f50a96a26775559`  
Google Drive file ID: `1q1zpjLUMqb53Gls4soprOntp8F30UBHo`

The canonical skill was updated in place and round-trip verified after download from Drive.

## New durable QA rules

1. ForgeGradle `runClient` / `forgeclientuserdev` is not authoritative for production-SRG namespace defects. For namespace-sensitive ports, launch the exact packaged JARs through the real `forgeclient` target.
2. Audit direct Mixin `@Shadow`, `@Accessor`, `@Invoker`, and string selectors against the production target owner/descriptor. Ordinary bytecode remapping cannot infer every Mixin target.
3. Reflection member-name strings are outside ordinary bytecode remapping. Prefer stable APIs; when reflection is required, deliberately support mapped + production aliases and exercise the path in the packaged client.
4. `invokedynamic` / LambdaMetafactory SAM names must be resolved by functional-interface owner plus bootstrap SAM descriptor. Name-only or unique-method heuristics are insufficient and can produce packaged-client `AbstractMethodError` failures.
5. Restore production clients from exact launcher/profile artifacts: client SRG, client-extra, Forge patched-client/universal, FML language providers, BootstrapLauncher/loader libraries, platform-correct native classifiers, and the official asset index/object cache.
6. Offline ForgeGradle can still trigger network metadata tasks. Stage verified outputs first and disable only already-satisfied network tasks; pin Java/toolchain and exact missing build artifacts rather than weakening QA.
7. Package/resource smoke is a release gate: validate `pack.mcmeta`, item/block models, sounds, blockstates, and model-bake logs in the real production client. Intentionally invisible blocks can use a valid zero-geometry model.
8. Optional external providers require three same-hash modes: embedded only, external installed with a real semantic gameplay action, and external removed from the same save with automatic fallback.
9. If a bridge has uninstall cleanup, seed real inventory/ender/dropped/projectile states and assert exact cleanup counts from server-side command/log output.
10. Dependency-owned inherited methods must be remapped by the actual declaring owner + exact descriptor. Compile visibility and runtime ownership are separate concerns; do not repackage dependency classes just to simplify remapping.

The expanded canonical reference is preserved in the Drive skill under `references/forge-production-client-qa.md`.
