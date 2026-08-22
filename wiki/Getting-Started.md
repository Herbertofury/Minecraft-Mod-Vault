# 🚀 Getting Started

> [!TIP]
> Start with the workflow you actually need. Minecraft Mod Vault is intentionally broader than a launcher, so the best entry point depends on whether you are **managing**, **repairing**, **porting**, **converting**, or **editing a world**.

## Choose a lane

| Goal | Primary workspace | Status |
|---|---|---|
| Identify and manage installed content | [OmniManager](OmniManager) | 🟢 Current project identity |
| Repair a broken/untrusted source project | [Repair Lab](Repair-Lab) | ✅ Verified in 0.9.0 evidence |
| Inspect/plan a JAR version port | [Porting Lab](Porting-Lab) | ✅ Verified in 0.9.0 evidence |
| Search reusable compatibility knowledge | [Compatibility Brain](Compatibility-Brain) | ✅ Verified in 0.9.0 evidence |
| Convert Java/Bedrock/loaders/assets | [OmniBridge](OmniBridge) | 📋 0.11.0 roadmap |
| Run fast automated Minecraft QA | [TestGrid](TestGrid) | 📋 0.11.0 roadmap |
| Let an agent test like a real player | [Agent Test Driver](Agent-Test-Driver) | 📋 0.11.0 roadmap |
| Edit/convert/repair/prune worlds | [WorldForge](WorldForge) | 📋 0.11.0 roadmap |

## Mental model

1. **Import or discover** the local artifact/world.
2. **Resolve identity and provenance** before changing anything.
3. **Inspect dependencies and target compatibility.**
4. **Choose management / repair / port / conversion / world-edit action.**
5. **Preview destructive or fidelity-sensitive changes.**
6. **Create a recoverable transaction/checkpoint.**
7. **Execute.**
8. **Validate structurally.**
9. **Escalate to real Minecraft runtime testing when behavior matters.**
10. **Keep the output plus evidence and provenance together.**

## Safety defaults

- Originals remain immutable wherever practical.
- Patched/local builds must not be silently overwritten by provider updates.
- Fuzzy identity matches are suggestions, not authorization to replace files.
- Unknown world/mod data should be preserved opaquely where safe rather than discarded.
- A world conversion that merely opens is **not** automatically considered correct.

## Next reads

- [Release Status](Release-Status)
- [Feature Overview](Feature-Overview)
- [Validation & Fidelity](Validation-and-Fidelity)
- [Troubleshooting](Troubleshooting)
