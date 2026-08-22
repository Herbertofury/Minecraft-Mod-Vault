# 🎨 Blockbench & Blender

> **Status:** 📋 OmniBridge roadmap for deep integration.

## Blockbench

Target deep support includes:

- `.bbmodel` parsing;
- Bedrock/Java model projects;
- bones, pivots, groups and UVs;
- animation timelines;
- validation and bone repair;
- hitbox assistance;
- rig retargeting;
- round-trip stable IDs;
- **Open in Blockbench / Send to OmniBridge** workflows where integration permits.

## Blender

Target workflows include:

- `.blend` automation through Blender Python;
- glTF/GLB first-class interchange;
- FBX/OBJ when appropriate;
- armatures and animation;
- UV/material/texture handling;
- voxelization and build conversion;
- Minecraft ↔ Blender and Blockbench ↔ Blender round trips.

## Round-trip rule

OmniBridge should preserve stable asset/component identity and provenance so an external edit does not break every relationship on reimport.
