#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:-$(pwd)}"
OUT="$ROOT/.malum-port-work"
REPORT_DIR="$ROOT/projects/malum-1.20.1-backport/reports"
UPSTREAM="https://github.com/SammySemicolon/Malum-Mod.git"

# Immutable source authorities.
# 1.20.1 = proven Forge/Java 17 architectural base (Malum 1.6.7).
BASE_SHA="c62ce3e2b51ac8daa4d700bed2df9404a68bbce1"
# Last Dec-08 source state that still declares the released Malum 1.8.2 line.
RELEASE_SHA="03b743a37f3eeb0cc7f4364f0730e1f135f78408"
# Current 1.21.1 development authority captured when this backport was started.
DEV_SHA="5472ca3deac8f47a2a7b61ca98b4debe5f9482e0"

rm -rf "$OUT" "$REPORT_DIR"
mkdir -p "$OUT" "$REPORT_DIR"

fetch_exact() {
  local sha="$1" dest="$2"
  git init -q "$dest"
  git -C "$dest" remote add origin "$UPSTREAM"
  git -C "$dest" fetch -q --depth 1 origin "$sha"
  git -C "$dest" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dest" rev-parse HEAD)" = "$sha"
}

fetch_exact "$BASE_SHA" "$OUT/base"
fetch_exact "$RELEASE_SHA" "$OUT/release"
fetch_exact "$DEV_SHA" "$OUT/dev"

BASE_VERSION="$(sed -n 's/^modVersion=//p' "$OUT/base/gradle.properties")"
RELEASE_VERSION="$(sed -n 's/^mod_version=//p' "$OUT/release/gradle.properties")"
DEV_VERSION="$(sed -n 's/^mod_version=//p' "$OUT/dev/gradle.properties")"

[[ "$BASE_VERSION" == "1.6.7" ]] || { echo "Unexpected base version: $BASE_VERSION"; exit 90; }
[[ "$RELEASE_VERSION" == "1.8.2" ]] || { echo "Unexpected released version: $RELEASE_VERSION"; exit 91; }
[[ "$DEV_VERSION" == "1.9.0" ]] || { echo "Unexpected dev version: $DEV_VERSION"; exit 92; }

# Three-lane inventory: released 1.8.2 is mandatory parity; 1.9.0 is an explicitly
# separate follow-up lane so unreleased behavior never silently replaces release parity.
git diff --no-index --name-status "$OUT/base/src/main" "$OUT/release/src/main" > "$REPORT_DIR/base-to-1.8.2-name-status.txt" || true
git diff --no-index --name-status "$OUT/release/src/main" "$OUT/dev/src/main" > "$REPORT_DIR/1.8.2-to-1.9.0-name-status.txt" || true
{
  echo "Malum Forge 1.20.1 backport authority ledger"
  echo "Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17"
  echo "Forge base: Malum $BASE_VERSION @ $BASE_SHA"
  echo "Released parity authority: Malum $RELEASE_VERSION @ $RELEASE_SHA"
  echo "Optional dev-forward authority: Malum $DEV_VERSION @ $DEV_SHA"
  echo
  for lane in base release dev; do
    printf '%s java: ' "$lane"; find "$OUT/$lane/src/main/java" -type f -name '*.java' | wc -l
    printf '%s resources: ' "$lane"; find "$OUT/$lane/src/main/resources" -type f | wc -l
  done
  echo
  echo "1.8.2 subsystems of special interest:"
  find "$OUT/release/src/main/java" -type f -name '*.java' \
    | sed "s#^$OUT/release/src/main/java/##" \
    | grep -Ei '/(geas|rite|spirit|codex|component|soul|weeping|augment|curio|totem|ritual|parallel)/' \
    | sort | head -400 || true
} > "$REPORT_DIR/inventory.txt"

# Build the RELEASE-PARITY candidate first. The old 1.20.1 project supplies ForgeGradle,
# Java 17 and 1.20-native dependency wiring; all Malum-owned 1.8.2 source/resources are
# overlaid wholesale. No gameplay subsystem is intentionally discarded here.
cp -a "$OUT/base" "$OUT/port"
rm -rf "$OUT/port/.git"
rm -rf "$OUT/port/src/main/java" "$OUT/port/src/main/resources"
mkdir -p "$OUT/port/src/main"
cp -a "$OUT/release/src/main/java" "$OUT/port/src/main/java"
cp -a "$OUT/release/src/main/resources" "$OUT/port/src/main/resources"

# Restore Forge 1.20 loader metadata after the 1.21 resource overlay.
rm -f "$OUT/port/src/main/resources/META-INF/neoforge.mods.toml"
mkdir -p "$OUT/port/src/main/resources/META-INF"
cp "$OUT/base/src/main/resources/META-INF/mods.toml" "$OUT/port/src/main/resources/META-INF/mods.toml"

# Target-native build identity/dependencies. We deliberately begin with the known-good
# 1.20 Lodestone/Curios line; only proven API blockers can widen into a Lodestone sub-port.
sed -i \
  -e 's/^forgeVersion=.*/forgeVersion=47.4.23/' \
  -e 's/^modVersion=.*/modVersion=1.8.2-backport.1/' \
  -e 's/^lodestoneVersion=.*/lodestoneVersion=1.6.4.1.256/' \
  -e 's/^curiosVersion=.*/curiosVersion=5.14.1+1.20.1/' \
  "$OUT/port/gradle.properties"

# Mechanical loader namespace bridge only. Semantic systems are kept and will be
# translated faithfully from compiler evidence instead of being stubbed out.
find "$OUT/port/src/main/java" -type f -name '*.java' -print0 | xargs -0 sed -i \
  -e 's/net\.neoforged\.api\.distmarker/net.minecraftforge.api.distmarker/g' \
  -e 's/net\.neoforged\.bus\.api/net.minecraftforge.eventbus.api/g' \
  -e 's/net\.neoforged\.fml/net.minecraftforge.fml/g' \
  -e 's/net\.neoforged\.neoforge/net.minecraftforge/g' \
  -e 's/import net\.minecraftforge\.fml\.common\.EventBusSubscriber;/import net.minecraftforge.fml.common.Mod.EventBusSubscriber;/g' \
  -e 's/ResourceLocation\.fromNamespaceAndPath(\([^,]*\), \([^)]*\))/new ResourceLocation(\1, \2)/g' \
  -e 's/ResourceLocation\.parse(\([^)]*\))/new ResourceLocation(\1)/g' \
  -e 's/\.getFirst()/\.get(0)/g'

# NeoForge 1.21 typed deferred wrappers -> Forge 1.20 RegistryObject equivalents.
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

while IFS= read -r -d '' f; do
  if grep -q 'RegistryObject<' "$f" && ! grep -q 'net.minecraftforge.registries.RegistryObject' "$f"; then
    sed -i '/^package .*;/a import net.minecraftforge.registries.RegistryObject;' "$f"
  fi
  if grep -q 'ForgeRegistries\.' "$f" && ! grep -q 'net.minecraftforge.registries.ForgeRegistries' "$f"; then
    sed -i '/^package .*;/a import net.minecraftforge.registries.ForgeRegistries;' "$f"
  fi
done < <(find "$OUT/port/src/main/java" -type f -name '*.java' -print0)

# Java 21 pattern-switch syntax -> Java 17 control flow. These rewrites preserve the
# exact branch order and outputs from released 1.8.2; they only remove language syntax
# unavailable to the Forge 1.20.1 / Java 17 target.
python3 - "$OUT/port/src/main/java" <<'PY'
from pathlib import Path
import sys

root = Path(sys.argv[1])

crafting = root / "com/sammy/malum/client/screen/codex/pages/recipe/vanilla/CraftingPage.java"
crafting_text = crafting.read_text()
crafting_old = '''        return switch (tool.getItem()) {
            case SwordItem swordItem ->
                    new CraftingPage(tool, empty, metal, empty, empty, metal, empty, empty, stick, empty);
            case AxeItem axeItem ->
                    new CraftingPage(tool, metal, metal, empty, metal, stick, empty, empty, stick, empty);
            case HoeItem hoeItem ->
                    new CraftingPage(tool, metal, metal, empty, empty, stick, empty, empty, stick, empty);
            case ShovelItem shovelItem ->
                    new CraftingPage(tool, empty, metal, empty, empty, stick, empty, empty, stick, empty);
            case PickaxeItem pickaxeItem ->
                    new CraftingPage(tool, metal, metal, metal, empty, stick, empty, empty, stick, empty);
            default -> null;
        };'''
crafting_new = '''        Item toolItem = tool.getItem();
        if (toolItem instanceof SwordItem) {
            return new CraftingPage(tool, empty, metal, empty, empty, metal, empty, empty, stick, empty);
        }
        if (toolItem instanceof AxeItem) {
            return new CraftingPage(tool, metal, metal, empty, metal, stick, empty, empty, stick, empty);
        }
        if (toolItem instanceof HoeItem) {
            return new CraftingPage(tool, metal, metal, empty, empty, stick, empty, empty, stick, empty);
        }
        if (toolItem instanceof ShovelItem) {
            return new CraftingPage(tool, empty, metal, empty, empty, stick, empty, empty, stick, empty);
        }
        if (toolItem instanceof PickaxeItem) {
            return new CraftingPage(tool, metal, metal, metal, empty, stick, empty, empty, stick, empty);
        }
        return null;'''
if crafting_old not in crafting_text:
    raise SystemExit("CraftingPage pattern switch signature changed; refusing a lossy rewrite")
crafting.write_text(crafting_text.replace(crafting_old, crafting_new, 1))

ether = root / "com/sammy/malum/common/block/ether/EtherBlockEntity.java"
ether_text = ether.read_text()
ether_old = '''        switch (getBlockState().getBlock()) { //TODO: this sucks
            case EtherWallTorchBlock etherWallTorchBlock -> {
                float offset = 0.15f;
                Direction direction = getBlockState().getValue(WallTorchBlock.FACING);
                x -= direction.getNormal().getX() * offset;
                y += 0.4f;
                z -= direction.getNormal().getZ() * offset;
            }
            case EtherTorchBlock etherTorchBlock -> y += 0.3f;
            case EtherBrazierBlock etherBrazierBlock -> y -= 0.05f;
            default -> {
            }
        }'''
ether_new = '''        Block block = getBlockState().getBlock();
        if (block instanceof EtherWallTorchBlock) {
            float offset = 0.15f;
            Direction direction = getBlockState().getValue(WallTorchBlock.FACING);
            x -= direction.getNormal().getX() * offset;
            y += 0.4f;
            z -= direction.getNormal().getZ() * offset;
        } else if (block instanceof EtherTorchBlock) {
            y += 0.3f;
        } else if (block instanceof EtherBrazierBlock) {
            y -= 0.05f;
        }'''
if ether_old not in ether_text:
    raise SystemExit("EtherBlockEntity pattern switch signature changed; refusing a lossy rewrite")
ether.write_text(ether_text.replace(ether_old, ether_new, 1))
PY

# Resource-layout bridge that is unambiguously version-specific. 1.21 uses singular
# data directories while 1.20.1 expects plural recipe/loot-table directories.
while IFS= read -r -d '' d; do
  parent="$(dirname "$d")"
  name="$(basename "$d")"
  case "$name" in
    recipe) mv "$d" "$parent/recipes" ;;
    loot_table) mv "$d" "$parent/loot_tables" ;;
  esac
done < <(find "$OUT/port/src/main/resources/data" -type d \( -name recipe -o -name loot_table \) -print0 2>/dev/null || true)

# High-risk residuals are an acceptance map, not a deletion list.
{
  echo "Residual future-version/API surfaces after mechanical bridge:"
  grep -RhoE 'DataComponents\.[A-Z0-9_]+|net\.minecraft\.core\.component\.[A-Za-z0-9_$.]+|StreamCodec|HolderLookup|DataComponentType|DeferredRegister\.[A-Za-z]+' \
    "$OUT/port/src/main/java" 2>/dev/null | sort | uniq -c | sort -nr || true
} > "$REPORT_DIR/residual-surfaces.txt"

# Reproducible source delta: this compact patch is enough to recreate the current
# mechanical port from the pinned 1.8.2 source without committing Malum's full assets.
diff -ruN "$OUT/release/src/main/java" "$OUT/port/src/main/java" > "$REPORT_DIR/mechanical-java-bridge.patch" || true

# First semantic compile gate.
set +e
(
  cd "$OUT/port"
  chmod +x gradlew
  ./gradlew --no-daemon compileJava --stacktrace --console=plain
) > "$REPORT_DIR/compile.log" 2>&1
STATUS=$?
set -e

echo "$STATUS" > "$REPORT_DIR/compile-exit-code.txt"
grep -E '(^|: )error:|cannot find symbol|does not exist|incompatible types|no suitable method|method .* cannot be applied' \
  "$REPORT_DIR/compile.log" > "$REPORT_DIR/error-index.txt" || true

exit "$STATUS"
