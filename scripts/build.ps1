$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Out = Join-Path $Root 'dist'
$Version = '0.13.0'
function Assert-NativeSuccess([string]$Step) {
  if ($LASTEXITCODE -ne 0) { throw "$Step failed with exit code $LASTEXITCODE" }
}
New-Item -ItemType Directory -Force -Path $Out | Out-Null
Push-Location $Root
try {
  $GoFiles = Get-ChildItem -LiteralPath $Root -Filter '*.go' -File | Select-Object -ExpandProperty FullName
  gofmt -w $GoFiles
  Assert-NativeSuccess 'gofmt'
  go test -mod=vendor -count=1 ./...
  Assert-NativeSuccess 'go test'
  go vet -mod=vendor ./...
  Assert-NativeSuccess 'go vet'
  if (Get-Command node -ErrorAction SilentlyContinue) {
    node --check web/app.js
    Assert-NativeSuccess 'web/app.js syntax check'
    node --check web/catalog.js
    Assert-NativeSuccess 'web/catalog.js syntax check'
    node --check web/repair-lab.js
    Assert-NativeSuccess 'web/repair-lab.js syntax check'
    node --check web/omnimanager.js
    Assert-NativeSuccess 'web/omnimanager.js syntax check'
    node --check web/conversion.js
    Assert-NativeSuccess 'web/conversion.js syntax check'
    node --check web/testgrid.js
    Assert-NativeSuccess 'web/testgrid.js syntax check'
  }
  $env:CGO_ENABLED = '0'
  $env:GOARCH = 'amd64'

  $env:GOOS = 'windows'
  $WindowsOutput = Join-Path $Out "Minecraft-Mod-Vault-$Version-windows-x64.exe"
  $WindowsCLIOutput = Join-Path $Out "Minecraft-Mod-Vault-$Version-windows-x64-cli.exe"
  go build -trimpath -buildvcs=false -ldflags='-H=windowsgui -s -w -buildid=' -o $WindowsOutput .
  Assert-NativeSuccess 'Windows build'
  go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o $WindowsCLIOutput .
  Assert-NativeSuccess 'Windows CLI build'

  $env:GOOS = 'linux'
  $LinuxOutput = Join-Path $Out "Minecraft-Mod-Vault-$Version-linux-x64"
  go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o $LinuxOutput .
  Assert-NativeSuccess 'Linux build'

  Write-Host "Built $WindowsOutput"
  Write-Host "Built $WindowsCLIOutput"
  Write-Host "Built $LinuxOutput"
}
finally {
  Pop-Location
}
