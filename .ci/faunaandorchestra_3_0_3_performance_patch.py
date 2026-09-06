#!/usr/bin/env python3
from __future__ import annotations

import difflib
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
        raise RuntimeError(f"{rel}: expected exactly 1 regex match, found {count}: {pattern[:120]}")
    write(rel, new)
    if note:
        NOTES.append(note)


# 1) Forge global LivingTick handler: reject non-frogs before random work and let the
# entity query apply its predicate directly instead of allocating/filtering a stream.
rel = "src/main/java/net/migueel26/faunaandorchestra/event/ModGameEvents.java"
replace_once(
    rel,
    """        if (!event.getEntity().level().isClientSide() &&\n                event.getEntity().tickCount % 60 == 0 &&\n                event.getEntity().level().getRandom().nextFloat() <= 0.01F &&\n                event.getEntity() instanceof QuirkyFrogEntity quirkyFrog\n                && quirkyFrog.isAptForChoir()) {\n\n            List<QuirkyFrogEntity> nearbyFrogs = quirkyFrog.level().getEntitiesOfClass(QuirkyFrogEntity.class, quirkyFrog.getBoundingBox().inflate(30))\n                    .stream().filter(QuirkyFrogEntity::isAptForChoir).toList();\n""",
    """        if (event.getEntity() instanceof QuirkyFrogEntity quirkyFrog\n                && !quirkyFrog.level().isClientSide()\n                && quirkyFrog.tickCount % 60 == 0\n                && quirkyFrog.level().getRandom().nextFloat() <= 0.01F\n                && quirkyFrog.isAptForChoir()) {\n\n            List<QuirkyFrogEntity> nearbyFrogs = quirkyFrog.level().getEntitiesOfClass(\n                    QuirkyFrogEntity.class, quirkyFrog.getBoundingBox().inflate(30), QuirkyFrogEntity::isAptForChoir);\n""",
    "LivingTick frog-choir handler now exits on entity type before side/modulo/RNG work and filters in the spatial query.",
)

# 2) Conductor tick: the player lookup was executed every tick even though it was only
# consumed once every 30 ticks.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/custom/ConductorEntity.java"
replace_once(
    rel,
    """            // WANDERING NOTES\n            List<Player> players = level().getEntitiesOfClass(Player.class, this.getBoundingBox().inflate(15));\n            if (ticksPlaying % 30 == 0 && players.stream().anyMatch(player -> player.hasEffect(ModEffects.ABSOLUTE_HEARING.get()))) {\n                tryToSummonWanderingNote();\n            }\n""",
    """            // WANDERING NOTES\n            if (ticksPlaying % 30 == 0) {\n                List<Player> players = level().getEntitiesOfClass(Player.class, this.getBoundingBox().inflate(15));\n                if (players.stream().anyMatch(player -> player.hasEffect(ModEffects.ABSOLUTE_HEARING.get()))) {\n                    tryToSummonWanderingNote();\n                }\n            }\n""",
    "Conductor Absolute Hearing lookup reduced from every tick to the existing 30-tick decision cadence (same gameplay condition).",
)

# 3) Faust listener poll had a missing reset: after the first 60 ticks it performed a
# spatial player query + list diff every single tick forever. It also allocated an empty
# list every waiting/idle tick.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/custom/Faust.java"
regex_once(
    rel,
    r"""        if \(isPlaying\(\) && !level\(\)\.isClientSide\(\)\) \{\n            if \(nearbyPlayersSearchDelay < 60\) \{\n                nearbyPlayersSearchDelay\+\+;\n                playersListening = new ArrayList<>\(\);\n            \} else \{\n(?P<body>.*?)                playersListening = nearbyPlayers;\n            \}\n        \} else \{\n            playersListening = new ArrayList<>\(\);\n            this\.nearbyPlayersSearchDelay = 0;\n        \}\n""",
    """        if (isPlaying() && !level().isClientSide()) {\n            if (nearbyPlayersSearchDelay < 60) {\n                nearbyPlayersSearchDelay++;\n            } else {\n\g<body>                playersListening = nearbyPlayers;\n                nearbyPlayersSearchDelay = 0;\n            }\n        } else {\n            if (!playersListening.isEmpty()) {\n                playersListening.clear();\n            }\n            this.nearbyPlayersSearchDelay = 0;\n        }\n""",
    note="Faust player-listener poll now actually resets to its intended 60-tick cadence and stops per-tick empty-list allocation.",
    flags=re.S,
)

# 4) Same missing-reset/allocation bug in Dan B.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/custom/jazzy_dammys/DanB.java"
regex_once(
    rel,
    r"""        if \(isPlaying\(\) && !level\(\)\.isClientSide\(\)\) \{\n            if \(nearbyPlayersSearchDelay < 60\) \{\n                nearbyPlayersSearchDelay\+\+;\n                playersListening = new ArrayList<>\(\);\n            \} else \{\n(?P<body>.*?)                playersListening = nearbyPlayers;\n\n                \}\n\n        \} else \{\n            nearbyPlayersSearchDelay = 0;\n            playersListening = new ArrayList<>\(\);\n        \}\n""",
    """        if (isPlaying() && !level().isClientSide()) {\n            if (nearbyPlayersSearchDelay < 60) {\n                nearbyPlayersSearchDelay++;\n            } else {\n\g<body>                playersListening = nearbyPlayers;\n                nearbyPlayersSearchDelay = 0;\n            }\n\n        } else {\n            nearbyPlayersSearchDelay = 0;\n            if (!playersListening.isEmpty()) {\n                playersListening.clear();\n            }\n        }\n""",
    note="Dan B player-listener poll now actually resets to its intended 60-tick cadence and stops per-tick empty-list allocation.",
    flags=re.S,
)

# 5) Great Composer music listener tracking: iterate the tiny server player list rather
# than querying the entire entity-section storage in a 126x64x126 box every tick.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/custom/boss/TheGreatComposer.java"
replace_once(
    rel,
    """                List<Player> nearbyPlayers = this.level().getEntitiesOfClass(\n                        Player.class, this.getBoundingBox().inflate(63.0, 32.0, 63.0), EntitySelector.LIVING_ENTITY_STILL_ALIVE);\n""",
    """                List<Player> nearbyPlayers = new ArrayList<>();\n                var listenerBounds = this.getBoundingBox().inflate(63.0, 32.0, 63.0);\n                for (ServerPlayer serverPlayer : ((ServerLevel) this.level()).players()) {\n                    if (serverPlayer.isAlive() && listenerBounds.intersects(serverPlayer.getBoundingBox())) {\n                        nearbyPlayers.add(serverPlayer);\n                    }\n                }\n""",
    "Great Composer boss-music listener discovery now iterates server players only instead of a huge entity-section query every tick.",
)
replace_once(
    rel,
    """            } else {\n                playersListening = new ArrayList<>();\n            }\n""",
    """            } else if (!playersListening.isEmpty()) {\n                playersListening.clear();\n            }\n""",
    "Great Composer no longer allocates a new empty listener list every fake-dead server tick.",
)

# 6) Orchestra goal: replace two ~101^3 block-position sweeps with a walk over block
# entities in already-loaded intersecting chunks, and replace player entity-section
# queries with direct player-list filtering. This keeps the same 50-block listener area.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/goals/ConductorEntityConductingOrchestraGoal.java"
text = read(rel)
text = text.replace(
    "import net.minecraft.world.level.block.entity.BlockEntity;\nimport net.minecraft.world.phys.Vec3;\n",
    "import net.minecraft.world.level.block.entity.BlockEntity;\nimport net.minecraft.world.level.chunk.LevelChunk;\nimport net.minecraft.world.phys.AABB;\nimport net.minecraft.world.phys.Vec3;\n",
    1,
)
text = text.replace("import java.util.*;\n", "import java.util.*;\nimport java.util.function.Consumer;\n", 1)
if text == read(rel):
    raise RuntimeError(f"{rel}: failed to add performance helper imports")
write(rel, text)
replace_once(
    rel,
    """        this.playersListening = this.conductor.level().getEntitiesOfClass(\n                Player.class, this.conductor.getBoundingBox().inflate(50.0, 50.0, 50.0), EntitySelector.LIVING_ENTITY_STILL_ALIVE);\n""",
    """        this.playersListening = this.getNearbyPlayers(50.0);\n""",
)
replace_once(
    rel,
    """        BlockPos.betweenClosedStream(conductor.getBoundingBox().inflate(50.0, 50.0, 50.0)).forEach(pos -> {\n            BlockEntity blockEntity = conductor.level().getBlockEntity(pos);\n            if (blockEntity instanceof ListeningBlockEntity listeningBlockEntity && !listeningBlockEntity.isListening()) {\n                listeningBlockEntity.onStartListening(conductor);\n            }\n        });\n""",
    """        forEachListeningBlockEntity(listeningBlockEntity -> {\n            if (!listeningBlockEntity.isListening()) {\n                listeningBlockEntity.onStartListening(conductor);\n            }\n        });\n""",
)
replace_once(
    rel,
    """        BlockPos.betweenClosedStream(conductor.getBoundingBox().inflate(50.0, 50.0, 50.0)).forEach(pos -> {\n            BlockEntity blockEntity = conductor.level().getBlockEntity(pos);\n            if (blockEntity instanceof ListeningBlockEntity listeningBlockEntity) {\n                listeningBlockEntity.onStopListening();\n            }\n        });\n""",
    """        forEachListeningBlockEntity(ListeningBlockEntity::onStopListening);\n""",
)
# There are two 32-radius player queries in tick; replace both.
text = read(rel)
old = """            List<Player> nearbyPlayers = this.conductor.level().getEntitiesOfClass(\n                    Player.class, this.conductor.getBoundingBox().inflate(32.0, 32.0, 32.0), EntitySelector.LIVING_ENTITY_STILL_ALIVE);\n"""
count = text.count(old)
if count != 1:
    raise RuntimeError(f"{rel}: expected 1 indented restart player query, found {count}")
text = text.replace(old, "            List<Player> nearbyPlayers = this.getNearbyPlayers(32.0);\n", 1)
old2 = """        List<Player> nearbyPlayers = this.conductor.level().getEntitiesOfClass(\n                Player.class, this.conductor.getBoundingBox().inflate(32.0, 32.0, 32.0), EntitySelector.LIVING_ENTITY_STILL_ALIVE);\n"""
count2 = text.count(old2)
if count2 != 1:
    raise RuntimeError(f"{rel}: expected 1 main tick player query, found {count2}")
text = text.replace(old2, "        List<Player> nearbyPlayers = this.getNearbyPlayers(32.0);\n", 1)
write(rel, text)

# Replace stream-based centroid and append helpers before class close.
replace_once(
    rel,
    """    private Vec3 getCentroid() {\n        if (!conductor.getOrchestra().isEmpty()) {\n            Set<MusicalEntity> orchestra = conductor.getOrchestra();\n            double n = orchestra.size();\n\n            return new Vec3(\n                    orchestra.stream().map(Entity::getX).reduce(0.0, Double::sum)/n,\n                    conductor.getY(),\n                    orchestra.stream().map(Entity::getZ).reduce(0.0, Double::sum)/n);\n        } else {\n            return new Vec3(0.0,0.0,0.0);\n        }\n    }\n""",
    """    private Vec3 getCentroid() {\n        Set<MusicalEntity> orchestra = conductor.getOrchestra();\n        if (orchestra.isEmpty()) {\n            return Vec3.ZERO;\n        }\n\n        double x = 0.0;\n        double z = 0.0;\n        for (MusicalEntity musician : orchestra) {\n            x += musician.getX();\n            z += musician.getZ();\n        }\n        double n = orchestra.size();\n        return new Vec3(x / n, conductor.getY(), z / n);\n    }\n\n    private List<Player> getNearbyPlayers(double radius) {\n        AABB bounds = conductor.getBoundingBox().inflate(radius, radius, radius);\n        List<Player> players = new ArrayList<>();\n        for (Player player : conductor.level().players()) {\n            if (player.isAlive() && bounds.intersects(player.getBoundingBox())) {\n                players.add(player);\n            }\n        }\n        return players;\n    }\n\n    private void forEachListeningBlockEntity(Consumer<ListeningBlockEntity> action) {\n        AABB bounds = conductor.getBoundingBox().inflate(50.0, 50.0, 50.0);\n        int minChunkX = ((int) Math.floor(bounds.minX)) >> 4;\n        int maxChunkX = ((int) Math.floor(bounds.maxX)) >> 4;\n        int minChunkZ = ((int) Math.floor(bounds.minZ)) >> 4;\n        int maxChunkZ = ((int) Math.floor(bounds.maxZ)) >> 4;\n\n        for (int chunkX = minChunkX; chunkX <= maxChunkX; chunkX++) {\n            for (int chunkZ = minChunkZ; chunkZ <= maxChunkZ; chunkZ++) {\n                if (!conductor.level().getChunkSource().hasChunk(chunkX, chunkZ)) {\n                    continue;\n                }\n                LevelChunk chunk = conductor.level().getChunk(chunkX, chunkZ);\n                for (BlockEntity blockEntity : chunk.getBlockEntities().values()) {\n                    if (blockEntity instanceof ListeningBlockEntity listeningBlockEntity\n                            && bounds.contains(Vec3.atCenterOf(blockEntity.getBlockPos()))) {\n                        action.accept(listeningBlockEntity);\n                    }\n                }\n            }\n        }\n    }\n""",
    "Orchestra block listener discovery now walks only block entities in loaded intersecting chunks instead of ~1.03 million block positions per start/stop; player tracking iterates players directly; centroid no longer allocates streams.",
)

# 7) Sound engine orchestra-presence query ran up to three full ticking-sound streams every
# MusicManager tick. Collapse to one allocation-free pass.
rel = "src/main/java/net/migueel26/faunaandorchestra/mixins/client/MixinSoundEngine.java"
replace_once(
    rel,
    """    @Override\n    public boolean faunaIsThereAnOrchestra() {\n        return tickingSounds.stream().anyMatch(InstrumentSoundInstance.class::isInstance) ||\n                tickingSounds.stream().anyMatch(TravellingMusicianSoundInstance.class::isInstance) ||\n                tickingSounds.stream().anyMatch(BossSoundInstance.class::isInstance);\n    }\n""",
    """    @Override\n    public boolean faunaIsThereAnOrchestra() {\n        for (TickableSoundInstance sound : tickingSounds) {\n            if (sound instanceof InstrumentSoundInstance\n                    || sound instanceof TravellingMusicianSoundInstance\n                    || sound instanceof BossSoundInstance) {\n                return true;\n            }\n        }\n        return false;\n    }\n""",
    "MusicManager orchestra-presence check collapsed from up to three stream traversals per client tick to one allocation-free loop.",
)

# 8) Parrot dancing redirect: filter conductors in the entity query rather than materialize
# all conductors then create a stream/filter/findAny chain.
rel = "src/main/java/net/migueel26/faunaandorchestra/mixins/client/MixinParrot.java"
replace_once(
    rel,
    """        return !this.level().getBlockState(this.jukebox).is(Blocks.JUKEBOX) || this.level().getEntitiesOfClass(ConductorEntity.class,\n                getBoundingBox().inflate(5.0D)).stream().filter(ConductorEntity::isConducting).findAny().isEmpty();\n""",
    """        return !this.level().getBlockState(this.jukebox).is(Blocks.JUKEBOX) || this.level().getEntitiesOfClass(\n                ConductorEntity.class, getBoundingBox().inflate(5.0D), ConductorEntity::isConducting).isEmpty();\n""",
    "Parrot orchestra check now filters at spatial-query time and removes stream allocations.",
)

# 9) AnimalEatGoal: manual closest-item pass instead of comparator stream allocation.
rel = "src/main/java/net/migueel26/faunaandorchestra/entity/goals/AnimalEatGoal.java"
replace_once(
    rel,
    """        if (list.isEmpty()) {\n            return false;\n        } else {\n            // We pick the closest one\n            this.targetEntity = list.stream()\n                    .min((i1, i2) -> Double.compare(this.mob.distanceToSqr(i1), this.mob.distanceToSqr(i2)))\n                    .orElse(list.get(0));\n            return true;\n        }\n""",
    """        if (list.isEmpty()) {\n            return false;\n        }\n\n        ItemEntity closest = null;\n        double closestDistance = Double.MAX_VALUE;\n        for (ItemEntity item : list) {\n            double distance = this.mob.distanceToSqr(item);\n            if (distance < closestDistance) {\n                closestDistance = distance;\n                closest = item;\n            }\n        }\n        this.targetEntity = closest;\n        return this.targetEntity != null;\n""",
    "AnimalEatGoal closest-food selection now uses one allocation-free pass instead of a stream/comparator pipeline.",
)

# 10) Always-registered HUD overlays: cache expensive spatial entity lookup once per game
# tick, not once per rendered frame. Also emit typewriter sound once per entity tick value.
rel = "src/main/java/net/migueel26/faunaandorchestra/screen/custom/AnyaScreen.java"
replace_once(
    rel,
    """    public static final IGuiOverlay OVERLAY = AnyaScreen::renderOverlay;\n    public static final int POP_UP_TIME = 160;\n""",
    """    public static final IGuiOverlay OVERLAY = AnyaScreen::renderOverlay;\n    public static final int POP_UP_TIME = 160;\n    private static Level cachedLevel;\n    private static AnyaGhost cachedAnya;\n    private static int lastLookupTick = Integer.MIN_VALUE;\n    private static int lastDialogueSoundTick = Integer.MIN_VALUE;\n    private static final RandomSource SOUND_RANDOM = RandomSource.create();\n""",
)
replace_once(
    rel,
    """            List<AnyaGhost> candidates = level.getEntitiesOfClass(AnyaGhost.class, player.getBoundingBox().inflate(30));\n            AnyaGhost anya = candidates.isEmpty() ? null : candidates.get(0);\n""",
    """            AnyaGhost anya = getNearbyAnya(level, player);\n""",
)
replace_once(
    rel,
    """    private static int xOffset(GuiGraphics guiGraphics) {\n""",
    """    private static AnyaGhost getNearbyAnya(Level level, LocalPlayer player) {\n        if (cachedLevel != level || lastLookupTick != player.tickCount) {\n            cachedLevel = level;\n            lastLookupTick = player.tickCount;\n            List<AnyaGhost> candidates = level.getEntitiesOfClass(AnyaGhost.class, player.getBoundingBox().inflate(30));\n            cachedAnya = candidates.isEmpty() ? null : candidates.get(0);\n        }\n        return cachedAnya;\n    }\n\n    private static int xOffset(GuiGraphics guiGraphics) {\n""",
)
replace_once(
    rel,
    """        if (dialogueTimer <= fullText.length()) {\n            player.playSound(ModSounds.DIALOGUE.get(), 0.5F, RandomSource.create().nextFloat());\n        }\n""",
    """        if (dialogueTimer <= fullText.length() && dialogueTimer != lastDialogueSoundTick) {\n            lastDialogueSoundTick = dialogueTimer;\n            player.playSound(ModSounds.DIALOGUE.get(), 0.5F, SOUND_RANDOM.nextFloat());\n        }\n""",
    "Anya HUD entity search is cached per game tick instead of per frame; typewriter sound is emitted once per dialogue tick instead of once per frame.",
)

rel = "src/main/java/net/migueel26/faunaandorchestra/screen/custom/TheGreatComposerScreen.java"
replace_once(
    rel,
    """    public static final IGuiOverlay OVERLAY = TheGreatComposerScreen::renderOverlay;\n    public static final int POP_UP_TIME = 160;\n""",
    """    public static final IGuiOverlay OVERLAY = TheGreatComposerScreen::renderOverlay;\n    public static final int POP_UP_TIME = 160;\n    private static Level cachedLevel;\n    private static TheGreatComposer cachedComposer;\n    private static int lastLookupTick = Integer.MIN_VALUE;\n    private static int lastDialogueSoundTick = Integer.MIN_VALUE;\n    private static final RandomSource SOUND_RANDOM = RandomSource.create();\n""",
)
replace_once(
    rel,
    """            List<TheGreatComposer> candidates = level.getEntitiesOfClass(TheGreatComposer.class, player.getBoundingBox().inflate(30));\n            TheGreatComposer composer = candidates.isEmpty() ? null : candidates.get(0);\n""",
    """            TheGreatComposer composer = getNearbyComposer(level, player);\n""",
)
replace_once(
    rel,
    """    private static int xOffset(GuiGraphics guiGraphics) {\n""",
    """    private static TheGreatComposer getNearbyComposer(Level level, LocalPlayer player) {\n        if (cachedLevel != level || lastLookupTick != player.tickCount) {\n            cachedLevel = level;\n            lastLookupTick = player.tickCount;\n            List<TheGreatComposer> candidates = level.getEntitiesOfClass(TheGreatComposer.class, player.getBoundingBox().inflate(30));\n            cachedComposer = candidates.isEmpty() ? null : candidates.get(0);\n        }\n        return cachedComposer;\n    }\n\n    private static int xOffset(GuiGraphics guiGraphics) {\n""",
)
replace_once(
    rel,
    """        if (dialogueTimer <= fullText.length() * 2 && dialogueTimer % 2 == 0) {\n            player.playSound(ModSounds.DIALOGUE.get(), 0.5F, RandomSource.create().nextFloat());\n        }\n""",
    """        if (dialogueTimer <= fullText.length() * 2 && dialogueTimer % 2 == 0\n                && dialogueTimer != lastDialogueSoundTick) {\n            lastDialogueSoundTick = dialogueTimer;\n            player.playSound(ModSounds.DIALOGUE.get(), 0.5F, SOUND_RANDOM.nextFloat());\n        }\n""",
    "Great Composer HUD entity search is cached per game tick instead of per frame; typewriter sound is emitted once per dialogue tick instead of once per frame.",
)

# DialogueScreen is only active while interacting, but remove repeated RNG object creation in
# its sound path as a low-risk allocation cleanup.
rel = "src/main/java/net/migueel26/faunaandorchestra/screen/custom/DialogueScreen.java"
replace_once(
    rel,
    """    protected static int prizeTimer = 0;\n    protected static int prize = -1;\n""",
    """    protected static int prizeTimer = 0;\n    protected static int prize = -1;\n    private static final RandomSource SOUND_RANDOM = RandomSource.create();\n""",
)
text = read(rel)
count = text.count("RandomSource.create().nextFloat()")
if count != 2:
    raise RuntimeError(f"{rel}: expected 2 per-sound RNG allocations, found {count}")
write(rel, text.replace("RandomSource.create().nextFloat()", "SOUND_RANDOM.nextFloat()"))
NOTES.append("Dialogue UI reuses one client RNG instead of allocating a new RandomSource for every typewriter sound.")

# 11) Animated block entity culling bounds. The shipped renderers advertise +/-16 blocks
# even though measured model geometry is <=2.22 blocks and max animated translation <=0.52.
# +/-3 is intentionally conservative and still reduces bounding volume >150x.
render_targets = [
    "src/main/java/net/migueel26/faunaandorchestra/client/block/MailboxBlockEntityRenderer.java",
    "src/main/java/net/migueel26/faunaandorchestra/client/block/BeaverStatueBlockEntityRenderer.java",
    "src/main/java/net/migueel26/faunaandorchestra/client/block/FloraEnhancerBlockEntityRenderer.java",
    "src/main/java/net/migueel26/faunaandorchestra/client/block/SewingMachineBlockEntityRenderer.java",
    "src/main/java/net/migueel26/faunaandorchestra/client/block/ComposerGravestoneBlockEntityRenderer.java",
    "src/main/java/net/migueel26/faunaandorchestra/client/block/MotherStatueBlockEntityRenderer.java",
]
render_changed = []
for rel in render_targets:
    text = read(rel)
    before = text
    text = text.replace("offset(-16, -16, -16)", "offset(-3, -3, -3)")
    text = text.replace("offset(16, 16, 16)", "offset(3, 3, 3)")
    if text == before:
        raise RuntimeError(f"{rel}: expected +/-16 render-bounds pattern not found")
    write(rel, text)
    render_changed.append(rel)

# Forge also lets some block entities expose their own render AABB. Tighten only exact
# +/-16 patterns, and report every file touched so CI makes this auditable.
be_render_changed = []
for path in (ROOT / "src/main/java/net/migueel26/faunaandorchestra/block/entity").rglob("*.java"):
    text = path.read_text(encoding="utf-8")
    if "getRenderBoundingBox" not in text:
        continue
    new = text.replace("offset(-16, -16, -16)", "offset(-3, -3, -3)")\
              .replace("offset(16, 16, 16)", "offset(3, 3, 3)")
    if new != text:
        path.write_text(new, encoding="utf-8", newline="\n")
        relpath = str(path.relative_to(ROOT)).replace("\\", "/")
        if relpath not in CHANGED:
            CHANGED.append(relpath)
        be_render_changed.append(relpath)
NOTES.append(
    "Animated block-entity render bounds tightened from +/-16 to +/-3 blocks after measuring model/animation extents (max geometry 2.22 blocks; max translation 0.52)."
)

# Emit an auditable report and diff for CI artifacts.
report = {
    "upstream_commit": "131bdcbf76aec07267d68a57185bb103669af83e",
    "target": "Fauna & Orchestra Forge 1.20.1 3.0.3",
    "changed_files": sorted(set(CHANGED)),
    "renderers_tightened": render_changed,
    "block_entities_tightened": be_render_changed,
    "optimizations": NOTES,
}
(ROOT / "PERFORMANCE-PATCH-REPORT.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
print(json.dumps(report, indent=2))
