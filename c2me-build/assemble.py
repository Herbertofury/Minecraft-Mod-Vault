import hashlib, json, sys, zipfile
from pathlib import Path
base, dfc, opencl, out = map(Path, sys.argv[1:5])
with zipfile.ZipFile(base) as z:
    entries={i.filename:z.read(i.filename) for i in z.infolist()}
    meta=json.loads(entries['META-INF/jarjar/metadata.json'])
original={x['path']:hashlib.sha256(entries[x['path']]).hexdigest() for x in meta['jars']}
def addmod(artifact,p):
    dst='META-INF/jarjar/'+p.name
    entries[dst]=p.read_bytes()
    meta['jars'].append({'identifier':{'group':'c2me','artifact':artifact},'version':{'range':'[0.2.0,)','artifactVersion':'0.2.0+alpha.12.1'},'path':dst,'isObfuscated':True})
addmod('c2me_opts_dfc',dfc)
addmod('c2me_opts_accel_opencl',opencl)
entries['META-INF/jarjar/metadata.json']=json.dumps(meta,indent=2).encode()
entries['META-INF/c2me-opencl-radium-compat.txt']=b'Base c2meF 0.2.0+alpha.12 SHA1 ad44f615a4b15afd1d6a4d907ccab1c3855451a1\nOpenCL source 342c5035d7251ca987c962a14997a21a36eace44\nRadium 0.12.4+git.26c9d8e compatibility: disable mixin.world.chunk_access and mixin.alloc.nbt\n'
with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED) as z:
    for n,b in entries.items(): z.writestr(n,b)
with zipfile.ZipFile(out) as z:
    names=set(z.namelist()); m=json.loads(z.read('META-INF/jarjar/metadata.json'))
    assert all(x['path'] in names for x in m['jars'])
    assert all(hashlib.sha256(z.read(p)).hexdigest()==h for p,h in original.items())
print(hashlib.sha256(out.read_bytes()).hexdigest(), out)
print('preserved_base_nested=',len(original),'final_nested=',len(m['jars']))
