#!/usr/bin/env python3
from pathlib import Path
import json, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6d.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
common_java = root / 'common/src/main/java'
common_res = root / 'common/src/main/resources'
forge_java = root / 'forge/src/main/java/net/spell_engine/forge'

# 1.21.x renamed this class to ServerChunkLoadingManager. On the exact 1.20.1 Yarn target the
# same vanilla tracker implementation is ThreadedAnvilChunkStorage; use the target-native names.
accessor_dir = common_java / 'net/spell_engine/mixin/server'
accessor_dir.mkdir(parents=True, exist_ok=True)
accessor_dir.joinpath('ThreadedAnvilChunkStorageAccessor.java').write_text(r'''package net.spell_engine.mixin.server;

import it.unimi.dsi.fastutil.ints.Int2ObjectMap;
import net.minecraft.server.world.ThreadedAnvilChunkStorage;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;

/** Exact vanilla entity-tracker map access for the Forge platform bridge. */
@Mixin(ThreadedAnvilChunkStorage.class)
public interface ThreadedAnvilChunkStorageAccessor {
    @Accessor("entityTrackers")
    Int2ObjectMap<Object> spellEngine_entityTrackers();
}
''')
accessor_dir.joinpath('ThreadedAnvilEntityTrackerAccessor.java').write_text(r'''package net.spell_engine.mixin.server;

import net.minecraft.server.world.EntityTrackingListener;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;

import java.util.Set;

/** Exact listener set from ThreadedAnvilChunkStorage.EntityTracker. */
@Mixin(targets = "net.minecraft.server.world.ThreadedAnvilChunkStorage$EntityTracker")
public interface ThreadedAnvilEntityTrackerAccessor {
    @Accessor("listeners")
    Set<EntityTrackingListener> spellEngine_listeners();
}
''')

mixins = common_res / 'spell_engine.mixins.json'
data = json.loads(mixins.read_text())
# Remove only names from the earlier 1.21-class-name diagnostic attempt if present.
for stale in ('server.ServerChunkLoadingManagerAccessor', 'server.ServerChunkEntityTrackerAccessor'):
    if stale in data.get('mixins', []):
        data['mixins'].remove(stale)
for name in ('server.ThreadedAnvilChunkStorageAccessor', 'server.ThreadedAnvilEntityTrackerAccessor'):
    if name not in data.get('mixins', []):
        data.setdefault('mixins', []).append(name)
mixins.write_text(json.dumps(data, indent=2) + '\n')

forge_java.joinpath('ForgeTracking.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.entity.Entity;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.server.world.ServerWorld;
import net.spell_engine.mixin.server.ThreadedAnvilChunkStorageAccessor;
import net.spell_engine.mixin.server.ThreadedAnvilEntityTrackerAccessor;

import java.util.ArrayList;
import java.util.Collection;
import java.util.List;

/**
 * Forge 47 implementation of PlayerLookup.tracking(Entity) semantics.
 * Reads the real 1.20.1 vanilla tracker/listener set, not a distance or chunk approximation.
 */
final class ForgeTracking {
    private ForgeTracking() { }

    static Collection<ServerPlayerEntity> players(Entity entity) {
        if (!(entity.getWorld() instanceof ServerWorld world)) {
            return List.of();
        }
        var storage = world.getChunkManager().threadedAnvilChunkStorage;
        var trackers = ((ThreadedAnvilChunkStorageAccessor) (Object) storage).spellEngine_entityTrackers();
        var tracker = trackers.get(entity.getId());
        if (tracker == null) {
            return List.of();
        }
        var listeners = ((ThreadedAnvilEntityTrackerAccessor) tracker).spellEngine_listeners();
        var players = new ArrayList<ServerPlayerEntity>(listeners.size());
        for (var listener : listeners) {
            players.add(listener.getPlayer());
        }
        return players;
    }
}
''')

tracking = forge_java.joinpath('ForgeTracking.java').read_text()
if 'return new ArrayList<>()' in tracking or 'compile scaffold' in tracking:
    raise SystemExit('ForgeTracking scaffold survived pass 6d')
final_mixins = json.loads(mixins.read_text()).get('mixins', [])
for name in ('server.ThreadedAnvilChunkStorageAccessor', 'server.ThreadedAnvilEntityTrackerAccessor'):
    if name not in final_mixins:
        raise SystemExit(f'missing tracker accessor mixin: {name}')
print('Spell Engine compatibility pass 6d applied: exact Forge 47/MC 1.20.1 entity-tracker recipients')
