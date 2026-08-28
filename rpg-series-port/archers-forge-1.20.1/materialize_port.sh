#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
UPSTREAM="$ROOT/.upstream"
GENERATED="$ROOT/generated"

bash "$ROOT/prepare_sources.sh" "$UPSTREAM"
# shellcheck disable=SC1091
source "$ROOT/UPSTREAM_PINS.env"
CURRENT="$UPSTREAM/current-$ARCHERS_CURRENT_SHA"
LEGACY="$UPSTREAM/legacy-1.20.1-$ARCHERS_LEGACY_1201_SHA"

rm -rf "$GENERATED"
mkdir -p "$GENERATED/common/java" "$GENERATED/common/resources" \
         "$GENERATED/reference/current-neoforge" "$GENERATED/reference/legacy-1.20.1"

cp -a "$CURRENT/common/src/main/java/." "$GENERATED/common/java/"
if [[ -d "$CURRENT/common/src/main/resources" ]]; then
  cp -a "$CURRENT/common/src/main/resources/." "$GENERATED/common/resources/"
fi
if [[ -d "$CURRENT/common/src/main/generated" ]]; then
  cp -a "$CURRENT/common/src/main/generated/." "$GENERATED/common/resources/"
fi
cp -a "$CURRENT/neoforge/src/main/." "$GENERATED/reference/current-neoforge/"
cp -a "$LEGACY/src/main/." "$GENERATED/reference/legacy-1.20.1/"

python3 "$ROOT/apply_1201_transforms.py" "$GENERATED/common/java" "$GENERATED/common/resources"
python3 "$ROOT/apply_1201_api_transforms.py" "$GENERATED/common/java"
python3 "$ROOT/apply_1201_api_wave2.py" "$GENERATED/common/java" "$GENERATED/common/resources"
python3 "$ROOT/apply_1201_runtime_transforms.py" "$GENERATED/common/java"
python3 "$ROOT/apply_1201_forge_registration.py" "$GENERATED/common/java"

if grep -R -nE '(^|[^A-Za-z0-9_.])Registry\.register(Reference)?\(' "$GENERATED/common/java"; then
  echo '[Archers materialize] unbridged vanilla registry mutation survived native Forge registration transform' >&2
  exit 2
fi
grep -F 'ArcherBlocks.registerBlocks();' "$GENERATED/common/java/net/archers/ArchersMod.java" >/dev/null
grep -F 'public static void registerItemGroup()' "$GENERATED/common/java/net/archers/ArchersMod.java" >/dev/null
grep -F 'public static void registerItems()' "$GENERATED/common/java/net/archers/block/ArcherBlocks.java" >/dev/null
if grep -nF 'Registry.register(' "$ROOT/forge/src/main/java/net/archers/forge/ForgeMod.java"; then
  echo '[Archers materialize] Forge bridge bypasses RegisterEvent helper' >&2
  exit 2
fi
grep -F 'RegistryKeys.ENTITY_TYPE' "$ROOT/forge/src/main/java/net/archers/forge/ForgeMod.java" >/dev/null
grep -F 'RegistrationBridge.withRegistrar' "$ROOT/forge/src/main/java/net/archers/forge/ForgeMod.java" >/dev/null

(
  cd "$GENERATED"
  find common -type f -print0 | sort -z | xargs -0 sha256sum > CURRENT_PORT_INPUTS.sha256
  find reference -type f -print0 | sort -z | xargs -0 sha256sum > REFERENCE_INPUTS.sha256
)

printf '[Archers materialize] current 3.1.1 Java files: %s\n' "$(find "$GENERATED/common/java" -type f -name '*.java' | wc -l | tr -d ' ')"
printf '[Archers materialize] current resource/data files: %s\n' "$(find "$GENERATED/common/resources" -type f | wc -l | tr -d ' ')"
echo '[Archers materialize] all current common content staged; explicit 1.20.1 transforms and native Forge registration seam applied; references retained outside compilation.'