# Botania Expanded

Forge 1.20.1 addon and high-fidelity visual expansion for Botania magical flora.

## Endoflame family — 0.1.0-dev22 concept-match pass

The active production slice is still the Endoflame family. dev22 is a narrow visual follow-up to the real-client dev21 checkpoint: gameplay/content remain fixed while the renderer is pushed closer to the supplied Default / Warped / Forgotten concept sheets.

- **Endoflame (Default)** — Botania's original Endoflame gameplay with the premium purple/magenta renderer.
- **Warped Endoflame** — cyan/teal variant with equivalent Endoflame generating behavior.
- **Forgotten Endoflame** — ash/ivory/gold variant with equivalent Endoflame generating behavior.

## dev22 visual delta from dev21

The dev21 native captures proved the renderer worked, but also made the remaining mismatch obvious: the outer bloom still read as thin blade/ribbon geometry and the dominant inner flame was too skinny compared with the concept art. dev22 changes only that visual slice:

- broadens the six outer petals substantially and keeps their large faces visible instead of rolling them edge-on;
- extends the low, almost-horizontal shoulder sweep before each petal curls upward;
- tightens and twists the terminal hooks so the top view reads as curled petals rather than a six-point star;
- rebuilds the dominant central tongue as a much broader flame mass with a hooked crown;
- keeps the two supporting inner tongues lower and subordinate to the hero tongue;
- reshapes the core into a compact pear chamber and stops forcing its structural/base pass full-bright;
- keeps darker normally-lit petal bodies with selective emissive accents rather than flattening the whole flower into glow;
- retains the clean two-branch stem and allocation-free mutable geometry scratch path.

## Gameplay parity

No gameplay class, recipe, tag, block/item registry, loot table, fuel/mana rule, or networking payload changed in dev22. The production JAR is byte-identical to the native-green dev21 artifact except for:

1. `com/herbertofury/botaniaexpanded/client/render/EndoflameRenderer.class`
2. `META-INF/mods.toml` version `0.1.0-dev21` -> `0.1.0-dev22`
3. `META-INF/MANIFEST.MF` implementation version `0.1.0-dev21` -> `0.1.0-dev22`

## Build / verification target

- Minecraft 1.20.1
- Forge 47.4.23
- Java / Eclipse Temurin 17.0.20.1
- Botania 1.20.1-455
- Patchouli 1.20.1-85
- Curios 5.14.1+1.20.1

`tools/validate_release.py` validates all 32 JSON resources and dev22 renderer guards. The changed renderer class is compiled against the exact Forge 47.4.23 official-mapped development artifact, then remapped through ForgeGradle 6.0.54 using the recovered offline Forge cache before being inserted into the already native-green dev21 production JAR. Exact binary delta auditing confirms only the three entries above changed.

The next acceptance gate is the same real packaged Forge client used by dev21: capture new front and top views and compare the exact live geometry against the concept sheets before approving the Endoflame family.
