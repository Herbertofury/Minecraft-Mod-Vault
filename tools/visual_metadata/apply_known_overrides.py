#!/usr/bin/env python3
import base64,gzip,json
from pathlib import Path
from fetch_visuals import square,strip

KNOWN_B64='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'
root=Path(__file__).resolve().parent; build=root/'build'; assets=build/'assets'
manifest=json.load(open(build/'visual_manifest.json',encoding='utf-8')); items={int(x['row']):x for x in manifest['items']}
known=json.loads(gzip.decompress(base64.b64decode(KNOWN_B64)).decode())
for k in known:
    x=items[int(k['row'])]
    for f in ('author','author_url','author_avatar_url','icon_url'):
        if k.get(f) and not x.get(f): x[f]=k[f]
    if k.get('gallery_urls'):
        cur=x.get('gallery_urls') or []
        x['gallery_urls']=(cur+[u for u in k['gallery_urls'] if u not in cur])[:6]
    r=int(x['row'])
    if not x.get('icon_asset') and x.get('icon_url'):
        p=assets/f'{r:03d}_icon.jpg'
        if square(x['icon_url'],p): x['icon_asset']=f'assets/{p.name}'
    if not x.get('author_avatar_asset') and x.get('author_avatar_url'):
        p=assets/f'{r:03d}_author.jpg'
        if square(x['author_avatar_url'],p): x['author_avatar_asset']=f'assets/{p.name}'
    if not x.get('gallery_asset') and x.get('gallery_urls'):
        p=assets/f'{r:03d}_gallery.jpg'
        if strip(x['gallery_urls'],p): x['gallery_asset']=f'assets/{p.name}'
arr=[items[r] for r in sorted(items)]
manifest['items']=arr
manifest['coverage']={'projects':len(arr),'authors':sum(bool(x.get('author')) for x in arr),'author_avatars':sum(bool(x.get('author_avatar_url')) for x in arr),'icons':sum(bool(x.get('icon_url')) for x in arr),'gallery_sources':sum(bool(x.get('gallery_urls')) for x in arr),'icon_assets':sum(bool(x.get('icon_asset')) for x in arr),'author_avatar_assets':sum(bool(x.get('author_avatar_asset')) for x in arr),'gallery_assets':sum(bool(x.get('gallery_asset')) for x in arr)}
json.dump(manifest,open(build/'visual_manifest.json','w'),ensure_ascii=False,indent=2)
print('COVERAGE_AFTER_OVERRIDES',json.dumps(manifest['coverage']))
