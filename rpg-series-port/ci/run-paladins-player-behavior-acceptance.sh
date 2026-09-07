#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/paladins-forge-1.20.1"
TMP="${RUNNER_TEMP:-/tmp}/paladins-first-compile"
RUN="$PORT/forge/run"
FRESH="$PORT/.fresh-paladins-forge-server"
CLIENT_LOG="$PORT/paladins-player-behavior.log"
CLIENT_LATEST="$RUN/logs/latest.log"
SERVER_LOG="$PORT/paladins-player-server.log"
SERVER_LATEST="$FRESH/logs/latest.log"
FIFO="$FRESH/paladins-player-console.fifo"
ROBOT_DIR="${RUNNER_TEMP:-/tmp}/paladins-input-robot"
QA_HELPER="$PORT/forge/src/main/java/net/paladins/forge/client/PaladinsQaAutoConnect.java"
QA_HELPER_CLASS="$PORT/forge/build/classes/java/main/net/paladins/forge/client/PaladinsQaAutoConnect.class"
CLIENT_PID=''
SERVER_PID=''
XVFB_PID=''
cleanup(){
  exec 9>&- 2>/dev/null || true
  if [[ -n "$CLIENT_PID" ]] && kill -0 "$CLIENT_PID" 2>/dev/null; then kill -TERM "$CLIENT_PID" 2>/dev/null || true; sleep 1; kill -KILL "$CLIENT_PID" 2>/dev/null || true; wait "$CLIENT_PID" 2>/dev/null || true; fi
  if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then kill -TERM "$SERVER_PID" 2>/dev/null || true; sleep 1; kill -KILL "$SERVER_PID" 2>/dev/null || true; wait "$SERVER_PID" 2>/dev/null || true; fi
  if [[ -n "$XVFB_PID" ]] && kill -0 "$XVFB_PID" 2>/dev/null; then kill -TERM "$XVFB_PID" 2>/dev/null || true; wait "$XVFB_PID" 2>/dev/null || true; fi
  rm -f "$FIFO" "$QA_HELPER" "$QA_HELPER_CLASS"
}
trap cleanup EXIT
wait_marker(){
  local marker="$1" timeout_seconds="$2" label="$3"
  local deadline=$((SECONDS + timeout_seconds))
  local fatal='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Exception in thread "Render thread"|The game crashed|Failed to load datapacks|Missing or unsupported mandatory dependencies|Failed to start the minecraft server|Exception in server tick loop|Incompatible FML modded server|mismatched mod channel|Connection refused'
  while (( SECONDS < deadline )); do
    if grep -Eiq "$fatal" "$CLIENT_LOG" "$CLIENT_LATEST" "$SERVER_LOG" "$SERVER_LATEST" 2>/dev/null; then
      echo "[Paladins player QA] fatal runtime signature during $label" >&2
      tail -n 400 "$CLIENT_LOG" "$CLIENT_LATEST" "$SERVER_LOG" "$SERVER_LATEST" 2>/dev/null || true
      return 1
    fi
    if grep -Fq "$marker" "$CLIENT_LOG" 2>/dev/null || grep -Fq "$marker" "$CLIENT_LATEST" 2>/dev/null || grep -Fq "$marker" "$SERVER_LOG" 2>/dev/null || grep -Fq "$marker" "$SERVER_LATEST" 2>/dev/null; then
      return 0
    fi
    if [[ -n "$SERVER_PID" ]] && ! kill -0 "$SERVER_PID" 2>/dev/null; then
      wait "$SERVER_PID" 2>/dev/null || true
      tail -n 400 "$SERVER_LOG" "$SERVER_LATEST" 2>/dev/null || true
      echo "[Paladins player QA] packaged server exited before $label marker: $marker" >&2
      return 1
    fi
    if [[ -n "$CLIENT_PID" ]] && ! kill -0 "$CLIENT_PID" 2>/dev/null; then
      wait "$CLIENT_PID" 2>/dev/null || true
      tail -n 400 "$CLIENT_LOG" "$CLIENT_LATEST" 2>/dev/null || true
      echo "[Paladins player QA] native client exited before $label marker: $marker" >&2
      return 1
    fi
    sleep 1
  done
  tail -n 400 "$CLIENT_LOG" "$CLIENT_LATEST" "$SERVER_LOG" "$SERVER_LATEST" 2>/dev/null || true
  echo "[Paladins player QA] timed out waiting for $label marker: $marker" >&2
  return 1
}
send_cmd(){ printf '%s\n' "$1" >&9; }
pick_jar(){
  local dir="$1" glob="$2"
  local -a jars=()
  mapfile -t jars < <(find "$dir" -maxdepth 1 -type f -name "$glob" ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -print | sort)
  (( ${#jars[@]} == 1 )) || { echo "[Paladins player QA] expected exactly one $glob in $dir, found ${#jars[@]}" >&2; return 1; }
  printf '%s\n' "${jars[0]}"
}

[[ -x "$FRESH/run.sh" ]] || { echo '[Paladins player QA] certified packaged Forge server missing; run baseline first' >&2; exit 1; }
[[ -f "$FRESH/world/level.dat" ]] || { echo '[Paladins player QA] packaged-server world missing' >&2; exit 1; }
[[ "$(find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 11 ]] || { echo '[Paladins player QA] expected exact 11-mod packaged runtime' >&2; exit 1; }
EXPECTED_SHA="$(awk '{print $1}' "$PORT/paladins.sha256")"
[[ "$EXPECTED_SHA" = '95e8f9e074dd1c432f9486c922173b76a3b3bc57667b28fe154e47dba5c374ee' ]] || { echo '[Paladins player QA] certified release ledger missing/drifted' >&2; exit 1; }
PAL_JAR="$(find "$FRESH/mods" -maxdepth 1 -type f -name 'paladins-forge-3.1.1+1.20.1.jar' -print -quit)"
[[ -f "$PAL_JAR" ]] || { echo '[Paladins player QA] packaged Paladins release missing' >&2; exit 1; }
[[ "$(sha256sum "$PAL_JAR" | awk '{print $1}')" = "$EXPECTED_SHA" ]] || { echo '[Paladins player QA] packaged Paladins release identity mismatch' >&2; exit 1; }
grep -Fq 'online-mode=false' "$FRESH/server.properties" || { echo '[Paladins player QA] packaged server must be offline-mode for deterministic CI player login' >&2; exit 1; }

# Reuse the exact foundation artifacts already built and validated by baseline acceptance.
# runClient evaluates common/forge build files before Minecraft starts, so every required
# project property must be carried into the native client invocation.
SHIELD_COMMON="$(pick_jar "$ROOT/rpg-series-port/shield-api-forge-1.20.1/common/build/libs" '*-common-*.jar')"
SHIELD_FORGE="$(pick_jar "$ROOT/rpg-series-port/shield-api-forge-1.20.1/forge/build/libs" '*.jar')"
SPELL_POWER_COMMON="$(pick_jar "$ROOT/rpg-series-port/spell_power-forge-1.20.1/common/build/libs" '*-common-*.jar')"
SPELL_POWER_FORGE="$(pick_jar "$ROOT/rpg-series-port/spell_power-forge-1.20.1/forge/build/libs" '*-forge-*.jar')"
TINY_COMMON="$(pick_jar "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/common/build/libs" '*-common-*.jar')"
TINY_FORGE="$(pick_jar "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" '*-forge-*.jar')"
SPELL_ENGINE_COMMON="$(pick_jar "$ROOT/.spell-engine-build/common/build/libs" '*-common-*.jar')"
SPELL_ENGINE_FORGE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"
RUNES_COMMON="$(pick_jar "$ROOT/rpg-series-port/runes-forge-1.20.1/common/build/libs" '*-common-*.jar')"
RUNES_FORGE="$(pick_jar "$ROOT/rpg-series-port/runes-forge-1.20.1/forge/build/libs" '*-forge-*.jar')"
STRUCTURE_COMMON="$(pick_jar "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1/common/build/libs" '*-common-*.jar')"
STRUCTURE_FORGE="$(pick_jar "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1/forge/build/libs" '*-forge-*.jar')"
ARMOR_COMMON="$(pick_jar "$ROOT/rpg-series-port/armor-model-api-forge-1.20.1/generated/common/build/libs" '*-common-*.jar')"
ARMOR_FORGE="$(pick_jar "$ROOT/rpg-series-port/armor-model-api-forge-1.20.1/generated/forge/build/libs" '*-forge-*.jar')"
CLOTH_FORGE="$TMP/ext/cloth-config-forge-11.1.136.jar"
PLAYER_FORGE="$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"
CURIOS_FORGE="$TMP/ext/curios-forge-5.14.1+1.20.1.jar"
for dep in "$SHIELD_COMMON" "$SHIELD_FORGE" "$SPELL_POWER_COMMON" "$SPELL_POWER_FORGE" "$TINY_COMMON" "$TINY_FORGE" \
  "$SPELL_ENGINE_COMMON" "$SPELL_ENGINE_FORGE" "$RUNES_COMMON" "$RUNES_FORGE" "$STRUCTURE_COMMON" "$STRUCTURE_FORGE" \
  "$ARMOR_COMMON" "$ARMOR_FORGE" "$CLOTH_FORGE" "$PLAYER_FORGE" "$CURIOS_FORGE"; do
  [[ -f "$dep" ]] || { echo "[Paladins player QA] missing baseline foundation artifact: $dep" >&2; exit 1; }
  unzip -tq "$dep" >/dev/null
done
ARGS=(
  "-Pshield_api_common_jar=$SHIELD_COMMON"
  "-Parmor_model_api_common_jar=$ARMOR_COMMON"
  "-Prunes_common_jar=$RUNES_COMMON"
  "-Pstructure_pool_api_common_jar=$STRUCTURE_COMMON"
  "-Pspell_power_common_jar=$SPELL_POWER_COMMON"
  "-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON"
  "-Ptiny_config_common_jar=$TINY_COMMON"
  "-Pshield_api_forge_jar=$SHIELD_FORGE"
  "-Parmor_model_api_forge_jar=$ARMOR_FORGE"
  "-Prunes_forge_jar=$RUNES_FORGE"
  "-Pstructure_pool_api_forge_jar=$STRUCTURE_FORGE"
  "-Pspell_power_forge_jar=$SPELL_POWER_FORGE"
  "-Pspell_engine_forge_jar=$SPELL_ENGINE_FORGE"
  "-Ptiny_config_forge_jar=$TINY_FORGE"
  "-Pcloth_config_forge_jar=$CLOTH_FORGE"
  "-Pplayer_animator_forge_jar=$PLAYER_FORGE"
  "-Pcurios_jar=$CURIOS_FORGE"
)
echo "[Paladins player QA] native client foundation property gate passed: ${#ARGS[@]} exact Gradle properties"
echo '[Paladins player QA] exact packaged Paladins server release identity gate passed.'

mkdir -p "$ROBOT_DIR"
cat > "$ROBOT_DIR/PaladinsInputRobot.java" <<'JAVA'
import java.awt.Robot;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
public final class PaladinsInputRobot {
  public static void main(String[] args) throws Exception {
    Robot r = new Robot(); r.setAutoDelay(60); r.mouseMove(640,360);
    r.mousePress(InputEvent.BUTTON1_DOWN_MASK); r.mouseRelease(InputEvent.BUTTON1_DOWN_MASK); Thread.sleep(350);
    r.keyPress(KeyEvent.VK_W); Thread.sleep(300);
    for(int i=0;i<3;i++){ r.mousePress(InputEvent.BUTTON1_DOWN_MASK); r.mouseRelease(InputEvent.BUTTON1_DOWN_MASK); Thread.sleep(250); }
    Thread.sleep(600); r.keyPress(KeyEvent.VK_SPACE); Thread.sleep(160); r.keyRelease(KeyEvent.VK_SPACE); r.keyRelease(KeyEvent.VK_W);
  }
}
JAVA
javac "$ROBOT_DIR/PaladinsInputRobot.java"

# CI-only bootstrap: compile a disposable client event subscriber into the dev run. It invokes
# Minecraft 1.20.1's real ConnectScreen network path after client initialization, bypassing the
# flaky title-screen Quick Play state observed in runs #206/#207. The helper is never committed
# into product sources, never added to the certified release JAR/source ZIP, and is removed on exit.
mkdir -p "$(dirname "$QA_HELPER")"
cat > "$QA_HELPER" <<'JAVA'
package net.paladins.forge.client;

import net.minecraft.client.MinecraftClient;
import net.minecraft.client.gui.screen.ConnectScreen;
import net.minecraft.client.network.ServerAddress;
import net.minecraft.client.network.ServerInfo;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.event.TickEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;
import net.paladins.PaladinsMod;

@Mod.EventBusSubscriber(modid = PaladinsMod.ID, value = Dist.CLIENT)
public final class PaladinsQaAutoConnect {
    private static int ticks;
    private static boolean triggered;

    private PaladinsQaAutoConnect() { }

    @SubscribeEvent
    public static void onClientTick(TickEvent.ClientTickEvent event) {
        if (event.phase != TickEvent.Phase.END || triggered) return;
        MinecraftClient client = MinecraftClient.getInstance();
        if (client.currentScreen == null || ++ticks < 40) return;
        triggered = true;
        String target = "127.0.0.1:25565";
        System.out.println("[Paladins player QA] PALADINS_QA_AUTOCONNECT_TRIGGERED: " + target);
        ConnectScreen.connect(
                client.currentScreen,
                client,
                ServerAddress.parse(target),
                new ServerInfo("Paladins Player QA", target, false),
                false);
    }
}
JAVA

grep -Fq 'PALADINS_QA_AUTOCONNECT_TRIGGERED' "$QA_HELPER"
grep -Fq 'ConnectScreen.connect(' "$QA_HELPER"
echo '[Paladins player QA] deterministic dev-client autoconnect helper staged.'

# Start the exact certified packaged Forge server first. The player assertions are driven through
# its real console against a genuine ServerPlayer; this avoids synthetic LivingEntity fixtures.
rm -rf "$FRESH/logs" "$RUN/logs"
mkdir -p "$RUN/logs" "$RUN/config"
printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"
: > "$SERVER_LOG"; : > "$CLIENT_LOG"
rm -f "$FIFO"; mkfifo "$FIFO"; exec 9<> "$FIFO"
( cd "$FRESH" && exec ./run.sh nogui < "$FIFO" ) > "$SERVER_LOG" 2>&1 & SERVER_PID=$!
wait_marker 'Done (' 180 'exact packaged server readiness'
echo '[Paladins player QA] exact certified packaged Forge server ready for real-player connection.'

export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe ALSOFT_DRIVERS=null
Xvfb "$DISPLAY" -screen 0 1280x720x24 -nolisten tcp > "$PORT/paladins-player-xvfb.log" 2>&1 & XVFB_PID=$!
sleep 1
kill -0 "$XVFB_PID" 2>/dev/null || { echo '[Paladins player QA] Xvfb failed to remain alive' >&2; cat "$PORT/paladins-player-xvfb.log" >&2 || true; exit 1; }
( gradle --no-daemon -p "$PORT" :forge:runClient "${ARGS[@]}" --args='--width 1280 --height 720' </dev/null ) > "$CLIENT_LOG" 2>&1 & CLIENT_PID=$!
wait_marker 'PALADINS_QA_AUTOCONNECT_TRIGGERED' 210 'deterministic native-client autoconnect trigger'
wait_marker 'joined the game' 90 'native Forge client joining exact packaged server'
grep -Fq 'Backend library: LWJGL' "$CLIENT_LATEST" || { echo '[Paladins player QA] real player joined without LWJGL evidence' >&2; exit 1; }
[[ "$(sha256sum "$PAL_JAR" | awk '{print $1}')" = "$EXPECTED_SHA" ]] || { echo '[Paladins player QA] packaged Paladins release changed during QA bootstrap' >&2; exit 1; }
echo '[Paladins player QA] PALADINS_REAL_PLAYER_JOIN_PASS: native LWJGL Forge client joined exact certified packaged server through the real ConnectScreen network stack.'

send_cmd 'gamerule doMobSpawning false'
send_cmd 'gamerule doDaylightCycle false'
send_cmd 'time set noon'
send_cmd 'weather clear'
send_cmd 'scoreboard objectives add palp dummy'
send_cmd 'gamemode survival @a'
send_cmd 'effect clear @a'
send_cmd 'clear @a'
send_cmd 'effect give @a minecraft:instant_health 1 10 true'
send_cmd 'effect give @a paladins:divine_protection 30 0 true'
send_cmd 'damage @a 5 minecraft:generic'
sleep 1
send_cmd 'execute store result score divine1_hp palp run data get entity @a[limit=1] Health 10'
send_cmd 'execute store success score divine1_left palp run effect clear @a paladins:divine_protection'
send_cmd 'execute if score divine1_hp palp matches 200 if score divine1_left palp matches 0 run say PALADINS_DIVINE_BLOCK_CONSUME_PASS'
wait_marker 'PALADINS_DIVINE_BLOCK_CONSUME_PASS' 30 'Divine Protection one-charge real-player block/consume'

send_cmd 'effect clear @a'
send_cmd 'effect give @a minecraft:instant_health 1 10 true'
send_cmd 'effect give @a paladins:divine_protection 30 1 true'
send_cmd 'damage @a 5 minecraft:generic'
sleep 1
send_cmd 'execute store result score divine2_first_hp palp run data get entity @a[limit=1] Health 10'
send_cmd 'damage @a 5 minecraft:generic'
sleep 1
send_cmd 'execute store result score divine2_hp palp run data get entity @a[limit=1] Health 10'
send_cmd 'execute store success score divine2_left palp run effect clear @a paladins:divine_protection'
send_cmd 'execute if score divine2_first_hp palp matches 200 if score divine2_hp palp matches 200 if score divine2_left palp matches 0 run say PALADINS_DIVINE_TWO_CHARGE_PASS'
wait_marker 'PALADINS_DIVINE_TWO_CHARGE_PASS' 30 'Divine Protection two-charge real-player decrement/consume'

send_cmd 'kill @e[tag=palqa_judgement_target]'
send_cmd 'effect clear @a'
send_cmd 'clear @a'
send_cmd 'fill -2 99 -2 2 99 8 minecraft:stone'
send_cmd 'fill -2 100 -2 2 103 8 minecraft:air'
send_cmd 'tp @a 0.5 100 0.5 0 0'
send_cmd 'give @a minecraft:stone_sword 1'
send_cmd 'summon minecraft:husk 0.5 100 3.5 {Tags:["palqa_judgement_target"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'effect give @a paladins:judgement 20 0 true'
send_cmd 'say PALADINS_JUDGEMENT_STUN_INPUT_READY'
wait_marker 'PALADINS_JUDGEMENT_STUN_INPUT_READY' 30 'real-player Judgement input readiness'
java -cp "$ROBOT_DIR" PaladinsInputRobot
sleep 1
send_cmd 'execute positioned 0.5 100 0.5 if entity @a[distance=..0.25] run say PALADINS_JUDGEMENT_STUN_MOVE_BLOCK_PASS'
send_cmd 'execute store result score stun_hp palp run data get entity @e[tag=palqa_judgement_target,limit=1] Health 10'
send_cmd 'execute if score stun_hp palp matches 200 run say PALADINS_JUDGEMENT_STUN_ATTACK_BLOCK_PASS'
wait_marker 'PALADINS_JUDGEMENT_STUN_MOVE_BLOCK_PASS' 20 'Judgement movement block'
wait_marker 'PALADINS_JUDGEMENT_STUN_ATTACK_BLOCK_PASS' 20 'Judgement attack block'

send_cmd 'effect clear @a paladins:judgement'
send_cmd 'kill @e[tag=palqa_judgement_target]'
send_cmd 'tp @a 0.5 100 0.5 0 0'
send_cmd 'summon minecraft:husk 0.5 100 3.5 {Tags:["palqa_judgement_target"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'say PALADINS_JUDGEMENT_CONTROL_INPUT_READY'
wait_marker 'PALADINS_JUDGEMENT_CONTROL_INPUT_READY' 20 'control input readiness'
java -cp "$ROBOT_DIR" PaladinsInputRobot
sleep 1
send_cmd 'execute positioned 0.5 100 0.5 unless entity @a[distance=..0.50] run say PALADINS_JUDGEMENT_CONTROL_MOVE_PASS'
send_cmd 'execute store result score control_hp palp run data get entity @e[tag=palqa_judgement_target,limit=1] Health 10'
send_cmd 'execute if score control_hp palp matches ..199 run say PALADINS_JUDGEMENT_CONTROL_ATTACK_PASS'
wait_marker 'PALADINS_JUDGEMENT_CONTROL_MOVE_PASS' 20 'control movement proof'
wait_marker 'PALADINS_JUDGEMENT_CONTROL_ATTACK_PASS' 20 'control attack proof'
send_cmd 'say PALADINS_PLAYER_BEHAVIOR_QA_FINISHED'
wait_marker 'PALADINS_PLAYER_BEHAVIOR_QA_FINISHED' 20 'player QA completion'

send_cmd 'kill @e[tag=palqa_judgement_target]'
send_cmd 'stop'
exec 9>&-
for _ in $(seq 1 30); do kill -0 "$SERVER_PID" 2>/dev/null || break; sleep 1; done
if kill -0 "$SERVER_PID" 2>/dev/null; then echo '[Paladins player QA] packaged server did not stop cleanly after real-player QA' >&2; exit 1; fi
wait "$SERVER_PID"; SERVER_PID=''

[[ "$(sha256sum "$PAL_JAR" | awk '{print $1}')" = "$EXPECTED_SHA" ]] || { echo '[Paladins player QA] packaged Paladins release identity drifted after real-player QA' >&2; exit 1; }
echo '[Paladins player QA] PLAYER_BEHAVIOR_ACCEPTANCE_PASS: exact packaged-server + native LWJGL client runtime proved Divine Protection one/two-charge damage interception and Judgement/STUN input blocking with positive post-clear controls.'
