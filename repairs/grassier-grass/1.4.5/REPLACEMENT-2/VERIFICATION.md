# Grassier Grass 1.4.5 Render Performance Fix — REPLACEMENT 2 Verification

Target: Minecraft 1.20.1 / Forge 47.x / Java 17

## Root cause corrected

REPLACEMENT 1 introduced a regression in the `SectionOffset` uniform cache helper. The bytecode transformer added `sectionOffsetUniform(ShaderInstance)` and then accidentally processed that helper in the same uniform-rewrite pass. Its real `ShaderInstance.getUniform("SectionOffset")` call was rewritten into a recursive call to `sectionOffsetUniform` itself.

On the first shader use, the helper stored the shader object, recursively re-entered itself, observed the same shader, and returned the still-null cached uniform. `drawSections` immediately dereferenced that null `AbstractUniform`, producing the render-thread `NullPointerException`.

## REPLACEMENT 2 correction

- Rebuilt from the untouched upstream Grassier Grass 1.4.5 Forge 1.20.1 JAR.
- The transformer now explicitly excludes `sectionOffsetUniform` from the per-section lookup rewrite pass.
- The two real hot-loop `SectionOffset` lookups are still replaced with the cache helper.
- The helper itself performs exactly one real `ShaderInstance.getUniform("SectionOffset")` lookup and contains zero self-calls.
- All other render-performance changes from REPLACEMENT 1 remain in place.

## Preserved performance work

- Persistent pending section-build queue instead of clearing/reconstructing it every render frame.
- Full section-grid scan only when camera section/radius/vertical-range/dirty state requires it.
- Duplicate in-flight build suppression and stale queued-build retirement.
- Frustum culling before section-coordinate decoding in the hot draw loops.
- Cached `SectionOffset` uniform per shader instance, now without recursion.
- Existing grass density, geometry, tinting, wind, shaders, render radius, LOD thresholds, resources, and configuration defaults are unchanged.

## Compatibility identity

- Install mode: **REPLACEMENT** — remove the previous/original Grassier Grass JAR and use this JAR instead.
- Shipped filename: `Grassier-Grass-1.4.5-Render-Performance-Fix-REPLACEMENT-2.jar`
- Internal mod ID: `grassiergrass`
- Advertised version: `1.4.5`
- Display name: `Grassier Grass`
- Loader icon entry: `icon.png`
- `META-INF/mods.toml` content: byte-identical to upstream.
- `META-INF/MANIFEST.MF` content: byte-identical to upstream.
- Neutral/private-label scan: PASS.

## Static/package verification

- Upstream SHA-256: `0bb02d32a840944feaba7f606823a0bf775f648f134749f988aab4a76a85ca3e`
- REPLACEMENT 2 unmarked SHA-256: `f4d8d65d381422522577e6e668f1ebbfc2548e425c207b1b9bb6880a8d920dbd`
- REPLACEMENT 2 final SHA-256: `91006e8220e65572d7e9e1e54b6058526c4adf908faedc750c9f19f13623df8c`
- Final size: `2300791` bytes
- ZIP/JAR integrity: PASS.
- ASM verifier: PASS — `GrassDrawDispatcher` 13 methods; `GrassSectionCache` 33 methods.
- Uniform-cache regression guard: PASS — helper self-calls `0`, helper real uniform lookups `1`, optimized external call sites `2`.
- Upstream -> unmarked entry diff: exactly two content changes and zero added/removed/metadata-only entries:
  - `com/leonardoinc22/shortgrass/client/render/GrassDrawDispatcher.class`
  - `com/leonardoinc22/shortgrass/client/render/GrassSectionCache.class`
- Unmarked -> final Repair Mark diff: exactly `icon.png`; zero added/removed/metadata-only entries.
- Deterministic unmarked rebuild from pristine upstream: byte-for-byte PASS.

## Repair Mark v2

- Exact upstream-authored `icon.png` basis: 958x958.
- Unmarked artwork SHA-256: `ac002748371d861809fb528cc563520cae209729fec87c5a03f79503203768ab`
- Marked artwork SHA-256: `7e3abc159d75fe4ac7f2fc0a0ff3eaa3ff0a34bdf3b9f86d15726fc8d5a4f659`
- 48x48 preview SHA-256: `b9621401e33079da77f6b33df1dbf90e1b5e565d3fe4056e8378dbcede9b52d1`
- 48x48 visual QA: PASS — upstream grass-block identity remains recognizable and the red frame/check remain legible.
- Integration: `embedded` in the existing upstream `logoFile="icon.png"` slot.
- No runtime code is added by the visual mark.

## Remaining runtime gate

The exact crash-causing recursive bytecode is removed and statically verified. Final runtime certification still requires launching the full target instance with REPLACEMENT 2, entering the same world/render path, and confirming the prior `sectionOffsetUniform` null crash is absent. After that, rerun the 90-second Spark profile to measure the performance delta.
