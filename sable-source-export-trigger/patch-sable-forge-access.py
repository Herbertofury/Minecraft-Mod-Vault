from pathlib import Path

path = Path("sable/forge/build.gradle")
text = path.read_text(encoding="utf-8")
old = '''                File forgeUniversalJar = new File(gradle.gradleUserHomeDir,
                        "caches/forge_gradle/maven_downloader/net/minecraftforge/forge/${forge_minecraft_version}-${forge_version}/forge-${forge_minecraft_version}-${forge_version}-universal.jar")
                List<File> minecraftVisibilityClasspath = [minecraftSrgJar]
                if (forgeBinpatchedJar.isFile()) {
                    minecraftVisibilityClasspath.add(forgeBinpatchedJar)
                }
'''
new = '''                File forgeUniversalJar = new File(gradle.gradleUserHomeDir,
                        "caches/forge_gradle/maven_downloader/net/minecraftforge/forge/${forge_minecraft_version}-${forge_version}/forge-${forge_minecraft_version}-${forge_version}-universal.jar")
                // Forge-added members keep their Forge names in production and therefore do not exist
                // in vanilla SRG mappings. Include ForgeGradle's generated mapped Forge artifact in
                // the structural visibility database so the audit proves exact owner/descriptor/access
                // rather than incorrectly classifying legitimate Forge-added members as unresolved.
                File forgeMappedJar = layout.buildDirectory.file(
                        "fg_cache/net/minecraftforge/forge/${forge_minecraft_version}-${forge_version}_mapped_official_${forge_minecraft_version}/forge-${forge_minecraft_version}-${forge_version}_mapped_official_${forge_minecraft_version}.jar").get().asFile
                List<File> minecraftVisibilityClasspath = [minecraftSrgJar]
                if (forgeMappedJar.isFile()) {
                    minecraftVisibilityClasspath.add(forgeMappedJar)
                }
                if (forgeBinpatchedJar.isFile()) {
                    minecraftVisibilityClasspath.add(forgeBinpatchedJar)
                }
'''
if old not in text:
    raise SystemExit("production Minecraft visibility classpath anchor not found")
text = text.replace(old, new, 1)
path.write_text(text, encoding="utf-8")
print("Patched production access verifier to structurally include ForgeGradle mapped Forge classes")
