# Grassier Grass 1.4.5 Render Performance Fix — REPLACEMENT 3 Safety Verification

Target: Minecraft 1.20.1 / Forge 47.x / Java 17

## Why REPLACEMENT 3 exists

REPLACEMENT 1 crashed because its added `sectionOffsetUniform(ShaderInstance)` helper recursively called itself and returned a null cached uniform. REPLACEMENT 2 fixed that recursion, but a second audit identified a remaining avoidable lifetime assumption: caching the `SectionOffset` uniform object across frames/resource reloads was behavior upstream did not use.

REPLACEMENT 3 removes the uniform cache optimization completely. The two `SectionOffset` lookups are once again the same direct `ShaderInstance.getUniform("SectionOffset")` operations used by upstream. The larger performance changes remain.

## Identity / install contract

- Install mode: **REPLACEMENT** — remove the original, REPLACEMENT-1, and REPLACEMENT-2 Grassier Grass JARs; install only REPLACEMENT-3.
- Shipped filename: `Grassier-Grass-1.4.5-Render-Performance-Fix-REPLACEMENT-3.jar`
- Internal mod ID: `grassiergrass`
- Advertised version: `1.4.5`
- Display name: `Grassier Grass`
- Loader icon entry: `icon.png`
- `META-INF/mods.toml`: byte-identical to upstream.
- `META-INF/MANIFEST.MF`: byte-identical to upstream.
- Public/private method and field surface in `GrassDrawDispatcher`: identical to upstream.
- `GrassSectionCache` adds only seven **private static primitive** scan-state fields; no existing method descriptor or public API is changed.

## Changes retained

1. Persistent pending build queue so the complete section grid is not reconstructed every steady frame.
2. Full grid scan only when first initialized, camera section/radius changes, upstream LOD recheck triggers, or dirty/light-dirty/reLOD state requires it.
3. In-flight duplicate build suppression and retirement of completed cached queue entries.
4. Frustum culling before section-key/coordinate decoding in the hot draw loops.
5. `disposeAll()` resets the added scan state.

No visual-density, grass-style, wind, texture, model, lighting, LOD threshold, render-radius, build-budget, shader code, or gameplay-content reduction was made.

## Safety audit results

- Final JAR ZIP integrity: **PASS**.
- ASM bytecode/frame analysis: **PASS** — `GrassDrawDispatcher` 12/12 methods; `GrassSectionCache` 33/33 methods.
- StackMapTable present after frame recomputation in both transformed classes.
- Deterministic rebuild from pristine upstream: **PASS** byte-for-byte for the unmarked replacement.
- Uniform-cache regression removal: **PASS** — no `sectionOffsetUniform`, no cached shader/uniform fields, exactly two direct `SectionOffset` lookups matching upstream.
- Queue index/removal/in-flight/dirty-state audit: **PASS**.
- No added thread or ExecutorService; upstream `BUILD_WORKER` is untouched.
- No OpenGL/GPU work moved off the render thread.
- No archive signature files exist to invalidate.
- Java class-major distribution remains identical to upstream: 47 Java-17 classes (major 61) plus the same one upstream Java-8 class (major 52).
- Neutral/private-label scan: **PASS**.

## Exact archive diff

Upstream -> unmarked R3 changes exactly two entries:

- `com/leonardoinc22/shortgrass/client/render/GrassDrawDispatcher.class`
- `com/leonardoinc22/shortgrass/client/render/GrassSectionCache.class`

No entries added or removed and no metadata-only changes.

Unmarked R3 -> final marked R3 changes exactly:

- `icon.png`

No entries added/removed and no metadata-only changes.

## Repair Mark v2

- Upstream icon: 958x958.
- Upstream icon SHA-256: `ac002748371d861809fb528cc563520cae209729fec87c5a03f79503203768ab`
- Marked icon SHA-256: `7e3abc159d75fe4ac7f2fc0a0ff3eaa3ff0a34bdf3b9f86d15726fc8d5a4f659`
- Integration: embedded in the existing upstream `logoFile="icon.png"` slot.
- Marker-only diff: **PASS**.
- Repair mark adds no runtime code.

## Hashes

- Pristine upstream: `0bb02d32a840944feaba7f606823a0bf775f648f134749f988aab4a76a85ca3e`
- Unmarked REPLACEMENT-3: `6684277e3b61b6ca2c42480f61077a9f800f96dc844ff6d2748b5d9554b9ff28`
- Final REPLACEMENT-3: `a2122e4402cc583087cc2262f91dde458c05be5d9fc4eabb06df2ef7a0ed0dc5`
- Final size: `2299859` bytes

## Remaining truthfully unrun gate

I cannot honestly guarantee that *any* Minecraft mod can never crash without launching this exact final SHA in the full target runtime. What is proven here is that the known R1 crash mechanism is absent, the residual R2 uniform-cache risk has been removed rather than merely guarded, the remaining changes pass bytecode/frame/package checks, and the queue/culling edits contain no identified crash path in static audit.

The next certification step is one real launch into the same world/render path with this exact REPLACEMENT-3 SHA, followed by the 90-second Spark profile if startup/gameplay remains stable.