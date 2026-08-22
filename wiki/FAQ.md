# ❓ FAQ

### Is Minecraft Mod Vault just another launcher?
No. Its project identity is broader: artifact identity/provenance, management, repair, compatibility knowledge, porting, and a roadmap for conversion/world tooling.

### Does it support Bedrock?
Current project documentation includes native Bedrock package/world management. OmniBridge/WorldForge expand Bedrock into deeper conversion and editing.

### Is OmniBridge already fully implemented?
This wiki does not claim that. OmniBridge is documented as the 0.11.0 feature-expansion roadmap and should be promoted only as real features pass their acceptance gates.

### Why so much emphasis on hashes and provenance?
Because Minecraft projects are frequently mirrored, renamed, patched or distributed from multiple hosts. Identity mistakes can overwrite the wrong local artifact.

### Why run real Minecraft tests if static tests pass?
Static checks cannot prove rendering, AI, menus, interactions, worldgen or cross-client/server behavior. TestGrid escalates only when runtime truth is necessary.

### Can WorldForge prune chunks normally without AI/build detection?
Yes—the roadmap explicitly preserves a simple select → prune → normal-regenerate workflow. Smart build-aware retrogen is optional.

### Will WorldForge edit `.mctemplate` directly?
That is a premium roadmap requirement: direct package workspace, 2D/3D/data editing, repackaging and real Bedrock validation without a mandatory install-first workflow.
