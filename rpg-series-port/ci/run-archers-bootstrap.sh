#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/archers-forge-1.20.1"
TMP="${RUNNER_TEMP:-/tmp}/archers-bootstrap"
ABI_MANIFEST="$PORT/archers-abi.sha256"
rm -rf "$TMP"; mkdir -p "$TMP"
: > "$ABI_MANIFEST"

resolve_jar() {
  local label="$1" dir="$2" name_glob="$3"
  local -a candidates=()
  if [[ ! -d "$dir" ]]; then
    echo "[Archers ABI] $label directory missing: $dir" >&2
    return 1
  fi
  mapfile -t candidates < <(find "$dir" -maxdepth 1 -type f -name "$name_glob" \
    ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' ! -name '*transformProduction*' -print | sort)
  if (( ${#candidates[@]} != 1 )); then
    echo "[Archers ABI] $label expected exactly one $name_glob release candidate, found ${#candidates[@]} in $dir" >&2
    find "$dir" -maxdepth 1 -type f -name '*.jar' -printf '[Archers ABI] candidate: %f\n' | sort >&2 || true
    return 1
  fi
  echo "[Archers ABI] resolved $label -> ${candidates[0]}" >&2
  printf '%s\n' "${candidates[0]}"
}

validate_jar() {
  local label="$1" jar="$2"
  if [[ -z "$jar" || ! -f "$jar" ]]; then
    echo "[Archers ABI] $label missing: ${jar:-<empty>}" >&2
    return 1
  fi
  if ! unzip -tq "$jar" >/dev/null; then
    echo "[Archers ABI] $label is not a valid JAR/ZIP: $jar" >&2
    return 1
  fi
  local hash
  hash="$(sha256sum "$jar" | awk '{print $1}')"
  printf '%s  %s\n' "$hash" "$jar" >> "$ABI_MANIFEST"
  echo "[Archers ABI] validated $label sha256=$hash"
}

download_jar() {
  local label="$1" url="$2" out="$3"
  echo "[Archers ABI] downloading $label"
  curl --retry 2 --retry-delay 1 --retry-connrefused -fsSL "$url" -o "$out"
  validate_jar "$label" "$out"
}

clone_exact() {
  local repo="$1" sha="$2" dst="$3"
  git init -q "$dst"
  git -C "$dst" remote add origin "https://github.com/$repo.git"
  git -C "$dst" fetch -q --depth=1 origin "$sha"
  git -C "$dst" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dst" rev-parse HEAD)" = "$sha"
}

echo '[Archers] Reconstructing proven RPG foundations (build/package only)'
bash "$ROOT/rpg-series-port/ci/build-spell-engine-foundation.sh"

RANGED="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"
TINY="$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated"
SPELL_ENGINE_BUILD="$ROOT/.spell-engine-build"
RANGED_COMMON="$(resolve_jar 'Ranged Weapon API common' "$RANGED/common/build/libs" '*-common-*.jar')"
RANGED_FORGE="$(resolve_jar 'Ranged Weapon API Forge' "$RANGED/forge/build/libs" '*-forge-*.jar')"
SPELL_POWER_COMMON="$(resolve_jar 'Spell Power common' "$SPELL_POWER/common/build/libs" '*-common-*.jar')"
SPELL_POWER_FORGE="$(resolve_jar 'Spell Power Forge' "$SPELL_POWER/forge/build/libs" '*-forge-*.jar')"
TINY_COMMON="$(resolve_jar 'TinyConfig common' "$TINY/common/build/libs" '*-common-*.jar')"
TINY_FORGE="$(resolve_jar 'TinyConfig Forge' "$TINY/forge/build/libs" '*-forge-*.jar')"
SPELL_ENGINE_COMMON="$(resolve_jar 'Spell Engine common' "$SPELL_ENGINE_BUILD/common/build/libs" '*-common-*.jar')"
SPELL_ENGINE_FORGE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"
validate_jar 'Ranged Weapon API common' "$RANGED_COMMON"
validate_jar 'Ranged Weapon API Forge' "$RANGED_FORGE"
validate_jar 'Spell Power common' "$SPELL_POWER_COMMON"
validate_jar 'Spell Power Forge' "$SPELL_POWER_FORGE"
validate_jar 'TinyConfig common' "$TINY_COMMON"
validate_jar 'TinyConfig Forge' "$TINY_FORGE"
validate_jar 'Spell Engine common' "$SPELL_ENGINE_COMMON"
validate_jar 'Spell Engine Forge' "$SPELL_ENGINE_FORGE"

STRUCTURE="$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$STRUCTURE" :common:jar :forge:build
STRUCTURE_COMMON="$(resolve_jar 'Structure Pool API common' "$STRUCTURE/common/build/libs" '*-common-*.jar')"
STRUCTURE_FORGE="$(resolve_jar 'Structure Pool API Forge' "$STRUCTURE/forge/build/libs" '*-forge-*.jar')"
validate_jar 'Structure Pool API common' "$STRUCTURE_COMMON"
validate_jar 'Structure Pool API Forge' "$STRUCTURE_FORGE"

BUNDLE="$ROOT/rpg-series-port/bundle-api-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$BUNDLE" :common:jar :forge:remapJar
BUNDLE_COMMON="$(resolve_jar 'Bundle API common' "$BUNDLE/common/build/libs" '*-common-*.jar')"
BUNDLE_FORGE="$(resolve_jar 'Bundle API Forge' "$BUNDLE/forge/build/libs" '*-forge-*.jar')"
validate_jar 'Bundle API common' "$BUNDLE_COMMON"
validate_jar 'Bundle API Forge' "$BUNDLE_FORGE"

ARMOR="$ROOT/rpg-series-port/armor-model-api-forge-1.20.1"
ARMOR_SHA=a664155a0aab3161cd7e4bf0c1f72512b4ec4949
clone_exact FabricExtras/ArmorModelAPI "$ARMOR_SHA" "$TMP/armor-target"
python3 "$ARMOR/tools/prepare_port.py" "$TMP/armor-target" "$ARMOR/generated"
gradle --no-daemon --stacktrace -p "$ARMOR/generated" :common:jar :forge:remapJar
ARMOR_COMMON="$(resolve_jar 'Armor Model API common' "$ARMOR/generated/common/build/libs" '*-common-*.jar')"
ARMOR_FORGE="$(resolve_jar 'Armor Model API Forge' "$ARMOR/generated/forge/build/libs" '*-forge-*.jar')"
validate_jar 'Armor Model API common' "$ARMOR_COMMON"
validate_jar 'Armor Model API Forge' "$ARMOR_FORGE"

mkdir -p "$TMP/ext"
CLOTH_FORGE="$TMP/ext/cloth-config-forge-11.1.136.jar"
PLAYER_FORGE="$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"
CURIOS_FORGE="$TMP/ext/curios-forge-5.14.1+1.20.1.jar"
download_jar 'Cloth Config Forge 11.1.136' 'https://maven.shedaniel.me/me/shedaniel/cloth/cloth-config-forge/11.1.136/cloth-config-forge-11.1.136.jar' "$CLOTH_FORGE"
download_jar 'Player Animator Forge 1.0.2+1.19.4' 'https://maven.kosmx.dev/dev/kosmx/player-anim/player-animation-lib-forge/1.0.2+1.19.4/player-animation-lib-forge-1.0.2+1.19.4.jar' "$PLAYER_FORGE"
download_jar 'Curios Forge 5.14.1+1.20.1' 'https://maven.theillusivec4.top/top/theillusivec4/curios/curios-forge/5.14.1+1.20.1/curios-forge-5.14.1+1.20.1.jar' "$CURIOS_FORGE"

echo '[Archers] Materializing exact current 3.1.1 content + explicit 1.20.1 transforms'
bash "$PORT/materialize_port.sh"
test -f "$PORT/generated/common/java/net/archers/ArchersMod.java"
test -f "$PORT/generated/common/java/net/archers/item/Quivers.java"
test -f "$PORT/generated/common/java/net/archers/item/misc/AutoFireHook.java"
test -f "$PORT/generated/common/resources/archers.mixins.json"

if grep -R -nE 'net\.minecraft\.component\.|DataComponentTypes|ComponentType<|PacketCodecs|Registry\.registerReference' "$PORT/generated/common/java"; then
  echo '[Archers] unported 1.21 component/reference-registry API survived compatibility transforms' >&2
  exit 2
fi
if grep -R -nE 'net\.neoforged|net\.fabricmc\.fabric' "$PORT/generated/common/java" "$PORT/forge/src/main/java"; then
  echo '[Archers] NeoForge/Fabric runtime API leaked into native Forge source' >&2
  exit 2
fi

ARGS=(
  "-Pbundle_api_common_jar=$BUNDLE_COMMON"
  "-Parmor_model_api_common_jar=$ARMOR_COMMON"
  "-Pranged_weapon_api_common_jar=$RANGED_COMMON"
  "-Pstructure_pool_api_common_jar=$STRUCTURE_COMMON"
  "-Pspell_power_common_jar=$SPELL_POWER_COMMON"
  "-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON"
  "-Ptiny_config_common_jar=$TINY_COMMON"
  "-Pbundle_api_forge_jar=$BUNDLE_FORGE"
  "-Parmor_model_api_forge_jar=$ARMOR_FORGE"
  "-Pranged_weapon_api_forge_jar=$RANGED_FORGE"
  "-Pstructure_pool_api_forge_jar=$STRUCTURE_FORGE"
  "-Pspell_power_forge_jar=$SPELL_POWER_FORGE"
  "-Pspell_engine_forge_jar=$SPELL_ENGINE_FORGE"
  "-Ptiny_config_forge_jar=$TINY_FORGE"
  "-Pcloth_config_forge_jar=$CLOTH_FORGE"
  "-Pplayer_animator_forge_jar=$PLAYER_FORGE"
  "-Pcurios_jar=$CURIOS_FORGE"
)

echo '[Archers] Common compile: separately built named common ABI artifacts'
gradle --no-daemon --stacktrace -p "$PORT" clean :common:compileJava "${ARGS[@]}"

echo '[Archers] Forge compile: exact separately packaged Forge ABI artifacts'
gradle --no-daemon --stacktrace -p "$PORT" :forge:compileJava "${ARGS[@]}"

echo '[Archers] Remapped package boundary'
gradle --no-daemon --stacktrace -p "$PORT" :forge:remapJar "${ARGS[@]}"
JAR="$(resolve_jar 'Archers Forge release' "$PORT/forge/build/libs" '*-forge-*.jar')"
validate_jar 'Archers Forge release' "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="archers"' >/dev/null
if unzip -Z1 "$JAR" | grep -E '^(com/github/theredbrain/bundleapi/|net/spell_engine/|net/spell_power/|net/fabric_extras/ranged_weapon/|net/fabric_extras/structure_pool/|net/rpg_foundation/armor_api/|net/tiny_config/)'; then
  echo '[Archers] external dependency classes leaked into Archers release JAR' >&2
  exit 3
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
        if n.startswith('net/archers/'):
            owned+=1
            if major!=61: bad.append(f'{n}=major{major}')
if not owned: raise SystemExit('[Archers] no owned classes packaged')
if bad: raise SystemExit('[Archers] invalid/non-Java17 owned classes: '+', '.join(bad[:30]))
if newer: raise SystemExit('[Archers] packaged classes newer than Java17: '+', '.join(f'{n}={m}' for n,m in newer[:30]))
print(f'[Archers] Java gate passed: {owned} owned classes major61; {total} packaged classes <=61.')
PY
sha256sum "$JAR" | tee "$PORT/archers.sha256"
echo "[Archers] First compile/package boundary passed: $JAR"
