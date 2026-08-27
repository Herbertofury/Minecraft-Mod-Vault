#!/usr/bin/env python3
from pathlib import Path
import sys

root = Path(sys.argv[1] if len(sys.argv) > 1 else "sable")
props = root / "gradle.properties"
build = root / "forge" / "build.gradle"
mapper = root / "buildSrc" / "src" / "main" / "java" / "RapierProductionJarMapper.java"

p = props.read_text(encoding="utf-8")
np = p.replace("forge_version=47.4.20", "forge_version=47.4.23").replace("org.gradle.parallel=true", "org.gradle.parallel=false")
if np == p:
    if "forge_version=47.4.23" not in p:
        raise SystemExit("unexpected forge_version in gradle.properties")
    if "org.gradle.parallel=false" not in p:
        raise SystemExit("unexpected org.gradle.parallel in gradle.properties")
props.write_text(np, encoding="utf-8")

text = build.read_text(encoding="utf-8")
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
if old_repo in text:
    text = text.replace(old_repo, new_repo, 1)
elif "url = file('build/targetModpackRepository')" not in text:
    raise SystemExit("targetModpackRepository repository block not found")

marker = """        expectedTargetArtifacts.each { key, spec ->
            logger.lifecycle(\"Staged ${spec.group}:${spec.artifact}:${spec.version} (${spec.sha256})\")
        }
"""
stage_maven = """        expectedTargetArtifacts.each { key, spec ->
            File flatJar = new File(repositoryDir, spec.file)
            if (!flatJar.isFile()) {
                throw new GradleException(\"Missing staged target jar for local Maven publication: ${flatJar}\")
            }
            File moduleDir = new File(repositoryDir, \"local/target/${spec.artifact}/${spec.version}\")
            if (!moduleDir.mkdirs() && !moduleDir.isDirectory()) {
                throw new GradleException(\"Unable to create local Maven module directory: ${moduleDir}\")
            }
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
if marker in text:
    text = text.replace(marker, stage_maven, 1)
elif "Published local.target:${spec.artifact}:${spec.version}" not in text:
    raise SystemExit("target artifact logger marker not found")

graph_marker = "ragdoll-port explicit refmap verification dependencies"
if graph_marker not in text:
    text += """

// ragdoll-port explicit refmap verification dependencies
[
    'verifySubLevelEntityCollisionBoundary',
    'verifyBasicSubLevelRenderLifecycle',
    'verifySableSpawnCommandBoundary',
    'verifyRapierPackagedArtifact',
    'verifyRapierProductionNamespace',
    'verifyCompanionProductionNamespace',
    'verifyProductionMinecraftAccessBoundary'
].each { verificationTask ->
    tasks.named(verificationTask) {
        dependsOn tasks.named('syncMixinRefmap')
    }
}
"""
build.write_text(text, encoding="utf-8")

m = mapper.read_text(encoding="utf-8")
needle = """        private boolean resolvesNamedMember(String owner, String name, String descriptor, boolean method) {
            if (method ? declaresMethod(owner, name, descriptor) : declaresField(owner, name, descriptor)) {
                return true;
            }
"""
replacement = """        private boolean resolvesNamedMember(String owner, String name, String descriptor, boolean method) {
            if (method ? declaresMethod(owner, name, descriptor) : declaresField(owner, name, descriptor)) {
                return true;
            }
            // JDK classes are deliberately outside the packaged ABI hierarchy. Enum methods are
            // inherited from java/lang/Enum, so recognize only those exact inherited methods.
            if (method && supertypes(owner).contains(\"java/lang/Enum\")) {
                if ((\"ordinal\".equals(name) && \"()I\".equals(descriptor))
                        || (\"name\".equals(name) && \"()Ljava/lang/String;\".equals(descriptor))) {
                    return true;
                }
            }
"""
if needle in m:
    m = m.replace(needle, replacement, 1)
elif "supertypes(owner).contains(\"java/lang/Enum\")" not in m:
    raise SystemExit("Rapier ABI resolver marker not found")
mapper.write_text(m, encoding="utf-8")

print("Sable Forge 1.20.1 patch applied: Forge 47.4.23, sequential Gradle, local Maven staging, refmap graph, strict Enum ABI inheritance")
