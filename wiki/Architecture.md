# 🏗️ Architecture

## Conceptual layers

```mermaid
flowchart TB
  UI[Professional UI / CLI / Automation]
  ID[Identity + Provenance Graph]
  OM[OmniManager]
  MD[Mod Doctor]
  RL[Repair Lab]
  PL[Porting Lab]
  CB[Compatibility / Migration Brain]
  OB[OmniBridge Semantic Conversion]
  WF[WorldForge / WorldGraph]
  TG[TestGrid]
  AT[Agent Test Driver]

  UI --> OM
  UI --> MD
  UI --> WF
  OM --> ID
  MD --> RL
  MD --> PL
  PL --> CB
  ID --> OB
  CB --> OB
  OB --> WF
  OB --> TG
  WF --> TG
  TG --> AT
```

## Architectural rules

1. **Identity before mutation.**
2. **Semantic intermediate representations over raw text substitution.**
3. **Versioned adapters.**
4. **Opaque passthrough for safe unknown data.**
5. **Transactions around destructive operations.**
6. **Confidence + provenance attached to generated results.**
7. **Runtime verification when behavior/rendering matters.**
8. **Plugins/adapters cannot silently weaken fidelity.**

## Performance direction

Rust/native acceleration, GPU compute, worker pools, content-addressed caches and copy-on-write runtimes are welcome only where benchmarks show real improvement without sacrificing correctness.
