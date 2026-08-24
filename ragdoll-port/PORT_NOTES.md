# Forge 1.20.1 Ragdoll Port

Target: Minecraft 1.20.1, Forge 47.4.23, Java 17.

Design goals:
- Preserve the `sable_player_ragdoll` compatibility/API surface required by add-ons while removing the unavailable Sable 1.21.1 runtime requirement.
- Keep `ragdoll_reactions` as a separate add-on.
- Use a proven Forge 1.20.1 ragdoll backend rather than metadata-only conversion.
- Prefer Ragdollified 1.0.0 (MIT) as the physics/runtime backend if its public API supports live player/mob ragdoll launch and force application.
- Preserve configurable hit/fall/explosion/lightning/impact reactions and multiplayer-safe server authority.
- Compile all delivered code for Java 17/classfile 61.

Original inputs (private repair inputs):
- ragdoll_reactions-1.21.1-0.7.0.jar
- sable_player_ragdoll-1.21.1-0.7.2.jar

No claim of runtime completion is made until Forge 1.20.1 launch and reaction smoke gates pass.
