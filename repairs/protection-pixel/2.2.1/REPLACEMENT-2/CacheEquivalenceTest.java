package test;
import java.util.*;
import net.mcreator.protectionpixel.repair.ArmorRenderTargetCache;
import net.mcreator.protectionpixel.init.ProtectionPixelModBlocks;
import net.minecraft.client.multiplayer.ClientLevel;
import net.minecraft.core.*;
import net.minecraft.world.level.block.*;
import net.minecraft.world.level.block.entity.*;
import net.minecraft.world.level.block.state.*;
import net.minecraft.world.level.chunk.*;

public final class CacheEquivalenceTest {
 static List<BlockPos> reference(ClientLevel level, BlockPos center, int radius){
   ArrayList<BlockPos> out=new ArrayList<>(); Object h=ProtectionPixelModBlocks.ARMORHANGER.get(), p=ProtectionPixelModBlocks.ARMORLOADPLATFORM.get();
   for(int z=-radius;z<=radius;z++) for(int x=-radius;x<=radius;x++){
     LevelChunk c=level.chunks.get((((long)SectionPos.m_123171_(center.m_123341_()+(x<<4)))<<32) ^ (SectionPos.m_123171_(center.m_123343_()+(z<<4)) & 0xffffffffL));
     if(c==null) continue;
     for(Map.Entry<BlockPos,BlockEntity> e:c.map.entrySet()){
       Object b=e.getValue().m_58900_().m_60734_(); if(b==h||b==p) out.add(e.getKey());
     }
   }
   return out;
 }
 static List<BlockPos> actual(ClientLevel l, BlockPos c, int r){ return new ArrayList<>(ArmorRenderTargetCache.targets(l,c,r).keySet()); }
 static void check(boolean v,String m){ if(!v) throw new AssertionError(m); }
 public static void main(String[] args){
   Random rnd=new Random(0x50505832L); int cases=10000;
   for(int t=0;t<cases;t++){
     ClientLevel level=new ClientLevel(); level.time=t+100;
     int cx=rnd.nextInt(41)-20, cz=rnd.nextInt(41)-20, radius=rnd.nextInt(4);
     BlockPos center=new BlockPos(cx*16+rnd.nextInt(16),64,cz*16+rnd.nextInt(16));
     for(int z=-radius;z<=radius;z++) for(int x=-radius;x<=radius;x++) if(rnd.nextBoolean()){
       LevelChunk ch=new LevelChunk(); int n=rnd.nextInt(9);
       for(int i=0;i<n;i++){
         Block b=switch(rnd.nextInt(5)){case 0->ProtectionPixelModBlocks.HANGER; case 1->ProtectionPixelModBlocks.PLATFORM; default->ProtectionPixelModBlocks.OTHER;};
         BlockPos pos=new BlockPos((cx+x)*16+rnd.nextInt(16),rnd.nextInt(256),(cz+z)*16+rnd.nextInt(16));
         ch.map.put(pos,new BlockEntity(new BlockState(b)));
       }
       level.put(cx+x,cz+z,ch);
     }
     List<BlockPos> exp=reference(level,center,radius), got=actual(level,center,radius);
     check(exp.equals(got),"equivalence mismatch case "+t+" exp="+exp+" got="+got);
     int chunkLookups=level.chunkLookups;
     int beLookups=level.blockEntityLookups;
     check(actual(level,center,radius).equals(exp),"same-key cache changed result");
     check(level.chunkLookups==chunkLookups,"same tick unexpectedly rescanned chunks");
     check(level.blockEntityLookups>=beLookups,"live target resolution counter regressed");
     level.time++;
     actual(level,center,radius);
     check(level.chunkLookups==chunkLookups+(2*radius+1)*(2*radius+1),"tick invalidation did not rescan exact grid");
   }
   // Explicit invalidation dimensions.
   ClientLevel a=new ClientLevel(); a.time=1; BlockPos c=new BlockPos(0,64,0); actual(a,c,1); int base=a.chunkLookups;
   actual(a,new BlockPos(16,64,0),1); check(a.chunkLookups==base+9,"center chunk change did not rebuild"); base=a.chunkLookups;
   actual(a,new BlockPos(16,64,0),2); check(a.chunkLookups==base+25,"radius change did not rebuild");
   ClientLevel b=new ClientLevel(); b.time=1; actual(b,new BlockPos(16,64,0),2); check(b.chunkLookups==25,"level identity change did not rebuild");

   // Same-tick removal/replacement at an already-known position must be observed without a full rescan.
   ClientLevel live=new ClientLevel(); live.time=77; LevelChunk ch=new LevelChunk(); BlockPos pos=new BlockPos(1,70,1);
   ch.map.put(pos,new BlockEntity(new BlockState(ProtectionPixelModBlocks.HANGER))); live.put(0,0,ch);
   check(actual(live,new BlockPos(0,64,0),0).equals(List.of(pos)),"initial live target missing");
   int liveScans=live.chunkLookups;
   ch.map.put(pos,new BlockEntity(new BlockState(ProtectionPixelModBlocks.OTHER)));
   check(actual(live,new BlockPos(0,64,0),0).isEmpty(),"same-tick replacement stayed stale");
   check(live.chunkLookups==liveScans,"same-tick live revalidation rescanned chunk grid");

   System.out.println("PASS cases="+cases+" stable-state target order/filter equivalence; same-tick chunk-scan reuse; live target revalidation; tick/chunk/radius/level invalidation");
 }
}
