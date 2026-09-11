# Backpacks 2.0.2.4 FINAL verification

**Status: NATIVE-CERTIFIED FINAL**

- Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17.0.20.1
- JAR: `Backpacks-2.0.2.4-FINAL-Forge-1.20.1.jar`
- JAR SHA-256: `252e83b9f7fbe20fd15a4de047d4b02f1abcf87852dbbe7944beb386a85b1342`
- JAR size: 2906589 bytes
- Source ZIP SHA-256: `d5798b9031fcf69412d1d1c08ce1a0f562b8cdd68465c11b92cd0327549c2052`
- Previous certified baseline: 2.0.2.3 / `89fb25992e4cc90e4afa6aab402619144dfd68aed8ff268af9e88a7fcab73eca`

All invalidated static, build, packaged-client/integrated-server, upgrade-persistence, visual-render, crafting, and packaged dedicated-server gates are green. Candidate5 fixes the final recipe-book warning without altering the previously certified Candidate4 render/gameplay implementation. The 2.0.2.4 baseline-delta policy is versioned and passing.

Known non-release-blocking environment/dependency behavior is documented in the evidence README; no Candidate5-owned severe runtime failures remain.
