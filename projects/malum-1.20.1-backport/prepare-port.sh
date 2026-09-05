#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:-$(pwd)}"
OUT="$ROOT/.malum-port-work"
REPORT_DIR="$ROOT/projects/malum-1.20.1-backport/reports"
BASE_REF="1.20.1"
LATEST_REF="1.21.1"
BASE_SHA="c62ce3e2b51ac8daa4d700bed2df9404a68bbce1"
LATEST_SHA="5472ca3deac8f47a2a7b61ca98b4debe5f9482e0"
UPSTREAM="https://github.com/SammySemicolon/Malum-Mod.git"

rm -rf "$OUT"
mkdir -p "$OUT" "$REPORT_DIR"

# Clone both authoritative lanes. 1.20.1 is the proven Forge/Java 17 foundation;
# 1.21.1 is the current feature/content authority (declares Malum 1.9.0).
git clone --depth 1 --branch "$BASE_REF" "$UPSTREAM" "$OUT/base"
git clone --depth 1 --branch "$LATEST_REF" "$UPSTREAM" "$OUT/latest"

ACTUAL_BASE_SHA="$(git -C "$OUT/base" rev-parse HEAD)"
ACTUAL_LATEST_SHA="$(git -C "$OUT/latest" rev-parse HEAD)"
if [[ "$ACTUAL_BASE_SHA" != "$BASE_SHA" ]]; then
  echo "WARNING: 1.20.1 moved: expected $BASE_SHA got $ACTUAL_BASE_SHA" | tee "$REPORT_DIR/source-drift.txt"
fi
if [[ "$ACTUAL_LATEST_SHA" != "$LATEST_SHA" ]]; then
  echo "WARNING: 1.21.1 moved: expected $LATEST_SHA got $ACTUAL_LATEST_SHA" | tee -a "$REPORT_DIR/source-drift.txt"
fi

# Inventory feature/content delta without guessing from compiler output.
git -C "$OUT/latest" diff --no-index --name-status "$OUT/base/src/main" "$OUT/latest/src/main" > "$REPORT_DIR/src-name-status.txt" || true
{
  echo "Malum backport source authority"
  echo "Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17"
  echo "Base branch: $BASE_REF @ $ACTUAL_BASE_SHA"
  echo "Latest branch: $LATEST_REF @ $ACTUAL_LATEST_SHA"
  echo
  echo "Source tree file counts:"
  printf 'base java: '; find "$OUT/base/src/main/java" -type f -name '*.java' | wc -l
  printf 'latest java: '; find "$OUT/latest/src/main/java" -type f -name '*.java' | wc -l
  printf 'base resources: '; find "$OUT/base/src/main/resources" -type f | wc -l
  printf 'latest resources: '; find "$OUT/latest/src/main/resources" -type f | wc -l
  echo
  echo "Latest-only Java packages of special interest:"
  find "$OUT/latest/src/main/java" -type f -name '*.java' | sed "s#^$OUT/latest/src/main/java/##" | grep -E '/(geas|rite|spirit|codex|component|soul|weeping|augment|curio|totem|ritual)/' | head -250 || true
} > "$REPORT_DIR/inventory.txt"

# Candidate: keep the proven 1.20.1 Forge build/tooling surface, overlay the entire
# current mod-owned source/resource tree. Nothing is intentionally dropped here.
cp -a "$OUT/base" "$OUT/port"
rm -rf "$OUT/port/.git"
rm -rf "$OUT/port/src/main/java" "$OUT/port/src/main/resources"
mkdir -p "$OUT/port/src/main"
cp -a "$OUT/latest/src/main/java" "$OUT/port/src/main/java"
cp -a "$OUT/latest/src/main/resources" "$OUT/port/src/main/resources"

# Retarget the proven base build identity to our backport lane.
sed -i \
  -e 's/^forgeVersion=.*/forgeVersion=47.4.23/' \
  -e 's/^modVersion=.*/modVersion=1.9.0-backport.1/' \
  "$OUT/port/gradle.properties"

# Loader/API mechanical bridge. These are syntax/namespace migrations only; semantic
# 1.21 systems (data components, registries, rites/geasa, codecs, etc.) remain intact
# and are deliberately left for compiler-guided faithful ports rather than stubs.
find "$OUT/port/src/main/java" -type f -name '*.java' -print0 | xargs -0 sed -i \
  -e 's/net\.neoforged\.api\.distmarker/net.minecraftforge.api.distmarker/g' \
  -e 's/net\.neoforged\.bus\.api/net.minecraftforge.eventbus.api/g' \
  -e 's/net\.neoforged\.fml/net.minecraftforge.fml/g' \
  -e 's/net\.neoforged\.neoforge/net.minecraftforge/g' \
  -e 's/import net\.minecraftforge\.fml\.common\.EventBusSubscriber;/import net.minecraftforge.fml.common.Mod.EventBusSubscriber;/g' \
  -e 's/ResourceLocation\.fromNamespaceAndPath(\([^,]*\), \([^)]*\))/new ResourceLocation(\1, \2)/g' \
  -e 's/ResourceLocation\.parse(\([^)]*\))/new ResourceLocation(\1)/g'

# NeoForge typed deferred wrappers -> Forge 1.20 RegistryObject equivalents.
find "$OUT/port/src/main/java" -type f -name '*.java' -print0 | xargs -0 perl -0pi -e '
  s/import net\.minecraftforge\.registries\.DeferredBlock;\n//g;
  s/import net\.minecraftforge\.registries\.DeferredItem;\n//g;
  s/import net\.minecraftforge\.registries\.DeferredHolder;\n//g;
  s/DeferredBlock<([^>]+)>/RegistryObject<$1>/g;
  s/DeferredItem<([^>]+)>/RegistryObject<$1>/g;
  s/DeferredHolder<[^,>]+,\s*([^>]+)>/RegistryObject<$1>/g;
  s/DeferredRegister\.Blocks/DeferredRegister<Block>/g;
  s/DeferredRegister\.Items/DeferredRegister<Item>/g;
  s/DeferredRegister\.createBlocks\(([^)]+)\)/DeferredRegister.create(ForgeRegistries.BLOCKS, $1)/g;
  s/DeferredRegister\.createItems\(([^)]+)\)/DeferredRegister.create(ForgeRegistries.ITEMS, $1)/g;
'

# Add common Forge registry imports only where the mechanical bridge produced usages.
while IFS= read -r -d '' f; do
  if grep -q 'RegistryObject<' "$f" && ! grep -q 'net.minecraftforge.registries.RegistryObject' "$f"; then
    sed -i '/^package .*;/a import net.minecraftforge.registries.RegistryObject;' "$f"
  fi
  if grep -q 'ForgeRegistries\.' "$f" && ! grep -q 'net.minecraftforge.registries.ForgeRegistries' "$f"; then
    sed -i '/^package .*;/a import net.minecraftforge.registries.ForgeRegistries;' "$f"
  fi
done < <(find "$OUT/port/src/main/java" -type f -name '*.java' -print0)

# Java 21 collection convenience methods that have direct Java 17 equivalents.
find "$OUT/port/src/main/java" -type f -name '*.java' -print0 | xargs -0 sed -i \
  -e 's/\.getFirst()/\.get(0)/g'

# Record residual high-risk 1.21 surfaces before compiling.
{
  echo "Residual future-version/API surfaces after mechanical bridge:"
  grep -RhoE 'DataComponents\.[A-Z0-9_]+|net\.minecraft\.core\.component\.[A-Za-z0-9_$.]+|StreamCodec|HolderLookup|DeferredRegister\.[A-Za-z]+' "$OUT/port/src/main/java" 2>/dev/null | sort | uniq -c | sort -nr || true
} > "$REPORT_DIR/residual-surfaces.txt"

# Compile with the target-native 1.20.1 build. Capture the complete first semantic map.
set +e
(
  cd "$OUT/port"
  chmod +x gradlew
  ./gradlew --no-daemon compileJava --stacktrace --console=plain
) > "$REPORT_DIR/compile.log" 2>&1
STATUS=$?
set -e

echo "$STATUS" > "$REPORT_DIR/compile-exit-code.txt"
# Compact error index for fast iteration while retaining the full log.
grep -E '(^|: )error:|cannot find symbol|does not exist|incompatible types|no suitable method|method .* cannot be applied' "$REPORT_DIR/compile.log" > "$REPORT_DIR/error-index.txt" || true

# Never claim a pass merely because reports were produced.
exit "$STATUS"
