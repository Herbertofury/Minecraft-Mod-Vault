#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
FOUNDATION="$PORT/.foundation"
OUT="$PORT/more_rpg_library-forge-2.7.2+1.20.1.jar"
SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.4+1.20.1.jar"
SPELL_POWER_JAR="$FOUNDATION/spell_power-forge-1.6.0+1.20.1-certified.jar"
RANGED_JAR="$FOUNDATION/ranged_weapon_api-forge-2.3.4+1.20.1-certified.jar"
TINY_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
RUNTIME_DEPS="$PORT/.more-rpg-runtime-deps"
CLOTH_FORGE_JAR="$RUNTIME_DEPS/cloth-config-forge-11.1.106.jar"
PLAYER_ANIM_FORGE_JAR="$RUNTIME_DEPS/player-animation-lib-forge-1.0.2+1.19.4.jar"
for f in "$OUT" "$SPELL_ENGINE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$TINY_JAR" "$CLOTH_FORGE_JAR" "$PLAYER_ANIM_FORGE_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done

echo '[More RPG 2.7.2] PRODUCTION_FORGE_CLIENT_BEGIN forge=47.4.23 minecraft=1.20.1 mappings=20230612.114412'
PROD="$PORT/.production-forge-client"; MCROOT="$PROD/minecraft-root"; RUN="$PROD/run"; NATIVES="$PROD/natives"
rm -rf "$PROD"; mkdir -p "$MCROOT" "$RUN/mods" "$RUN/config" "$RUN/saves" "$NATIVES"
printf '{"profiles":{},"settings":{},"version":3}\n' > "$MCROOT/launcher_profiles.json"
INSTALLER="$PROD/forge-1.20.1-47.4.23-installer.jar"
curl -fsSL 'https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar' -o "$INSTALLER"
java -jar "$INSTALLER" --installClient "$MCROOT" > "$PORT/more-rpg-production-client-install.log" 2>&1
VERSION_ID='1.20.1-forge-47.4.23'; FORGE_JSON="$MCROOT/versions/$VERSION_ID/$VERSION_ID.json"; test -f "$FORGE_JSON"
CLIENT_SRG="$(find "$MCROOT/libraries/net/minecraft/client" -type f -name 'client-1.20.1-20230612.114412-srg.jar' | head -n1)"
FORGE_CLIENT="$(find "$MCROOT/libraries/net/minecraftforge/forge/1.20.1-47.4.23" -type f -name 'forge-1.20.1-47.4.23-client.jar' | head -n1)"
FORGE_UNIVERSAL="$(find "$MCROOT/libraries/net/minecraftforge/forge/1.20.1-47.4.23" -type f -name 'forge-1.20.1-47.4.23-universal.jar' | head -n1)"
for f in "$CLIENT_SRG" "$FORGE_CLIENT" "$FORGE_UNIVERSAL"; do test -f "$f"; unzip -tq "$f" >/dev/null; done
printf '[More RPG 2.7.2] PRODUCTION_FORGE_NAMESPACE_ARTIFACTS_PASS client_srg=%s forge_client=%s forge_universal=%s\n' "$(sha256sum "$CLIENT_SRG" | awk '{print $1}')" "$(sha256sum "$FORGE_CLIENT" | awk '{print $1}')" "$(sha256sum "$FORGE_UNIVERSAL" | awk '{print $1}')"

python3 - "$MCROOT" <<'PY'
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
import hashlib, json, os, sys, urllib.request
root=Path(sys.argv[1]); assets=root/'assets'; (assets/'indexes').mkdir(parents=True,exist_ok=True); (assets/'objects').mkdir(parents=True,exist_ok=True)
def get(url):
    req=urllib.request.Request(url,headers={'User-Agent':'MoreRPG-Forge1201-QA/1.0'})
    with urllib.request.urlopen(req,timeout=60) as r: return r.read()
def sha1(data): return hashlib.sha1(data).hexdigest()
manifest_raw=get('https://piston-meta.mojang.com/mc/game/version_manifest_v2.json'); manifest=json.loads(manifest_raw); entry=next(v for v in manifest['versions'] if v['id']=='1.20.1')
version_raw=get(entry['url'])
if entry.get('sha1') and sha1(version_raw)!=entry['sha1']: raise SystemExit('1.20.1 version JSON SHA1 mismatch')
vanilla=root/'versions'/'1.20.1'; vanilla.mkdir(parents=True,exist_ok=True); (vanilla/'1.20.1.json').write_bytes(version_raw); version=json.loads(version_raw)
ai=version['assetIndex']; index_raw=get(ai['url'])
if ai.get('sha1') and sha1(index_raw)!=ai['sha1']: raise SystemExit('asset index SHA1 mismatch')
index_id=ai['id']; (assets/'indexes'/f'{index_id}.json').write_bytes(index_raw); index=json.loads(index_raw); hashes=sorted({o['hash'] for o in index['objects'].values()})
def ensure(h):
    p=assets/'objects'/h[:2]/h; p.parent.mkdir(parents=True,exist_ok=True)
    if p.is_file() and hashlib.sha1(p.read_bytes()).hexdigest()==h: return h
    data=get(f'https://resources.download.minecraft.net/{h[:2]}/{h}')
    if sha1(data)!=h: raise RuntimeError(f'asset SHA1 mismatch {h}')
    tmp=p.with_suffix('.tmp'); tmp.write_bytes(data); os.replace(tmp,p); return h
with ThreadPoolExecutor(max_workers=16) as ex:
    futs=[ex.submit(ensure,h) for h in hashes]
    for f in as_completed(futs): f.result()
for h in hashes:
    p=assets/'objects'/h[:2]/h
    if not p.is_file() or hashlib.sha1(p.read_bytes()).hexdigest()!=h: raise SystemExit(f'asset verify failed {h}')
(root/'asset-index-id.txt').write_text(index_id+'\n'); (root/'asset-object-count.txt').write_text(str(len(hashes))+'\n')
print(f'[More RPG 2.7.2] OFFICIAL_ASSET_CACHE_PASS index={index_id} objects={len(hashes)}')
PY
ASSET_INDEX_ID="$(tr -d '\r\n' < "$MCROOT/asset-index-id.txt")"; ASSET_COUNT="$(tr -d '\r\n' < "$MCROOT/asset-object-count.txt")"

python3 - "$MCROOT" "$FORGE_JSON" "$NATIVES" "$RUN" "$ASSET_INDEX_ID" <<'PY'
from pathlib import Path
import json, os, shutil, sys, zipfile
root=Path(sys.argv[1]); forge_json=Path(sys.argv[2]); natives=Path(sys.argv[3]); run=Path(sys.argv[4]); asset_index=sys.argv[5]
forge=json.loads(forge_json.read_text()); vanilla=json.loads((root/'versions'/'1.20.1'/'1.20.1.json').read_text()); libs=root/'libraries'
def rules_allow(item):
    rules=item.get('rules')
    if not rules: return True
    allowed=False
    for rule in rules:
        osrule=rule.get('os',{}); name=osrule.get('name')
        if name and name!='linux': continue
        arch=osrule.get('arch')
        if arch and arch not in ('x86_64','amd64'): continue
        if rule.get('features'): continue
        allowed=(rule.get('action')=='allow')
    return allowed
def artifact_path(lib):
    dl=lib.get('downloads',{}).get('artifact')
    if dl and dl.get('path'): return libs/dl['path']
    name=lib.get('name','').split('@')[0]; parts=name.split(':')
    if len(parts)<3: return None
    group,artifact,version=parts[:3]; classifier=parts[3] if len(parts)>3 else None; fn=f'{artifact}-{version}' + (f'-{classifier}' if classifier else '') + '.jar'
    return libs/Path(group.replace('.','/'))/artifact/version/fn
cp=[]; seen=set()
for doc in (vanilla,forge):
    for lib in doc.get('libraries',[]):
        if not rules_allow(lib): continue
        p=artifact_path(lib)
        if p and p.is_file() and str(p) not in seen: cp.append(p); seen.add(str(p))
        native_key=lib.get('natives',{}).get('linux')
        if native_key:
            native_key=native_key.replace('${arch}','64'); cl=lib.get('downloads',{}).get('classifiers',{}).get(native_key)
            if not cl or not cl.get('path'): raise SystemExit(f'missing native classifier metadata: {lib.get("name")} {native_key}')
            jar=libs/cl['path']
            if not jar.is_file(): raise SystemExit(f'missing native jar: {jar}')
            with zipfile.ZipFile(jar) as z:
                for info in z.infolist():
                    if info.is_dir() or info.filename.startswith('META-INF/'): continue
                    out=natives/Path(info.filename).name
                    with z.open(info) as src, out.open('wb') as dst: shutil.copyfileobj(src,dst)
vanilla_client=root/'versions'/'1.20.1'/'1.20.1.jar'
if vanilla_client.is_file() and str(vanilla_client) not in seen: cp.append(vanilla_client)
classpath=os.pathsep.join(str(p) for p in cp)
if 'client-1.20.1-20230612.114412-srg.jar' not in classpath: raise SystemExit('production SRG client absent from resolved classpath')
if 'forge-1.20.1-47.4.23-client.jar' not in classpath: raise SystemExit('Forge production client jar absent from resolved classpath')
subs={'${library_directory}':str(libs),'${classpath_separator}':os.pathsep,'${natives_directory}':str(natives),'${classpath}':classpath,'${launcher_name}':'MoreRPG-QA','${launcher_version}':'1.0'}
def expand(s):
    for k,v in subs.items(): s=s.replace(k,v)
    return s
def resolve_args(entries):
    out=[]
    for e in entries:
        if isinstance(e,str): out.append(expand(e)); continue
        if not rules_allow(e): continue
        v=e.get('value',[]); v=[v] if isinstance(v,str) else v; out.extend(expand(x) for x in v)
    return out
jvm=resolve_args(forge.get('arguments',{}).get('jvm',[])); jvm += [f'-Djava.library.path={natives}', '-cp', classpath]; jvm += ['-Xmx4G','-Dmixin.debug.export=true','-Dmixin.debug.export.filter=net.minecraft.client.gui.**','-Dmixin.debug.export.decompile=false','-Dmixin.debug.verbose=true']
main=forge.get('mainClass') or 'cpw.mods.bootstraplauncher.BootstrapLauncher'
game=['--username','MoreRPGQA','--version','1.20.1','--gameDir',str(run),'--assetsDir',str(root/'assets'),'--assetIndex',asset_index,'--uuid','00000000000000000000000000000001','--accessToken','0','--clientId','0','--xuid','0','--userType','legacy','--versionType','release','--launchTarget','forgeclient','--fml.forgeVersion','47.4.23','--fml.mcVersion','1.20.1','--fml.forgeGroup','net.minecraftforge','--fml.mcpVersion','20230612.114412','--width','1280','--height','720','--quickPlaySingleplayer','MRPG-QA']
for name,vals in [('jvm.nul',jvm),('game.nul',game)]:
    with (root/name).open('wb') as f:
        for v in vals: f.write(v.encode()+b'\0')
(root/'main-class.txt').write_text(main+'\n'); (root/'resolved-classpath.txt').write_text(classpath+'\n')
print(f'[More RPG 2.7.2] PRODUCTION_LAUNCH_PROFILE_RESOLVED_PASS main={main} classpath_entries={len(cp)}')
PY

cp "$OUT" "$RUN/mods/more-rpg-library-forge.jar"; cp "$SPELL_ENGINE_JAR" "$RUN/mods/spell-engine-forge.jar"; cp "$SPELL_POWER_JAR" "$RUN/mods/spell-power-forge.jar"; cp "$RANGED_JAR" "$RUN/mods/ranged-weapon-api-forge.jar"; cp "$TINY_JAR" "$RUN/mods/tiny-config-forge.jar"; cp "$CLOTH_FORGE_JAR" "$RUN/mods/cloth-config-forge.jar"; cp "$PLAYER_ANIM_FORGE_JAR" "$RUN/mods/player-animation-lib-forge.jar"
[[ "$(sha256sum "$RUN/mods/more-rpg-library-forge.jar" | awk '{print $1}')" = "$(sha256sum "$OUT" | awk '{print $1}')" ]]
SOURCE_WORLD="$PORT/.fresh-more-rpg-server/world"; [[ -d "$SOURCE_WORLD" ]] || { echo '[More RPG 2.7.2] Stage-1 source world missing for production client' >&2; exit 1; }; cp -a "$SOURCE_WORLD" "$RUN/saves/MRPG-QA"; printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"; rm -rf "$RUN/logs" "$RUN/.mixin.out"
mapfile -d '' -t JVM_ARGS < "$MCROOT/jvm.nul"; mapfile -d '' -t GAME_ARGS < "$MCROOT/game.nul"; MAIN_CLASS="$(tr -d '\r\n' < "$MCROOT/main-class.txt")"; LOG="$PORT/more-rpg-production-client.log"; : > "$LOG"; ACTIVE_PID=''
cleanup() { [[ -z "${ACTIVE_PID:-}" ]] || { kill -TERM "$ACTIVE_PID" 2>/dev/null || true; wait "$ACTIVE_PID" 2>/dev/null || true; }; }; trap cleanup EXIT INT TERM
env -u MOD_CLASSES LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe ALSOFT_DRIVERS=null xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' java "${JVM_ARGS[@]}" "$MAIN_CLASS" "${GAME_ARGS[@]}" </dev/null > "$LOG" 2>&1 &
ACTIVE_PID=$!; PID=$ACTIVE_PID; DEADLINE=$((SECONDS+300)); PASS=0
FATAL='ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Exception in thread "Render thread"|The game crashed|Missing or unsupported mandatory dependencies|Could not initialize GLFW|Failed to initialize graphics window'
while ((SECONDS<DEADLINE)); do LATEST="$RUN/logs/latest.log"; FILES=("$LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST"); if grep -Eiq "$FATAL" "${FILES[@]}"; then tail -n 700 "${FILES[@]}" 2>/dev/null || true; exit 1; fi; if [[ -f "$LATEST" ]] && grep -Fq 'ModLauncher launch target: forgeclient' "$LATEST" && grep -Fq 'ModLauncher naming: srg' "$LATEST" && grep -Fq 'Reloading ResourceManager' "$LATEST" && grep -Fq 'Backend library: LWJGL' "$LATEST" && grep -Fq '[More RPG QA] FATAL_POISON_APPLIED' "${FILES[@]}"; then PASS=1; break; fi; if ! kill -0 "$PID" 2>/dev/null; then wait "$PID" || true; ACTIVE_PID=''; tail -n 700 "${FILES[@]}" 2>/dev/null || true; exit 1; fi; sleep 1; done
[[ "$PASS" -eq 1 ]] || { tail -n 700 "$LOG" "$RUN/logs/latest.log" 2>/dev/null || true; exit 1; }
LATEST="$RUN/logs/latest.log"; grep -Eiq 'DrawHeartsMixin.*more-rpg-classes\.mixins\.json|more-rpg-classes\.mixins\.json.*DrawHeartsMixin' "$LOG" "$LATEST"; MIXIN_OUT="$RUN/.mixin.out"; [[ -d "$MIXIN_OUT" ]]; mapfile -t HEART_TARGETS < <(grep -aRl 'net/more_rpg_classes/client/heart/HeartRegistry' "$MIXIN_OUT" --include='*.class' 2>/dev/null | sort || true); ((${#HEART_TARGETS[@]} > 0)) || { echo '[More RPG 2.7.2] production SRG HUD transform missing HeartRegistry' >&2; exit 1; }
EVIDENCE="$PORT/production-transformed-hud-evidence"; rm -rf "$EVIDENCE"; mkdir -p "$EVIDENCE"; cp "${HEART_TARGETS[0]}" "$EVIDENCE/more-rpg-production-transformed-hud-target.class"; HUD_SHA="$(sha256sum "$EVIDENCE/more-rpg-production-transformed-hud-target.class" | awk '{print $1}')"; printf '%s  more-rpg-production-transformed-hud-target.class\n' "$HUD_SHA" > "$EVIDENCE/more-rpg-production-transformed-hud-target.sha256"; printf '%s\n' "${HEART_TARGETS[0]#$MIXIN_OUT/}" > "$EVIDENCE/more-rpg-production-transformed-hud-target.source-path.txt"; strings -a "$EVIDENCE/more-rpg-production-transformed-hud-target.class" | sort -u > "$EVIDENCE/more-rpg-production-transformed-hud-target.strings.txt"; grep -F 'net/more_rpg_classes/client/heart/HeartRegistry' "$EVIDENCE/more-rpg-production-transformed-hud-target.strings.txt" >/dev/null
kill -TERM "$PID" 2>/dev/null || true; wait "$PID" 2>/dev/null || true; ACTIVE_PID=''; printf '[More RPG 2.7.2] PRODUCTION_DRAW_HEARTS_MIXIN_TRANSFORMED_PASS targets=%s first=%s sha256=%s\n' "${#HEART_TARGETS[@]}" "${HEART_TARGETS[0]#$MIXIN_OUT/}" "$HUD_SHA"; printf '[More RPG 2.7.2] PRODUCTION_FORGE_CLIENT_SRG_PASS assets=%s mod_sha=%s launch_target=forgeclient naming=srg\n' "$ASSET_COUNT" "$(sha256sum "$OUT" | awk '{print $1}')"
