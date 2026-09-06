import re, pathlib
SRC=pathlib.Path('/mnt/data/foolish-decompiled/net/mcreator/foolish/procedures')
DST=pathlib.Path('/tmp/foolish-opt-work/project/optimized-src/net/mcreator/foolish/procedures')
FILES=['ScarlantSentinelOnEntityTickUpdateProcedure.java','ScarlantSteedTickUpdateProcedure.java','ScarlantOnEntityTickUpdateProcedure.java','FlareOnEntityTickUpdateProcedure.java','InfiltratorOnEntityTickUpdateProcedure.java','SeekingSaciliteOnEntityTickUpdateProcedure.java','FlorauderOnEntityTickUpdateProcedure.java','GiantIsopodOnEntityTickUpdateProcedure.java','SanguineOnEntityTickUpdateProcedure.java','VarmintKingSpawnerTickProcedure.java']
SORTED_FIRST = re.compile(r'''\.stream\(\)\s*\.sorted\(\s*(?P<cmp>\(new Object\(\)\s*\{\s*Comparator<Entity>\s+compareDistOf\(double _x, double _y, double _z\)\s*\{\s*return Comparator\.comparingDouble\(_entcnd\s*->\s*_entcnd\.m_20275_\(_x, _y, _z\)\);\s*\}\s*\}\)\.compareDistOf\(.*?\))\s*\)\s*\.findFirst\(\)''', re.S)
HELPER='''

   /**
    * Fast path for spatial player lookups used by MCreator tick procedures. Minecraft's
    * generic getEntitiesOfClass path traverses entity sections even when the requested type
    * is Player; EntityGetter already exposes the live player list directly. Non-player
    * queries deliberately delegate to the original implementation so mutation-sensitive
    * entity/item behavior stays byte-for-byte equivalent in ordering and visibility.
    */
   private static final class SpatialQueries {
      @SuppressWarnings("unchecked")
      static <T extends Entity> java.util.List<T> query(LevelAccessor world, Class<T> type, AABB box, java.util.function.Predicate<? super T> predicate) {
         if (type != net.minecraft.world.entity.player.Player.class) {
            return world.m_6443_(type, box, predicate);
         }

         java.util.ArrayList<T> out = new java.util.ArrayList<>();
         for (net.minecraft.world.entity.player.Player player : world.m_6907_()) {
            T typed = (T)player;
            if (box.m_82381_(player.m_20191_()) && predicate.test(typed)) {
               out.add(typed);
            }
         }
         return out;
      }
   }
'''

DST.mkdir(parents=True, exist_ok=True)
report=[]
for name in FILES:
    t=(SRC/name).read_text()
    raw=t.count('world.m_6443_(')
    players=len(re.findall(r'world\.m_6443_\(\s*Player\.class',t,re.S))
    before=len(re.findall(r'\.stream\(\)\s*\.sorted\(',t,re.S))
    t=t.replace('world.m_6443_(', 'SpatialQueries.query(world, ')
    t,nmin=SORTED_FIRST.subn(lambda m: '.stream().min('+m.group('cmp')+')', t)
    t=re.sub(r'if \((?P<expr>\(var\d+ instanceof Mob _mobEnt\w* \? _mobEnt\w*\.m_5448_\(\) : null\)) instanceof LivingEntity _ent\) \{',
             lambda m: 'LivingEntity _ent = '+m.group('expr')+';\n                  if (_ent != null) {', t)
    idx=t.rfind('}'); t=t[:idx]+HELPER+t[idx:]
    (DST/name).write_text(t)
    remaining=len(re.findall(r'\.stream\(\)\s*\.sorted\(',t,re.S))
    report.append(f'{name}: rawQueries={raw} playerQueries={players} sortedFirstBefore={before} minReplaced={nmin} remainingSorted={remaining} helperWorldDelegates={t.count("world.m_6443_(")}')
    if remaining or nmin != before: raise RuntimeError(report[-1])
out='\n'.join(report)+'\n'; print(out,end=''); pathlib.Path('/tmp/foolish-opt-work/evidence/tick-transform-report.txt').write_text(out)
