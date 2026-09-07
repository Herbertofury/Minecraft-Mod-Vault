#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/paladins-forge-1.20.1"
TMP="${RUNNER_TEMP:-/tmp}/paladins-first-compile"
ABI_MANIFEST="$PORT/paladins-foundations.sha256"
SHIELD_ACCEPTED_SHA256="bd6a2fbeb357c25953abfb14ba18d2c5344e5351c29d2cb082244bc48e8da48a"
rm -rf "$TMP"; mkdir -p "$TMP/ext"
: > "$ABI_MANIFEST"

resolve_jar() {
  local label="$1" dir="$2" name_glob="$3"
  local -a candidates=()
  [[ -d "$dir" ]] || { echo "[Paladins ABI] $label directory missing: $dir" >&2; return 1; }
  mapfile -t candidates < <(find "$dir" -maxdepth 1 -type f -name "$name_glob" \
    ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -print | sort)
  if (( ${#candidates[@]} != 1 )); then
    echo "[Paladins ABI] $label expected exactly one $name_glob release candidate, found ${#candidates[@]} in $dir" >&2
    find "$dir" -maxdepth 1 -type f -name '*.jar' -printf '[Paladins ABI] candidate: %f\n' | sort >&2 || true
    return 1
  fi
  printf '%s\n' "${candidates[0]}"
}

validate_jar() {
  local label="$1" jar="$2" hash
  [[ -f "$jar" ]] || { echo "[Paladins ABI] $label missing: $jar" >&2; return 1; }
  unzip -tq "$jar" >/dev/null
  hash="$(sha256sum "$jar" | awk '{print $1}')"
  printf '%s  %s\n' "$hash" "$jar" >> "$ABI_MANIFEST"
  echo "[Paladins ABI] $label sha256=$hash"
}

download_jar() {
  local label="$1" url="$2" out="$3"
  curl --retry 2 --retry-delay 1 --retry-connrefused -fsSL "$url" -o "$out"
  validate_jar "$label" "$out"
}

clone_exact() {
  local repo="$1" sha="$2" dst="$3"
  git init -q "$dst"
  git -C "$dst" remote add origin "https://github.com/$repo.git"
  git -C "$dst" fetch -q --depth=1 origin "$sha"
  git -C "$dst" checkout -q --detach FETCH_HEAD
  [[ "$(git -C "$dst" rev-parse HEAD)" = "$sha" ]]
}

# The workflow explicitly checks out the exact pull-request head. GitHub's GITHUB_SHA on
# pull_request events is the synthetic merge commit, so comparing HEAD to GITHUB_SHA there
# is incorrect. Preserve strict equality on push/other events and log the immutable PR head.
CHECKOUT_HEAD="$(git -C "$ROOT" rev-parse HEAD)"
if [[ "${GITHUB_EVENT_NAME:-}" != "pull_request" && -n "${GITHUB_SHA:-}" ]]; then
  [[ "$CHECKOUT_HEAD" = "$GITHUB_SHA" ]] || {
    echo "[Paladins] checkout mismatch: HEAD=$CHECKOUT_HEAD GITHUB_SHA=$GITHUB_SHA event=${GITHUB_EVENT_NAME:-unknown}" >&2
    exit 2
  }
fi
echo "[Paladins] exact checkout HEAD=$CHECKOUT_HEAD event=${GITHUB_EVENT_NAME:-unknown}"

echo '[Paladins] Reconstructing exact separate foundations'
bash "$ROOT/rpg-series-port/ci/build-shield-api-foundation.sh"
SHIELD="$ROOT/rpg-series-port/shield-api-forge-1.20.1"
SHIELD_COMMON="$(resolve_jar 'Shield API common' "$SHIELD/common/build/libs" '*-common-*.jar')"
SHIELD_FORGE="$(resolve_jar 'Shield API Forge' "$SHIELD/forge/build/libs" '*.jar')"
validate_jar 'Shield API common' "$SHIELD_COMMON"
validate_jar 'Shield API Forge' "$SHIELD_FORGE"
ACTUAL_SHIELD_SHA="$(sha256sum "$SHIELD_FORGE" | awk '{print $1}')"
[[ "$ACTUAL_SHIELD_SHA" = "$SHIELD_ACCEPTED_SHA256" ]] || {
  echo "[Paladins] Shield graduation identity mismatch: expected $SHIELD_ACCEPTED_SHA256 got $ACTUAL_SHIELD_SHA" >&2
  exit 3
}
echo "[Paladins] Shield graduation identity matched run #188: $ACTUAL_SHIELD_SHA"

bash "$ROOT/rpg-series-port/ci/build-spell-engine-foundation.sh"
SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"
TINY="$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated"
SPELL_ENGINE_BUILD="$ROOT/.spell-engine-build"
SPELL_POWER_COMMON="$(resolve_jar 'Spell Power common' "$SPELL_POWER/common/build/libs" '*-common-*.jar')"
SPELL_POWER_FORGE="$(resolve_jar 'Spell Power Forge' "$SPELL_POWER/forge/build/libs" '*-forge-*.jar')"
TINY_COMMON="$(resolve_jar 'TinyConfig common' "$TINY/common/build/libs" '*-common-*.jar')"
TINY_FORGE="$(resolve_jar 'TinyConfig Forge' "$TINY/forge/build/libs" '*-forge-*.jar')"
SPELL_ENGINE_COMMON="$(resolve_jar 'Spell Engine common' "$SPELL_ENGINE_BUILD/common/build/libs" '*-common-*.jar')"
SPELL_ENGINE_FORGE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"
for spec in \
  "Spell Power common|$SPELL_POWER_COMMON" \
  "Spell Power Forge|$SPELL_POWER_FORGE" \
  "TinyConfig common|$TINY_COMMON" \
  "TinyConfig Forge|$TINY_FORGE" \
  "Spell Engine common|$SPELL_ENGINE_COMMON" \
  "Spell Engine Forge|$SPELL_ENGINE_FORGE"; do
  validate_jar "${spec%%|*}" "${spec#*|}"
done

RUNES="$ROOT/rpg-series-port/runes-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$RUNES" clean :common:jar :forge:remapJar
RUNES_COMMON="$(resolve_jar 'Runes common' "$RUNES/common/build/libs" '*-common-*.jar')"
RUNES_FORGE="$(resolve_jar 'Runes Forge' "$RUNES/forge/build/libs" '*-forge-*.jar')"
validate_jar 'Runes common' "$RUNES_COMMON"
validate_jar 'Runes Forge' "$RUNES_FORGE"

STRUCTURE="$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$STRUCTURE" clean :common:jar :forge:remapJar
STRUCTURE_COMMON="$(resolve_jar 'Structure Pool API common' "$STRUCTURE/common/build/libs" '*-common-*.jar')"
STRUCTURE_FORGE="$(resolve_jar 'Structure Pool API Forge' "$STRUCTURE/forge/build/libs" '*-forge-*.jar')"
validate_jar 'Structure Pool API common' "$STRUCTURE_COMMON"
validate_jar 'Structure Pool API Forge' "$STRUCTURE_FORGE"

ARMOR="$ROOT/rpg-series-port/armor-model-api-forge-1.20.1"
ARMOR_SHA=a664155a0aab3161cd7e4bf0c1f72512b4ec4949
clone_exact FabricExtras/ArmorModelAPI "$ARMOR_SHA" "$TMP/armor-target"
python3 "$ARMOR/tools/prepare_port.py" "$TMP/armor-target" "$ARMOR/generated"
gradle --no-daemon --stacktrace -p "$ARMOR/generated" clean :common:jar :forge:remapJar
ARMOR_COMMON="$(resolve_jar 'Armor Model API common' "$ARMOR/generated/common/build/libs" '*-common-*.jar')"
ARMOR_FORGE="$(resolve_jar 'Armor Model API Forge' "$ARMOR/generated/forge/build/libs" '*-forge-*.jar')"
validate_jar 'Armor Model API common' "$ARMOR_COMMON"
validate_jar 'Armor Model API Forge' "$ARMOR_FORGE"

CLOTH_FORGE="$TMP/ext/cloth-config-forge-11.1.136.jar"
PLAYER_FORGE="$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"
CURIOS_FORGE="$TMP/ext/curios-forge-5.14.1+1.20.1.jar"
download_jar 'Cloth Config Forge 11.1.136' 'https://maven.shedaniel.me/me/shedaniel/cloth/cloth-config-forge/11.1.136/cloth-config-forge-11.1.136.jar' "$CLOTH_FORGE"
download_jar 'Player Animator Forge 1.0.2+1.19.4' 'https://maven.kosmx.dev/dev/kosmx/player-anim/player-animation-lib-forge/1.0.2+1.19.4/player-animation-lib-forge-1.0.2+1.19.4.jar' "$PLAYER_FORGE"
download_jar 'Curios Forge 5.14.1+1.20.1' 'https://maven.theillusivec4.top/top/theillusivec4/curios/curios-forge/5.14.1+1.20.1/curios-forge-5.14.1+1.20.1.jar' "$CURIOS_FORGE"

echo '[Paladins] Materializing exact current 3.1.1 authority over the pinned 1.20.1 substrate'
bash "$PORT/materialize_port.sh"
test -f "$PORT/generated/common/java/net/paladins/PaladinsMod.java"
test -f "$PORT/forge/src/main/java/net/paladins/forge/ForgeMod.java"
test -f "$PORT/forge/src/main/java/net/paladins/forge/client/ForgeClientMod.java"
if grep -R -nE 'net\.neoforged|net\.fabricmc\.fabric' "$PORT/generated/common/java" "$PORT/forge/src/main/java"; then
  echo '[Paladins] NeoForge/Fabric runtime API leaked into native Forge source' >&2
  exit 4
fi
if grep -R -nE 'net\.minecraft\.component\.|DataComponentTypes|ComponentType<|PacketCodecs|Registry\.registerReference' "$PORT/generated/common/java"; then
  echo '[Paladins] unported 1.21 API survived compatibility transforms' >&2
  exit 4
fi

ARGS=(
  "-Pshield_api_common_jar=$SHIELD_COMMON"
  "-Parmor_model_api_common_jar=$ARMOR_COMMON"
  "-Prunes_common_jar=$RUNES_COMMON"
  "-Pstructure_pool_api_common_jar=$STRUCTURE_COMMON"
  "-Pspell_power_common_jar=$SPELL_POWER_COMMON"
  "-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON"
  "-Ptiny_config_common_jar=$TINY_COMMON"
  "-Pshield_api_forge_jar=$SHIELD_FORGE"
  "-Parmor_model_api_forge_jar=$ARMOR_FORGE"
  "-Prunes_forge_jar=$RUNES_FORGE"
  "-Pstructure_pool_api_forge_jar=$STRUCTURE_FORGE"
  "-Pspell_power_forge_jar=$SPELL_POWER_FORGE"
  "-Pspell_engine_forge_jar=$SPELL_ENGINE_FORGE"
  "-Ptiny_config_forge_jar=$TINY_FORGE"
  "-Pcloth_config_forge_jar=$CLOTH_FORGE"
  "-Pplayer_animator_forge_jar=$PLAYER_FORGE"
  "-Pcurios_jar=$CURIOS_FORGE"
)

echo '[Paladins] Common compile against named foundation ABI artifacts'
gradle --no-daemon --stacktrace -p "$PORT" clean :common:compileJava "${ARGS[@]}"
echo '[Paladins] Native Forge compile'
gradle --no-daemon --stacktrace -p "$PORT" :forge:compileJava "${ARGS[@]}"
echo '[Paladins] Remapped package boundary'
gradle --no-daemon --stacktrace -p "$PORT" :forge:remapJar "${ARGS[@]}"
JAR="$(resolve_jar 'Paladins Forge release' "$PORT/forge/build/libs" '*-forge-*.jar')"
validate_jar 'Paladins Forge release' "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="paladins"' >/dev/null
if unzip -Z1 "$JAR" | grep -E '^(net/fabric_extras/shield_api/|net/rpg_foundation/armor_api/|net/runes/|net/fabric_extras/structure_pool/|net/spell_power/|net/spell_engine/|net/tiny_config/)'; then
  echo '[Paladins] external foundation classes leaked into Paladins release JAR' >&2
  exit 5
fi
python3 - "$JAR" <<'PY'
import struct,sys,zipfile
owned=total=0; bad=[]; newer=[]
with zipfile.ZipFile(sys.argv[1]) as z:
    for n in z.namelist():
        if not n.endswith('.class'): continue
        d=z.read(n); total+=1
        if len(d)<8 or d[:4] != b'\xca\xfe\xba\xbe': bad.append(n); continue
        major=struct.unpack('>H',d[6:8])[0]
        if major>61: newer.append((n,major))
        if n.startswith('net/paladins/'):
            owned+=1
            if major!=61: bad.append(f'{n}=major{major}')
if not owned: raise SystemExit('[Paladins] no owned classes packaged')
if bad: raise SystemExit('[Paladins] invalid/non-Java17 owned classes: '+', '.join(bad[:30]))
if newer: raise SystemExit('[Paladins] packaged classes newer than Java17: '+', '.join(f'{n}={m}' for n,m in newer[:30]))
print(f'[Paladins] Java gate passed: {owned} owned classes major61; {total} packaged classes <=61.')
PY
sha256sum "$JAR" | tee "$PORT/paladins-forge-1.20.1.sha256"
echo "[Paladins] First compile/package boundary passed: $JAR"
