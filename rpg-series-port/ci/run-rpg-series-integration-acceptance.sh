#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
CI="$ROOT/rpg-series-port/ci"
PAL="$ROOT/rpg-series-port/paladins-forge-1.20.1"
ROG="$ROOT/rpg-series-port/rogues-forge-1.20.1"
TMP="${RUNNER_TEMP:-/tmp}/paladins-first-compile"
source "$CI/ROGUES_GRADUATION.env"
PAL_EXPECTED_JAR_SHA='95e8f9e074dd1c432f9486c922173b76a3b3bc57667b28fe154e47dba5c374ee'
PAL_EXPECTED_SOURCE_SHA='fb0e5812857a2fd46de488cd17a80011ef5d18795ff96fa1a3ebed5fd19a4377'
OLD_ROGUES_JAR_SHA='9e8c880f55ab57d91148c0be702a431bad6e312900b25f65c9dbec266e3ca401'

pick_one(){
  local dir="$1" pattern="$2" label="$3"
  local -a files=()
  mapfile -t files < <(find "$dir" -maxdepth 1 -type f -name "$pattern" ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -print | sort)
  (( ${#files[@]} == 1 )) || { echo "[RPG integration] $label expected one $pattern in $dir, found ${#files[@]}" >&2; exit 1; }
  printf '%s\n' "${files[0]}"
}

stop_tree(){
  local root="$1" child kids
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do stop_tree "$child"; done
  kill -TERM "$root" 2>/dev/null || true
  sleep 1
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}

# Preserve the graduated Rogues identity while reusing its deterministic reconstruction lane.
# The deep runner normally performs these substitutions in-memory; integration owns the same
# exact identity seam explicitly so it never validates the obsolete pre-fix candidate.
for script in run-rogues-acceptance.sh run-rogues-release-certification.sh; do
  if grep -Fq "$OLD_ROGUES_JAR_SHA" "$CI/$script"; then
    sed -i "s/$OLD_ROGUES_JAR_SHA/$ROGUES_EXPECTED_JAR_SHA/g" "$CI/$script"
  fi
done
if grep -Fq '__CAPTURE_AFTER_FIRST_DEEP__' "$CI/run-rogues-release-certification.sh"; then
  sed -i "s/__CAPTURE_AFTER_FIRST_DEEP__/$ROGUES_EXPECTED_SOURCE_SHA/g" "$CI/run-rogues-release-certification.sh"
fi

# Reconstruct and certify both leaves independently before they are allowed to coexist.
# Rogues first-compile intentionally reuses Paladins' foundation lane, but that leaves a
# non-certified intermediate Paladins byte stream. Paladins acceptance + certification runs
# second and restores the exact graduated Paladins release identity before integration.
echo '[RPG integration] Reconstructing and certifying Rogues exact graduated identity.'
bash "$CI/run-rogues-acceptance.sh"
bash "$CI/run-rogues-release-certification.sh"
ROG_JAR="$(pick_one "$ROG/forge/build/libs" '*-forge-*.jar' 'Rogues release')"
ROG_SHA="$(sha256sum "$ROG_JAR" | awk '{print $1}')"
[[ "$ROG_SHA" = "$ROGUES_EXPECTED_JAR_SHA" ]] || { echo "[RPG integration] Rogues identity drift: $ROG_SHA" >&2; exit 1; }

# Paladins acceptance creates the superset packaged server (Paladins + ten dependencies).
# Certification canonicalizes its intermediate JAR to the frozen graduated identity.
echo '[RPG integration] Reconstructing and certifying Paladins exact graduated identity.'
bash "$CI/run-paladins-acceptance.sh"
bash "$CI/run-paladins-release-certification.sh"
PAL_JAR="$(pick_one "$PAL/forge/build/libs" '*-forge-*.jar' 'Paladins release')"
PAL_SHA="$(sha256sum "$PAL_JAR" | awk '{print $1}')"
[[ "$PAL_SHA" = "$PAL_EXPECTED_JAR_SHA" ]] || { echo "[RPG integration] Paladins identity drift: $PAL_SHA" >&2; exit 1; }
[[ "$(awk '{print $1}' "$PAL/paladins-source.sha256")" = "$PAL_EXPECTED_SOURCE_SHA" ]] || { echo '[RPG integration] Paladins source identity drift' >&2; exit 1; }
[[ "$(awk '{print $1}' "$ROG/rogues-source.sha256")" = "$ROGUES_EXPECTED_SOURCE_SHA" ]] || { echo '[RPG integration] Rogues source identity drift' >&2; exit 1; }
echo "[RPG integration] EXACT_LEAF_IDENTITIES_PASS paladins=$PAL_SHA rogues=$ROG_SHA"

# Build the combined packaged runtime from the already-certified Paladins server because it
# is the dependency superset: Paladins + Shield API + Runes + the eight dependencies shared
# with Rogues. Add only the exact Rogues release, yielding twelve separate mod JARs.
FRESH="$PAL/.fresh-paladins-forge-server"
[[ -x "$FRESH/run.sh" && -d "$FRESH/mods" ]] || { echo '[RPG integration] certified Paladins packaged server missing' >&2; exit 1; }
rm -f "$FRESH/mods/rogues-forge-"*.jar
cp -f "$ROG_JAR" "$FRESH/mods/rogues-forge-3.1.1+1.20.1.jar"
[[ "$(find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 12 ]] || { echo '[RPG integration] expected exact twelve-mod combined packaged runtime' >&2; find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' -printf '%f\n' | sort; exit 1; }
[[ "$(sha256sum "$FRESH/mods/paladins-forge-3.1.1+1.20.1.jar" | awk '{print $1}')" = "$PAL_EXPECTED_JAR_SHA" ]] || { echo '[RPG integration] installed Paladins bytes drifted' >&2; exit 1; }
[[ "$(sha256sum "$FRESH/mods/rogues-forge-3.1.1+1.20.1.jar" | awk '{print $1}')" = "$ROGUES_EXPECTED_JAR_SHA" ]] || { echo '[RPG integration] installed Rogues bytes drifted' >&2; exit 1; }

# Fail closed on duplicate mod IDs and owned-class leakage. Shared foundations stay separate;
# neither leaf may own the other's namespace. Also require each release to retain its own
# mixin activation metadata at the production boundary.
python3 - "$FRESH/mods" "$FRESH/mods/paladins-forge-3.1.1+1.20.1.jar" "$FRESH/mods/rogues-forge-3.1.1+1.20.1.jar" <<'PY'
from pathlib import Path
import re, sys, zipfile
mods=Path(sys.argv[1]); pal=Path(sys.argv[2]); rog=Path(sys.argv[3])
owners={}
for jar in sorted(mods.glob('*.jar')):
    with zipfile.ZipFile(jar) as z:
        try: toml=z.read('META-INF/mods.toml').decode('utf-8','replace')
        except KeyError: continue
        ids=re.findall(r'(?m)^\s*modId\s*=\s*["\']([^"\']+)',toml)
        for modid in ids:
            previous=owners.setdefault(modid,jar.name)
            if previous != jar.name:
                raise SystemExit(f'[RPG integration] duplicate modId {modid}: {previous} + {jar.name}')
with zipfile.ZipFile(pal) as z:
    names=set(z.namelist()); manifest=z.read('META-INF/MANIFEST.MF').decode('utf-8','replace')
    if any(n.startswith('net/rogues/') for n in names): raise SystemExit('[RPG integration] Rogues classes leaked into Paladins')
    if 'MixinConfigs:' not in manifest: raise SystemExit('[RPG integration] Paladins production mixin activation missing')
with zipfile.ZipFile(rog) as z:
    names=set(z.namelist()); manifest=z.read('META-INF/MANIFEST.MF').decode('utf-8','replace')
    if any(n.startswith('net/paladins/') for n in names): raise SystemExit('[RPG integration] Paladins classes leaked into Rogues')
    if 'MixinConfigs: rogues.mixins.json' not in manifest: raise SystemExit('[RPG integration] Rogues production mixin activation missing')
print('[RPG integration] MOD_OWNERSHIP_COLLISION_PASS ids='+','.join(sorted(owners)))
PY

# Combined packaged Forge dedicated-server gate and cross-mod game-thread semantics.
SERVER_LOG="$ROOT/rpg-series-integration-server.log"
SERVER_LATEST="$FRESH/logs/latest.log"
FIFO="$FRESH/rpg-integration-console.fifo"
SERVER_PID=''
CLIENT_PID=''
XVFB_PID=''
cleanup(){
  exec 9>&- 2>/dev/null || true
  if [[ -n "$CLIENT_PID" ]] && kill -0 "$CLIENT_PID" 2>/dev/null; then stop_tree "$CLIENT_PID"; fi
  if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then stop_tree "$SERVER_PID"; fi
  if [[ -n "$XVFB_PID" ]] && kill -0 "$XVFB_PID" 2>/dev/null; then kill -TERM "$XVFB_PID" 2>/dev/null || true; wait "$XVFB_PID" 2>/dev/null || true; fi
  rm -f "$FIFO" "$ROG/forge/src/main/java/net/rogues/forge/client/RpgSeriesIntegrationAutoConnect.java" "$ROG/forge/build/classes/java/main/net/rogues/forge/client/RpgSeriesIntegrationAutoConnect.class"
}
trap cleanup EXIT
wait_marker(){
  local marker="$1" timeout_seconds="$2" label="$3" deadline=$((SECONDS+timeout_seconds))
  local fatal='ModLoadingException|Failed to create mod instance|Failed to start the minecraft server|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Registry is already frozen|Can not register to a locked registry|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed|Incompatible FML modded server|mismatched mod channel|Connection refused'
  while (( SECONDS < deadline )); do
    if grep -Eiq "$fatal" "$SERVER_LOG" "$SERVER_LATEST" "$ROOT/rpg-series-integration-client.log" "$ROG/forge/run/logs/latest.log" 2>/dev/null; then
      echo "[RPG integration] fatal runtime signature during $label" >&2
      tail -n 500 "$SERVER_LOG" "$SERVER_LATEST" "$ROOT/rpg-series-integration-client.log" "$ROG/forge/run/logs/latest.log" 2>/dev/null || true
      return 1
    fi
    if grep -Fq "$marker" "$SERVER_LOG" 2>/dev/null || grep -Fq "$marker" "$SERVER_LATEST" 2>/dev/null || grep -Fq "$marker" "$ROOT/rpg-series-integration-client.log" 2>/dev/null || grep -Fq "$marker" "$ROG/forge/run/logs/latest.log" 2>/dev/null; then return 0; fi
    if [[ -n "$SERVER_PID" ]] && ! kill -0 "$SERVER_PID" 2>/dev/null; then echo "[RPG integration] server exited before $label: $marker" >&2; tail -n 500 "$SERVER_LOG" "$SERVER_LATEST" 2>/dev/null || true; return 1; fi
    if [[ -n "$CLIENT_PID" ]] && ! kill -0 "$CLIENT_PID" 2>/dev/null; then echo "[RPG integration] client exited before $label: $marker" >&2; tail -n 500 "$ROOT/rpg-series-integration-client.log" "$ROG/forge/run/logs/latest.log" 2>/dev/null || true; return 1; fi
    sleep 1
  done
  echo "[RPG integration] timed out waiting for $label: $marker" >&2
  tail -n 500 "$SERVER_LOG" "$SERVER_LATEST" "$ROOT/rpg-series-integration-client.log" "$ROG/forge/run/logs/latest.log" 2>/dev/null || true
  return 1
}
send_cmd(){ printf '%s\n' "$1" >&9; }
rm -rf "$FRESH/logs"; : > "$SERVER_LOG"; rm -f "$FIFO"; mkfifo "$FIFO"; exec 9<> "$FIFO"
( cd "$FRESH" && exec ./run.sh nogui < "$FIFO" ) > "$SERVER_LOG" 2>&1 & SERVER_PID=$!
wait_marker 'Done (' 180 'combined packaged server readiness'
echo '[RPG integration] COMBINED_PACKAGED_SERVER_READY_PASS'
send_cmd 'gamerule doMobSpawning false'; send_cmd 'scoreboard objectives add rpgi dummy'; send_cmd 'kill @e[tag=rpgi]'
send_cmd 'summon minecraft:cow 0 100 0 {Tags:["rpgi","rpgi_pal"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'effect give @e[tag=rpgi_pal,limit=1] paladins:priest_absorption 30 1 true'
send_cmd 'execute store result score pal_abs rpgi run data get entity @e[tag=rpgi_pal,limit=1] AbsorptionAmount 10'
send_cmd 'execute if score pal_abs rpgi matches 40 run say RPG_INTEGRATION_PALADINS_EFFECT_PASS'
send_cmd 'summon minecraft:cow 4 100 0 {Tags:["rpgi","rpgi_rog"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'effect give @e[tag=rpgi_rog,limit=1] rogues:stealth 30 0 true'
send_cmd 'execute store success score rog_effect rpgi run effect clear @e[tag=rpgi_rog,limit=1] rogues:stealth'
send_cmd 'execute if score rog_effect rpgi matches 1 run say RPG_INTEGRATION_ROGUES_EFFECT_PASS'
send_cmd 'summon rogues:bear_trap 8 100 0 {Tags:["rpgi","rpgi_trap"]}'
send_cmd 'execute if entity @e[type=rogues:bear_trap,tag=rpgi_trap,limit=1] run say RPG_INTEGRATION_ROGUES_ENTITY_PASS'
for marker in RPG_INTEGRATION_PALADINS_EFFECT_PASS RPG_INTEGRATION_ROGUES_EFFECT_PASS RPG_INTEGRATION_ROGUES_ENTITY_PASS; do wait_marker "$marker" 30 "$marker"; done
echo '[RPG integration] COMBINED_SERVER_SEMANTICS_PASS'

# Native LWJGL client coexistence gate. The Rogues userdev client remains the leaf under test;
# Paladins + its two additional runtime-only foundations are supplied through the normal mods
# directory, while the eight common foundations remain the same exact Gradle/runtime inputs.
SHIELD_FORGE="$(pick_one "$ROOT/rpg-series-port/shield-api-forge-1.20.1/forge/build/libs" '*.jar' 'Shield API Forge')"
RUNES_FORGE="$(pick_one "$ROOT/rpg-series-port/runes-forge-1.20.1/forge/build/libs" '*-forge-*.jar' 'Runes Forge')"
SPELL_POWER_COMMON="$(pick_one "$ROOT/rpg-series-port/spell_power-forge-1.20.1/common/build/libs" '*-common-*.jar' 'Spell Power common')"
SPELL_POWER_FORGE="$(pick_one "$ROOT/rpg-series-port/spell_power-forge-1.20.1/forge/build/libs" '*-forge-*.jar' 'Spell Power Forge')"
TINY_COMMON="$(pick_one "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/common/build/libs" '*-common-*.jar' 'TinyConfig common')"
TINY_FORGE="$(pick_one "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" '*-forge-*.jar' 'TinyConfig Forge')"
SPELL_ENGINE_COMMON="$(pick_one "$ROOT/.spell-engine-build/common/build/libs" '*-common-*.jar' 'Spell Engine common')"
SPELL_ENGINE_FORGE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"
STRUCTURE_COMMON="$(pick_one "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1/common/build/libs" '*-common-*.jar' 'Structure Pool common')"
STRUCTURE_FORGE="$(pick_one "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1/forge/build/libs" '*-forge-*.jar' 'Structure Pool Forge')"
ARMOR_COMMON="$(pick_one "$ROOT/rpg-series-port/armor-model-api-forge-1.20.1/generated/common/build/libs" '*-common-*.jar' 'Armor Model common')"
ARMOR_FORGE="$(pick_one "$ROOT/rpg-series-port/armor-model-api-forge-1.20.1/generated/forge/build/libs" '*-forge-*.jar' 'Armor Model Forge')"
CLOTH_FORGE="$TMP/ext/cloth-config-forge-11.1.136.jar"; PLAYER_FORGE="$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"; CURIOS_FORGE="$TMP/ext/curios-forge-5.14.1+1.20.1.jar"
ARGS=("-Parmor_model_api_common_jar=$ARMOR_COMMON" "-Pstructure_pool_api_common_jar=$STRUCTURE_COMMON" "-Pspell_power_common_jar=$SPELL_POWER_COMMON" "-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON" "-Ptiny_config_common_jar=$TINY_COMMON" "-Parmor_model_api_forge_jar=$ARMOR_FORGE" "-Pstructure_pool_api_forge_jar=$STRUCTURE_FORGE" "-Pspell_power_forge_jar=$SPELL_POWER_FORGE" "-Pspell_engine_forge_jar=$SPELL_ENGINE_FORGE" "-Ptiny_config_forge_jar=$TINY_FORGE" "-Pcloth_config_forge_jar=$CLOTH_FORGE" "-Pplayer_animator_forge_jar=$PLAYER_FORGE" "-Pcurios_jar=$CURIOS_FORGE")
RUN="$ROG/forge/run"; mkdir -p "$RUN/mods" "$RUN/config"; rm -f "$RUN/mods/"*.jar
cp -f "$PAL_JAR" "$RUN/mods/paladins-forge-3.1.1+1.20.1.jar"
cp -f "$SHIELD_FORGE" "$RUN/mods/$(basename "$SHIELD_FORGE")"
cp -f "$RUNES_FORGE" "$RUN/mods/$(basename "$RUNES_FORGE")"
printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"
QA_HELPER="$ROG/forge/src/main/java/net/rogues/forge/client/RpgSeriesIntegrationAutoConnect.java"
mkdir -p "$(dirname "$QA_HELPER")"
cat > "$QA_HELPER" <<'JAVA'
package net.rogues.forge.client;
import net.minecraft.client.MinecraftClient;
import net.minecraft.client.gui.screen.ConnectScreen;
import net.minecraft.client.network.ServerAddress;
import net.minecraft.client.network.ServerInfo;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.event.TickEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;
import net.rogues.RoguesMod;
@Mod.EventBusSubscriber(modid=RoguesMod.ID, value=Dist.CLIENT)
public final class RpgSeriesIntegrationAutoConnect {
  private static int ticks; private static boolean triggered;
  private RpgSeriesIntegrationAutoConnect(){}
  @SubscribeEvent public static void onClientTick(TickEvent.ClientTickEvent event){
    if(event.phase!=TickEvent.Phase.END||triggered)return;
    MinecraftClient client=MinecraftClient.getInstance();
    if(client.currentScreen==null||++ticks<40)return;
    triggered=true;
    String target="127.0.0.1:25565";
    System.out.println("[RPG integration] RPG_INTEGRATION_AUTOCONNECT_TRIGGERED: "+target);
    ConnectScreen.connect(client.currentScreen,client,ServerAddress.parse(target),new ServerInfo("RPG Integration",target,false),false);
  }
}
JAVA
rm -rf "$RUN/logs"; : > "$ROOT/rpg-series-integration-client.log"
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe ALSOFT_DRIVERS=null
Xvfb "$DISPLAY" -screen 0 1280x720x24 -nolisten tcp > "$ROOT/rpg-series-integration-xvfb.log" 2>&1 & XVFB_PID=$!; sleep 1
kill -0 "$XVFB_PID" 2>/dev/null || { cat "$ROOT/rpg-series-integration-xvfb.log" >&2 || true; exit 1; }
( gradle --no-daemon -p "$ROG" :forge:runClient "${ARGS[@]}" --args='--width 1280 --height 720' </dev/null ) > "$ROOT/rpg-series-integration-client.log" 2>&1 & CLIENT_PID=$!
wait_marker 'RPG_INTEGRATION_AUTOCONNECT_TRIGGERED' 210 'integration native-client autoconnect'
wait_marker 'joined the game' 90 'integration real client join'
grep -Fq 'Backend library: LWJGL' "$RUN/logs/latest.log" || { echo '[RPG integration] joined without LWJGL evidence' >&2; exit 1; }
# A successful modded handshake plus server-side execution of both namespaces proves one live
# client/server session negotiated the combined registry/channel set. Reject client resource or
# model failures tied to either leaf even if the connection technically succeeds.
if grep -Eiq 'Using missing texture.*(paladins|rogues)|Failed to load model.*(paladins|rogues)|Unable to load model.*(paladins|rogues)' "$ROOT/rpg-series-integration-client.log" "$RUN/logs/latest.log"; then
  echo '[RPG integration] leaf-owned resource/model failure in combined native client' >&2; exit 1
fi
send_cmd 'execute if entity @a[limit=1] run say RPG_INTEGRATION_REAL_PLAYER_JOIN_PASS'
wait_marker 'RPG_INTEGRATION_REAL_PLAYER_JOIN_PASS' 30 'combined real-player server acknowledgement'
echo '[RPG integration] COMBINED_NATIVE_CLIENT_JOIN_PASS: native LWJGL Forge client joined the twelve-mod packaged server with Paladins and Rogues active together.'

send_cmd 'kill @e[tag=rpgi]'; send_cmd 'stop'; exec 9>&-
for _ in $(seq 1 30); do kill -0 "$SERVER_PID" 2>/dev/null || break; sleep 1; done
if kill -0 "$SERVER_PID" 2>/dev/null; then echo '[RPG integration] combined packaged server did not stop cleanly' >&2; exit 1; fi
wait "$SERVER_PID"; SERVER_PID=''
if [[ -n "$CLIENT_PID" ]] && kill -0 "$CLIENT_PID" 2>/dev/null; then stop_tree "$CLIENT_PID"; CLIENT_PID=''; fi

echo '[RPG integration] FULL_RPG_SERIES_INTEGRATION_PASS: exact graduated Paladins + Rogues identities, twelve-mod ownership/collision audit, combined packaged Forge server semantics, and native LWJGL real-client join all passed.'
