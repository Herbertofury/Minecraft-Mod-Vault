# 👩‍💻 Developer Guide

## Core contribution rule

Do not add a feature that only renders UI or generates files. Every user-facing control must connect to real domain logic and observable results.

## Implementation checklist

1. identify exact source/target formats and versions;
2. preserve originals/provenance;
3. add the narrow semantic implementation;
4. add structural/unit tests;
5. add fixture/regression coverage;
6. run loader/world/runtime integration tests as needed;
7. run real Minecraft verification for player-facing behavior;
8. record exact limitations and fidelity status;
9. update the capability matrix;
10. update this wiki only after evidence exists.

## World/version adapters

Version-specific assumptions belong in versioned adapters or migration knowledge, not scattered magic constants.

## Testing

See [TestGrid](TestGrid), [Agent Test Driver](Agent-Test-Driver), and [Validation & Fidelity](Validation-and-Fidelity).
