import re, pathlib
srcroot=pathlib.Path('/mnt/data/foolish-decompiled/net/mcreator/foolish/procedures')
dstroot=pathlib.Path('/tmp/foolish-opt-work/project/optimized-src/net/mcreator/foolish/procedures')
files=['AOEResponseProcedure.java','AOEResponse1Procedure.java','AOEResponse2Procedure.java','AOEResponse3Procedure.java','AOEResponse4Procedure.java']
nearest_re=re.compile(r'''world\.m_6443_\(\s*(?P<type>[A-Za-z0-9_$.]+\.class)\s*,\s*(?P<box>AABB\.m_165882_\(new Vec3\(x, y, z\),\s*[0-9.]+\s*,\s*[0-9.]+\s*,\s*[0-9.]+\s*\))\s*,\s*e\s*->\s*true\s*\)\s*\.stream\(\)\s*\.sorted\(\s*\(new Object\(\)\s*\{\s*Comparator<Entity>\s+compareDistOf\(double _x, double _y, double _z\)\s*\{\s*return Comparator\.comparingDouble\(_entcnd\s*->\s*_entcnd\.m_20275_\(_x, _y, _z\)\);\s*\}\s*\}\)\.compareDistOf\(x, y, z\)\s*\)\s*\.findFirst\(\)\s*\.orElse\(null\)''', re.S)
query_re=re.compile(r'''world\.m_6443_\(\s*(?P<type>[A-Za-z0-9_$.]+\.class)\s*,\s*(?P<box>AABB\.m_165882_\(new Vec3\(x, y, z\),\s*[0-9.]+\s*,\s*[0-9.]+\s*,\s*[0-9.]+\s*\))\s*,\s*e\s*->\s*true\s*\)''', re.S)
query_pat=re.compile(r'world\.m_6443_\(([^,]+),\s*AABB\.m_165882_\(new Vec3\(x, y, z\),\s*([0-9.]+),\s*([0-9.]+),\s*([0-9.]+)\),\s*e -> true\)')
helper='''

   /**
    * Per-execution spatial query cache. MCreator generated the original AoE code with
    * dozens of repeated entity-section scans and full stream sorts for the same boxes.
    * This keeps the exact tick cadence and gameplay checks, but snapshots only the
    * relevant AoE source entities once and reuses exact-box results during this one
    * synchronous procedure invocation. Entity state is read live from the objects.
    */
   private static final class QueryCache {
      private final LevelAccessor world;
      private final java.util.Map<QueryKey, java.util.List<? extends Entity>> exact = new java.util.HashMap<>();
      private java.util.List<Entity> primedSources = java.util.List.of();
      private AABB primedBox;
      private boolean primed;

      QueryCache(LevelAccessor world) {
         this.world = world;
      }

      boolean prime(AABB box, java.util.function.Predicate<Entity> sourcePredicate) {
         this.primedBox = box;
         this.primedSources = new java.util.ArrayList<>(world.m_6443_(Entity.class, box, sourcePredicate));
         this.primed = true;
         return !this.primedSources.isEmpty();
      }

      @SuppressWarnings("unchecked")
      <T extends Entity> java.util.List<T> query(Class<T> type, AABB box) {
         QueryKey key = new QueryKey(type, box);
         java.util.List<? extends Entity> cached = exact.get(key);
         if (cached != null) {
            return (java.util.List<T>) cached;
         }

         java.util.ArrayList<T> out = new java.util.ArrayList<>();
         if (type == net.minecraft.world.entity.player.Player.class) {
            for (net.minecraft.world.entity.player.Player player : world.m_6907_()) {
               if (box.m_82381_(player.m_20191_())) {
                  out.add(type.cast(player));
               }
            }
         } else if (primed && contains(primedBox, box)) {
            for (Entity candidate : primedSources) {
               if (type.isInstance(candidate) && box.m_82381_(candidate.m_20191_())) {
                  out.add(type.cast(candidate));
               }
            }
         } else {
            out.addAll(world.m_6443_(type, box, e -> true));
         }
         exact.put(key, out);
         return out;
      }

      <T extends Entity> T nearest(Class<T> type, AABB box, double x, double y, double z) {
         T best = null;
         double bestDistance = Double.POSITIVE_INFINITY;
         for (T candidate : query(type, box)) {
            double distance = candidate.m_20275_(x, y, z);
            if (distance < bestDistance) {
               bestDistance = distance;
               best = candidate;
            }
         }
         return best;
      }

      private static boolean contains(AABB outer, AABB inner) {
         return outer.f_82288_ <= inner.f_82288_ && outer.f_82289_ <= inner.f_82289_ && outer.f_82290_ <= inner.f_82290_
            && outer.f_82291_ >= inner.f_82291_ && outer.f_82292_ >= inner.f_82292_ && outer.f_82293_ >= inner.f_82293_;
      }
   }

   private static record QueryKey(Class<?> type, AABB box) {
   }
'''
dstroot.mkdir(parents=True, exist_ok=True)
report=[]
for name in files:
    src=(srcroot/name).read_text()
    matches=list(query_pat.finditer(src))
    classes={}
    for m in matches:
        typ=m.group(1).strip().replace('.class','')
        r=float(m.group(2)); classes[typ]=max(classes.get(typ,0.0),r)
    player_max=classes.pop('Player',0.0)
    nonplayer=list(classes); global_max=max(classes.values(), default=0.0)
    t=src
    t,n_near=nearest_re.subn(lambda m:f'_q.nearest({m.group("type")}, {m.group("box")}, x, y, z)',t)
    t,n_query=query_re.subn(lambda m:f'_q.query({m.group("type")}, {m.group("box")})',t)
    random_gate=re.compile(r'(if \([^\n]*Mth\.m_216271_\(RandomSource\.m_216327_\(\), 1, 2\)[^\n]*\) \{)')
    pre='\n            QueryCache _q = new QueryCache(world);'
    if nonplayer:
        pred=' || '.join(f'e instanceof {typ}' for typ in nonplayer)
        pre += f'\n            boolean _hasNonPlayerSource = _q.prime(AABB.m_165882_(new Vec3(x, y, z), {global_max:.1f}, {global_max:.1f}, {global_max:.1f}), e -> {pred});'
        if player_max:
            pre += f'\n            if (!_hasNonPlayerSource && _q.query(net.minecraft.world.entity.player.Player.class, AABB.m_165882_(new Vec3(x, y, z), {player_max:.1f}, {player_max:.1f}, {player_max:.1f})).isEmpty()) {{\n               return;\n            }}'
        else:
            pre += '\n            if (!_hasNonPlayerSource) {\n               return;\n            }'
    t,count=random_gate.subn(lambda m:m.group(1)+pre,t,count=1)
    if count != 1: raise RuntimeError(f'{name}: random gate insertion failed')
    idx=t.rfind('}'); t=t[:idx]+helper+t[idx:]
    (dstroot/name).write_text(t)
    report.append(f'{name}: raw={len(matches)} nearest_replaced={n_near} queries_replaced={n_query} nonplayer={classes} playerMax={player_max}')
pathlib.Path('/tmp/foolish-opt-work/evidence/aoe-transform-report.txt').write_text('\n'.join(report)+'\n')
print('\n'.join(report))
