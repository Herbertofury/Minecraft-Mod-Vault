$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
Write-Host 'OmniPorter Forge 1.20.1 prewarmer revision r5'

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

function Get-JavaVersionText([string]$JavaExe) {
    $Stdout = Join-Path $env:TEMP ("omniporter-java-stdout-" + [guid]::NewGuid().ToString('N') + '.txt')
    $Stderr = Join-Path $env:TEMP ("omniporter-java-stderr-" + [guid]::NewGuid().ToString('N') + '.txt')
    try {
        $Process = Start-Process -FilePath $JavaExe -ArgumentList '-version' -NoNewWindow -Wait -PassThru -RedirectStandardOutput $Stdout -RedirectStandardError $Stderr
        $Text = ((Get-Content $Stdout -Raw -ErrorAction SilentlyContinue) + "`n" + (Get-Content $Stderr -Raw -ErrorAction SilentlyContinue)).Trim()
        if ($Process.ExitCode -ne 0) { throw "java -version failed with exit code $($Process.ExitCode): $Text" }
        return $Text
    }
    finally {
        Remove-Item $Stdout,$Stderr -Force -ErrorAction SilentlyContinue
    }
}

$BundledJava = Get-ChildItem -Path $JdkRoot -Filter java.exe -File -Recurse -ErrorAction SilentlyContinue |
    Where-Object { $_.FullName -match '\\bin\\java\.exe$' } |
    Select-Object -First 1 -ExpandProperty FullName

if ([string]::IsNullOrWhiteSpace($BundledJava)) {
    Write-Host 'Provisioning isolated Eclipse Temurin 17 JDK (Windows x64)...'
    if (!(Test-Path $JdkZip)) {
        Invoke-WebRequest -UseBasicParsing 'https://api.adoptium.net/v3/binary/latest/17/ga/windows/x64/jdk/hotspot/normal/eclipse?project=jdk' -OutFile $JdkZip
    }
    Remove-Item -Recurse -Force $JdkRoot -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Force -Path $JdkRoot | Out-Null
    Expand-Archive -Path $JdkZip -DestinationPath $JdkRoot -Force
    $BundledJava = Get-ChildItem -Path $JdkRoot -Filter java.exe -File -Recurse |
        Where-Object { $_.FullName -match '\\bin\\java\.exe$' } |
        Select-Object -First 1 -ExpandProperty FullName
}

if ([string]::IsNullOrWhiteSpace($BundledJava) -or !(Test-Path $BundledJava)) {
    throw 'Could not locate the isolated Java 17 runtime.'
}

$JavaVersionText = Get-JavaVersionText $BundledJava
if ($JavaVersionText -notmatch 'version\s+"17\.') {
    throw "Expected Java 17 but isolated runtime reports: $JavaVersionText"
}

$JavaBin = Split-Path $BundledJava -Parent
$JavaHome = Split-Path $JavaBin -Parent
$env:JAVA_HOME = $JavaHome
$env:PATH = "$JavaBin;$env:PATH"

Write-Host "Using isolated Java 17: $BundledJava"
Write-Host $JavaVersionText

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
