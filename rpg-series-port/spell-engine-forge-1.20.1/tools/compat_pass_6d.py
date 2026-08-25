#!/usr/bin/env python3
from pathlib import Path
import json, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6d.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
common_java = root / 'common/src/main/java'
common_res = root / 'common/src/main/resources'
forge_java = root / 'forge/src/main/java/net/spell_engine/forge'

accessor_dir = common_java / 'net/spell_engine/mixin/server'
accessor_dir.mkdir(parents=True, exist_ok=True)
accessor_dir.joinpath('ServerChunkLoadingManagerAccessor.java').write_text(r'''package net.spell_engine.mixin.server;

import it.unimi.dsi.fastutil.ints.Int2ObjectMap;
import net.minecraft.server.world.ServerChunkLoadingManager;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;

/** Exact vanilla entity-tracker map access for the Forge platform bridge. */
@Mixin(ServerChunkLoadingManager.class)
public interface ServerChunkLoadingManagerAccessor {
    @Accessor("entityTrackers")
    Int2ObjectMap<Object> spellEngine_entityTrackers();
}
''')
accessor_dir.joinpath('ServerChunkEntityTrackerAccessor.java').write_text(r'''package net.spell_engine.mixin.server;

import net.minecraft.server.world.EntityTrackingListener;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;

import java.util.Set;

/** Exact listener set from ServerChunkLoadingManager.EntityTracker. */
@Mixin(targets = "net.minecraft.server.world.ServerChunkLoadingManager$EntityTracker")
public interface ServerChunkEntityTrackerAccessor {
    @Accessor("listeners")
    Set<EntityTrackingListener> spellEngine_listeners();
}
''')

mixins = common_res / 'spell_engine.mixins.json'
data = json.loads(mixins.read_text())
for name in ('server.ServerChunkLoadingManagerAccessor', 'server.ServerChunkEntityTrackerAccessor'):
    if name not in data.get('mixins', []):
        data.setdefault('mixins', []).append(name)
mixins.write_text(json.dumps(data, indent=2) + '\n')

forge_java.joinpath('ForgeTracking.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.entity.Entity;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.server.world.ServerWorld;
import net.spell_engine.mixin.server.ServerChunkEntityTrackerAccessor;
import net.spell_engine.mixin.server.ServerChunkLoadingManagerAccessor;

import java.util.ArrayList;
import java.util.Collection;
import java.util.List;

/**
 * Forge 47 implementation of Fabric PlayerLookup.tracking(Entity) semantics.
 * Reads the real vanilla tracker/listener set, not a distance or chunk approximation.
 */
final class ForgeTracking {
    private ForgeTracking() { }

    static Collection<ServerPlayerEntity> players(Entity entity) {
        if (!(entity.getWorld() instanceof ServerWorld world)) {
            return List.of();
        }
        var manager = world.getChunkManager().chunkLoadingManager;
        var trackers = ((ServerChunkLoadingManagerAccessor) (Object) manager).spellEngine_entityTrackers();
        var tracker = trackers.get(entity.getId());
        if (tracker == null) {
            return List.of();
        }
        var listeners = ((ServerChunkEntityTrackerAccessor) tracker).spellEngine_listeners();
        var players = new ArrayList<ServerPlayerEntity>(listeners.size());
        for (var listener : listeners) {
            players.add(listener.getPlayer());
        }
        return players;
    }
}
''')

# Hard guards: no empty-recipient scaffold and both accessors must be registered.
tracking = forge_java.joinpath('ForgeTracking.java').read_text()
if 'return new ArrayList<>()' in tracking or 'compile scaffold' in tracking:
    raise SystemExit('ForgeTracking scaffold survived pass 6d')
final_mixins = json.loads(mixins.read_text()).get('mixins', [])
for name in ('server.ServerChunkLoadingManagerAccessor', 'server.ServerChunkEntityTrackerAccessor'):
    if name not in final_mixins:
        raise SystemExit(f'missing tracker accessor mixin: {name}')
print('Spell Engine compatibility pass 6d applied: exact Forge 47 entity-tracker recipients')
