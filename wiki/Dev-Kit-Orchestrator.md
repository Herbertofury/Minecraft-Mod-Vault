# Dev Kit Orchestrator

`mmv-devkit` is the capability resolver for the Minecraft Dev Kit. It does not replace Minecraft Mod Vault TestGrid; it gives TestGrid and humans the correct toolchain inputs.

## Responsibilities

1. Locate the synced Dev Kit.
2. Inventory real files/folders against stable tool IDs.
3. Resolve tools by capability + loader + Minecraft version + platform.
4. Select the correct Java generation.
5. Prepare portable archives into a local cache without mutating Drive.
6. Produce a deterministic workflow plan.
7. Hand explicit paths/argv to TestGrid when execution needs evidence capture.

## Commands

```text
mmv-devkit configure --root PATH
mmv-devkit scan [--json]
mmv-devkit doctor [--strict]
mmv-devkit list [--category NAME]
mmv-devkit catalog [--output FILE]
mmv-devkit resolve CAPABILITY [--loader LOADER] [--mc VERSION]
mmv-devkit plan TASK [--loader LOADER] [--mc VERSION] [--json]
mmv-devkit prepare TOOL_ID
mmv-devkit run TOOL_ID [--allow-installer] [-- ARGS...]
```

`TASK` may be `build`, `port`, `decompile`, `profile`, `render`, `world`, `model`, `bedrock`, `launch`, `mappings`, or `repair`.

## Java policy

| Minecraft target | Resolver Java |
|---|---:|
| 1.16.x and older legacy targets | 8 |
| 1.17 through 1.20.x | 17 |
| 1.21.x | 21 |
| 26.x | 25 |

The Dev Kit currently has JDK 17, 25, and 26 archives. `doctor` intentionally reports JDK 21 as an offline gap for 1.21.x instead of silently using a newer JDK and calling it equivalent.

## Safety

The orchestrator is intentionally conservative around destructive or executable artifacts:

- it never downloads software;
- installers require `--allow-installer`;
- ZIP/tar path traversal is rejected;
- prepared archives are cached outside the source Drive tree;
- child processes use explicit argv, not shell interpolation;
- the registry and plans can be consumed as structured data.

## TestGrid handoff

Use `plan --json`. Each step reports a stable tool ID and selected path. TestGrid can then launch that exact path with its normal log assertions, network probes, hashes, reports, and evidence capture. Selection stays centralized while execution stays observable.
