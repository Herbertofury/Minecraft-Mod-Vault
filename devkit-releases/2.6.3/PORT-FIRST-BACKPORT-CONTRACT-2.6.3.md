# Minecraft Dev Kit 2.6.3 — Port-First / Opt-In Vanilla Backport Contract

## Non-negotiable order

1. **Complete the full target-native mod port first.** Port all mod-owned content, code, assets, data, dependencies, integrations, save/network behavior, and target-loader behavior before adding future-vanilla content.
2. **Certify the base port independently.** Future-vanilla additions remain OFF. The base artifact must pass applicable content/source parity, build, `port-guard`, dedicated-server, native-client/integrated-server, restart/persistence, and release gates.
3. **Always offer the future-vanilla parity layer afterward.** The offer is produced even when the user did not explicitly ask for it. `NONE_REQUIRED` is a valid result.
4. **Show exactly what must be brought over.** For each cross-version vanilla feature, list the exact feature ID, source/first-seen Minecraft version, why the mod touches it, target presence, parity role, complete dependency closure, provider ownership, embedded fallback needs, assets/data/code provenance, client/server/network/save/world impact, and QA lanes.
5. **Default OFF; explicit opt-in only.** Never silently merge future-vanilla content into the certified base artifact. Preserve the base build intact.
6. **Certify the opted-in layer separately.** Test provider-present, provider-absent embedded fallback where supported, opt-out/disable recovery, same-world provider removal when persistence applies, sound/resource hashes, and exactly-once behavior where provider and fallback hooks can overlap.

## Parity wording

Classify every offered future-vanilla dependency as one of:

- `OPTIONAL_ENHANCEMENT` — improves source-era vanilla parity but is not needed to preserve the mod's intended target-native behavior.
- `REQUIRED_FOR_EXACT_SOURCE_PARITY` — exact behavior from the newer source-era mod/vanilla lane cannot be claimed without it.
- `NOT_APPLICABLE` — no cross-version vanilla work is needed for this surface.

A base artifact can be called a **complete target-native port** only when any removed future-vanilla dependency has a behaviorally valid target-native adaptation. If no valid adaptation exists, do not claim exact source-version parity until the offered backport layer is accepted and completed.

## Machine gate

`mmv-devkit vanilla-atlas plan-backport` is offer-only by default:

```text
stage=POST_PORT_OPTIONAL_VANILLA_PARITY
offerRequired=true
optIn=false
baselinePortGate=NOT_PROVIDED
implementationReady=false
```

Planning may happen before the base port is finished so requirements are known early. Implementation authorization requires both:

```text
--port-report <passing port-guard-report.json>
--opt-in
```

`--opt-in` with no report or a failing report is rejected. Provider presence is still not feature-ownership proof; unresolved ownership remains `UNRESOLVED` and cannot become implementation-ready.
