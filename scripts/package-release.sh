#!/usr/bin/env bash
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
VERSION=0.11.0
OUT=${1:-"$ROOT/release"}
STAGE="$OUT/.stage"
EPOCH="2026-08-21 00:00:00 UTC"
rm -rf "$OUT"
mkdir -p "$STAGE"

"$ROOT/scripts/build.sh" "$ROOT/dist"

DOCS=(
  README.md
  RELEASE-NOTES-0.11.0.md
  OMNIBRIDGE-ARCHITECTURE.md
  CONVERSION-CAPABILITY-MATRIX.md
  OMNIBRIDGE-TOOL-ADAPTERS.md
  RESEARCH-SOURCES.md
  ULTIMATE-TOOLS-AND-PORTING-KNOWLEDGE.md
  THIRD-PARTY-NOTICES.md
  Minecraft-Mod-Vault-0.11.0-BUILD-VERIFICATION.txt
)

make_runtime_stage() {
  local platform=$1 binary=$2
  local dir="$STAGE/Minecraft-Mod-Vault-$VERSION-$platform"
  mkdir -p "$dir/verification" "$dir/repair-brain"
  cp "$ROOT/dist/$binary" "$dir/"
  cp "$ROOT/repair-brain/repair-history.jsonl" "$dir/repair-brain/"
  cp -a "$ROOT/THIRD-PARTY-LICENSES" "$dir/"
  for f in "${DOCS[@]}"; do cp "$ROOT/$f" "$dir/"; done
  local evidence
  while IFS= read -r -d '' evidence; do
    cp "$evidence" "$dir/verification/"
  done < <(find "$ROOT/verification" -maxdepth 1 -type f -print0 | sort -z)
  (cd "$dir" && find . -type f ! -name PACKAGE-CONTENTS-SHA256.txt -print0 | sort -z | xargs -0 sha256sum > PACKAGE-CONTENTS-SHA256.txt)
}
make_runtime_stage windows-x64 "Minecraft-Mod-Vault-$VERSION-windows-x64.exe"
make_runtime_stage linux-x64 "Minecraft-Mod-Vault-$VERSION-linux-x64"

SRC="$STAGE/Minecraft-Mod-Vault-$VERSION-source"
mkdir -p "$SRC"
( cd "$ROOT" && tar \
    --exclude='./dist' \
    --exclude='./release' \
    --exclude='./.git' \
    --exclude='./*.zip' \
    --exclude='./*.tar.gz' \
    -cf - . ) | ( cd "$SRC" && tar -xf - )
(cd "$SRC" && find . -type f ! -name PACKAGE-CONTENTS-SHA256.txt -print0 | sort -z | xargs -0 sha256sum > PACKAGE-CONTENTS-SHA256.txt)

python3 - "$STAGE/Minecraft-Mod-Vault-$VERSION-windows-x64" "$OUT/Minecraft-Mod-Vault-$VERSION-windows-x64.zip" "$EPOCH" <<'PY'
import os,sys,zipfile,datetime
src,out,_=sys.argv[1:]
dt=(2026,8,21,0,0,0)
base=os.path.dirname(src)
with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED,compresslevel=9) as z:
    for root,dirs,files in os.walk(src):
        dirs.sort(); files.sort()
        for name in files:
            p=os.path.join(root,name); arc=os.path.relpath(p,base).replace(os.sep,'/')
            info=zipfile.ZipInfo(arc,dt); info.compress_type=zipfile.ZIP_DEFLATED
            mode=os.stat(p).st_mode & 0o777; info.external_attr=(mode or 0o644)<<16
            with open(p,'rb') as f: z.writestr(info,f.read(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
PY
python3 - "$SRC" "$OUT/Minecraft-Mod-Vault-$VERSION-source.zip" "$EPOCH" <<'PY'
import os,sys,zipfile
src,out,_=sys.argv[1:]
dt=(2026,8,21,0,0,0)
base=os.path.dirname(src)
with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED,compresslevel=9) as z:
    for root,dirs,files in os.walk(src):
        dirs.sort(); files.sort()
        for name in files:
            p=os.path.join(root,name); arc=os.path.relpath(p,base).replace(os.sep,'/')
            info=zipfile.ZipInfo(arc,dt); info.compress_type=zipfile.ZIP_DEFLATED
            mode=os.stat(p).st_mode & 0o777; info.external_attr=(mode or 0o644)<<16
            with open(p,'rb') as f: z.writestr(info,f.read(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
PY
( cd "$STAGE" && tar --sort=name --mtime="$EPOCH" --owner=0 --group=0 --numeric-owner -czf "$OUT/Minecraft-Mod-Vault-$VERSION-linux-x64.tar.gz" "Minecraft-Mod-Vault-$VERSION-linux-x64" )

unzip -tq "$OUT/Minecraft-Mod-Vault-$VERSION-windows-x64.zip" >/dev/null
unzip -tq "$OUT/Minecraft-Mod-Vault-$VERSION-source.zip" >/dev/null
tar -tzf "$OUT/Minecraft-Mod-Vault-$VERSION-linux-x64.tar.gz" >/dev/null
( cd "$OUT" && sha256sum Minecraft-Mod-Vault-$VERSION-* > "CHECKSUMS-SHA256-$VERSION.txt" )
rm -rf "$STAGE"
ls -lh "$OUT"
