#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
JAR="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more_rpg_library-forge-2.7.2+1.20.1.jar"
test -f "$JAR"
unzip -tq "$JAR" >/dev/null
MOD="$(mktemp)"; GAME="$(mktemp)"; COMMON="$(mktemp)"; FORGE="$(mktemp)"; INVENTORY="$(mktemp)"
trap 'rm -f "$MOD" "$GAME" "$COMMON" "$FORGE" "$INVENTORY"' EXIT
unzip -Z1 "$JAR" > "$INVENTORY"
for entry in \
  net/more_rpg_classes/forge/client/ForgeClientMod.class \
  net/more_rpg_classes/forge/client/ForgeClientEvents.class \
  net/more_rpg_classes/client/MoreRPGClassesClient.class \
  net/more_rpg_classes/client/heart/HeartRegistry.class \
  net/more_rpg_classes/client/heart/HeartTypes.class; do
  grep -Fxq "$entry" "$INVENTORY" || { echo "[More RPG 2.7.2] packaged client lifecycle entry missing: $entry" >&2; exit 1; }
done
javap -classpath "$JAR" -c -p net.more_rpg_classes.forge.client.ForgeClientMod > "$MOD"
javap -classpath "$JAR" -c -p net.more_rpg_classes.forge.client.ForgeClientEvents > "$GAME"
javap -classpath "$JAR" -c -p net.more_rpg_classes.client.MoreRPGClassesClient > "$COMMON"
javap -classpath "$JAR" -c -p net.more_rpg_classes.forge.ForgeMod > "$FORGE"
for needle in \
  'MoreRPGClassesClient.init' \
  'MoreRPGClassesClient.registerEntityRenderers' \
  'MoreRPGClassesClient.registerParticleAppearances'; do
  grep -Fq "$needle" "$MOD" || { echo "[More RPG 2.7.2] packaged MOD-bus client lifecycle missing $needle" >&2; cat "$MOD" >&2; exit 1; }
done
for needle in 'MobBeamWorldRenderer.render' 'MobBeamWorldRenderer.onDisconnect'; do
  grep -Fq "$needle" "$GAME" || { echo "[More RPG 2.7.2] packaged FORGE-bus client lifecycle missing $needle" >&2; cat "$GAME" >&2; exit 1; }
done
for needle in 'HeartTypes.getHeartTypes' 'HeartRegistry.register' 'SpellTooltip.addDescriptionMutator'; do
  grep -Fq "$needle" "$COMMON" || { echo "[More RPG 2.7.2] current client initialization bytecode missing $needle" >&2; cat "$COMMON" >&2; exit 1; }
done
if grep -Eq 'net/more_rpg_classes/client|net/minecraft/client' "$FORGE"; then
  echo '[More RPG 2.7.2] common ForgeMod hard-links physical-client classes' >&2
  cat "$FORGE" >&2
  exit 1
fi
echo '[More RPG 2.7.2] PACKAGED_FORGE_CLIENT_LIFECYCLE_PASS init=heart+tooltip+models renderers=true particles=true beam=true logout=true server_entrypoint_client_free=true'
