#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:-$(pwd)}"
OUT="$ROOT/.wayward-port-work"
REPORT_DIR="$ROOT/projects/wayward-attributes-1.20.1-backport/reports"
UPSTREAM="https://github.com/LodestarMC/WaywardAttributes.git"
SOURCE_SHA="e49971d895ecc4c0e53fa45b4f1c4e5b375775d7"

rm -rf "$OUT" "$REPORT_DIR"
mkdir -p "$OUT" "$REPORT_DIR"

git init -q "$OUT/source"
git -C "$OUT/source" remote add origin "$UPSTREAM"
git -C "$OUT/source" fetch -q --depth 1 origin "$SOURCE_SHA"
git -C "$OUT/source" checkout -q --detach FETCH_HEAD
test "$(git -C "$OUT/source" rev-parse HEAD)" = "$SOURCE_SHA"

SOURCE_VERSION="$(sed -n 's/^mod_version=//p' "$OUT/source/gradle.properties")"
[[ "$SOURCE_VERSION" == "1.1" ]] || { echo "Unexpected Wayward source version: $SOURCE_VERSION"; exit 90; }

mkdir -p "$OUT/port/src/main"
cp -a "$OUT/source/src/main/java" "$OUT/port/src/main/java"
cp -a "$OUT/source/src/main/resources" "$OUT/port/src/main/resources"
if [[ -d "$OUT/source/src/generated/resources" ]]; then
  mkdir -p "$OUT/port/src/generated"
  cp -a "$OUT/source/src/generated/resources" "$OUT/port/src/generated/resources"
fi

cat > "$OUT/port/settings.gradle" <<'EOF'
pluginManagement {
    repositories {
        gradlePluginPortal()
        maven { url = 'https://maven.minecraftforge.net/' }
    }
}
rootProject.name = 'WaywardAttributes-Forge-1.20.1'
EOF

cat > "$OUT/port/gradle.properties" <<'EOF'
org.gradle.jvmargs=-Xmx4G
org.gradle.daemon=false
minecraft_version=1.20.1
forge_version=47.4.23
mod_id=wayward_attributes
mod_version=1.1.30-backport.1
mod_group_id=team.lodestar.wayward_attributes
lodestone_version=1.6.4.1.256
EOF

cat > "$OUT/port/build.gradle" <<'EOF'
plugins {
    id 'java-library'
    id 'maven-publish'
    id 'net.minecraftforge.gradle' version '[6.0,6.2)'
}
version = "${minecraft_version}-${mod_version}"
group = mod_group_id
base { archivesName = mod_id }
java.toolchain.languageVersion = JavaLanguageVersion.of(17)

minecraft {
    mappings channel: 'official', version: minecraft_version
    copyIdeResources = true
    if (file('src/main/resources/META-INF/accesstransformer.cfg').exists()) {
        accessTransformer = file('src/main/resources/META-INF/accesstransformer.cfg')
    }
    runs {
        client { workingDirectory project.file('run'); mods { wayward_attributes { source sourceSets.main } } }
        server { workingDirectory project.file('run-server'); args '--nogui'; mods { wayward_attributes { source sourceSets.main } } }
    }
}
sourceSets.main.resources.srcDir 'src/generated/resources'

repositories {
    mavenCentral()
    maven { url = 'https://maven.blamejared.com/' }
}

dependencies {
    minecraft "net.minecraftforge:forge:${minecraft_version}-${forge_version}"
    implementation fg.deobf("team.lodestar.lodestone:lodestone:${minecraft_version}-${lodestone_version}")
}

tasks.withType(JavaCompile).configureEach { options.encoding = 'UTF-8' }
jar.finalizedBy('reobfJar')
EOF

# Replace NeoForge loader metadata with target-native Forge metadata.
rm -f "$OUT/port/src/main/resources/META-INF/neoforge.mods.toml"
rm -rf "$OUT/port/src/main/templates"
mkdir -p "$OUT/port/src/main/resources/META-INF"
cat > "$OUT/port/src/main/resources/META-INF/mods.toml" <<'EOF'
modLoader="javafml"
loaderVersion="[47,)"
license="LGPL-3.0-or-later"
[[mods]]
modId="wayward_attributes"
version="1.20.1-1.1.30-backport.1"
displayName="Wayward Attributes"
authors="Lodestar"
description='''Attribute presentation and ranged/tool attribute extensions backported for Malum on Forge 1.20.1.'''
[[dependencies.wayward_attributes]]
modId="forge"
mandatory=true
versionRange="[47.4.23,)"
ordering="NONE"
side="BOTH"
[[dependencies.wayward_attributes]]
modId="minecraft"
mandatory=true
versionRange="[1.20.1,1.21)"
ordering="NONE"
side="BOTH"
[[dependencies.wayward_attributes]]
modId="lodestone"
mandatory=true
versionRange="[1.6.4.1,)"
ordering="AFTER"
side="BOTH"
EOF

find "$OUT/port/src/main/java" -type f -name '*.java' -print0 | xargs -0 sed -i \
  -e 's/net\.neoforged\.api\.distmarker/net.minecraftforge.api.distmarker/g' \
  -e 's/net\.neoforged\.bus\.api/net.minecraftforge.eventbus.api/g' \
  -e 's/net\.neoforged\.fml/net.minecraftforge.fml/g' \
  -e 's/net\.neoforged\.neoforge/net.minecraftforge/g' \
  -e 's/ResourceLocation\.fromNamespaceAndPath(\([^,]*\), \([^)]*\))/new ResourceLocation(\1, \2)/g' \
  -e 's/ResourceLocation\.parse(\([^)]*\))/new ResourceLocation(\1)/g' \
  -e 's/\.getFirst()/\.get(0)/g' \
  -e 's/NeoForge\.EVENT_BUS/MinecraftForge.EVENT_BUS/g' \
  -e 's/NeoForgeMod/ForgeMod/g'

# Main mod constructor: Forge 1.20 gets its mod bus from FMLJavaModLoadingContext.
python3 - <<'PY'
from pathlib import Path
p = Path('.wayward-port-work/port/src/main/java/team/lodestar/wayward_attributes/WaywardAttributes.java')
s = p.read_text()
s = s.replace('import net.minecraftforge.eventbus.api.IEventBus;\n', '')
s = s.replace('import net.minecraftforge.forge.common.ForgeMod;\n', 'import net.minecraftforge.common.ForgeMod;\nimport net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;\n')
s = s.replace('public WaywardAttributes(IEventBus modEventBus) {', 'public WaywardAttributes() {\n        var modEventBus = FMLJavaModLoadingContext.get().getModEventBus();')
s = s.replace('        ForgeMod.enableMergedAttributeTooltips();\n', '')
p.write_text(s)
PY

# Import MinecraftForge only where the event bus replacement is actually used.
while IFS= read -r -d '' f; do
  if grep -q 'MinecraftForge\.EVENT_BUS' "$f" && ! grep -q 'net.minecraftforge.common.MinecraftForge' "$f"; then
    sed -i '/^package .*;/a import net.minecraftforge.common.MinecraftForge;' "$f"
  fi
done < <(find "$OUT/port/src/main/java" -type f -name '*.java' -print0)

# 1.21 singular tag folders -> 1.20.1 plural folders.
for root in "$OUT/port/src/main/resources/data" "$OUT/port/src/generated/resources/data"; do
  [[ -d "$root" ]] || continue
  while IFS= read -r -d '' d; do
    parent="$(dirname "$d")"; name="$(basename "$d")"
    case "$name" in
      item) mv "$d" "$parent/items" ;;
      block) mv "$d" "$parent/blocks" ;;
      entity_type) mv "$d" "$parent/entity_types" ;;
    esac
  done < <(find "$root" -type d \( -path '*/tags/item' -o -path '*/tags/block' -o -path '*/tags/entity_type' \) -print0 2>/dev/null || true)
done

{
  echo "Wayward Attributes release-era backport ledger"
  echo "Source: 1.1 build line @ $SOURCE_SHA (Malum 1.8.2 consumes build 1.1.30)"
  echo "Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17 / Lodestone 1.6.4.1"
  printf 'Java sources: '; find "$OUT/port/src/main/java" -type f -name '*.java' | wc -l
  printf 'Main resources: '; find "$OUT/port/src/main/resources" -type f | wc -l
  echo
  echo "Known semantic bridge: NeoForge AttachmentType draw-speed state -> Forge 1.20 capability/runtime state."
} > "$REPORT_DIR/inventory.txt"

{
  echo "Residual future-version/API surfaces after mechanical bridge:"
  grep -RhoE 'AttachmentType|ATTACHMENT_TYPES|DataComponents\.[A-Z0-9_]+|net\.minecraft\.core\.component\.[A-Za-z0-9_$.]+|StreamCodec|HolderLookup' \
    "$OUT/port/src/main/java" 2>/dev/null | sort | uniq -c | sort -nr || true
} > "$REPORT_DIR/residual-surfaces.txt"

diff -ruN "$OUT/source/src/main/java" "$OUT/port/src/main/java" > "$REPORT_DIR/mechanical-java-bridge.patch" || true

set +e
(
  cd "$OUT/port"
  gradle --no-daemon compileJava --stacktrace --console=plain
) > "$REPORT_DIR/compile.log" 2>&1
STATUS=$?
set -e

echo "$STATUS" > "$REPORT_DIR/compile-exit-code.txt"
grep -E '(^|: )error:|cannot find symbol|does not exist|incompatible types|no suitable method|method .* cannot be applied' \
  "$REPORT_DIR/compile.log" > "$REPORT_DIR/error-index.txt" || true
exit "$STATUS"
