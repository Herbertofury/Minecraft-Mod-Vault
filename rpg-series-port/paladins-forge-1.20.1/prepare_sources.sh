#!/usr/bin/env bash
set -euo pipefail
if [[ $# -ne 3 ]]; then
  echo 'usage: prepare_sources.sh <current-3.1.1-tree> <legacy-1.20.1-tree> <generated-root>' >&2
  exit 2
fi
CURRENT="$(cd "$1" && pwd)"
LEGACY="$(cd "$2" && pwd)"
OUT="$3"

rm -rf "$OUT"
mkdir -p "$OUT/common/java" "$OUT/common/resources" \
         "$OUT/reference/current-neoforge" "$OUT/reference/legacy-1.20.1"

# Current 3.1.1 is the sole feature/content authority.
cp -a "$CURRENT/common/src/main/java/." "$OUT/common/java/"
if [[ -d "$CURRENT/common/src/main/resources" ]]; then
  cp -a "$CURRENT/common/src/main/resources/." "$OUT/common/resources/"
fi
# Paladins keeps generated recipes/tags/lang/sounds under common/src/main/generated.
if [[ -d "$CURRENT/common/src/main/generated" ]]; then
  cp -a "$CURRENT/common/src/main/generated/." "$OUT/common/resources/"
fi

# Loader and historical trees are retained strictly outside compilation for translation work.
if [[ -d "$CURRENT/neoforge/src/main" ]]; then
  cp -a "$CURRENT/neoforge/src/main/." "$OUT/reference/current-neoforge/"
fi
cp -a "$LEGACY/src/main/." "$OUT/reference/legacy-1.20.1/"

# Common compilation must never inherit loader metadata from upstream.
rm -f "$OUT/common/resources/fabric.mod.json" \
      "$OUT/common/resources/META-INF/neoforge.mods.toml" \
      "$OUT/common/resources/META-INF/mods.toml"

# No symlinked input is allowed into deterministic release materialization.
if find "$OUT/common" -type l -print -quit | grep -q .; then
  echo '[Paladins materialize] symlink detected in common release inputs' >&2
  exit 2
fi

# Validate every staged JSON file now, before Gradle can obscure a content corruption.
python3 - "$OUT/common/resources" <<'PY'
import json, pathlib, sys
root=pathlib.Path(sys.argv[1]); count=0
for p in sorted(root.rglob('*.json')):
    with p.open('r',encoding='utf-8') as f: json.load(f)
    count += 1
print(f'[Paladins materialize] validated {count} JSON resource/data files')
PY

# Freeze the native-Forge registration frontier as machine-readable evidence.  This is an
# inventory, not a transform: the next port wave will replace these mutations with the proven
# Forge RegisterEvent/DeferredRegister bridge while common code keeps definition ownership.
REGISTRY="$OUT/REGISTRY_FRONTIER.txt"
{
  echo '# Paladins 3.1.1 direct registry mutation frontier'
  echo '# Generated from immutable current source; sorted; line numbers relative to generated/common/java.'
  grep -R -nE '(^|[^A-Za-z0-9_.])Registry\.register(Reference)?\(' "$OUT/common/java" || true
} | LC_ALL=C sort > "$REGISTRY"
DIRECT="$(grep -cE 'Registry\.register(Reference)?\(' "$REGISTRY" || true)"
printf '%s\n' "$DIRECT" > "$OUT/REGISTRY_FRONTIER.count"
if [[ "$DIRECT" -eq 0 ]]; then
  echo '[Paladins materialize] expected current direct-registry frontier disappeared; upstream assumption changed' >&2
  exit 2
fi

(
  cd "$OUT"
  find common -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > CURRENT_PORT_INPUTS.sha256
  find reference -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > REFERENCE_INPUTS.sha256
)

printf '[Paladins materialize] current 3.1.1 Java files: %s\n' "$(find "$OUT/common/java" -type f -name '*.java' | wc -l | tr -d ' ')"
printf '[Paladins materialize] current resource/data files: %s\n' "$(find "$OUT/common/resources" -type f | wc -l | tr -d ' ')"
printf '[Paladins materialize] direct registry mutations inventoried for Forge translation: %s\n' "$DIRECT"
echo '[Paladins materialize] current content staged intact; historical 1.20.1 retained as mapping/API reference only.'
