#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-spell-engine.sh"
PATCHED="$ROOT/rpg-series-port/ci/.run-spell-engine-graduation.generated.sh"
ENV_FILE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/SPELL_ENGINE_GRADUATION.env"

test -f "$BASE"
test -f "$ENV_FILE"
source "$ENV_FILE"

python3 - "$BASE" "$PATCHED" <<'PY'
from pathlib import Path
import sys
src=Path(sys.argv[1]).read_text()
out=Path(sys.argv[2])

# Clone the exact TinyConfig 3.1.0 authority alongside the existing exact-source fan-in.
needle="clone_exact FabricExtras/RangedWeaponAPI \"$RANGED_TARGET\" \"$UP/ranged-234\" & P6=$!\nwait \"$P1\" \"$P2\" \"$P3\" \"$P4\" \"$P5\" \"$P6\"\n"
replacement="clone_exact FabricExtras/RangedWeaponAPI \"$RANGED_TARGET\" \"$UP/ranged-234\" & P6=$!\nclone_exact ZsoltMolnarrr/TinyConfig e20fc8ac72fde8274f0df72de2ebb81ffe6f8727 \"$UP/tiny-config-310\" & P7=$!\nwait \"$P1\" \"$P2\" \"$P3\" \"$P4\" \"$P5\" \"$P6\" \"$P7\"\n"
if src.count(needle) != 1:
    raise SystemExit(f'[Spell Engine graduation] expected one exact-source fan-in seam, found {src.count(needle)}')
src=src.replace(needle,replacement)

# Reconstruct the already-certified TinyConfig port before Spell Engine is prepared.
needle='''# Build the two already-verified foundation dependencies as actual separate mods. Spell Engine compiles\n# against their named common JARs and runs against their Forge JARs; their source trees are never added\n# to Spell Engine sourceSets, so dependency implementation classes cannot leak into its release JAR.\nSPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"\n'''
replacement='''# Reconstruct certified TinyConfig 3.1.0. Spell Engine 1.10.2 compiles against its common API and\n# intentionally embeds the Forge artifact via JarJar, matching upstream 1.10.2 packaging.\nTINY_CONFIG="$ROOT/rpg-series-port/tiny-config-forge-1.20.1"\nTINY_GEN="$TINY_CONFIG/generated"\npython3 "$TINY_CONFIG/tools/prepare_port.py" "$UP/tiny-config-310" "$TINY_GEN"\ngradle --no-daemon --stacktrace -p "$TINY_GEN" clean :common:jar :forge:remapJar\nTINY_CONFIG_COMMON_JAR="$(find "$TINY_GEN/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | head -n 1)"\nTINY_CONFIG_FORGE_JAR="$(find "$TINY_GEN/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | head -n 1)"\ntest -f "$TINY_CONFIG_COMMON_JAR" -a -f "$TINY_CONFIG_FORGE_JAR"\nTINY_CONFIG_ACTUAL_SHA="$(sha256sum "$TINY_CONFIG_FORGE_JAR" | awk '{print $1}')"\necho "[Spell Engine graduation] TinyConfig SHA=$TINY_CONFIG_ACTUAL_SHA expected=$TINY_CONFIG_310_EXPECTED_JAR_SHA"\n[[ "$TINY_CONFIG_ACTUAL_SHA" = "$TINY_CONFIG_310_EXPECTED_JAR_SHA" ]]\necho '[Spell Engine graduation] CERTIFIED_TINY_CONFIG_310_IDENTITY_PASS'\n\n# Build the already-verified external RPG foundation dependencies as actual separate mods. Spell Engine\n# compiles against their named common JARs and runs against their Forge JARs; their source trees are\n# never added to Spell Engine sourceSets, so dependency implementation classes cannot leak into release.\nSPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"\n'''
if src.count(needle) != 1:
    raise SystemExit(f'[Spell Engine graduation] expected one foundation-build seam, found {src.count(needle)}')
src=src.replace(needle,replacement)

needle='export SPELL_POWER_COMMON_JAR RANGED_COMMON_JAR SPELL_POWER_FORGE_JAR RANGED_FORGE_JAR\n'
replacement='export TINY_CONFIG_COMMON_JAR TINY_CONFIG_FORGE_JAR SPELL_POWER_COMMON_JAR RANGED_COMMON_JAR SPELL_POWER_FORGE_JAR RANGED_FORGE_JAR\n'
if src.count(needle) != 1:
    raise SystemExit(f'[Spell Engine graduation] expected one dependency export seam, found {src.count(needle)}')
src=src.replace(needle,replacement)

# The old runner only checked that some TinyConfig was nested. Graduation requires the certified 3.1.0
# payload and explicitly rejects the historical 2.x workaround.
needle="unzip -l \"$OUT_JAR\" | grep -F 'META-INF/jars/TinyConfig-'\n"
replacement='''if unzip -Z1 "$OUT_JAR" | grep -Eiq 'META-INF/jars/[^/]*(tiny.?config|TinyConfig)[^/]*2\\.'; then\n  echo 'Obsolete TinyConfig 2.x leaked into Spell Engine 1.10.2 release' >&2; exit 1\nfi\nTINY_NESTED="$(unzip -Z1 "$OUT_JAR" | grep -Ei '^META-INF/jars/.*tiny.?config.*\\.jar$' | head -n 1 || true)"\n[[ -n "$TINY_NESTED" ]] || { echo 'Spell Engine release lost upstream TinyConfig JarJar payload' >&2; exit 1; }\nunzip -p "$OUT_JAR" "$TINY_NESTED" > "$PORT/tiny-config-embedded-check.jar"\nunzip -tq "$PORT/tiny-config-embedded-check.jar" >/dev/null\n[[ "$(sha256sum "$PORT/tiny-config-embedded-check.jar" | awk '{print $1}')" = "$TINY_CONFIG_310_EXPECTED_JAR_SHA" ]] || {\n  echo 'Spell Engine embedded TinyConfig is not the certified 3.1.0 Forge artifact' >&2; exit 1;\n}\necho '[Spell Engine graduation] CERTIFIED_TINY_CONFIG_310_EMBEDDED_PASS'\n'''
if src.count(needle) != 1:
    raise SystemExit(f'[Spell Engine graduation] expected one legacy TinyConfig package gate, found {src.count(needle)}')
src=src.replace(needle,replacement)

out.write_text(src)
PY

chmod +x "$PATCHED"
bash -n "$PATCHED"
echo '[Spell Engine graduation] TINY_CONFIG_310_PARITY_WRAPPER_READY'
exec bash "$PATCHED"
