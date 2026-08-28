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

(
  cd "$GENERATED"
  find common -type f -print0 | sort -z | xargs -0 sha256sum > CURRENT_PORT_INPUTS.sha256
  find reference -type f -print0 | sort -z | xargs -0 sha256sum > REFERENCE_INPUTS.sha256
)

printf '[Archers materialize] current 3.1.1 Java files: %s\n' "$(find "$GENERATED/common/java" -type f -name '*.java' | wc -l | tr -d ' ')"
printf '[Archers materialize] current resource/data files: %s\n' "$(find "$GENERATED/common/resources" -type f | wc -l | tr -d ' ')"
echo '[Archers materialize] all current common content staged; explicit 1.20.1 transforms applied; references retained outside compilation.'
