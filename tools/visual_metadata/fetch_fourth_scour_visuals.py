#!/usr/bin/env python3
import io, json, re
from pathlib import Path
from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
from PIL import Image, ImageOps, ImageDraw

ROOT=Path(__file__).resolve().parent
INPUT=ROOT/'fourth_scour_2026-08-24.json'
BUILD=ROOT/'build-fourth-scour'
ASSETS=BUILD/'assets'
ASSETS.mkdir(parents=True,exist_ok=True)
S=requests.Session(); S.headers.update({'User-Agent':'Mozilla/5.0 Chrome/151 Safari/537.36','Accept-Language':'en-US,en;q=0.9'})

def get(url, timeout=25):
    try:
        r=S.get(url,timeout=timeout); r.raise_for_status(); return r
    except Exception as e:
        print('GET FAIL',url,type(e).__name__,e,flush=True); return None

def clean(u,base=''):
    if not u:return ''
    if u.startswith('//'):u='https:'+u
    if base:u=urljoin(base,u)
    return u if u.startswith('http') else ''

def img(url):
    r=get(url)
    if not r:return None
    try:
        im=Image.open(io.BytesIO(r.content)).convert('RGB')
        if im.width<32 or im.height<32:return None
        return im
    except Exception:return None

def save_square(url,path,size=96):
    im=img(url)
    if not im:return False
    im=ImageOps.fit(im,(size,size),method=Image.Resampling.LANCZOS)
    im.save(path,'JPEG',quality=82,optimize=True); return True

def save_strip(urls,path,w=480,h=90):
    ims=[]
    for u in urls[:6]:
        im=img(u)
        if im:ims.append(im)
        if len(ims)>=4:break
    if not ims:return False
    canvas=Image.new('RGB',(w,h),'white'); gap=4
    cellw=(w-gap*(len(ims)-1))//len(ims)
    x=0
    for im in ims:
        tile=ImageOps.fit(im,(cellw,h),method=Image.Resampling.LANCZOS)
        canvas.paste(tile,(x,0)); x+=cellw+gap
    canvas.save(path,'JPEG',quality=80,optimize=True); return True

def cf_meta(it):
    u=it['project_url']; out={'author':'','icon_url':'','author_avatar_url':'','gallery_urls':[]}
    # cfwidget is useful for icon/owner metadata and works for both Java and Bedrock pages.
    r=get('https://api.cfwidget.com/'+urlparse(u).path.lstrip('/'))
    if r:
        try:
            d=r.json(); out['icon_url']=clean(d.get('thumbnail',''))
            ms=d.get('members') or []
            owner=next((m for m in ms if str(m.get('title','')).lower()=='owner'),ms[0] if ms else {})
            out['author']=owner.get('username','') or owner.get('name','') or ''
        except Exception:pass
    r=get(u)
    if not r:return out
    soup=BeautifulSoup(r.text,'html.parser')
    if not out['icon_url']:
        m=soup.find('meta',attrs={'property':'og:image'}); out['icon_url']=clean(m.get('content','') if m else '',u)
    # Page image order normally includes project image, gallery, then profile avatar. Filter obvious UI assets.
    seen=[]
    for tag in soup.find_all('img'):
        src=clean(tag.get('src') or tag.get('data-src') or '',u)
        if not src or 'media.forgecdn.net' not in src or src in seen:continue
        alt=(tag.get('alt') or '').lower()
        seen.append(src)
        if 'profile avatar' in alt and not out['author_avatar_url']:
            out['author_avatar_url']=src; continue
        if any(k in alt for k in ('logo','icon')) and not out['icon_url']:
            out['icon_url']=src
        if src!=out['icon_url'] and len(out['gallery_urls'])<6:out['gallery_urls'].append(src)
    # Author fallback from visible 'By' block.
    if not out['author']:
        txt=soup.get_text(' ',strip=True)
        m=re.search(r'\bBy\s+([A-Za-z0-9_\-]+)',txt)
        if m:out['author']=m.group(1)
    return out

def mr_meta(it):
    u=it['project_url']; slug=urlparse(u).path.rstrip('/').split('/')[-1]
    out={'author':'','icon_url':'','author_avatar_url':'','gallery_urls':[]}
    r=get('https://api.modrinth.com/v2/project/'+slug)
    if not r:return out
    try:d=r.json()
    except Exception:return out
    out['icon_url']=clean(d.get('icon_url',''))
    out['gallery_urls']=[clean(g.get('url','')) for g in d.get('gallery',[]) if g.get('url')][:6]
    team=d.get('team')
    if team:
        tr=get('https://api.modrinth.com/v2/team/'+str(team)+'/members')
        if tr:
            try:
                ms=sorted(tr.json(),key=lambda m:m.get('ordering',999))
                owner=next((m for m in ms if 'owner' in str(m.get('role','')).lower()),ms[0] if ms else {})
                user=owner.get('user') or {}; out['author']=user.get('name') or user.get('username') or ''
                out['author_avatar_url']=clean(user.get('avatar_url',''))
            except Exception:pass
    return out

def main():
    items=json.loads(INPUT.read_text(encoding='utf-8')); manifest=[]
    for i,it in enumerate(items,1):
        print(f'[{i:02d}/{len(items)}] {it["name"]}',flush=True)
        meta=mr_meta(it) if it['provider']=='modrinth' else cf_meta(it)
        rec={**it,**meta,'icon_asset':'','author_avatar_asset':'','gallery_asset':''}
        stem=f'{i:03d}'
        ip=ASSETS/f'{stem}_icon.jpg'; ap=ASSETS/f'{stem}_author.jpg'; gp=ASSETS/f'{stem}_gallery.jpg'
        if save_square(meta.get('icon_url',''),ip):rec['icon_asset']=f'assets/{ip.name}'
        if save_square(meta.get('author_avatar_url',''),ap):rec['author_avatar_asset']=f'assets/{ap.name}'
        if save_strip(meta.get('gallery_urls',[]),gp):rec['gallery_asset']=f'assets/{gp.name}'
        manifest.append(rec)
    (BUILD/'manifest.json').write_text(json.dumps(manifest,indent=2),encoding='utf-8')
    cov={k:sum(bool(r[k]) for r in manifest) for k in ('icon_asset','author_avatar_asset','gallery_asset')}
    (BUILD/'coverage.json').write_text(json.dumps({'items':len(manifest),**cov},indent=2),encoding='utf-8')
    print('coverage',cov,flush=True)

if __name__=='__main__':main()
