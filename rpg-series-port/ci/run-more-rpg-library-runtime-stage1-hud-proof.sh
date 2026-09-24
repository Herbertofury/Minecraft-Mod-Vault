#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
DIAGNOSTIC="$ROOT/rpg-series-port/ci/run-more-rpg-library-runtime-stage1-diagnostic.sh"
CLIENT_RUN="$ROOT/.more-rpg-library-build/forge/run"
CLIENT_LOG="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more-rpg-native-client-integrated.log"

[[ -f "$DIAGNOSTIC" ]]
bash -n "$DIAGNOSTIC"

# Preserve the unchanged Stage-1/world-entry gate, but ask Mixin to export only the transformed
# vanilla GUI classes needed for a direct DrawHeartsMixin bytecode proof.
rm -rf "$CLIENT_RUN/.mixin.out"
MIXIN_JAVA_TOOL_OPTIONS='-Dmixin.debug.export=true -Dmixin.debug.export.filter=net.minecraft.client.gui.** -Dmixin.debug.export.decompile=false -Dmixin.debug.verbose=true'
export JAVA_TOOL_OPTIONS="$MIXIN_JAVA_TOOL_OPTIONS${JAVA_TOOL_OPTIONS:+ $JAVA_TOOL_OPTIONS}"
echo '[More RPG 2.7.2] DRAW_HEARTS_MIXIN_EXPORT_ARMED filter=net.minecraft.client.gui.**'

# The diagnostic remains authoritative for packaged server + mapped world entry + fatal poison.
# Do not attempt transformed-bytecode proof if that real gameplay gate fails.
bash "$DIAGNOSTIC"

LATEST="$CLIENT_RUN/logs/latest.log"
MIXIN_OUT="$CLIENT_RUN/.mixin.out"
[[ -f "$CLIENT_LOG" && -f "$LATEST" ]] || {
  echo '[More RPG 2.7.2] direct HUD proof missing client runtime logs' >&2
  exit 1
}
if ! grep -Eiq 'DrawHeartsMixin.*more-rpg-classes\.mixins\.json|more-rpg-classes\.mixins\.json.*DrawHeartsMixin|DrawHeartsMixin' "$CLIENT_LOG" "$LATEST"; then
  echo '[More RPG 2.7.2] DrawHeartsMixin verbose application record missing' >&2
  exit 1
fi
[[ -d "$MIXIN_OUT" ]] || {
  echo '[More RPG 2.7.2] Mixin post-transform export directory missing' >&2
  exit 1
}
mapfile -t HEART_TARGETS < <(grep -aRl 'net/more_rpg_classes/client/heart/HeartRegistry' "$MIXIN_OUT" --include='*.class' 2>/dev/null | sort || true)
if ((${#HEART_TARGETS[@]} == 0)); then
  echo '[More RPG 2.7.2] No exported transformed GUI class contains HeartRegistry reference' >&2
  find "$MIXIN_OUT" -type f -name '*.class' -print | sort | head -n 100 || true
  exit 1
fi
printf '[More RPG 2.7.2] DRAW_HEARTS_MIXIN_TRANSFORMED_CLIENT_PASS targets=%s first=%s\n' \
  "${#HEART_TARGETS[@]}" "${HEART_TARGETS[0]#$MIXIN_OUT/}"
echo '[More RPG 2.7.2] RUNTIME_STAGE1_HARDENED_PASS world_entry=true transformed_hud=true production_client_pending=true'
