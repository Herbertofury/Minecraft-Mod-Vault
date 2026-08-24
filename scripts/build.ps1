$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Out = Join-Path $Root 'dist'
$Version = '0.9.0'
New-Item -ItemType Directory -Force -Path $Out | Out-Null
Push-Location $Root
try {
  gofmt -w *.go
  go test -mod=vendor -count=1 ./...
  go vet -mod=vendor ./...
  if (Get-Command node -ErrorAction SilentlyContinue) {
    node --check web/app.js
    node --check web/catalog.js
    node --check web/repair-lab.js
  }
  $env:CGO_ENABLED = '0'
  $env:GOARCH = 'amd64'

  $env:GOOS = 'windows'
  $WindowsOutput = Join-Path $Out "Minecraft-Mod-Vault-$Version-windows-x64.exe"
  go build -trimpath -buildvcs=false -ldflags='-H=windowsgui -s -w -buildid=' -o $WindowsOutput .

  $env:GOOS = 'linux'
  $LinuxOutput = Join-Path $Out "Minecraft-Mod-Vault-$Version-linux-x64"
  go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o $LinuxOutput .

  Write-Host "Built $WindowsOutput"
  Write-Host "Built $LinuxOutput"
}
finally {
  Pop-Location
}
