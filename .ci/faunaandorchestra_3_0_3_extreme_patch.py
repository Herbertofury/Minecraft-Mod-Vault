#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
CHANGED: list[str] = []
NOTES: list[str] = []


def read(rel: str) -> str:
    return (ROOT / rel).read_text(encoding="utf-8")


def write(rel: str, text: str) -> None:
    path = ROOT / rel
    old = path.read_text(encoding="utf-8")
    if old == text:
        return
    path.write_text(text, encoding="utf-8", newline="\n")
    CHANGED.append(rel)


def replace_once(rel: str, old: str, new: str, note: str | None = None) -> None:
    text = read(rel)
    count = text.count(old)
    if count != 1:
        raise RuntimeError(f"{rel}: expected exactly 1 literal match, found {count}")
    write(rel, text.replace(old, new, 1))
    if note:
        NOTES.append(note)


def regex_once(rel: str, pattern: str, repl: str, note: str | None = None, flags: int = 0) -> None:
    text = read(rel)
    new, count = re.subn(pattern, repl, text, count=1, flags=flags)
    if count != 1:
        raise RuntimeError(f"{rel}: expected exactly 1 regex match, found {count}: {pattern[:160]}")
    write(rel, new)
    if note:
        NOTES.append(note)


# Round-two / Wave 2A: exact-behavior hardening and hot-path allocation/network removal.

# 1) Living Music: the 80-tick despawn countdown was SynchedEntityData and dirtied
# entity metadata every server tick. No client renderer/logic reads it and the only
# external setter (BambooTrapBlock) executes on both logical sides. Keep the same field,
# persistence key, 80->0 timing and trap sentinel, but make it ordinary entity state.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/custom/LivingMusicEntity.java"
text = read(rel)
text = text.replace("import net.minecraft.network.syncher.EntityDataAccessor;\n", "")
text = text.replace("import net.minecraft.network.syncher.EntityDataSerializers;\n", "")
text = text.replace("import net.minecraft.network.syncher.SynchedEntityData;\n", "")
text = text.replace(
    "    protected static final EntityDataAccessor<Integer> TICKS_UNTIL_DEATH = SynchedEntityData.defineId(LivingMusicEntity.class, EntityDataSerializers.INT);\n",
    "",
)
text = text.replace(
    "    public static final int BASE_TICKS_UNTIL_DEATH = 80;\n",
    "    public static final int BASE_TICKS_UNTIL_DEATH = 80;\n    private int ticksUntilDeath = BASE_TICKS_UNTIL_DEATH;\n",
    1,
)
text = text.replace(
    """    @Override\n    protected void defineSynchedData() {\n        super.defineSynchedData();\n\n        entityData.define(TICKS_UNTIL_DEATH, BASE_TICKS_UNTIL_DEATH);\n    }\n\n""",
    "",
    1,
)
text = text.replace(
    """    public int getTicksUntilDeath() {\n        return entityData.get(TICKS_UNTIL_DEATH);\n    }\n\n    public void setTicksUntilDeath(int ticks) {\n        entityData.set(TICKS_UNTIL_DEATH, ticks);\n    }\n""",
    """    public int getTicksUntilDeath() {\n        return ticksUntilDeath;\n    }\n\n    public void setTicksUntilDeath(int ticks) {\n        this.ticksUntilDeath = ticks;\n    }\n""",
    1,
)
if "TICKS_UNTIL_DEATH" in text or "EntityDataAccessor" in text or "EntityDataSerializers" in text:
    raise RuntimeError(f"{rel}: synced countdown remnants remain")
write(rel, text)
NOTES.append("LivingMusicEntity countdown is now ordinary persisted state instead of dirtying SynchedEntityData every server tick; lifetime/trap timing is unchanged.")

# 2) Active frog choir: preserve the exact 45-block inflated-AABB / living-player stop
# condition while iterating the dimension player list directly. This removes an entity
# section query and temporary result list every goal tick.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/goals/QuirkyFrogConductingChoirGoal.java"
text = read(rel)
text = text.replace("import net.minecraft.world.entity.EntitySelector;\n", "")
text = text.replace("import net.minecraft.world.phys.Vec3;\n", "import net.minecraft.world.phys.AABB;\nimport net.minecraft.world.phys.Vec3;\n", 1)
old = """        if (times == 8 ||\n                conductor.isTame() ||\n                this.conductor.level().getEntitiesOfClass(\n                Player.class, this.conductor.getBoundingBox().inflate(45.0, 45.0, 45.0), EntitySelector.LIVING_ENTITY_STILL_ALIVE).isEmpty()) {\n            this.stop();\n        }\n"""
new = """        if (times == 8 || conductor.isTame() || !hasNearbyLivingPlayer()) {\n            this.stop();\n        }\n"""
if old not in text:
    raise RuntimeError(f"{rel}: choir player-query block not found")
text = text.replace(old, new, 1)
anchor = """    private void croac(QuirkyFrogEntity chorister) {\n"""
helper = """    private boolean hasNearbyLivingPlayer() {\n        AABB bounds = conductor.getBoundingBox().inflate(45.0, 45.0, 45.0);\n        for (Player player : conductor.level().players()) {\n            if (player.isAlive() && bounds.intersects(player.getBoundingBox())) {\n                return true;\n            }\n        }\n        return false;\n    }\n\n"""
if anchor not in text:
    raise RuntimeError(f"{rel}: croac anchor not found")
text = text.replace(anchor, helper + anchor, 1)
write(rel, text)
NOTES.append("Active frog choir no longer performs a radius entity-section query/list allocation every goal tick; exact 45-block living-player condition is preserved.")

# 3) Great Composer listener tracking: first-pass already replaced the huge entity query
# with direct server-player iteration. Round two removes the remaining three transient
# player lists per active boss tick by reusing one scratch list and diffing in-place.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/custom/boss/TheGreatComposer.java"
replace_once(
    rel,
    "    private List<Player> playersListening = new ArrayList<>();\n",
    "    private final List<Player> playersListening = new ArrayList<>();\n    private final List<Player> nearbyPlayersScratch = new ArrayList<>();\n",
)
old = """                List<Player> nearbyPlayers = new ArrayList<>();\n                var listenerBounds = this.getBoundingBox().inflate(63.0, 32.0, 63.0);\n                for (ServerPlayer serverPlayer : ((ServerLevel) this.level()).players()) {\n                    if (serverPlayer.isAlive() && listenerBounds.intersects(serverPlayer.getBoundingBox())) {\n                        nearbyPlayers.add(serverPlayer);\n                    }\n                }\n\n                List<Player> newPlayers = new ArrayList<>(nearbyPlayers);\n                List<Player> exitPlayers = new ArrayList<>(playersListening);\n                exitPlayers.removeAll(nearbyPlayers);\n                newPlayers.removeAll(playersListening);\n\n                for (Player player : newPlayers) {\n                    if (player instanceof  ServerPlayer serverPlayer) {\n                        ModNetwork.sendToPlayer(new StartAmbientMusicS2CPacket(this.uuid), serverPlayer);\n                    }\n                }\n\n                for (Player player : exitPlayers) {\n                    if (player instanceof  ServerPlayer serverPlayer) {\n                        ModNetwork.sendToPlayer(new StopMusicS2CPacket(this.uuid), serverPlayer);\n                    }\n                }\n\n                playersListening = nearbyPlayers;\n"""
new = """                nearbyPlayersScratch.clear();\n                var listenerBounds = this.getBoundingBox().inflate(63.0, 32.0, 63.0);\n                for (ServerPlayer serverPlayer : ((ServerLevel) this.level()).players()) {\n                    if (serverPlayer.isAlive() && listenerBounds.intersects(serverPlayer.getBoundingBox())) {\n                        nearbyPlayersScratch.add(serverPlayer);\n                    }\n                }\n\n                for (Player player : nearbyPlayersScratch) {\n                    if (!playersListening.contains(player) && player instanceof ServerPlayer serverPlayer) {\n                        ModNetwork.sendToPlayer(new StartAmbientMusicS2CPacket(this.uuid), serverPlayer);\n                    }\n                }\n\n                for (Player player : playersListening) {\n                    if (!nearbyPlayersScratch.contains(player) && player instanceof ServerPlayer serverPlayer) {\n                        ModNetwork.sendToPlayer(new StopMusicS2CPacket(this.uuid), serverPlayer);\n                    }\n                }\n\n                playersListening.clear();\n                playersListening.addAll(nearbyPlayersScratch);\n"""
if old not in read(rel):
    raise RuntimeError(f"{rel}: first-pass listener diff block not found")
replace_once(rel, old, new, "Great Composer listener diff reuses scratch/player lists instead of allocating three lists every active server tick.")

# 4) Orchestra goal: keep first-pass loaded-chunk listener walk, but reuse player buffers
# and compute the musician UUID snapshot only once when membership actually changes.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/goals/ConductorEntityConductingOrchestraGoal.java"
text = read(rel)
text = text.replace(
    "    private List<Player> playersListening;\n",
    "    private final List<Player> playersListening = new ArrayList<>();\n    private final List<Player> nearbyPlayersScratch = new ArrayList<>();\n",
    1,
)
text = text.replace(
    "        this.playersListening = this.getNearbyPlayers(50.0);\n",
    "        this.collectNearbyPlayers(50.0, this.playersListening);\n",
    1,
)
text = text.replace(
    "            List<Player> nearbyPlayers = this.getNearbyPlayers(32.0);\n",
    "            this.collectNearbyPlayers(32.0, this.nearbyPlayersScratch);\n            List<UUID> orchestraIds = snapshotOrchestraIds();\n",
    1,
)
text = text.replace(
    "            for (Player player : nearbyPlayers) {\n",
    "            for (Player player : nearbyPlayersScratch) {\n",
    1,
)
text = text.replace(
    "                                conductor.getOrchestra().stream().map(Entity::getUUID).toList(),\n",
    "                                orchestraIds,\n",
    1,
)
old = """        List<Player> nearbyPlayers = this.getNearbyPlayers(32.0);\n\n        // We find which players weren't nearby before and now are and send Packets to them\n        List<Player> newPlayers = new ArrayList<>(nearbyPlayers);\n        List<Player> exitPlayers = new ArrayList<>(playersListening);\n        exitPlayers.removeAll(nearbyPlayers);\n        newPlayers.removeAll(playersListening);\n        for (Player player : newPlayers) {\n            if (player instanceof  ServerPlayer serverPlayer) {\n                ModNetwork.sendToPlayer(new RestartOrchestraMusicS2CPacket(\n                                conductor.getUUID(),\n                                conductor.getOrchestra().stream().map(Entity::getUUID).toList(),\n                                conductor.getTicksPlaying(),\n                                conductor.getCurrentVolume(),\n                                conductor.getSheetMusic().toString()),\n                        serverPlayer);\n            }\n        }\n\n        for (Player player : exitPlayers) {\n            if (player instanceof ServerPlayer serverPlayer) {\n                ModNetwork.sendToPlayer(new StopOrchestraMusicS2CPacket(\n                            conductor.getOrchestra().stream().map(Entity::getUUID).toList()),\n                        serverPlayer);\n            }\n        }\n\n        playersListening = nearbyPlayers;\n"""
new = """        this.collectNearbyPlayers(32.0, this.nearbyPlayersScratch);\n\n        // Find which players crossed the listening boundary. Build the orchestra UUID\n        // snapshot lazily and only once for this tick if at least one packet needs it.\n        List<UUID> orchestraIds = null;\n        for (Player player : nearbyPlayersScratch) {\n            if (!playersListening.contains(player) && player instanceof ServerPlayer serverPlayer) {\n                if (orchestraIds == null) orchestraIds = snapshotOrchestraIds();\n                ModNetwork.sendToPlayer(new RestartOrchestraMusicS2CPacket(\n                                conductor.getUUID(),\n                                orchestraIds,\n                                conductor.getTicksPlaying(),\n                                conductor.getCurrentVolume(),\n                                conductor.getSheetMusic().toString()),\n                        serverPlayer);\n            }\n        }\n\n        for (Player player : playersListening) {\n            if (!nearbyPlayersScratch.contains(player) && player instanceof ServerPlayer serverPlayer) {\n                if (orchestraIds == null) orchestraIds = snapshotOrchestraIds();\n                ModNetwork.sendToPlayer(new StopOrchestraMusicS2CPacket(orchestraIds), serverPlayer);\n            }\n        }\n\n        playersListening.clear();\n        playersListening.addAll(nearbyPlayersScratch);\n"""
if old not in text:
    raise RuntimeError(f"{rel}: first-pass normal player diff block not found")
text = text.replace(old, new, 1)
old_helper = """    private List<Player> getNearbyPlayers(double radius) {\n        AABB bounds = conductor.getBoundingBox().inflate(radius, radius, radius);\n        List<Player> players = new ArrayList<>();\n        for (Player player : conductor.level().players()) {\n            if (player.isAlive() && bounds.intersects(player.getBoundingBox())) {\n                players.add(player);\n            }\n        }\n        return players;\n    }\n\n"""
new_helper = """    private void collectNearbyPlayers(double radius, List<Player> target) {\n        AABB bounds = conductor.getBoundingBox().inflate(radius, radius, radius);\n        target.clear();\n        for (Player player : conductor.level().players()) {\n            if (player.isAlive() && bounds.intersects(player.getBoundingBox())) {\n                target.add(player);\n            }\n        }\n    }\n\n    private List<UUID> snapshotOrchestraIds() {\n        List<UUID> ids = new ArrayList<>(conductor.getOrchestra().size());\n        for (MusicalEntity musician : conductor.getOrchestra()) {\n            ids.add(musician.getUUID());\n        }\n        return ids;\n    }\n\n"""
if old_helper not in text:
    raise RuntimeError(f"{rel}: first-pass getNearbyPlayers helper not found")
text = text.replace(old_helper, new_helper, 1)
if "getOrchestra().stream().map(Entity::getUUID).toList()" in text:
    raise RuntimeError(f"{rel}: repeated orchestra UUID stream remains")
write(rel, text)
NOTES.append("Orchestra listener tracking now reuses player buffers and creates at most one musician-UUID snapshot per membership-change tick instead of list/stream churn per player.")

# 5) Musician goal hardening: stop() dereferenced conductor after a null guard. Move the
# empty-orchestra check inside the guard; normal behavior is identical, invalid state no
# longer crashes.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/goals/MusicalEntityPlayingInstrumentGoal.java"
replace_once(
    rel,
    """        if (conductor != null) {\n            conductor.removeMusician(musician);\n        }\n\n        if (conductor.isOrchestraEmpty()) {\n            conductor.setTicksPlaying(0);\n        }\n\n        musician.setConductor(null);\n""",
    """        if (conductor != null) {\n            conductor.removeMusician(musician);\n            if (conductor.isOrchestraEmpty()) {\n                conductor.setTicksPlaying(0);\n            }\n        }\n\n        musician.setConductor(null);\n""",
    "MusicalEntityPlayingInstrumentGoal.stop no longer dereferences a null conductor; normal orchestra shutdown behavior is unchanged.",
)

# 6) Anya Ghost: do not write the exact same UUID to SynchedEntityData every tick.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/custom/AnyaGhost.java"
replace_once(
    rel,
    """            if (player != null) {\n                this.setPlayerUUID(player.getUUID());\n                if (level().isClientSide()) {\n""",
    """            if (player != null) {\n                UUID nearestPlayerUUID = player.getUUID();\n                if (!nearestPlayerUUID.equals(playerUUID)) {\n                    this.setPlayerUUID(nearestPlayerUUID);\n                }\n                if (level().isClientSide()) {\n""",
    "Anya Ghost avoids redundant synchronized UUID writes when the nearest player has not changed.",
)

# 7) Listener Container: cache the above block state once in the server tick instead of
# fetching the exact same block position twice every tick.
rel = "src/main/java/net/migueel26/faunaandorchestra/block/entity/ListenerContainerBlockEntity.java"
replace_once(
    rel,
    """        boolean isBottle = state.getValue(ListenerContainerBlock.BOTTLE);\n        boolean isAssembled = state.getValue(ListenerContainerBlock.LISTENING) && level.getBlockState(pos.above()).getOptionalValue(ListenerBlock.LISTENING).orElse(false);\n        int drops = entity.getDroplets();\n        boolean hasListenerAbove = level.getBlockState(pos.above()).is(ModBlocks.LISTENER.get());\n""",
    """        boolean isBottle = state.getValue(ListenerContainerBlock.BOTTLE);\n        BlockPos abovePos = pos.above();\n        BlockState aboveState = level.getBlockState(abovePos);\n        boolean isAssembled = state.getValue(ListenerContainerBlock.LISTENING) && aboveState.getOptionalValue(ListenerBlock.LISTENING).orElse(false);\n        int drops = entity.getDroplets();\n        boolean hasListenerAbove = aboveState.is(ModBlocks.LISTENER.get());\n""",
    "ListenerContainerBlockEntity caches its above-block state once per tick instead of reading the same position twice.",
)

report = {
    "upstream_commit": "131bdcbf76aec07267d68a57185bb103669af83e",
    "target": "Fauna & Orchestra Forge 1.20.1 3.0.3 - extreme performance Wave 2A",
    "changed_files": sorted(set(CHANGED)),
    "optimizations": NOTES,
}
(ROOT / "EXTREME-PERFORMANCE-PATCH-REPORT.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
print(json.dumps(report, indent=2))
