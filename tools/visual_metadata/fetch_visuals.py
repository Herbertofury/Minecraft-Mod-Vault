#!/usr/bin/env python3
import base64, concurrent.futures as cf, gzip, html as hm, io, json, re, time
from pathlib import Path
from urllib.parse import urlparse, urljoin, parse_qs, unquote
import requests
from bs4 import BeautifulSoup
from PIL import Image, ImageOps, ImageDraw

DATA_B64='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'
ROOT=Path(__file__).resolve().parent; BUILD=ROOT/'build'; ASSETS=BUILD/'assets'; ASSETS.mkdir(parents=True,exist_ok=True)
UA='Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/151 Safari/537.36'
S=requests.Session(); S.headers.update({'User-Agent':UA,'Accept-Language':'en-US,en;q=0.9'})
IMG_EXT=re.compile(r'\.(?:png|jpe?g|webp|gif)(?:\?|$)',re.I)

def clean(u,base=''):
    if not u:return ''
    u=hm.unescape(str(u)).replace('\\u002F','/').replace('\\/','/')
    if u.startswith('//'):u='https:'+u
    if u.startswith('/') and base:u=urljoin(base,u)
    try:
        q=parse_qs(urlparse(u).query)
        if q.get('url') and unquote(q['url'][0]).startswith('http'):u=unquote(q['url'][0])
    except:pass
    return u

def jget(u):
    for n in range(3):
        try:
            r=S.get(u,timeout=35)
            if r.status_code==200:return r.json()
            if r.status_code in (429,500,502,503,504):time.sleep(n+1)
            else:return None
        except:time.sleep(n+1)
    return None

def hget(u):
    try:
        r=S.get(u,timeout=30)
        return r.text if r.status_code==200 else ''
    except:return ''

def imgsrc(x,b):
    for a in ('src','data-src','data-lazy-src'):
        v=x.get(a)
        if v and not str(v).startswith('data:'):return clean(v,b)
    s=x.get('srcset') or x.get('data-srcset')
    return clean(str(s).split(',')[-1].strip().split()[0],b) if s else ''

def avatar(page,base,user=''):
    if not page:return ''
    s=BeautifulSoup(page,'html.parser')
    for x in s.find_all('img'):
        a=(x.get('alt') or '').lower()
        if 'profile avatar' in a or (user and user.lower() in a and 'avatar' in a):
            u=imgsrc(x,base)
            if u:return u
    t=page.replace('\\u002F','/').replace('\\/','/')
    for u in re.findall(r'https://[^"\'<> ]+',t):
        l=u.lower()
        if 'jtvnw.net' in l and ('profile_image' in l or 'user-default-pictures' in l):return clean(u)
    return ''

def gallery(page,base,icon=''):
    if not page:return []
    s=BeautifulSoup(page,'html.parser'); c=[]
    for x in s.find_all(['img','a']):
        vals=[imgsrc(x,base)] if x.name=='img' else [clean(x.get('href',''),base)]
        for u in vals:
            if u and u.startswith('http') and IMG_EXT.search(u):c.append(u)
    t=page.replace('\\u002F','/').replace('\\/','/')
    c+=re.findall(r'https://[^"\'<> ]+?(?:png|jpe?g|webp|gif)(?:\?[^"\'<> ]*)?',t,re.I)
    def pri(u):
        l=u.lower()
        if 'media.forgecdn.net/attachments/' in l:return 0
        if 'cdn.modrinth.com/data/' in l and '/images/' in l:return 0
        if 'i.imgur.com/' in l:return 2
        if 'githubusercontent' in l or 'github.com/user-attachments' in l or 'cdn.tropicraft.net/' in l:return 3
        return 9
    out=[]
    for u in sorted(map(clean,c),key=pri):
        l=u.lower()
        if not u or u==clean(icon) or u in out or pri(u)>=9:continue
        if any(z in l for z in ('profile_image','user-default-pictures','/avatars/','tier-frame','tier-icon')):continue
        out.append(u)
        if len(out)>=6:break
    return out

def enrich(it):
    u=it['project_url']; p=it['provider']; it.update(author='',author_url='',author_avatar_url='',icon_url='',gallery_urls=[])
    try:
        if p=='curseforge':
            d=jget('https://api.cfwidget.com/'+urlparse(u).path.lstrip('/')) or {}
            it['icon_url']=clean(d.get('thumbnail',''))
            ms=d.get('members') or []; o=next((m for m in ms if str(m.get('title','')).lower()=='owner'),ms[0] if ms else {})
            it['author']=o.get('username',''); it['author_cf_id']=o.get('id','')
            if it['author']:it['author_url']=f"https://www.curseforge.com/members/{it['author'].lower()}/projects"
            h=hget(u); it['author_avatar_url']=avatar(h,u,it['author']); it['gallery_urls']=gallery(h,u,it['icon_url'])
            if not it['author_avatar_url'] and it['author_url']:it['author_avatar_url']=avatar(hget(it['author_url']),it['author_url'],it['author'])
        elif p=='modrinth':
            slug=urlparse(u).path.rstrip('/').split('/')[-1]; d=jget('https://api.modrinth.com/v2/project/'+slug) or {}
            it['icon_url']=clean(d.get('icon_url','')); it['gallery_urls']=[clean(g.get('url','')) for g in d.get('gallery',[]) if g.get('url')][:6]
            ms=jget('https://api.modrinth.com/v2/team/'+str(d.get('team',''))+'/members') or []; ms=sorted(ms,key=lambda m:m.get('ordering',999))
            o=next((m for m in ms if 'owner' in str(m.get('role','')).lower()),ms[0] if ms else {}); x=o.get('user') or {}
            it['author']=x.get('name') or x.get('username',''); it['author_avatar_url']=clean(x.get('avatar_url',''))
            if x.get('username'):it['author_url']='https://modrinth.com/user/'+x['username']
        else:
            h=hget(u); s=BeautifulSoup(h,'html.parser') if h else None; m=s.find('meta',attrs={'property':'og:image'}) if s else None
            it['icon_url']=clean(m.get('content',''),u) if m else ''; it['author_avatar_url']=avatar(h,u); it['gallery_urls']=gallery(h,u,it['icon_url'])
    except Exception as e:it['error']=repr(e)
    it['visual_source_url']=u; return it

def fetchimg(u):
    if not u:return None
    try:
        r=S.get(u,timeout=35,headers={'User-Agent':UA,'Referer':'https://www.curseforge.com/'})
        if r.status_code!=200 or len(r.content)<200:return None
        z=Image.open(io.BytesIO(r.content)); z.seek(0); return z.convert('RGB')
    except:return None

def square(u,p,n=72):
    z=fetchimg(u)
    if z is None:return False
    b=Image.new('RGB',(n,n),'white'); q=ImageOps.contain(z,(n-4,n-4),Image.Resampling.LANCZOS); b.paste(q,((n-q.width)//2,(n-q.height)//2)); b.save(p,'JPEG',quality=82,optimize=True); return True

def strip(urls,p,w=600,h=170):
    a=[x for x in (fetchimg(u) for u in urls[:3]) if x is not None]
    if not a:return False
    b=Image.new('RGB',(w,h),(245,245,245)); gap=5; cw=(w-gap*(len(a)-1))//len(a)
    for i,z in enumerate(a):
        q=ImageOps.contain(z,(cw,h),Image.Resampling.LANCZOS); b.paste(q,(i*(cw+gap)+(cw-q.width)//2,(h-q.height)//2))
    b.save(p,'JPEG',quality=78,optimize=True); return True

def assets(it):
    r=int(it['row']); o=dict(it); ip=ASSETS/f'{r:03d}_icon.jpg'; ap=ASSETS/f'{r:03d}_author.jpg'; gp=ASSETS/f'{r:03d}_gallery.jpg'
    o['icon_asset']=f'assets/{ip.name}' if square(it.get('icon_url',''),ip) else ''
    o['author_avatar_asset']=f'assets/{ap.name}' if square(it.get('author_avatar_url',''),ap) else ''
    o['gallery_asset']=f'assets/{gp.name}' if strip(it.get('gallery_urls',[]),gp) else ''
    return o

def main():
    items=json.loads(gzip.decompress(base64.b64decode(DATA_B64)).decode()); print('loaded',len(items),flush=True)
    a=[]
    with cf.ThreadPoolExecutor(max_workers=3) as ex:
        for i,x in enumerate(ex.map(enrich,items),1):
            a.append(x)
            if i%25==0:print('metadata',i,flush=True)
    b=[]
    with cf.ThreadPoolExecutor(max_workers=8) as ex:
        for i,x in enumerate(ex.map(assets,a),1):
            b.append(x)
            if i%25==0:print('assets',i,flush=True)
    b.sort(key=lambda x:int(x['row']))
    cov={'projects':len(b),'authors':sum(bool(x.get('author')) for x in b),'author_avatars':sum(bool(x.get('author_avatar_url')) for x in b),'icons':sum(bool(x.get('icon_url')) for x in b),'gallery_sources':sum(bool(x.get('gallery_urls')) for x in b),'icon_assets':sum(bool(x.get('icon_asset')) for x in b),'author_avatar_assets':sum(bool(x.get('author_avatar_asset')) for x in b),'gallery_assets':sum(bool(x.get('gallery_asset')) for x in b)}
    BUILD.mkdir(exist_ok=True); json.dump({'generated':time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime()),'coverage':cov,'items':b},open(BUILD/'visual_manifest.json','w'),ensure_ascii=False,indent=2); print('COVERAGE',json.dumps(cov),flush=True)
    if len(b)!=253:raise SystemExit('expected 253')
if __name__=='__main__':main()
