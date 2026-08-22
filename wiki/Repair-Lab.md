# 🩺 Repair Lab

> **Status:** ✅ Verified in the 0.9.0 release evidence.

Repair Lab is the source/project repair lane for broken or untrusted mod projects.

## Safety model

- treat source archives as hostile/untrusted input;
- defend against traversal, symlinks, duplicate paths, archive bombs and extraction escape;
- preserve immutable originals;
- detect the real loader/build contract;
- stage recognized migration edits instead of blindly rewriting everything;
- keep build/test/clean execution gated and explicit;
- hash outputs and preserve proof;
- support rollback.

## Repair loop

```mermaid
flowchart LR
  A[Input project/archive] --> B[Safe extraction + fingerprint]
  B --> C[Detect loader/build contract]
  C --> D[Diagnose root cause]
  D --> E[Scoped repair]
  E --> F[Build/test]
  F -->|fail| D
  F -->|pass| G[Artifact + proof + rollback data]
```

## See also

- [Porting Lab](Porting-Lab)
- [Compatibility Brain](Compatibility-Brain)
- [TestGrid](TestGrid)
