#!/usr/bin/env bash
set -euo pipefail

ROOT="$PWD"
SABLE="$ROOT/sable"
EXPORT="$ROOT/export"
mkdir -p "$ROOT/target_modpack/mods" "$EXPORT"

CREATE="$ROOT/target_modpack/mods/create-1.20.1-6.0.8.jar"
curl -fL --retry 4 --retry-delay 2 -o "$CREATE" 'https://edge.forgecdn.net/files/7178/761/create-1.20.1-6.0.8.jar'
echo "6fbb910c367dbce8e4fc7e5bf64b6edd4de980906ed00af8e47e4af843c0d9b0  $CREATE" | sha256sum -c -

LZ4="$HOME/.gradle/caches/modules-2/files-2.1/at.yawk.lz4/lz4-java/1.11.0/b9669fb5e3ccf50a579c9a4f750cab37ce8fce1b/lz4-java-1.11.0.jar"
mkdir -p "$(dirname "$LZ4")"
curl -fL --retry 4 --retry-delay 2 -o "$LZ4" 'https://repo1.maven.org/maven2/at/yawk/lz4/lz4-java/1.11.0/lz4-java-1.11.0.jar'
echo "535c5578cab5dcd0a438e202df80091632b873c0370c25d9b1c1ad1d73577207  $LZ4" | sha256sum -c -

python3 - <<'PY'
from pathlib import Path
import hashlib, zipfile
create = Path('target_modpack/mods/create-1.20.1-6.0.8.jar')
repo = Path('sable/forge/build/targetModpackRepository')
repo.mkdir(parents=True, exist_ok=True)
specs = [
    ('create-1.20.1','6.0.8',None,'create-1.20.1-6.0.8.jar','6fbb910c367dbce8e4fc7e5bf64b6edd4de980906ed00af8e47e4af843c0d9b0'),
    ('flywheel-forge-1.20.1','1.0.5','META-INF/jarjar/flywheel-forge-1.20.1-1.0.5.jar','flywheel-forge-1.20.1-1.0.5.jar','316ca250f19244956b5f0cd75329309ea65a77b4b8da854389b6a9222e7f427c'),
    ('Registrate','MC1.20-1.3.3','META-INF/jarjar/Registrate-MC1.20-1.3.3.jar','Registrate-MC1.20-1.3.3.jar','226862d4638b77273f4627fbac871aa0b3af584dde377f4ce2cb0c7cc228cf00'),
    ('Ponder-Forge-1.20.1','1.0.91','META-INF/jarjar/Ponder-Forge-1.20.1-1.0.91.jar','Ponder-Forge-1.20.1-1.0.91.jar','86e6b64372aba6d9c56f2c35725ea26d8febf2c75eed9950566e7f2849443b34'),
]
with zipfile.ZipFile(create) as zf:
    for artifact, version, member, filename, expected in specs:
        data = create.read_bytes() if member is None else zf.read(member)
        actual = hashlib.sha256(data).hexdigest()
        if actual != expected:
            raise SystemExit(f'{artifact}:{version} SHA mismatch: {actual}')
        (repo/filename).write_bytes(data)
        module = repo/'local'/'target'/artifact/version
        module.mkdir(parents=True, exist_ok=True)
        (module/filename).write_bytes(data)
        (module/f'{artifact}-{version}.pom').write_text(f'''<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>local.target</groupId>
  <artifactId>{artifact}</artifactId>
  <version>{version}</version>
  <packaging>jar</packaging>
</project>
''', encoding='utf-8')
        print(f'PRESTAGED local.target:{artifact}:{version} sha256={actual}')
PY

cd "$SABLE"
sed -i 's/^forge_version=47\.4\.20$/forge_version=47.4.23/' gradle.properties
sed -i 's/^org\.gradle\.parallel=true$/org.gradle.parallel=false/' gradle.properties

python3 - <<'PY'
from pathlib import Path
path = Path('forge/build.gradle')
text = path.read_text(encoding='utf-8')
old_repo = """    flatDir {
        dirs layout.buildDirectory.dir('targetModpackRepository')
    }
"""
new_repo = """    maven {
        url = file('build/targetModpackRepository')
        metadataSources {
            mavenPom()
            artifact()
        }
    }
    flatDir {
        dirs layout.buildDirectory.dir('targetModpackRepository')
    }
"""
if old_repo not in text:
    raise SystemExit('target repository block not found')
text = text.replace(old_repo, new_repo, 1)

marker = """        expectedTargetArtifacts.each { key, spec ->
            logger.lifecycle(\"Staged ${spec.group}:${spec.artifact}:${spec.version} (${spec.sha256})\")
        }
"""
stage = """        expectedTargetArtifacts.each { key, spec ->
            File flatJar = new File(repositoryDir, spec.file)
            if (!flatJar.isFile()) throw new GradleException(\"Missing staged target jar: ${flatJar}\")
            File moduleDir = new File(repositoryDir, \"local/target/${spec.artifact}/${spec.version}\")
            if (!moduleDir.mkdirs() && !moduleDir.isDirectory()) throw new GradleException(\"Unable to create ${moduleDir}\")
            File moduleJar = new File(moduleDir, \"${spec.artifact}-${spec.version}.jar\")
            Files.copy(flatJar.toPath(), moduleJar.toPath(), StandardCopyOption.REPLACE_EXISTING)
            File modulePom = new File(moduleDir, \"${spec.artifact}-${spec.version}.pom\")
            modulePom.setText(\"\"\"<?xml version=\\\"1.0\\\" encoding=\\\"UTF-8\\\"?>
<project xmlns=\\\"http://maven.apache.org/POM/4.0.0\\\">
  <modelVersion>4.0.0</modelVersion>
  <groupId>local.target</groupId>
  <artifactId>${spec.artifact}</artifactId>
  <version>${spec.version}</version>
  <packaging>jar</packaging>
</project>
\"\"\", 'UTF-8')
            logger.lifecycle(\"Published local.target:${spec.artifact}:${spec.version} -> ${moduleJar}\")
        }

        expectedTargetArtifacts.each { key, spec ->
            logger.lifecycle(\"Staged ${spec.group}:${spec.artifact}:${spec.version} (${spec.sha256})\")
        }
"""
if marker not in text:
    raise SystemExit('target artifact logger marker not found')
text = text.replace(marker, stage, 1)

for task in ('verifySubLevelEntityCollisionBoundary', 'verifyBasicSubLevelRenderLifecycle'):
    token = f"tasks.register('{task}') {{"
    start = text.find(token)
    if start < 0: raise SystemExit(f'{task} not found')
    end = text.find("\ntasks.register('", start + len(token))
    if end < 0: end = len(text)
    block = text[start:end]
    needle = "    dependsOn tasks.named('compileJava')"
    if needle not in block: raise SystemExit(f'{task} compileJava dependency not found')
    replacement = needle + "\n    dependsOn tasks.named('syncMixinRefmap')"
    block = block.replace(needle, replacement, 1)
    text = text[:start] + block + text[end:]

path.write_text(text, encoding='utf-8')
PY

chmod +x gradlew
grep '^forge_version=' gradle.properties
grep '^org.gradle.parallel=' gradle.properties

./gradlew --no-daemon --no-parallel :sable_companion_1_20:build
./gradlew --no-daemon --no-parallel :forge:spotlessApply
./gradlew --no-daemon --no-parallel \
  :forge:verifyTargetModpackDependencies \
  :forge:verifySubLevelEntityCollisionBoundary \
  :forge:verifyBasicSubLevelRenderLifecycle \
  :forge:verifySableSpawnCommandBoundary \
  :forge:build \
  :forge:verifyRapierPackagedArtifact \
  :forge:verifyRapierProductionNamespace \
  :forge:verifyCompanionProductionNamespace \
  :forge:verifyProductionMinecraftAccessBoundary

cd "$ROOT"
find sable/forge/build/libs -maxdepth 1 -type f -name '*-all.jar' -print -exec cp '{}' "$EXPORT/" \;
test "$(find "$EXPORT" -type f -name '*-all.jar' | wc -l)" -ge 1
sha256sum "$EXPORT"/*-all.jar > "$EXPORT/SHA256SUMS.txt"
printf '%s\n' \
  'source=6dd48142a62347f1454559b546cbaa63eca648ac' \
  'forge=47.4.23' \
  'create_sha256=6fbb910c367dbce8e4fc7e5bf64b6edd4de980906ed00af8e47e4af843c0d9b0' \
  'ci_patch=prestage-local-maven+sequential-forgegradle+explicit-refmap-verifier-deps' \
  > "$EXPORT/PROVENANCE.txt"
