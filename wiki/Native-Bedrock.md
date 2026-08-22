# 🧱 Native Bedrock Management

> **Status:** 🟢 Documented as part of current OmniManager on `main`.

Bedrock content is first-class. OmniManager's documented management surface includes:

- `.mcpack`
- `.mcaddon`
- `.mcworld`
- `.mctemplate`
- behavior packs
- resource packs
- skin packs
- world templates
- development packs
- installed worlds
- Bedrock Stable and Preview/Beta roots
- custom `com.mojang` roots

## Metadata preserved

Localized names/descriptions, UUIDs, versions, minimum engine version, modules, scripts, capabilities, dependencies, authors, license, project links, icons and world artwork should remain attached to the content identity.

## World activation safety

World pack activation changes should edit the correct behavior/resource pack JSON and retain enough transaction evidence to restore the exact previous bytes.

## Roadmap expansion

OmniBridge and WorldForge extend this into conversion, direct `.mctemplate` editing, world repair, Java ↔ Bedrock world migration, add-on awareness and real Bedrock runtime testing.
