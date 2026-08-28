#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
RUN="$ROOT/rpg-series-port/paladins-forge-1.20.1/forge/run"
OPTIONS="$RUN/options.txt"
SOURCE="$ROOT/rpg-series-port/ci/run-paladins-player-behavior-acceptance.sh"
PATCHED="${RUNNER_TEMP:-/tmp}/run-paladins-player-behavior-ui.sh"

mkdir -p "$RUN"
touch "$OPTIONS"
ensure_option(){
  local key="$1" value="$2"
  if grep -q "^${key}:" "$OPTIONS"; then
    sed -i "s/^${key}:.*/${key}:${value}/" "$OPTIONS"
  else
    printf '%s:%s\n' "$key" "$value" >> "$OPTIONS"
  fi
}

# Deterministic native-client state. GUI scale 3 makes the requested 1280x720 window
# a 426x240 logical canvas, so vanilla menu button centers are stable for Robot input.
ensure_option onboardAccessibility false
ensure_option skipMultiplayerWarning true
ensure_option joinedFirstServer true
ensure_option pauseOnLostFocus false
ensure_option guiScale 3
for expected in \
  'onboardAccessibility:false' \
  'skipMultiplayerWarning:true' \
  'joinedFirstServer:true' \
  'pauseOnLostFocus:false' \
  'guiScale:3'; do
  grep -Fxq "$expected" "$OPTIONS"
done
echo '[Paladins player QA] FIRST_RUN_UI_PREFLIGHT_PASS: onboarding disabled and deterministic GUI scale 3 seeded.'

python3 - "$SOURCE" "$PATCHED" <<'PY'
from pathlib import Path
import sys
src = Path(sys.argv[1]).read_text(encoding='utf-8')
old_compile = 'javac "$ROBOT_DIR/PaladinsInputRobot.java"\n'
new_compile = r'''javac "$ROBOT_DIR/PaladinsInputRobot.java"
cat > "$ROBOT_DIR/PaladinsConnectRobot.java" <<'JAVA'
import java.awt.Robot;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
public final class PaladinsConnectRobot {
  private static void click(Robot r, int x, int y) throws Exception {
    r.mouseMove(x, y);
    Thread.sleep(200);
    r.mousePress(InputEvent.BUTTON1_DOWN_MASK);
    r.mouseRelease(InputEvent.BUTTON1_DOWN_MASK);
  }
  public static void main(String[] args) throws Exception {
    Robot r = new Robot();
    r.setAutoDelay(80);
    // 1280x720, guiScale=3 -> 426x240 logical. Title multiplayer button:
    // title y=height/4+48=108, multiplayer top y=132, 20px high => physical center y=426.
    click(r, 640, 426);
    Thread.sleep(1400);
    // Multiplayer screen Direct Connection is the centered first-row action near height-52;
    // logical center y=198 -> physical y=594. This opens DirectConnectScreen.
    click(r, 640, 594);
    Thread.sleep(1000);
    r.keyPress(KeyEvent.VK_CONTROL); r.keyPress(KeyEvent.VK_A); r.keyRelease(KeyEvent.VK_A); r.keyRelease(KeyEvent.VK_CONTROL);
    for (char c : "127.0.0.1:25565".toCharArray()) {
      int key = KeyEvent.getExtendedKeyCodeForChar(c);
      if (c == ':') {
        r.keyPress(KeyEvent.VK_SHIFT); r.keyPress(KeyEvent.VK_SEMICOLON); r.keyRelease(KeyEvent.VK_SEMICOLON); r.keyRelease(KeyEvent.VK_SHIFT);
      } else {
        r.keyPress(key); r.keyRelease(key);
      }
    }
    Thread.sleep(250);
    r.keyPress(KeyEvent.VK_ENTER); r.keyRelease(KeyEvent.VK_ENTER);
    System.out.println("PALADINS_DIRECT_CONNECT_ROBOT_DISPATCHED");
  }
}
JAVA
javac "$ROBOT_DIR/PaladinsConnectRobot.java"
'''
old_launch = "( gradle --no-daemon -p \"$PORT\" :forge:runClient \"${ARGS[@]}\" --args='--width 1280 --height 720 --quickPlayMultiplayer 127.0.0.1:25565' </dev/null ) > \"$CLIENT_LOG\" 2>&1 & CLIENT_PID=$!\nwait_marker 'joined the game' 210 'native Forge client joining exact packaged server'\n"
new_launch = r'''( gradle --no-daemon -p "$PORT" :forge:runClient "${ARGS[@]}" --args='--width 1280 --height 720' </dev/null ) > "$CLIENT_LOG" 2>&1 & CLIENT_PID=$!
wait_marker 'Backend library: LWJGL' 180 'native Forge LWJGL initialization'
wait_marker 'Created: 256x128x0 minecraft:textures/atlas/mob_effects.png-atlas' 120 'native Forge title-screen resource readiness'
sleep 2
java -cp "$ROBOT_DIR" PaladinsConnectRobot
wait_marker 'joined the game' 120 'native Forge client joining exact packaged server through vanilla Direct Connection UI'
'''
if src.count(old_compile) != 1:
    raise SystemExit(f'expected one input Robot compile anchor, got {src.count(old_compile)}')
if src.count(old_launch) != 1:
    raise SystemExit(f'expected one Quick Play launch anchor, got {src.count(old_launch)}')
src = src.replace(old_compile, new_compile).replace(old_launch, new_launch)
Path(sys.argv[2]).write_text(src, encoding='utf-8')
PY
chmod +x "$PATCHED"
bash -n "$PATCHED"
grep -Fq 'PALADINS_DIRECT_CONNECT_ROBOT_DISPATCHED' "$PATCHED"
if grep -Fq -- '--quickPlayMultiplayer' "$PATCHED"; then
  echo '[Paladins player QA] stale Quick Play launch path survived QA patch' >&2
  exit 1
fi
echo '[Paladins player QA] DIRECT_CONNECT_UI_PATCH_PASS: temporary QA script uses vanilla Multiplayer -> Direct Connection UI; tracked product/release source remains untouched.'
exec bash "$PATCHED"
