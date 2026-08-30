#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-more-rpg-library-production-client.sh"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
PATCHED="$PORT/.run-more-rpg-library-production-client.generated.sh"
test -f "$BASE"
python3 - "$BASE" "$PATCHED" <<'PY'
from pathlib import Path
import sys
src=Path(sys.argv[1]).read_text(); out=Path(sys.argv[2])

anchor='''for f in "$OUT" "$SPELL_ENGINE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$TINY_JAR" "$CLOTH_FORGE_JAR" "$PLAYER_ANIM_FORGE_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done\n'''
insert='''for f in "$OUT" "$SPELL_ENGINE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$TINY_JAR" "$CLOTH_FORGE_JAR" "$PLAYER_ANIM_FORGE_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done\nprintf '[More RPG 2.7.2] PRODUCTION_MODSET_HASH more_rpg=%s spell_engine=%s spell_power=%s ranged=%s tiny_config=%s cloth=%s player_anim=%s java=%s\\n' \\\n  "$(sha256sum "$OUT" | awk '{print $1}')" "$(sha256sum "$SPELL_ENGINE_JAR" | awk '{print $1}')" "$(sha256sum "$SPELL_POWER_JAR" | awk '{print $1}')" \\\n  "$(sha256sum "$RANGED_JAR" | awk '{print $1}')" "$(sha256sum "$TINY_JAR" | awk '{print $1}')" "$(sha256sum "$CLOTH_FORGE_JAR" | awk '{print $1}')" \\\n  "$(sha256sum "$PLAYER_ANIM_FORGE_JAR" | awk '{print $1}')" "$(java -version 2>&1 | head -n1 | tr ' ' '_')"\n'''
if src.count(anchor)!=1: raise SystemExit(f'production modset hash seam drifted: {src.count(anchor)}')
src=src.replace(anchor,insert,1)

anchor='''CLIENT_SRG="$(find "$MCROOT/libraries/net/minecraft/client" -type f -name 'client-1.20.1-20230612.114412-srg.jar' | head -n1)"\nFORGE_CLIENT="$(find "$MCROOT/libraries/net/minecraftforge/forge/1.20.1-47.4.23" -type f -name 'forge-1.20.1-47.4.23-client.jar' | head -n1)"\nFORGE_UNIVERSAL="$(find "$MCROOT/libraries/net/minecraftforge/forge/1.20.1-47.4.23" -type f -name 'forge-1.20.1-47.4.23-universal.jar' | head -n1)"\nfor f in "$CLIENT_SRG" "$FORGE_CLIENT" "$FORGE_UNIVERSAL"; do test -f "$f"; unzip -tq "$f" >/dev/null; done\nprintf '[More RPG 2.7.2] PRODUCTION_FORGE_NAMESPACE_ARTIFACTS_PASS client_srg=%s forge_client=%s forge_universal=%s\\n' "$(sha256sum "$CLIENT_SRG" | awk '{print $1}')" "$(sha256sum "$FORGE_CLIENT" | awk '{print $1}')" "$(sha256sum "$FORGE_UNIVERSAL" | awk '{print $1}')"\n'''
replacement='''CLIENT_SRG="$(find "$MCROOT/libraries/net/minecraft/client" -type f -name 'client-1.20.1-20230612.114412-srg.jar' | head -n1)"\nCLIENT_EXTRA="$(find "$MCROOT/libraries/net/minecraft/client" -type f -name 'client-1.20.1-20230612.114412-extra.jar' | head -n1)"\nCLIENT_SLIM="$(find "$MCROOT/libraries/net/minecraft/client" -type f -name 'client-1.20.1-20230612.114412-slim.jar' | head -n1)"\nFORGE_CLIENT="$(find "$MCROOT/libraries/net/minecraftforge/forge/1.20.1-47.4.23" -type f -name 'forge-1.20.1-47.4.23-client.jar' | head -n1)"\nFORGE_UNIVERSAL="$(find "$MCROOT/libraries/net/minecraftforge/forge/1.20.1-47.4.23" -type f -name 'forge-1.20.1-47.4.23-universal.jar' | head -n1)"\nFMLCORE="$(find "$MCROOT/libraries/net/minecraftforge/fmlcore/1.20.1-47.4.23" -type f -name 'fmlcore-1.20.1-47.4.23.jar' | head -n1)"\nJAVAFML="$(find "$MCROOT/libraries/net/minecraftforge/javafmllanguage/1.20.1-47.4.23" -type f -name 'javafmllanguage-1.20.1-47.4.23.jar' | head -n1)"\nLOWCODE="$(find "$MCROOT/libraries/net/minecraftforge/lowcodelanguage/1.20.1-47.4.23" -type f -name 'lowcodelanguage-1.20.1-47.4.23.jar' | head -n1)"\nMCLANG="$(find "$MCROOT/libraries/net/minecraftforge/mclanguage/1.20.1-47.4.23" -type f -name 'mclanguage-1.20.1-47.4.23.jar' | head -n1)"\nBOOTSTRAP="$(find "$MCROOT/libraries/cpw/mods/bootstraplauncher" -type f -name 'bootstraplauncher-*.jar' | sort | head -n1)"\nMODLAUNCHER="$(find "$MCROOT/libraries/cpw/mods/modlauncher" -type f -name 'modlauncher-*.jar' | sort | head -n1)"\nfor f in "$CLIENT_SRG" "$CLIENT_EXTRA" "$CLIENT_SLIM" "$FORGE_CLIENT" "$FORGE_UNIVERSAL" "$FMLCORE" "$JAVAFML" "$LOWCODE" "$MCLANG" "$BOOTSTRAP" "$MODLAUNCHER"; do test -f "$f"; unzip -tq "$f" >/dev/null; done\nprintf '[More RPG 2.7.2] PRODUCTION_FORGE_LIBRARY_TREE_PASS srg=%s extra=%s slim=%s forge_client=%s universal=%s fmlcore=%s javafml=%s lowcode=%s mclang=%s bootstrap=%s modlauncher=%s\\n' \\\n  "$(sha256sum "$CLIENT_SRG" | awk '{print $1}')" "$(sha256sum "$CLIENT_EXTRA" | awk '{print $1}')" "$(sha256sum "$CLIENT_SLIM" | awk '{print $1}')" \\\n  "$(sha256sum "$FORGE_CLIENT" | awk '{print $1}')" "$(sha256sum "$FORGE_UNIVERSAL" | awk '{print $1}')" "$(sha256sum "$FMLCORE" | awk '{print $1}')" \\\n  "$(sha256sum "$JAVAFML" | awk '{print $1}')" "$(sha256sum "$LOWCODE" | awk '{print $1}')" "$(sha256sum "$MCLANG" | awk '{print $1}')" \\\n  "$(sha256sum "$BOOTSTRAP" | awk '{print $1}')" "$(sha256sum "$MODLAUNCHER" | awk '{print $1}')"\n'''
if src.count(anchor)!=1: raise SystemExit(f'production library tree seam drifted: {src.count(anchor)}')
src=src.replace(anchor,replacement,1)

anchor="jvm += [f'-Djava.library.path={natives}', '-cp', classpath]\n"
replacement="jvm += [f'-Dorg.lwjgl.librarypath={natives}', f'-Djava.library.path={natives}', '-cp', classpath]\n"
if src.count(anchor)!=1: raise SystemExit(f'LWJGL native-path seam drifted: {src.count(anchor)}')
src=src.replace(anchor,replacement,1)

if src.count('-Dorg.lwjgl.librarypath={natives}')!=1 or src.count('-Djava.library.path={natives}')!=1:
    raise SystemExit('dual native library path contract missing')
if 'CLIENT_EXTRA=' not in src or 'BOOTSTRAP=' not in src or 'MODLAUNCHER=' not in src:
    raise SystemExit('production Forge library-tree contract missing')
out.write_text(src)
PY
bash -n "$PATCHED"
echo '[More RPG 2.7.2] PRODUCTION_FORGE_CLIENT_HARDENED_WRAPPER_READY library_tree=true dual_native_paths=true'
exec bash "$PATCHED"
