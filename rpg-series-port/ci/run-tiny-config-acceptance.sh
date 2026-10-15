#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/tiny-config-forge-1.20.1"
UP="$ROOT/.rpg-upstream/tiny-config-3.1.0"
GEN="$PORT/generated"
TARGET_SHA="e20fc8ac72fde8274f0df72de2ebb81ffe6f8727"
ENV_FILE="$PORT/TINY_CONFIG_GRADUATION.env"
ACTIVE_PID=""

source "$ENV_FILE"

stop_tree() {
  local root="${1:-}" child kids
  [[ -n "$root" ]] || return 0
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do stop_tree "$child"; done
  kill -TERM "$root" 2>/dev/null || true
  sleep 1
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}
cleanup() {
  [[ -z "${ACTIVE_PID:-}" ]] || stop_tree "$ACTIVE_PID"
  ACTIVE_PID=""
}
trap cleanup EXIT INT TERM

clone_exact() {
  local repo="$1" sha="$2" dest="$3"
  rm -rf "$dest"
  mkdir -p "$(dirname "$dest")"
  git init -q "$dest"
  git -C "$dest" remote add origin "https://github.com/${repo}.git"
  git -C "$dest" fetch -q --depth=1 origin "$sha"
  git -C "$dest" checkout -q --detach FETCH_HEAD
  [[ "$(git -C "$dest" rev-parse HEAD)" = "$sha" ]]
}

pick_release_jar() {
  find "$GEN/forge/build/libs" -maxdepth 1 -type f -name '*.jar' \
    ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1
}

prepare() {
  python3 "$PORT/tools/prepare_port.py" "$UP" "$GEN"
  grep -Fx "target=$TARGET_SHA" "$GEN/PORT-PINS.txt" >/dev/null
  grep -Fx 'target_version=3.1.0' "$GEN/PORT-PINS.txt" >/dev/null
  grep -Fx 'minecraft=1.20.1' "$GEN/PORT-PINS.txt" >/dev/null
  grep -Fx 'forge=47.4.23' "$GEN/PORT-PINS.txt" >/dev/null
  grep -Fx 'java=17' "$GEN/PORT-PINS.txt" >/dev/null
}

rm -rf "$ROOT/.rpg-upstream"
clone_exact ZsoltMolnarrr/TinyConfig "$TARGET_SHA" "$UP"
prepare

# Fail closed on the modern 3.1.0 API surface. The legacy 2.3.2 net.tinyconfig
# package is intentionally NOT the authority for this graduation lane.
for required in \
  'common/src/main/java/net/tiny_config/ConfigManager.java' \
  'common/src/main/java/net/tiny_config/Platform.java' \
  'common/src/main/java/net/tiny_config/versioning/Versionable.java' \
  'common/src/main/java/net/tiny_config/versioning/VersionableConfig.java' \
  'forge/src/main/java/net/tiny_config/forge/PlatformImpl.java' \
  'forge/src/main/java/net/tiny_config/forge/ExampleModForge.java'; do
  [[ -f "$GEN/$required" ]] || { echo "[TinyConfig] required 3.1.0 surface missing: $required" >&2; exit 1; }
done
if grep -R -E 'net\.neoforged|Platform\.Type\.NEOFORGE|package net\.tinyconfig([.;]|$)' \
  "$GEN/common/src/main/java" "$GEN/forge/src/main/java"; then
  echo '[TinyConfig] stale NeoForge or legacy 2.3.2 namespace leaked into the modern Forge port' >&2
  exit 1
fi

grep -Fq 'private final Object ioLock = new Object();' "$GEN/common/src/main/java/net/tiny_config/ConfigManager.java"
grep -Fq 'public Config safeValue()' "$GEN/common/src/main/java/net/tiny_config/ConfigManager.java"
grep -Fq 'public Builder schemaVersion(int required)' "$GEN/common/src/main/java/net/tiny_config/ConfigManager.java"
grep -Fq 'public Builder validate(Function<Config, Boolean> validator)' "$GEN/common/src/main/java/net/tiny_config/ConfigManager.java"
grep -Fq 'public Builder constrain(Function<Config, Config> constraint)' "$GEN/common/src/main/java/net/tiny_config/ConfigManager.java"

echo '[TinyConfig] Deterministic source package gate'
SOURCE_ZIP="$ROOT/tiny-config-3.1.0-forge-1.20.1-source-ci.zip"
rm -f "$SOURCE_ZIP"
python3 - "$GEN" "$SOURCE_ZIP" <<'PY_SOURCE'
from pathlib import Path
import stat, sys, zipfile
src=Path(sys.argv[1]).resolve(); out=Path(sys.argv[2]).resolve()
skip={'.gradle','build','run','runs','.git'}
files=[]
for p in src.rglob('*'):
    rel=p.relative_to(src)
    if any(part in skip for part in rel.parts):
        continue
    if p.is_file():
        files.append((rel.as_posix(),p))
with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED,compresslevel=9) as z:
    for name,p in sorted(files):
        info=zipfile.ZipInfo(name,(1980,1,1,0,0,0))
        info.compress_type=zipfile.ZIP_DEFLATED
        info.external_attr=(stat.S_IFREG|0o644)<<16
        info.create_system=3
        z.writestr(info,p.read_bytes(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
PY_SOURCE
unzip -tq "$SOURCE_ZIP" >/dev/null
SOURCE_SHA="$(sha256sum "$SOURCE_ZIP" | awk '{print $1}')"
printf '%s  %s\n' "$SOURCE_SHA" "$SOURCE_ZIP" | tee "$PORT/tiny-config-source.sha256"

echo '[TinyConfig] Compile + remapped package gate'
gradle --no-daemon --stacktrace -p "$GEN" clean :common:jar :forge:remapJar
JAR="$(pick_release_jar)"
[[ -n "$JAR" && -f "$JAR" ]]
unzip -tq "$JAR" >/dev/null
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="tiny_config"' >/dev/null
unzip -p "$JAR" META-INF/mods.toml | grep -F 'versionRange="[1.20.1,1.20.2)"' >/dev/null
unzip -p "$JAR" META-INF/MANIFEST.MF | grep -F 'MixinConfigs: tiny_config.mixins.json' >/dev/null
unzip -Z1 "$JAR" | grep -Fx 'net/tiny_config/ConfigManager.class' >/dev/null
unzip -Z1 "$JAR" | grep -Fx 'net/tiny_config/versioning/VersionableConfig.class' >/dev/null
unzip -Z1 "$JAR" | grep -Fx 'net/tiny_config/forge/PlatformImpl.class' >/dev/null
if unzip -Z1 "$JAR" | grep -E '(^|/)(net/neoforged|net/tinyconfig)/' >/dev/null; then
  echo '[TinyConfig] legacy/NeoForge implementation leaked into release JAR' >&2; exit 1
fi

python3 - "$JAR" <<'PY_CLASS'
import struct,sys,zipfile
jar=sys.argv[1]; owned=total=0; bad=[]; newer=[]
with zipfile.ZipFile(jar) as zf:
    for name in zf.namelist():
        if not name.endswith('.class'): continue
        data=zf.read(name); total += 1
        if len(data)<8 or data[:4]!=b'\xca\xfe\xba\xbe': bad.append(name); continue
        major=struct.unpack('>H',data[6:8])[0]
        if major>61: newer.append((name,major))
        if name.startswith('net/tiny_config/'):
            owned += 1
            if major!=61: bad.append(f'{name}=major{major}')
if owned < 8: raise SystemExit(f'[TinyConfig] incomplete owned class inventory: {owned}')
if bad: raise SystemExit('[TinyConfig] invalid/non-Java17 owned classes: '+', '.join(bad[:20]))
if newer: raise SystemExit('[TinyConfig] packaged class newer than Java17: '+', '.join(f'{n}={m}' for n,m in newer[:20]))
print(f'[TinyConfig] JAVA17_PACKAGE_PASS owned={owned} total={total}')
PY_CLASS
FIRST_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$FIRST_SHA" "$JAR" | tee "$PORT/tiny-config-first-build.sha256"
RELEASE_JAR="$PORT/tiny-config-forge-3.1.0+1.20.1-release.jar"
cp -f "$JAR" "$RELEASE_JAR"

# A second clean remap must be byte-identical. Deterministic Architectury IDs are
# part of prepare_port.py specifically so a build-only namespace cannot masquerade
# as product drift.
echo '[TinyConfig] Clean reproducibility gate'
gradle --no-daemon --stacktrace -p "$GEN" clean :common:jar :forge:remapJar
JAR2="$(pick_release_jar)"
SECOND_SHA="$(sha256sum "$JAR2" | awk '{print $1}')"
[[ "$FIRST_SHA" = "$SECOND_SHA" ]] || { echo "[TinyConfig] non-deterministic release: $FIRST_SHA != $SECOND_SHA" >&2; exit 1; }
printf '%s  %s\n' "$SECOND_SHA" "$RELEASE_JAR" | tee "$PORT/tiny-config-forge.sha256"
cmp -s "$JAR2" "$RELEASE_JAR"

# Real semantic test inside initialized Forge. The QA hook is injected only into
# the disposable generated source AFTER the untouched release JAR has been sealed.
# It is never copied into RELEASE_JAR.
echo '[TinyConfig] Real Forge ConfigManager semantic gate'
QA_JAVA="$GEN/forge/src/main/java/net/tiny_config/TinyConfigGraduationSelfTest.java"
cat > "$QA_JAVA" <<'JAVA_QA'
package net.tiny_config;

import net.minecraftforge.fml.loading.FMLPaths;
import net.tiny_config.versioning.VersionableConfig;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.atomic.AtomicReference;

public final class TinyConfigGraduationSelfTest {
    private TinyConfigGraduationSelfTest() {}

    public static final class Versioned extends VersionableConfig {
        public int value;
        public Versioned() { this.value = 5; }
        public Versioned(int value) { this.value = value; }
    }
    public static final class Plain {
        public int value;
        public Plain() { this.value = 42; }
    }

    private static void require(boolean ok, String message) {
        if (!ok) throw new IllegalStateException("[TinyConfig CI] " + message);
    }

    public static void run() throws Exception {
        require(Platform.Forge, "Platform.Forge is false in Forge runtime");
        require(!Platform.NeoForge, "Platform.NeoForge leaked into Forge runtime");
        Path forgeConfig = FMLPaths.CONFIGDIR.get().toAbsolutePath().normalize();
        Path platformConfig = Platform.util().getConfigDir().toAbsolutePath().normalize();
        require(forgeConfig.equals(platformConfig), "Platform config directory does not match FMLPaths.CONFIGDIR");

        Path qaDir = forgeConfig.resolve("tiny-config-graduation-qa");
        Files.createDirectories(qaDir);
        Path versionedFile = qaDir.resolve("versioned.json");
        Files.writeString(versionedFile, "{\"schema_version\":3,\"value\":99}\n");

        ConfigManager<Versioned> manager = new ConfigManager<>("versioned", new Versioned(5))
                .builder()
                .setDirectory("tiny-config-graduation-qa")
                .sanitize(true)
                .schemaVersion(3)
                .validate(v -> v.value >= 0)
                .constrain(v -> { v.value = Math.min(v.value, 10); return v; })
                .build();
        manager.refresh();
        require(manager.value.value == 10, "constraint did not clamp valid config to 10");
        require(manager.value.schema_version == 3, "schema version did not remain 3");
        String sanitized = Files.readString(versionedFile);
        require(sanitized.contains("\"value\": 10"), "sanitize did not persist constrained value");

        Files.writeString(versionedFile, "{\"schema_version\":2,\"value\":7}\n");
        manager.refresh();
        require(manager.value.value == 10, "stale schema replaced the accepted value");
        require(manager.value.schema_version == 3, "stale schema was not restored to required schema");
        require(Files.readString(versionedFile).contains("\"value\": 10"), "stale schema rewrite lost safe value");

        Files.writeString(versionedFile, "{\"schema_version\":3,\"value\":-5}\n");
        manager.refresh();
        require(manager.value.value == 10, "validator accepted an invalid negative value");
        require(Files.readString(versionedFile).contains("\"value\": 10"), "invalid value was not sanitized back to safe state");

        boolean schemaGuard = false;
        try {
            new ConfigManager<>("plain-schema-invalid", new Plain()).builder().schemaVersion(1);
        } catch (ExceptionInInitializerError expected) {
            schemaGuard = true;
        }
        require(schemaGuard, "schemaVersion accepted a non-Versionable config type");

        Path threadFile = qaDir.resolve("threadsafe.json");
        Files.deleteIfExists(threadFile);
        ConfigManager<Plain> lazy = new ConfigManager<>("threadsafe", new Plain())
                .builder().setDirectory("tiny-config-graduation-qa").sanitize(true).build();
        int workers = 12;
        CountDownLatch ready = new CountDownLatch(workers);
        CountDownLatch start = new CountDownLatch(1);
        AtomicReference<Throwable> failure = new AtomicReference<>();
        List<Thread> threads = new ArrayList<>();
        for (int i=0;i<workers;i++) {
            Thread t = new Thread(() -> {
                ready.countDown();
                try {
                    start.await();
                    Plain v = lazy.safeValue();
                    if (v == null || v.value != 42) throw new IllegalStateException("lazy safeValue returned bad value");
                } catch (Throwable t1) {
                    failure.compareAndSet(null, t1);
                }
            }, "tiny-config-qa-" + i);
            threads.add(t); t.start();
        }
        ready.await(); start.countDown();
        for (Thread t : threads) t.join();
        require(failure.get() == null, "concurrent safeValue failed: " + failure.get());
        require(Files.exists(threadFile), "concurrent safeValue did not create sanitized config file");
        require(Files.readString(threadFile).contains("\"value\": 42"), "concurrent safeValue persisted corrupt data");

        System.out.println("TINY_CONFIG_SELF_TEST_PASS");
    }
}
JAVA_QA
ENTRY="$GEN/forge/src/main/java/net/tiny_config/forge/ExampleModForge.java"
python3 - "$ENTRY" <<'PY_HOOK'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text()
old='        ExampleMod.init();\n'
new='        ExampleMod.init();\n        if ("1".equals(System.getenv("TINY_CONFIG_SELF_TEST"))) {\n            try { net.tiny_config.TinyConfigGraduationSelfTest.run(); }\n            catch (Exception e) { throw new RuntimeException(e); }\n        }\n'
if s.count(old)!=1: raise SystemExit('[TinyConfig] expected exactly one ExampleMod.init hook')
p.write_text(s.replace(old,new))
PY_HOOK

rm -rf "$GEN/forge/run/logs"
mkdir -p "$GEN/forge/run"
printf 'eula=true\n' > "$GEN/forge/run/eula.txt"
SERVER_SMOKE="$PORT/tiny-config-server-smoke.log"; : > "$SERVER_SMOKE"
env TINY_CONFIG_SELF_TEST=1 gradle --no-daemon -p "$GEN" :forge:runServer > "$SERVER_SMOKE" 2>&1 &
ACTIVE_PID=$!
DEADLINE=$((SECONDS+150)); PASS=0
FATAL='\[TinyConfig CI\]|MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Exception in server tick loop|The game crashed'
while ((SECONDS<DEADLINE)); do
  LOG=$(find "$GEN/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true)
  FILES=("$SERVER_SMOKE"); [[ -n "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  if grep -Fq 'TINY_CONFIG_SELF_TEST_PASS' "${FILES[@]}" && [[ -n "$LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LOG"; then PASS=1; break; fi
  if ! kill -0 "$ACTIVE_PID" 2>/dev/null; then wait "$ACTIVE_PID" || true; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "$SERVER_SMOKE"; echo '[TinyConfig] semantic Forge server timed out' >&2; exit 1; }
stop_tree "$ACTIVE_PID"; ACTIVE_PID=""
echo '[TinyConfig] FORGE_CONFIG_MANAGER_SEMANTICS_PASS'

# Reconstruct pristine source so the client lane cannot accidentally depend on QA code.
prepare

echo '[TinyConfig] Headless Forge client bootstrap gate'
rm -rf "$GEN/forge/run/logs"
mkdir -p "$GEN/forge/run/config"
printf 'earlyWindowControl = false\n' > "$GEN/forge/run/config/fml.toml"
CLIENT_SMOKE="$PORT/tiny-config-client-smoke.log"; : > "$CLIENT_SMOKE"
env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$GEN" :forge:runClient </dev/null > "$CLIENT_SMOKE" 2>&1 &
ACTIVE_PID=$!
DEADLINE=$((SECONDS+180)); READY=0; PASS=0
FATAL_CLIENT='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError.*tiny_config|ClassNotFoundException.*tiny_config|The game crashed whilst initializing game|Exception in thread "Render thread"|Failed to initialize graphics window|Could not initialize GLFW'
while ((SECONDS<DEADLINE)); do
  LOG=$(find "$GEN/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true)
  FILES=("$CLIENT_SMOKE"); [[ -n "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq "$FATAL_CLIENT" "${FILES[@]}"; then stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  if [[ -n "$LOG" ]] && grep -Fq 'Backend library: LWJGL' "$LOG" && grep -Fq 'Reloading ResourceManager' "$LOG"; then
    [[ "$READY" -ne 0 ]] || READY=$SECONDS
    if ((SECONDS-READY>=5)); then PASS=1; break; fi
  fi
  if ! kill -0 "$ACTIVE_PID" 2>/dev/null; then wait "$ACTIVE_PID" || true; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "$CLIENT_SMOKE"; echo '[TinyConfig] client did not reach post-bootstrap state' >&2; exit 1; }
stop_tree "$ACTIVE_PID"; ACTIVE_PID=""
echo '[TinyConfig] NATIVE_FORGE_CLIENT_BOOTSTRAP_PASS'

# Exact untouched release artifact in a fresh official Forge 47.4.23 server.
echo '[TinyConfig] Fresh packaged Forge server gate'
FRESH="$PORT/.fresh-tiny-config-forge-server"
rm -rf "$FRESH"; mkdir -p "$FRESH/mods"
curl -fsSL 'https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar' -o "$FRESH/forge-installer.jar"
(
  cd "$FRESH"
  java -jar forge-installer.jar --installServer >/dev/null
  printf 'eula=true\n' > eula.txt
  printf '%s\n' '-Xmx2G' > user_jvm_args.txt
  cat > server.properties <<'PROPS'
level-type=minecraft:flat
generate-structures=false
view-distance=3
simulation-distance=3
spawn-protection=0
online-mode=false
PROPS
  cp "$RELEASE_JAR" mods/
)
INSTALLED="$FRESH/mods/$(basename "$RELEASE_JAR")"
cmp -s "$RELEASE_JAR" "$INSTALLED"
[[ "$(sha256sum "$INSTALLED" | awk '{print $1}')" = "$FIRST_SHA" ]]
PACKAGE_SMOKE="$PORT/tiny-config-package-server-smoke.log"; : > "$PACKAGE_SMOKE"
( cd "$FRESH" && exec ./run.sh nogui ) > "$PACKAGE_SMOKE" 2>&1 &
ACTIVE_PID=$!
DEADLINE=$((SECONDS+150)); PASS=0
FATAL_PACKAGE='ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|Exception in server tick loop|The game crashed|Attempted to load class .* for invalid dist DEDICATED_SERVER'
while ((SECONDS<DEADLINE)); do
  LOG="$FRESH/logs/latest.log"
  FILES=("$PACKAGE_SMOKE"); [[ -f "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq "$FATAL_PACKAGE" "${FILES[@]}"; then stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  if [[ -f "$LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LOG"; then PASS=1; break; fi
  if ! kill -0 "$ACTIVE_PID" 2>/dev/null; then wait "$ACTIVE_PID" || true; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "$PACKAGE_SMOKE"; echo '[TinyConfig] packaged server did not reach ready state' >&2; exit 1; }
stop_tree "$ACTIVE_PID"; ACTIVE_PID=""
echo '[TinyConfig] CANONICAL_PACKAGED_SERVER_PASS'

if [[ "$TINY_CONFIG_EXPECTED_JAR_SHA" = '__CAPTURE_AFTER_FIRST_GREEN__' || "$TINY_CONFIG_EXPECTED_SOURCE_SHA" = '__CAPTURE_AFTER_FIRST_GREEN__' ]]; then
  echo "[TinyConfig] TINY_CONFIG_FIRST_GREEN_CAPTURE jar=$FIRST_SHA source=$SOURCE_SHA"
  echo '[TinyConfig] Acceptance is green; freeze both identities and replay before graduation.'
else
  [[ "$FIRST_SHA" = "$TINY_CONFIG_EXPECTED_JAR_SHA" ]] || { echo "[TinyConfig] frozen JAR drifted: $FIRST_SHA != $TINY_CONFIG_EXPECTED_JAR_SHA" >&2; exit 1; }
  [[ "$SOURCE_SHA" = "$TINY_CONFIG_EXPECTED_SOURCE_SHA" ]] || { echo "[TinyConfig] frozen source drifted: $SOURCE_SHA != $TINY_CONFIG_EXPECTED_SOURCE_SHA" >&2; exit 1; }
  echo "[TinyConfig] CROSS_RUN_RELEASE_IDENTITY_PASS jar=$FIRST_SHA source=$SOURCE_SHA"
  echo '[TinyConfig] TINY_CONFIG_GRADUATION_PASS'
fi
