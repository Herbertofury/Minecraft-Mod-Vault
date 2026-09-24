#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
GAME_DIR="$PORT/.production-client-run"
LAUNCH_SCRIPT="$GAME_DIR/launch-forgeclient.sh"
WORLD="$GAME_DIR/saves/MRPG-QA"
QA_TICK="$WORLD/datapacks/more-rpg-runtime-qa/data/more_rpg_qa/functions/tick.mcfunction"
PLAYERDATA="$WORLD/playerdata"
RESTART_LOG="$PORT/more-rpg-production-restart.log"
LATEST="$GAME_DIR/logs/latest.log"
SESSION_LOCK="$WORLD/session.lock"

for f in "$LAUNCH_SCRIPT" "$QA_TICK"; do test -f "$f"; done
test -d "$PLAYERDATA"
bash -n "$LAUNCH_SCRIPT"
grep -Fq -- '--quickPlaySingleplayer' "$LAUNCH_SCRIPT"
grep -Fq 'forgeclient' "$LAUNCH_SCRIPT"
! grep -Fq 'forgeclientuserdev' "$LAUNCH_SCRIPT"
grep -Fxq 'showLoadWarnings = false' "$GAME_DIR/config/forge-client.toml"

# Stage-2 must have persisted the original QA tag to disk. Search decompressed player NBT directly;
# this is independent of the second launch and prevents a same-tick reapplication false positive.
python3 - "$PLAYERDATA" <<'PY'
import gzip
from pathlib import Path
import sys
root = Path(sys.argv[1])
needle = b'mrpg_qa_fatal_poison'
hits = []
for path in sorted(root.glob('*.dat')):
    try:
        raw = gzip.open(path, 'rb').read()
    except OSError as exc:
        raise SystemExit(f'[More RPG 2.7.2] invalid gzip player NBT {path}: {exc}')
    if needle in raw:
        hits.append(path.name)
if not hits:
    raise SystemExit('[More RPG 2.7.2] persisted fatal-poison player tag missing from on-disk NBT')
print('[More RPG 2.7.2] PRODUCTION_PLAYER_NBT_PERSISTENCE_PASS files=' + ','.join(hits))
PY

# Install a restart-only probe BEFORE the original QA commands. If the persisted tag is absent on
# launch two, the probe cannot emit; the original function would instead reapply fatal_poison and
# emit FATAL_POISON_APPLIED, which this gate treats as a hard persistence failure.
if grep -Fq 'PERSISTED_FATAL_POISON_TAG_SEEN' "$QA_TICK"; then
  echo '[More RPG 2.7.2] restart probe unexpectedly already present' >&2
  exit 1
fi
grep -Fq 'effect give @s more_rpg_classes:fatal_poison' "$QA_TICK"
grep -Fq 'tag @a[tag=!mrpg_qa_fatal_poison] add mrpg_qa_fatal_poison' "$QA_TICK"
QA_TMP="$QA_TICK.restart"
cat > "$QA_TMP" <<'MCFUNCTION'
execute as @a[tag=mrpg_qa_fatal_poison,tag=!mrpg_qa_restart_probe] run tellraw @a {"text":"[More RPG QA] PERSISTED_FATAL_POISON_TAG_SEEN"}
tag @a[tag=mrpg_qa_fatal_poison,tag=!mrpg_qa_restart_probe] add mrpg_qa_restart_probe
MCFUNCTION
cat "$QA_TICK" >> "$QA_TMP"
mv "$QA_TMP" "$QA_TICK"
[[ "$(grep -Fc 'PERSISTED_FATAL_POISON_TAG_SEEN' "$QA_TICK")" -eq 1 ]]
[[ "$(grep -Fc 'effect give @s more_rpg_classes:fatal_poison' "$QA_TICK")" -eq 1 ]]
echo '[More RPG 2.7.2] PRODUCTION_RESTART_PERSISTENCE_PROBE_ARMED source=player_tag_preexisting_only'

LOCK_BEFORE='missing'
if [[ -f "$SESSION_LOCK" ]]; then LOCK_BEFORE="$(stat -c '%Y:%s' "$SESSION_LOCK")"; fi
rm -rf "$GAME_DIR/logs"
: > "$RESTART_LOG"

ACTIVE_PID=''
descendants() {
  local parent="$1" child
  while IFS= read -r child; do
    [[ -z "$child" ]] && continue
    descendants "$child"
    printf '%s\n' "$child"
  done < <(pgrep -P "$parent" 2>/dev/null || true)
}
stop_tree() {
  local root="${1:-}"; [[ -n "$root" ]] || return 0
  local -a children=(); mapfile -t children < <(descendants "$root")
  ((${#children[@]})) && kill -TERM "${children[@]}" 2>/dev/null || true
  kill -TERM "$root" 2>/dev/null || true
  for _ in {1..50}; do kill -0 "$root" 2>/dev/null || break; sleep 0.1; done
  mapfile -t children < <(descendants "$root")
  ((${#children[@]})) && kill -KILL "${children[@]}" 2>/dev/null || true
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}
cleanup() { [[ -n "${ACTIVE_PID:-}" ]] && stop_tree "$ACTIVE_PID" || true; }
trap cleanup EXIT INT TERM

env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe ALSOFT_DRIVERS=null \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  "$LAUNCH_SCRIPT" </dev/null > "$RESTART_LOG" 2>&1 &
ACTIVE_PID=$!
CLIENT_PID=$ACTIVE_PID
DEADLINE=$((SECONDS+300))
PASS=0
FATAL='ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Exception in thread "Render thread"|The game crashed|Missing or unsupported mandatory dependencies|Could not initialize GLFW|Failed to initialize graphics window|Timed out trying to setup the Game Window'
while ((SECONDS<DEADLINE)); do
  FILES=("$RESTART_LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then
    stop_tree "$CLIENT_PID"; ACTIVE_PID=''; tail -n 700 "${FILES[@]}" || true; exit 1
  fi
  if grep -Fq '[More RPG QA] FATAL_POISON_APPLIED' "${FILES[@]}"; then
    echo '[More RPG 2.7.2] restart re-applied fatal_poison; persisted player tag was not honored' >&2
    stop_tree "$CLIENT_PID"; ACTIVE_PID=''; tail -n 700 "${FILES[@]}" || true; exit 1
  fi
  if [[ -f "$LATEST" ]] \
    && grep -Fq 'Backend library: LWJGL' "$LATEST" \
    && grep -Fq 'more_rpg_classes' "${FILES[@]}" \
    && grep -Fq '[More RPG QA] PERSISTED_FATAL_POISON_TAG_SEEN' "${FILES[@]}"; then
    PASS=1; break
  fi
  if ! kill -0 "$CLIENT_PID" 2>/dev/null; then
    wait "$CLIENT_PID" || true; ACTIVE_PID=''; tail -n 700 "${FILES[@]}" || true; exit 1
  fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || {
  stop_tree "$CLIENT_PID"; ACTIVE_PID=''; tail -n 700 "$RESTART_LOG" "$LATEST" 2>/dev/null || true; exit 1
}

[[ -f "$SESSION_LOCK" ]] || {
  echo '[More RPG 2.7.2] second production world did not create session.lock' >&2
  stop_tree "$CLIENT_PID"; ACTIVE_PID=''; exit 1
}
LOCK_AFTER="$(stat -c '%Y:%s' "$SESSION_LOCK")"
[[ "$LOCK_AFTER" != "$LOCK_BEFORE" ]] || {
  echo "[More RPG 2.7.2] session.lock did not change across production restart before=$LOCK_BEFORE after=$LOCK_AFTER" >&2
  stop_tree "$CLIENT_PID"; ACTIVE_PID=''; exit 1
}

stop_tree "$CLIENT_PID"; ACTIVE_PID=''
echo "[More RPG 2.7.2] PRODUCTION_RESTART_WORLD_REOPEN_PASS session_lock_before=$LOCK_BEFORE session_lock_after=$LOCK_AFTER"
echo '[More RPG 2.7.2] PRODUCTION_RESTART_PERSISTENCE_PASS same_game_dir=true same_world=true player_nbt_tag=true datapack_preexisting_tag=true no_reapply=true'
