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
$JdkRoot = Join-Path $Root 'jdk17'
$JdkZip = Join-Path $Downloads 'temurin17-windows-x64-jdk.zip'
$OutZip = Join-Path $ScriptRoot "omniporter-forge-$MinecraftVersion-$ForgeVersion-gradle-cache.zip"

New-Item -ItemType Directory -Force -Path $Root,$GradleHome,$Work,$Downloads | Out-Null

function Test-Java17([string]$JavaExe) {
    if ([string]::IsNullOrWhiteSpace($JavaExe) -or !(Test-Path $JavaExe)) { return $false }
    $Text = (& $JavaExe -version 2>&1 | Out-String)
    return $Text -match 'version\s+"17\.'
}

function Resolve-Java17 {
    $Candidates = @()

    if (![string]::IsNullOrWhiteSpace($env:OMNIPORTER_JAVA17)) {
        $Candidates += (Join-Path $env:OMNIPORTER_JAVA17 'bin\java.exe')
    }
    if (![string]::IsNullOrWhiteSpace($env:JAVA_HOME)) {
        $Candidates += (Join-Path $env:JAVA_HOME 'bin\java.exe')
    }

    $Roots = @(
        "$env:ProgramFiles\Eclipse Adoptium",
        "$env:ProgramFiles\Microsoft",
        "$env:ProgramFiles\Java",
        "$env:LOCALAPPDATA\Programs\Eclipse Adoptium"
    )
    foreach ($SearchRoot in $Roots) {
        if (Test-Path $SearchRoot) {
            $Candidates += Get-ChildItem -Path $SearchRoot -Filter java.exe -File -Recurse -ErrorAction SilentlyContinue |
                Where-Object { $_.FullName -match '\\bin\\java\.exe$' } |
                Select-Object -ExpandProperty FullName
        }
    }

    foreach ($Candidate in ($Candidates | Select-Object -Unique)) {
        if (Test-Java17 $Candidate) { return $Candidate }
    }

    Write-Host 'Java 17 was not found. Downloading latest Eclipse Temurin 17 JDK (Windows x64)...'
    if (!(Test-Path $JdkZip)) {
        Invoke-WebRequest -UseBasicParsing 'https://api.adoptium.net/v3/binary/latest/17/ga/windows/x64/jdk/hotspot/normal/eclipse?project=jdk' -OutFile $JdkZip
    }
    Remove-Item -Recurse -Force $JdkRoot -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Force -Path $JdkRoot | Out-Null
    Expand-Archive -Path $JdkZip -DestinationPath $JdkRoot -Force
    $BundledJava = Get-ChildItem -Path $JdkRoot -Filter java.exe -File -Recurse |
        Where-Object { $_.FullName -match '\\bin\\java\.exe$' } |
        Select-Object -First 1 -ExpandProperty FullName
    if (!(Test-Java17 $BundledJava)) {
        throw 'Temurin Java 17 provisioning failed: downloaded archive did not expose a usable Java 17 runtime.'
    }
    return $BundledJava
}

$JavaExe = Resolve-Java17
$JavaBin = Split-Path $JavaExe -Parent
$JavaHome = Split-Path $JavaBin -Parent
$env:JAVA_HOME = $JavaHome
$env:PATH = "$JavaBin;$env:PATH"

Write-Host "Using Java 17: $JavaExe"
& $JavaExe -version
if ($LASTEXITCODE -ne 0) { throw 'Java 17 validation failed.' }

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

& $GradleExe --version
if ($LASTEXITCODE -ne 0) { throw 'Gradle runtime validation failed.' }

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
