#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
ORIGINAL="$ROOT/rpg-series-port/ci/run-tiny-config-acceptance.sh"
PATCHED="$(mktemp -t tiny-config-acceptance-v2.XXXXXX.sh)"
cleanup_patch() { rm -f "$PATCHED"; }
trap cleanup_patch EXIT INT TERM

python3 - "$ORIGINAL" "$PATCHED" <<'PY'
from pathlib import Path
import sys

src = Path(sys.argv[1]).read_text()
replacements = [
    (
        'QA_JAVA="$GEN/forge/src/main/java/net/tiny_config/TinyConfigGraduationSelfTest.java"\ncat > "$QA_JAVA" <<\'JAVA_QA\'\npackage net.tiny_config;\n\nimport net.minecraftforge.fml.loading.FMLPaths;\nimport net.tiny_config.versioning.VersionableConfig;\n',
        'QA_JAVA="$GEN/forge/src/main/java/net/tiny_config/forge/qa/TinyConfigGraduationSelfTest.java"\nmkdir -p "$(dirname "$QA_JAVA")"\ncat > "$QA_JAVA" <<\'JAVA_QA\'\npackage net.tiny_config.forge.qa;\n\nimport net.minecraftforge.fml.loading.FMLPaths;\nimport net.tiny_config.ConfigManager;\nimport net.tiny_config.Platform;\nimport net.tiny_config.versioning.VersionableConfig;\n'
    ),
    (
        'try { net.tiny_config.TinyConfigGraduationSelfTest.run(); }',
        'try { net.tiny_config.forge.qa.TinyConfigGraduationSelfTest.run(); }'
    ),
    (
        "FATAL='\\[TinyConfig CI\\]|MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Exception in server tick loop|The game crashed'",
        "FATAL='\\[TinyConfig CI\\]|ResolutionException|export package|MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Exception in server tick loop|The game crashed'"
    ),
    (
        '    public static final class Plain {\n',
        '    private static final class PlatformProbe extends Platform {\n        private static Path configDir() { return util().getConfigDir(); }\n    }\n\n    public static final class Plain {\n'
    ),
    (
        'Path platformConfig = Platform.util().getConfigDir().toAbsolutePath().normalize();',
        'Path platformConfig = PlatformProbe.configDir().toAbsolutePath().normalize();'
    ),
]
for old, new in replacements:
    count = src.count(old)
    if count != 1:
        raise SystemExit(f'[TinyConfig v2] expected exactly one harness seam, found {count}: {old[:90]!r}')
    src = src.replace(old, new, 1)
Path(sys.argv[2]).write_text(src)
PY

chmod +x "$PATCHED"
echo '[TinyConfig v2] QA harness isolated from product package; protected platform probe uses legal subclass access'
exec bash "$PATCHED"
