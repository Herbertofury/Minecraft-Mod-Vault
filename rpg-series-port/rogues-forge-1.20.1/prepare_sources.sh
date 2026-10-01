#!/usr/bin/env bash
set -euo pipefail
[[ $# -eq 3 ]] || { echo 'usage: prepare_sources.sh <current> <legacy> <out>' >&2; exit 2; }
CURRENT="$(cd "$1" && pwd)"; LEGACY="$(cd "$2" && pwd)"; OUT="$3"
rm -rf "$OUT"; mkdir -p "$OUT/common/java" "$OUT/common/resources" "$OUT/reference/current-neoforge" "$OUT/reference/legacy-1.20.1"
cp -a "$CURRENT/common/src/main/java/." "$OUT/common/java/"
[[ ! -d "$CURRENT/common/src/main/resources" ]] || cp -a "$CURRENT/common/src/main/resources/." "$OUT/common/resources/"
[[ ! -d "$CURRENT/common/src/main/generated" ]] || cp -a "$CURRENT/common/src/main/generated/." "$OUT/common/resources/"
[[ ! -d "$CURRENT/neoforge/src/main" ]] || cp -a "$CURRENT/neoforge/src/main/." "$OUT/reference/current-neoforge/"
cp -a "$LEGACY/src/main/." "$OUT/reference/legacy-1.20.1/"
rm -f "$OUT/common/resources/fabric.mod.json" "$OUT/common/resources/META-INF/neoforge.mods.toml" "$OUT/common/resources/META-INF/mods.toml"
if find "$OUT/common" -type l -print -quit | grep -q .; then echo '[Rogues materialize] symlink in release input' >&2; exit 2; fi
python3 - "$OUT/common/resources" <<'PY'
import json,pathlib,sys
n=0
for p in sorted(pathlib.Path(sys.argv[1]).rglob('*.json')):
    with p.open(encoding='utf-8') as f: json.load(f)
    n+=1
print(f'[Rogues materialize] validated {n} JSON files')
PY
REG="$OUT/REGISTRY_FRONTIER.txt"; { echo '# Rogues 3.1.1 registry frontier'; grep -R -nE 'Registry\.register(Reference)?\(' "$OUT/common/java" || true; } | LC_ALL=C sort > "$REG"
N="$(grep -cE 'Registry\.register(Reference)?\(' "$REG" || true)"; printf '%s\n' "$N" > "$OUT/REGISTRY_FRONTIER.count"; (( N > 0 )) || { echo '[Rogues materialize] registry frontier vanished' >&2; exit 2; }
(cd "$OUT"; find common -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > CURRENT_PORT_INPUTS.sha256; find reference -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > REFERENCE_INPUTS.sha256)
echo "[Rogues materialize] current Java=$(find "$OUT/common/java" -name '*.java' -type f | wc -l) registry_frontier=$N; current content staged, legacy reference-only"
