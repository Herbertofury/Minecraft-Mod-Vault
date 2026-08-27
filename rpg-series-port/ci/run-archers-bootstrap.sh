#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/archers-forge-1.20.1"
TMP="${RUNNER_TEMP:-/tmp}/archers-bootstrap"
rm -rf "$TMP"; mkdir -p "$TMP"

pick_jar() { find "$1" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1; }
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

STRUCTURE="$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$STRUCTURE" :forge:build
STRUCTURE_JAR="$(pick_jar "$STRUCTURE/forge/build/libs")"

BUNDLE="$ROOT/rpg-series-port/bundle-api-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$BUNDLE" :forge:remapJar
BUNDLE_JAR="$(pick_jar "$BUNDLE/forge/build/libs")"

ARMOR="$ROOT/rpg-series-port/armor-model-api-forge-1.20.1"
ARMOR_SHA=a664155a0aab3161cd7e4bf0c1f72512b4ec4949
clone_exact FabricExtras/ArmorModelAPI "$ARMOR_SHA" "$TMP/armor-target"
python3 "$ARMOR/tools/prepare_port.py" "$TMP/armor-target" "$ARMOR/generated"
gradle --no-daemon --stacktrace -p "$ARMOR/generated" :forge:remapJar
ARMOR_JAR="$(pick_jar "$ARMOR/generated/forge/build/libs")"

RANGED_JAR="$(pick_jar "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1/forge/build/libs")"
SPELL_POWER_JAR="$(pick_jar "$ROOT/rpg-series-port/spell_power-forge-1.20.1/forge/build/libs")"
SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"
TINY_JAR="$(pick_jar "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs")"

# Existing external Forge dependencies pinned by the compatibility ledger.
mkdir -p "$TMP/ext"
curl -fsSL 'https://maven.shedaniel.me/me/shedaniel/cloth/cloth-config-forge/11.1.136/cloth-config-forge-11.1.136.jar' -o "$TMP/ext/cloth-config-forge-11.1.136.jar"
curl -fsSL 'https://maven.kosmx.dev/dev/kosmx/player-anim/player-animation-lib-forge/1.0.2+1.19.4/player-animation-lib-forge-1.0.2+1.19.4.jar' -o "$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"
curl -fsSL 'https://maven.theillusivec4.top/top/theillusivec4/curios/curios-forge/5.14.1+1.20.1/curios-forge-5.14.1+1.20.1.jar' -o "$TMP/ext/curios-forge-5.14.1+1.20.1.jar"
CLOTH_JAR="$TMP/ext/cloth-config-forge-11.1.136.jar"
PLAYER_JAR="$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"
CURIOS_JAR="$TMP/ext/curios-forge-5.14.1+1.20.1.jar"

for jar in "$STRUCTURE_JAR" "$BUNDLE_JAR" "$ARMOR_JAR" "$RANGED_JAR" "$SPELL_POWER_JAR" "$SPELL_ENGINE_JAR" "$TINY_JAR" "$CLOTH_JAR" "$PLAYER_JAR" "$CURIOS_JAR"; do
  test -n "$jar" && test -f "$jar"; unzip -tq "$jar"
done

# Raw file dependencies are opaque to Loom's mod remapper. Materialize the exact packaged
# Forge JARs as commit-keyed Maven artifacts so modCompileOnly resolves/remaps Mojmap/SRG
# signatures into the common Yarn named namespace. The copied bytes remain exact.
ABI_REPO="$TMP/abi-repo"
ABI_VERSION="${GITHUB_SHA:0:12}"
mkdir -p "$ABI_REPO"
stage_abi() {
  local artifact="$1" jar="$2"
  local dir="$ABI_REPO/local/archersabi/$artifact/$ABI_VERSION"
  mkdir -p "$dir"
  cp "$jar" "$dir/$artifact-$ABI_VERSION.jar"
  cat > "$dir/$artifact-$ABI_VERSION.pom" <<POM
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>local.archersabi</groupId>
  <artifactId>$artifact</artifactId>
  <version>$ABI_VERSION</version>
</project>
POM
  cmp -s "$jar" "$dir/$artifact-$ABI_VERSION.jar"
  printf '[Archers ABI] %s %s\n' "$artifact" "$(sha256sum "$jar" | cut -d' ' -f1)"
}
stage_abi bundle-api "$BUNDLE_JAR"
stage_abi armor-model-api "$ARMOR_JAR"
stage_abi ranged-weapon-api "$RANGED_JAR"
stage_abi structure-pool-api "$STRUCTURE_JAR"
stage_abi spell-power "$SPELL_POWER_JAR"
stage_abi spell-engine "$SPELL_ENGINE_JAR"
stage_abi tiny-config "$TINY_JAR"
stage_abi cloth-config "$CLOTH_JAR"
stage_abi player-animator "$PLAYER_JAR"

echo '[Archers] Materializing exact current 3.1.1 content + explicit 1.20.1 transforms'
bash "$PORT/materialize_port.sh"
test -f "$PORT/generated/common/java/net/archers/ArchersMod.java"
test -f "$PORT/generated/common/java/net/archers/item/Quivers.java"
test -f "$PORT/generated/common/java/net/archers/item/misc/AutoFireHook.java"
test -f "$PORT/generated/common/resources/archers.mixins.json"

# Known 1.21-only systems already translated by the deterministic transform must not survive.
if grep -R -nE 'net\.minecraft\.component\.|DataComponentTypes|ComponentType<|PacketCodecs|Registry\.registerReference' "$PORT/generated/common/java"; then
  echo '[Archers] unported 1.21 component/reference-registry API survived compatibility transforms' >&2
  exit 2
fi
if grep -R -nE 'net\.neoforged|net\.fabricmc\.fabric' "$PORT/generated/common/java" "$PORT/forge/src/main/java"; then
  echo '[Archers] NeoForge/Fabric runtime API leaked into native Forge source' >&2
  exit 2
fi

echo '[Archers] Compiling current 3.1.1 against Loom-remapped views of separate packaged Forge dependencies'
ARGS=(
  "-Parchers_abi_repo=$ABI_REPO"
  "-Parchers_abi_version=$ABI_VERSION"
  "-Pcurios_jar=$CURIOS_JAR"
)
gradle --no-daemon --stacktrace -p "$PORT" clean :common:compileJava :forge:compileJava "${ARGS[@]}"

echo '[Archers] Remapped package boundary'
gradle --no-daemon --stacktrace -p "$PORT" :forge:remapJar "${ARGS[@]}"
JAR="$(pick_jar "$PORT/forge/build/libs")"
test -f "$JAR"; unzip -tq "$JAR"
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
