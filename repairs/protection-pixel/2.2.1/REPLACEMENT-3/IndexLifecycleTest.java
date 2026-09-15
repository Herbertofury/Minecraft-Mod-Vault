import java.util.*;
import net.minecraft.core.BlockPos;
import net.minecraft.world.level.Level;
import net.minecraft.world.level.block.entity.BlockEntity;
import net.mcreator.protectionpixel.repair.ArmorRenderTargetIndex;

public final class IndexLifecycleTest {
    private static final Random RNG = new Random(0x50505233L);

    public static void main(String[] args) {
        Level level = new Level();
        LinkedHashMap<BlockPos, BlockEntity> reference = new LinkedHashMap<>();
        LinkedHashSet<BlockPos> indexedReference = new LinkedHashSet<>();
        ArrayList<BlockPos> universe = new ArrayList<>();
        for (int i = 0; i < 600; ++i) {
            universe.add(new BlockPos(RNG.nextInt(1024)-512, RNG.nextInt(128), RNG.nextInt(1024)-512));
        }

        // Server-side calls must be inert.
        Level server = new Level();
        server.client = false;
        BlockPos serverPos = new BlockPos(0,64,0);
        server.entities.put(serverPos, new BlockEntity());
        ArmorRenderTargetIndex.register(server, serverPos);
        if (!ArmorRenderTargetIndex.targets(server, serverPos, 32).isEmpty()) fail("server index was not inert");

        // Establish client level after the server no-op path.
        Map<BlockPos, BlockEntity> firstView = ArmorRenderTargetIndex.targets(level, new BlockPos(0,64,0), 8);
        Map<BlockPos, BlockEntity> secondView = ArmorRenderTargetIndex.targets(level, new BlockPos(0,64,0), 8);
        if (firstView != secondView) fail("render map view allocated/replaced between calls");
        Iterator<Map.Entry<BlockPos, BlockEntity>> i1 = firstView.entrySet().iterator();
        Iterator<Map.Entry<BlockPos, BlockEntity>> i2 = firstView.entrySet().iterator();
        if (i1 != i2) fail("render iterator allocated/replaced between calls");

        int queries = 0;
        int staleSelfHeals = 0;
        for (int step = 0; step < 50000; ++step) {
            int op = RNG.nextInt(100);
            BlockPos pos = universe.get(RNG.nextInt(universe.size()));
            if (op < 38) {
                BlockEntity be = new BlockEntity();
                level.entities.put(pos, be);
                ArmorRenderTargetIndex.register(level, pos);
                reference.putIfAbsent(pos, be);
                indexedReference.add(pos);
                // Replacing the live BE at an indexed position must be seen without re-registering.
                if (RNG.nextInt(7) == 0) {
                    BlockEntity replacement = new BlockEntity();
                    level.entities.put(pos, replacement);
                    if (reference.containsKey(pos)) reference.put(pos, replacement);
                }
            } else if (op < 62) {
                level.entities.remove(pos);
                ArmorRenderTargetIndex.unregister(level, pos);
                reference.remove(pos);
                indexedReference.remove(pos);
            } else if (op < 67 && reference.containsKey(pos)) {
                // Simulate a third party bypassing the normal lifecycle; query must self-heal nulls.
                level.entities.remove(pos);
                staleSelfHeals++;
            } else {
                BlockPos center = universe.get(RNG.nextInt(universe.size()));
                int radius = RNG.nextInt(10);
                assertQuery(level, reference, indexedReference, center, radius);
                // Self-heal only stale indexed positions actually visited by this radius.
                indexedReference.removeIf(p -> inRange(p, center, radius) && !level.entities.containsKey(p));
                reference.entrySet().removeIf(e -> inRange(e.getKey(), center, radius) && !level.entities.containsKey(e.getKey()));
                queries++;
            }
        }

        // World-size independence: 20k unrelated BEs are never registered and must never be scanned.
        for (int n = 0; n < 20000; ++n) {
            BlockPos p = new BlockPos(100000 + n, 64, 100000 + (n * 3));
            level.entities.put(p, new BlockEntity());
        }
        BlockPos center = new BlockPos(0,64,0);
        level.lookupCount = 0;
        int expectedCandidates = countInRange(indexedReference, center, 64);
        Map<BlockPos, BlockEntity> view = ArmorRenderTargetIndex.targets(level, center, 64);
        int seen = 0;
        for (Map.Entry<BlockPos, BlockEntity> ignored : view.entrySet()) seen++;
        if (level.lookupCount != expectedCandidates) {
            fail("lookup count depended on non-target world size: lookups=" + level.lookupCount + " expected indexed candidates=" + expectedCandidates);
        }

        // Level transition clears old-world positions; delayed old-level removal cannot erase the new level.
        Level nextLevel = new Level();
        LinkedHashMap<BlockPos, BlockEntity> nextRef = new LinkedHashMap<>();
        LinkedHashSet<BlockPos> nextIndexedRef = new LinkedHashSet<>();
        if (!ArmorRenderTargetIndex.targets(nextLevel, center, 64).isEmpty()) fail("new level inherited old targets");
        for (int n = 0; n < 25; ++n) {
            BlockPos p = new BlockPos(n * 17, 70, n * -19);
            BlockEntity be = new BlockEntity();
            nextLevel.entities.put(p, be);
            ArmorRenderTargetIndex.register(nextLevel, p);
            nextRef.put(p, be);
            nextIndexedRef.add(p);
        }
        // Delayed callback from old world should be ignored.
        ArmorRenderTargetIndex.unregister(level, reference.keySet().stream().findFirst().orElse(center));
        assertQuery(nextLevel, nextRef, nextIndexedRef, center, 64);

        System.out.println("PASS randomized_steps=50000 queries=" + queries +
                " stale_self_heals=" + staleSelfHeals +
                " unrelated_block_entities=20000 indexed_lookup_bound=" + expectedCandidates +
                " reusable_map_view=true reusable_iterator=true server_noop=true world_transition=true");
    }

    private static void assertQuery(Level level, LinkedHashMap<BlockPos, BlockEntity> reference, LinkedHashSet<BlockPos> indexedReference, BlockPos center, int radius) {
        ArrayList<BlockPos> expected = new ArrayList<>();
        for (BlockPos p : indexedReference) {
            if (!level.entities.containsKey(p)) continue;
            if (inRange(p, center, radius)) expected.add(p);
        }
        level.lookupCount = 0;
        Map<BlockPos, BlockEntity> actualMap = ArmorRenderTargetIndex.targets(level, center, radius);
        ArrayList<BlockPos> actual = new ArrayList<>();
        for (Map.Entry<BlockPos, BlockEntity> e : actualMap.entrySet()) {
            actual.add(e.getKey());
            if (e.getValue() != level.entities.get(e.getKey())) fail("value not live for " + e.getKey());
        }
        if (!actual.equals(expected)) {
            fail("query mismatch expected=" + expected.size() + " actual=" + actual.size());
        }
        int candidates = countInRange(indexedReference, center, radius);
        if (level.lookupCount > candidates) {
            fail("too many BE lookups: " + level.lookupCount + " > " + candidates);
        }
    }

    private static int countInRange(Collection<BlockPos> positions, BlockPos center, int radius) {
        int count = 0;
        for (BlockPos p : positions) if (inRange(p, center, radius)) count++;
        return count;
    }

    private static boolean inRange(BlockPos p, BlockPos center, int radius) {
        int pcx = p.m_123341_() >> 4;
        int pcz = p.m_123343_() >> 4;
        int ccx = center.m_123341_() >> 4;
        int ccz = center.m_123343_() >> 4;
        return Math.abs(pcx - ccx) <= radius && Math.abs(pcz - ccz) <= radius;
    }

    private static void fail(String message) { throw new AssertionError(message); }
}
