# More RPG Library 2.7.2 -> native Forge 1.20.1 graduation acceptance

Authority:
- Minecraft 1.20.1 substrate `7da3c766ef5aebd850a0eb2f6a26bde2409f626f`
- supplied mandatory feature floor 2.7.1
- current 2.7.2 feature authority `bfa35e55133bd795676b6beb40c334d2904bf0ea`
- frozen Spell Engine 1.10.4 JAR `47bd98e6812a65d8e5b7137869c81b1b2168365a74410d86b1e46e5170eb935e`

A build is not graduation. Final acceptance requires all of the following on Java 17 + Forge 47.4.23:

1. Exact-authority source assembly is deterministic and machine-inventoried.
2. No Fabric API, Fabric Loader runtime, Forgified Fabric API, Connector, NeoForge classes/metadata, or accidental loader bridge in the packaged mod.
3. Modern 2.7.2 behavior retained for controlled entity relations, Control Enemy, stealth, mob spellcasting/channel/back-away behavior, custom spell impacts/predicates, functional attributes, elemental weakness, loot/config/functions, modern spell schools, particles/sounds/player animations, shared passives, Duelist's Focus/Siren's Tear, and worldgen processors.
4. Exact graduated foundations are used: Spell Engine 1.10.4, Spell Power 1.6.0, Ranged Weapon API 2.3.4, TinyConfig 3.1.0. Critical Strike and Armory remain optional unless a behavior lane explicitly activates them.
5. Dedicated-server semantic tests prove registration and behavior without client-class leakage.
6. Native Forge physical client proves resources, particles/render/HUD, animation loading, and representative gameplay paths.
7. Integrated-world tests exercise representative spells/effects/AI/loot/worldgen and optional-provider lanes.
8. Save/restart/config persistence is clean.
9. Untouched release builds twice with identical payload and outer-JAR identities.
10. Fresh packaged Forge server starts using the untouched release and reaches real `Done (...)!` plus semantic self-test markers.
11. First-green release/source identities are frozen and reproduced by a separate exact-head replay.
12. Evidence is preserved to GitHub + canonical Drive, then the RPG workbench is promoted only by non-force fast-forward.

Never replace a missing modern behavior with deletion merely because the API changed between 1.21.1 and 1.20.1; adapt it to the native target API.
