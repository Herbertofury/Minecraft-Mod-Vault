#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/archers-forge-1.20.1"
TMP="${RUNNER_TEMP:-/tmp}/archers-bootstrap"
rm -rf "$TMP"; mkdir -p "$TMP"

pick_release_jar() { find "$1" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' ! -name '*transformProduction*' | sort | head -n1; }
pick_common_jar() { find "$1" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*transformProduction*' ! -name '*javadoc*' | sort | head -n1; }
clone_exact() {
  local repo="$1" sha="$2" dst="$3"
  git init -q "$dst"
  git -C "$dst" remote add origin "https://github.com/$repo.git"
  git -C "$dst" fetch -q --depth=1 origin "$sha"
  git -C "$dst" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dst" rev-parse HEAD)" = "$sha"
}
check_jar() { test -n "$1" && test -f "$1" && unzip -tq "$1" >/dev/null; }

echo '[Archers] Reconstructing proven RPG foundations (build/package only)'
bash "$ROOT/rpg-series-port/ci/build-spell-engine-foundation.sh"

STRUCTURE="$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$STRUCTURE" :common:jar :forge:build
STRUCTURE_COMMON="$(pick_common_jar "$STRUCTURE/common/build/libs")"
STRUCTURE_FORGE="$(pick_release_jar "$STRUCTURE/forge/build/libs")"

BUNDLE="$ROOT/rpg-series-port/bundle-api-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$BUNDLE" :common:jar :forge:remapJar
BUNDLE_COMMON="$(pick_common_jar "$BUNDLE/common/build/libs")"
BUNDLE_FORGE="$(pick_release_jar "$BUNDLE/forge/build/libs")"

ARMOR="$ROOT/rpg-series-port/armor-model-api-forge-1.20.1"
ARMOR_SHA=a664155a0aab3161cd7e4bf0c1f72512b4ec4949
clone_exact FabricExtras/ArmorModelAPI "$ARMOR_SHA" "$TMP/armor-target"
python3 "$ARMOR/tools/prepare_port.py" "$TMP/armor-target" "$ARMOR/generated"
gradle --no-daemon --stacktrace -p "$ARMOR/generated" :common:jar :forge:remapJar
ARMOR_COMMON="$(pick_common_jar "$ARMOR/generated/common/build/libs")"
ARMOR_FORGE="$(pick_release_jar "$ARMOR/generated/forge/build/libs")"

RANGED="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"
TINY="$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated"
SPELL_ENGINE_BUILD="$ROOT/.spell-engine-build"
RANGED_COMMON="$(pick_common_jar "$RANGED/common/build/libs")"
RANGED_FORGE="$(pick_release_jar "$RANGED/forge/build/libs")"
SPELL_POWER_COMMON="$(pick_common_jar "$SPELL_POWER/common/build/libs")"
SPELL_POWER_FORGE="$(pick_release_jar "$SPELL_POWER/forge/build/libs")"
TINY_COMMON="$(pick_common_jar "$TINY/common/build/libs")"
TINY_FORGE="$(pick_release_jar "$TINY/forge/build/libs")"
SPELL_ENGINE_COMMON="$(pick_common_jar "$SPELL_ENGINE_BUILD/common/build/libs")"
SPELL_ENGINE_FORGE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"

# Exact external Forge runtime/loader artifacts. Common compiles against their Fabric/intermediary
# counterparts through Loom; the Forge module independently checks these packaged Forge ABIs.
mkdir -p "$TMP/ext"
curl -fsSL 'https://maven.shedaniel.me/me/shedaniel/cloth/cloth-config-forge/11.1.136/cloth-config-forge-11.1.136.jar' -o "$TMP/ext/cloth-config-forge-11.1.136.jar"
curl -fsSL 'https://maven.kosmx.dev/dev/kosmx/player-anim/player-animation-lib-forge/1.0.2+1.19.4/player-animation-lib-forge-1.0.2+1.19.4.jar' -o "$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"
curl -fsSL 'https://maven.theillusivec4.top/top/theillusivec4/curios/curios-forge/5.14.1+1.20.1/curios-forge-5.14.1+1.20.1.jar' -o "$TMP/ext/curios-forge-5.14.1+1.20.1.jar"
CLOTH_FORGE="$TMP/ext/cloth-config-forge-11.1.136.jar"
PLAYER_FORGE="$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"
CURIOS_FORGE="$TMP/ext/curios-forge-5.14.1+1.20.1.jar"

COMMON_JARS=("$BUNDLE_COMMON" "$ARMOR_COMMON" "$RANGED_COMMON" "$STRUCTURE_COMMON" "$SPELL_POWER_COMMON" "$SPELL_ENGINE_COMMON" "$TINY_COMMON")
FORGE_JARS=("$BUNDLE_FORGE" "$ARMOR_FORGE" "$RANGED_FORGE" "$STRUCTURE_FORGE" "$SPELL_POWER_FORGE" "$SPELL_ENGINE_FORGE" "$TINY_FORGE" "$CLOTH_FORGE" "$PLAYER_FORGE" "$CURIOS_FORGE")
for jar in "${COMMON_JARS[@]}" "${FORGE_JARS[@]}"; do check_jar "$jar"; done

echo '[Archers ABI] common named artifacts:'
for jar in "${COMMON_JARS[@]}"; do sha256sum "$jar"; done
echo '[Archers ABI] packaged Forge artifacts:'
for jar in "${FORGE_JARS[@]}"; do sha256sum "$jar"; done

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
JAR="$(pick_release_jar "$PORT/forge/build/libs")"
check_jar "$JAR"
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
