#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
CI="$ROOT/rpg-series-port/ci"
TARGET="$CI/run-rpg-series-integration-acceptance.sh"
NEXT="$CI/run-rpg-series-integration-deep-acceptance.sh"
[[ -f "$TARGET" && -f "$NEXT" ]] || { echo '[RPG integration client remap] required runner missing' >&2; exit 2; }
python3 - "$TARGET" <<'PY'
from pathlib import Path
import sys
path=Path(sys.argv[1])
text=path.read_text()
old_pal_jar="95e8f9e074dd1c432f9486c922173b76a3b3bc57667b28fe154e47dba5c374ee"
new_pal_jar="38d011a0435cda0385485b18f7f2736b8cf4f98fec01ebc12d69ea74fc0fd8e7"
old_pal_source="fb0e5812857a2fd46de488cd17a80011ef5d18795ff96fa1a3ebed5fd19a4377"
new_pal_source="a49b351b085cb7a423244fb8da74ac1f5b4b389d3849d89d61ea0047e0964ade"
if text.count(old_pal_jar) != 1:
    raise SystemExit(f'[RPG integration client remap] expected one stale Paladins JAR authority pin, found {text.count(old_pal_jar)}')
if text.count(old_pal_source) != 1:
    raise SystemExit(f'[RPG integration client remap] expected one stale Paladins source authority pin, found {text.count(old_pal_source)}')
text=text.replace(old_pal_jar,new_pal_jar,1)
text=text.replace(old_pal_source,new_pal_source,1)
old=r'''RUN="$ROG/forge/run"; mkdir -p "$RUN/mods" "$RUN/config"; rm -f "$RUN/mods/"*.jar
cp -f "$PAL_JAR" "$RUN/mods/paladins-forge-3.1.1+1.20.1.jar"
cp -f "$SHIELD_FORGE" "$RUN/mods/$(basename "$SHIELD_FORGE")"
cp -f "$RUNES_FORGE" "$RUN/mods/$(basename "$RUNES_FORGE")"
printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"'''
new=r'''RUN="$ROG/forge/run"; mkdir -p "$RUN/mods" "$RUN/config"; rm -f "$RUN/mods/"*.jar
CLIENT_INIT="${RUNNER_TEMP:-/tmp}/rpg-series-integration-extra-runtime.gradle"
export RPG_INTEGRATION_PALADINS_JAR="$PAL_JAR"
export RPG_INTEGRATION_SHIELD_JAR="$SHIELD_FORGE"
export RPG_INTEGRATION_RUNES_JAR="$RUNES_FORGE"
cat > "$CLIENT_INIT" <<'GRADLE_INIT'
allprojects {
    afterEvaluate { p ->
        if (p.path == ':forge') {
            ['RPG_INTEGRATION_PALADINS_JAR','RPG_INTEGRATION_SHIELD_JAR','RPG_INTEGRATION_RUNES_JAR'].each { key ->
                def raw = System.getenv(key)
                if (raw == null || raw.trim().isEmpty()) throw new GradleException("Missing integration runtime input: ${key}")
                def jar = p.file(raw)
                if (!jar.isFile()) throw new GradleException("Integration runtime input does not exist: ${jar}")
                p.dependencies.add('modRuntimeOnly', p.files(jar))
            }
        }
    }
}
GRADLE_INIT
printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"
echo "[RPG integration] CLIENT_NAMESPACE_REMAP_HARDENING_PASS paladins=$PAL_SHA rogues=$ROG_SHA"'''
run_old=r'''( gradle --no-daemon -p "$ROG" :forge:runClient "${ARGS[@]}" --args='--width 1280 --height 720' </dev/null ) > "$ROOT/rpg-series-integration-client.log" 2>&1 & CLIENT_PID=$!'''
run_new=r'''( gradle --no-daemon --init-script "$CLIENT_INIT" -p "$ROG" :forge:runClient "${ARGS[@]}" --args='--width 1280 --height 720' </dev/null ) > "$ROOT/rpg-series-integration-client.log" 2>&1 & CLIENT_PID=$!'''
if text.count(old) != 1:
    raise SystemExit(f'[RPG integration client remap] expected one production-JAR run/mods seam, found {text.count(old)}')
if text.count(run_old) != 1:
    raise SystemExit(f'[RPG integration client remap] expected one userdev launch seam, found {text.count(run_old)}')
text=text.replace(old,new,1)
text=text.replace(run_old,run_new,1)
path.write_text(text)
PY
grep -Fq '38d011a0435cda0385485b18f7f2736b8cf4f98fec01ebc12d69ea74fc0fd8e7' "$TARGET" || { echo '[RPG integration client remap] repaired Paladins JAR authority pin missing' >&2; exit 2; }
grep -Fq 'a49b351b085cb7a423244fb8da74ac1f5b4b389d3849d89d61ea0047e0964ade' "$TARGET" || { echo '[RPG integration client remap] repaired Paladins source authority pin missing' >&2; exit 2; }
grep -Fq 'CLIENT_NAMESPACE_REMAP_HARDENING_PASS' "$TARGET" || { echo '[RPG integration client remap] remap marker missing after patch' >&2; exit 2; }
grep -Fq -- '--init-script "$CLIENT_INIT"' "$TARGET" || { echo '[RPG integration client remap] init-script launch hook missing after patch' >&2; exit 2; }
if grep -Fq 'cp -f "$PAL_JAR" "$RUN/mods/paladins-forge-3.1.1+1.20.1.jar"' "$TARGET"; then echo '[RPG integration client remap] production Paladins JAR still bypasses Loom remap' >&2; exit 2; fi
bash -n "$TARGET"
echo '[RPG integration client remap] PALADINS_REPAIRED_AUTHORITY_PIN_PASS'
echo '[RPG integration client remap] CLIENT_NAMESPACE_REMAP_WRAPPER_PASS'
bash "$NEXT"
