$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$MinecraftVersion = '1.20.1'
$ForgeVersion = '47.4.23'
$GradleVersion = '8.8'
$ScriptRoot = if ([string]::IsNullOrWhiteSpace($PSScriptRoot)) { (Get-Location).Path } else { $PSScriptRoot }
$Root = Join-Path $ScriptRoot 'OmniPorter-Forge1201-Prewarm'
$GradleHome = Join-Path $Root 'gradle-home'
$Work = Join-Path $Root 'mdk'
$Downloads = Join-Path $Root 'downloads'
$GradleZip = Join-Path $Downloads "gradle-$GradleVersion-bin.zip"
$MdkZip = Join-Path $Downloads "forge-$MinecraftVersion-$ForgeVersion-mdk.zip"
$GradleDir = Join-Path $Root "gradle-$GradleVersion"
$GradleExe = Join-Path $GradleDir 'bin\gradle.bat'
$OutZip = Join-Path $ScriptRoot "omniporter-forge-$MinecraftVersion-$ForgeVersion-gradle-cache.zip"

New-Item -ItemType Directory -Force -Path $Root,$GradleHome,$Work,$Downloads | Out-Null

if (!(Test-Path $GradleZip)) {
    Invoke-WebRequest -UseBasicParsing "https://services.gradle.org/distributions/gradle-$GradleVersion-bin.zip" -OutFile $GradleZip
}
if (!(Test-Path $MdkZip)) {
    Invoke-WebRequest -UseBasicParsing "https://maven.minecraftforge.net/net/minecraftforge/forge/$MinecraftVersion-$ForgeVersion/forge-$MinecraftVersion-$ForgeVersion-mdk.zip" -OutFile $MdkZip
}

if (!(Test-Path $GradleExe)) {
    Expand-Archive -Path $GradleZip -DestinationPath $Root -Force
}
Remove-Item -Recurse -Force $Work -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $Work | Out-Null
Expand-Archive -Path $MdkZip -DestinationPath $Work -Force

$env:GRADLE_USER_HOME = $GradleHome
Write-Host "Using isolated GRADLE_USER_HOME: $GradleHome"

Push-Location $Work
try {
    & $GradleExe --no-daemon --refresh-dependencies tasks
    if ($LASTEXITCODE -ne 0) { throw "Gradle tasks failed with exit code $LASTEXITCODE" }

    # A real MDK build resolves ForgeGradle, Forge userdev, mappings, Minecraft libraries,
    # reobfuscation tooling, and the normal Java compile/runtime graph needed by OmniPorter.
    & $GradleExe --no-daemon --refresh-dependencies build
    if ($LASTEXITCODE -ne 0) { throw "Gradle build failed with exit code $LASTEXITCODE" }
}
finally {
    Pop-Location
}

if (Test-Path $OutZip) { Remove-Item $OutZip -Force }
Compress-Archive -Path (Join-Path $GradleHome '*') -DestinationPath $OutZip -CompressionLevel Optimal

$Hash = (Get-FileHash -Algorithm SHA256 $OutZip).Hash.ToLowerInvariant()
$Size = (Get-Item $OutZip).Length
Write-Host ''
Write-Host 'DONE. Upload this file into Minecraft Dev Kit -> 01 Toolchains & SDKs -> Gradle:'
Write-Host $OutZip
Write-Host "Size: $Size bytes"
Write-Host "SHA-256: $Hash"
