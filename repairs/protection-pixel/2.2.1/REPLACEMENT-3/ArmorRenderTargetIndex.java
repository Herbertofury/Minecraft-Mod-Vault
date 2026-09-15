package net.mcreator.protectionpixel.repair;

import java.lang.ref.WeakReference;
import java.util.AbstractMap;
import java.util.AbstractSet;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;
import java.util.NoSuchElementException;
import java.util.Set;
import net.minecraft.core.BlockPos;
import net.minecraft.core.SectionPos;
import net.minecraft.world.level.Level;
import net.minecraft.world.level.block.entity.BlockEntity;

/**
 * Client-side lifecycle index for Protection Pixel armor-preview targets.
 *
 * Target block entities register when Forge calls onLoad() and unregister when
 * BlockEntity#setRemoved() runs. Rendering therefore never scans chunks or
 * unrelated block entities. The steady-state render path only walks the sparse
 * set of known Protection Pixel targets and resolves each live block entity.
 *
 * This class intentionally references only common Minecraft types so the two
 * target block-entity classes can call it safely on a dedicated server. Server
 * calls immediately no-op via Level#isClientSide().
 */
public final class ArmorRenderTargetIndex {
    private static final ArrayList<Target> TARGETS = new ArrayList<>();
    private static final HashMap<BlockPos, Target> BY_POS = new HashMap<>();
    private static WeakReference<Level> activeLevelRef = new WeakReference<>(null);
    private static final TargetMapView VIEW = new TargetMapView();

    private ArmorRenderTargetIndex() {}

    public static void register(Level level, BlockPos pos) {
        if (level == null || !level.m_5776_() || pos == null) {
            return;
        }
        ensureLevel(level);
        BlockPos immutable = pos.m_7949_();
        if (BY_POS.containsKey(immutable)) {
            return;
        }
        Target target = new Target(
                immutable,
                SectionPos.m_123171_(immutable.m_123341_()),
                SectionPos.m_123171_(immutable.m_123343_()));
        BY_POS.put(immutable, target);
        TARGETS.add(target);
    }

    public static void unregister(Level level, BlockPos pos) {
        if (level == null || !level.m_5776_() || pos == null || activeLevelRef.get() != level) {
            return;
        }
        Target target = BY_POS.remove(pos);
        if (target != null) {
            TARGETS.remove(target); // lifecycle changes are rare; keep render iteration array-dense.
        }
    }

    public static Map<BlockPos, BlockEntity> targets(Level level, BlockPos center, int radius) {
        if (level == null || !level.m_5776_() || center == null) {
            return java.util.Collections.emptyMap();
        }
        ensureLevel(level);
        VIEW.bind(
                SectionPos.m_123171_(center.m_123341_()),
                SectionPos.m_123171_(center.m_123343_()),
                radius);
        return VIEW;
    }

    private static void ensureLevel(Level level) {
        if (activeLevelRef.get() != level) {
            TARGETS.clear();
            BY_POS.clear();
            activeLevelRef = new WeakReference<>(level);
        }
    }

    private static void removeStale(Target target) {
        if (BY_POS.remove(target.pos) != null) {
            TARGETS.remove(target);
        }
    }

    private static final class Target {
        final BlockPos pos;
        final int chunkX;
        final int chunkZ;

        Target(BlockPos pos, int chunkX, int chunkZ) {
            this.pos = pos;
            this.chunkX = chunkX;
            this.chunkZ = chunkZ;
        }
    }

    /** Reusable, allocation-free map facade for the existing renderer loop. */
    private static final class TargetMapView extends AbstractMap<BlockPos, BlockEntity> {
        private int centerChunkX;
        private int centerChunkZ;
        private int radius;
        private final TargetEntrySet entrySet = new TargetEntrySet(this);

        void bind(int centerChunkX, int centerChunkZ, int radius) {
            this.centerChunkX = centerChunkX;
            this.centerChunkZ = centerChunkZ;
            this.radius = radius;
        }

        boolean inRange(Target target) {
            return Math.abs(target.chunkX - centerChunkX) <= radius &&
                    Math.abs(target.chunkZ - centerChunkZ) <= radius;
        }

        Level level() {
            return activeLevelRef.get();
        }

        @Override
        public Set<Entry<BlockPos, BlockEntity>> entrySet() {
            return entrySet;
        }
    }

    private static final class TargetEntrySet extends AbstractSet<Map.Entry<BlockPos, BlockEntity>> {
        private final TargetMapView view;
        private final TargetCursor cursor;

        TargetEntrySet(TargetMapView view) {
            this.view = view;
            this.cursor = new TargetCursor(view);
        }

        @Override
        public Iterator<Map.Entry<BlockPos, BlockEntity>> iterator() {
            return cursor.reset();
        }

        @Override
        public int size() {
            int count = 0;
            Level level = view.level();
            if (level == null) {
                return 0;
            }
            for (int i = 0; i < TARGETS.size(); ++i) {
                Target target = TARGETS.get(i);
                if (view.inRange(target) && level.m_7702_(target.pos) != null) {
                    ++count;
                }
            }
            return count;
        }
    }

    /**
     * One reusable cursor implements both Iterator and Map.Entry, eliminating
     * steady-frame iterator/entry allocation. Rendering is client-thread confined.
     */
    private static final class TargetCursor implements Iterator<Map.Entry<BlockPos, BlockEntity>>, Map.Entry<BlockPos, BlockEntity> {
        private final TargetMapView view;
        private int index;
        private Target nextTarget;
        private BlockEntity nextValue;
        private Target currentTarget;
        private BlockEntity currentValue;

        TargetCursor(TargetMapView view) {
            this.view = view;
        }

        TargetCursor reset() {
            index = 0;
            nextTarget = null;
            nextValue = null;
            currentTarget = null;
            currentValue = null;
            return this;
        }

        @Override
        public boolean hasNext() {
            if (nextTarget != null) {
                return true;
            }
            currentTarget = null;
            currentValue = null;
            Level level = view.level();
            if (level == null) {
                return false;
            }
            while (index < TARGETS.size()) {
                Target target = TARGETS.get(index++);
                if (!view.inRange(target)) {
                    continue;
                }
                BlockEntity blockEntity = level.m_7702_(target.pos);
                if (blockEntity == null) {
                    // Exact lifecycle hooks should normally remove this first; self-heal if another mod bypasses them.
                    --index;
                    removeStale(target);
                    continue;
                }
                nextTarget = target;
                nextValue = blockEntity;
                return true;
            }
            return false;
        }

        @Override
        public Map.Entry<BlockPos, BlockEntity> next() {
            if (!hasNext()) {
                throw new NoSuchElementException();
            }
            currentTarget = nextTarget;
            currentValue = nextValue;
            nextTarget = null;
            nextValue = null;
            return this;
        }

        @Override
        public BlockPos getKey() {
            return currentTarget == null ? null : currentTarget.pos;
        }

        @Override
        public BlockEntity getValue() {
            return currentValue;
        }

        @Override
        public BlockEntity setValue(BlockEntity value) {
            throw new UnsupportedOperationException();
        }
    }
}
