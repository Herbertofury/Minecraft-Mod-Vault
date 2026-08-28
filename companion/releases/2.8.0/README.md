# Minecraft Catalog Companion 2.8.0

2.8.0 is the provider-generic creator identity + typed post-media release.

## Cross-site creator and media ownership

- Creator avatars are now an independent semantic role and discovery lane rather than a side effect of project/gallery discovery.
- 23 first-class provider families carry exact project/author path rules and post-media semantics, while generic structured-data fallbacks cover future providers.
- Planet Minecraft binds the exact project author/member URL before accepting creator media. The verified Enderwoman case resolves `redstonae` to `https://www.planetminecraft.com/member/redstonae/` and uses the member avatar without allowing comment avatars or sibling-project imagery into the card.
- AFDIAN post media recovers the clean original asset from transformed/watermarked preview URLs and supports image, GIF, and video entries.
- Generic JSON-LD author/image/VideoObject extraction, linked originals, direct video/source elements, and semantic post-media hints feed the same typed media model.
- Project icon, creator avatar, gallery image, GIF, and video remain distinct roles all the way through parsing, cache sanitization, IPC, renderer merging, and lightbox display.
- Creator requests are creator-wide single-flights and run independently of project/gallery delivery, so a slow profile page cannot hold up an otherwise ready card.
- Live-media cache schema 10 invalidates pre-2.8 role-polluted records.

## Performance contract

- The provider expansion does not add a result, gallery, source, or provider cap.
- Existing progressive head delivery, worker parsing, shared Chromium/session cache, native HTTP lanes, provider API seeds, speculative media warming, parallel priming, uBlock Origin, and translator integration remain intact.
- Author discovery is parallel to project/gallery discovery rather than serialized behind it.

## Exact-package QA

All 36 release suites pass against source, staged Windows `resources/app`, and a fresh extraction of the final Windows ZIP. The suites cover the 23-provider registry, exact Planet Minecraft member/avatar binding, AFDIAN original-media recovery, typed image/GIF/video rendering, independent author delivery, media identity gates, parser pool, startup overlap, real-socket stress, and the existing catalog/browser regression matrix.

## Native Windows gate

The build host is Linux, so the Windows Electron `.exe` was not falsely claimed as executed here. The final Windows package ships `Self-Test.cmd`; native Windows Electron/Chromium smoke plus visual sampling of session-heavy providers remains the final platform-specific gate.

Final Windows ZIP SHA-256:

`6172cad335f55f38ce281104ebcce2966def793abd5f9902330c2ff88350ad12`

Source ZIP SHA-256:

`3332f33a9539701a16ba4e8e5650e22476f280dc44486d6bd87d7cf5b99e914c`
