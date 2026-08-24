#!/usr/bin/env sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
OUT=${1:-"$ROOT/dist"}
VERSION=0.11.0
mkdir -p "$OUT"
cd "$ROOT"

gofmt -w *.go
go test -mod=vendor -count=1 ./...
go vet -mod=vendor ./...
if command -v node >/dev/null 2>&1; then
  node --check web/app.js
  node --check web/catalog.js
  node --check web/repair-lab.js
  node --check web/omnimanager.js
  node --check web/conversion.js
fi

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags="-H=windowsgui -s -w -buildid=" -o "$OUT/Minecraft-Mod-Vault-$VERSION-windows-x64.exe" .
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o "$OUT/Minecraft-Mod-Vault-$VERSION-linux-x64" .

echo "Built Minecraft Mod Vault $VERSION release binaries in $OUT"
