#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/wizards-forge-1.20.1"
GEN="$PORT/generated"
UPLIFT="$ROOT/rpg-series-port/ci/run-wizards-spell-engine-1.10.4-acceptance.sh"
QA="$GEN/qa"
QA_WORLD='WIZARDS-QA'
QA_LOG="$PORT/wizards-arcane-blast-gameplay.log"

# Preserve the complete 1.10.4 build/package/server/client acceptance as the prerequisite.
bash "$UPLIFT"

test -d "$GEN"
test -f "$GEN/libs/spell-engine-common.jar"
test -f "$GEN/libs/spell-engine-forge.jar"
unzip -tq "$GEN/libs/spell-engine-forge.jar"

# Ground the gameplay oracle in the exact ported 3.1.1 content before launching Minecraft.
grep -F 'Identifier.of(WizardsMod.ID, "arcane_blast")' "$GEN/common/src/main/java/net/wizards/content/WizardSpells.java" >/dev/null
grep -F 'effectAdd(WizardsEffects.arcaneCharge.id.toString()' "$GEN/common/src/main/java/net/wizards/content/WizardSpells.java" >/dev/null
grep -F 'NAMESPACE, "staff_arcane"' "$GEN/common/src/main/java/net/wizards/item/WizardWeapons.java" >/dev/null
grep -F 'withSpellId(WizardSpells.arcane_blast.id())' "$GEN/common/src/main/java/net/wizards/item/WizardWeapons.java" >/dev/null
grep -F 'Identifier.of(WizardsMod.ID, "arcane_charge")' "$GEN/common/src/main/java/net/wizards/effect/WizardsEffects.java" >/dev/null
echo '[Wizards QA] ARCANE_BLAST_SOURCE_ORACLE_PASS spell=wizards:arcane_blast weapon=wizards:staff_arcane effect=wizards:arcane_charge'

# Add a disposable Loom/Forge QA subproject. It is a separate mod JAR and never enters the Wizards
# production source set or release artifact.
if ! grep -Eq "^[[:space:]]*include[[:space:]]+['\"]qa['\"]" "$GEN/settings.gradle"; then
  printf "\ninclude 'qa'\n" >> "$GEN/settings.gradle"
fi
rm -rf "$QA"
mkdir -p "$QA/src/main/java/net/wizards/qa" "$QA/src/main/resources/META-INF"
cat > "$QA/build.gradle" <<'GRADLE'
architectury { platformSetupLoomIde(); forge() }

dependencies {
    forge "net.minecraftforge:forge:$rootProject.forge_version"
    modImplementation files('../libs/spell-engine-common.jar')
    modImplementation files('../libs/spell-engine-forge.jar')
}

processResources {
    inputs.property 'version', project.version
    filesMatching('META-INF/mods.toml') { expand(version: project.version) }
}
GRADLE
cat > "$QA/src/main/resources/META-INF/mods.toml" <<'TOML'
modLoader="javafml"
loaderVersion="[47,)"
license="All Rights Reserved"
[[mods]]
modId="wizards_qa"
version="${version}"
displayName="Wizards Runtime QA"
description="CI-only native gameplay verifier; never shipped."
[[dependencies.wizards_qa]]
modId="forge"
mandatory=true
versionRange="[47.4,48)"
ordering="NONE"
side="BOTH"
[[dependencies.wizards_qa]]
modId="minecraft"
mandatory=true
versionRange="[1.20.1,1.20.2)"
ordering="NONE"
side="BOTH"
[[dependencies.wizards_qa]]
modId="wizards"
mandatory=true
versionRange="[3.1.1,)"
ordering="AFTER"
side="BOTH"
[[dependencies.wizards_qa]]
modId="spell_engine"
mandatory=true
versionRange="[1.10.4,)"
ordering="AFTER"
side="BOTH"
TOML
cat > "$QA/src/main/java/net/wizards/qa/WizardsGameplayQA.java" <<'JAVA'
package net.wizards.qa;

import net.minecraft.entity.EntityType;
import net.minecraft.entity.effect.StatusEffect;
import net.minecraft.entity.mob.ZombieEntity;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.registry.Registries;
import net.minecraft.server.MinecraftServer;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.server.world.ServerWorld;
import net.minecraft.util.Hand;
import net.minecraft.util.Identifier;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.entity.player.PlayerEvent;
import net.minecraftforge.fml.common.Mod;
import net.spell_engine.api.spell.registry.SpellRegistry;
import net.spell_engine.internals.SpellExecution;
import net.spell_engine.internals.casting.SpellCast;
import net.spell_engine.internals.target.SpellTarget;

import java.util.concurrent.atomic.AtomicBoolean;

@Mod(WizardsGameplayQA.MOD_ID)
public final class WizardsGameplayQA {
    public static final String MOD_ID = "wizards_qa";
    private static final AtomicBoolean STARTED = new AtomicBoolean(false);

    public WizardsGameplayQA() {
        MinecraftForge.EVENT_BUS.addListener(this::onPlayerLogin);
    }

    private void onPlayerLogin(PlayerEvent.PlayerLoggedInEvent event) {
        PlayerEntity entity = event.getEntity();
        if (!(entity instanceof ServerPlayerEntity player)) return;
        if (!STARTED.compareAndSet(false, true)) return;
        MinecraftServer server = player.getServer();
        if (server == null) {
            fail("server_null");
            return;
        }
        server.execute(() -> runArcaneBlast(player));
    }

    private static void runArcaneBlast(ServerPlayerEntity player) {
        try {
            ServerWorld world = player.getServerWorld();
            Identifier spellId = new Identifier("wizards", "arcane_blast");
            Identifier staffId = new Identifier("wizards", "staff_arcane");
            Identifier runeId = new Identifier("runes", "arcane_stone");
            Identifier chargeId = new Identifier("wizards", "arcane_charge");

            var spellEntry = SpellRegistry.from(world).getEntry(spellId).orElse(null);
            if (spellEntry == null) fail("missing_spell_registry_entry:" + spellId);

            Item staff = Registries.ITEM.get(staffId);
            Item rune = Registries.ITEM.get(runeId);
            StatusEffect charge = Registries.STATUS_EFFECT.get(chargeId);
            if (staff == null || staff == Items.AIR) fail("missing_staff:" + staffId);
            if (rune == null || rune == Items.AIR) fail("missing_rune:" + runeId);
            if (charge == null) fail("missing_effect:" + chargeId);

            player.setStackInHand(Hand.MAIN_HAND, new ItemStack(staff));
            if (!player.getInventory().insertStack(new ItemStack(rune, 64))) {
                fail("could_not_insert_arcane_runes");
            }

            ZombieEntity target = EntityType.ZOMBIE.create(world);
            if (target == null) fail("zombie_create_failed");
            target.setAiDisabled(true);
            target.setPersistent();
            target.refreshPositionAndAngles(player.getX() + 2.0D, player.getY(), player.getZ(), 0.0F, 0.0F);
            if (!world.spawnEntity(target)) fail("zombie_spawn_failed");

            float beforeHealth = target.getHealth();
            SpellExecution.performSpell(
                    world,
                    player,
                    spellEntry,
                    SpellTarget.SearchResult.of(target),
                    SpellCast.Action.RELEASE,
                    1.0F
            );
            float afterHealth = target.getHealth();
            boolean arcaneCharge = player.hasStatusEffect(charge);

            if (!(afterHealth < beforeHealth)) {
                fail("target_health_not_reduced:before=" + beforeHealth + ",after=" + afterHealth);
            }
            if (!arcaneCharge) {
                fail("caster_missing_arcane_charge");
            }

            System.out.println("[Wizards QA] ARCANE_BLAST_GAMEPLAY_PASS"
                    + " spell=wizards:arcane_blast"
                    + " weapon=wizards:staff_arcane"
                    + " target=minecraft:zombie"
                    + " before=" + beforeHealth
                    + " after=" + afterHealth
                    + " caster_effect=wizards:arcane_charge");
            System.out.flush();
            System.exit(0);
        } catch (Throwable t) {
            fail(t.getClass().getName() + ":" + String.valueOf(t.getMessage()));
        }
    }

    private static void fail(String reason) {
        System.err.println("[Wizards QA] ARCANE_BLAST_GAMEPLAY_FAIL reason=" + reason);
        System.err.flush();
        System.exit(17);
    }
}
JAVA

gradle --no-daemon --stacktrace -p "$GEN" :qa:compileJava :qa:remapJar
QA_JAR="$(find "$QA/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev*' | sort | head -n1)"
[[ -n "$QA_JAR" && -f "$QA_JAR" ]]
unzip -tq "$QA_JAR"
unzip -p "$QA_JAR" META-INF/mods.toml | grep -F 'modId="wizards_qa"' >/dev/null
if unzip -Z1 "$QA_JAR" | grep -q '^net/wizards/[^q]'; then
  echo '[Wizards QA] production Wizards classes leaked into QA JAR' >&2
  exit 2
fi
echo "[Wizards QA] QA_JAR_READY $(sha256sum "$QA_JAR" | awk '{print $1}')"

# Reuse the already-proven packaged-server world from the prerequisite acceptance. This avoids a
# second world-generation lease while still entering a real integrated-server save.
WORLD_SRC="$GEN/.fresh-wizards-forge-server/world"
[[ -f "$WORLD_SRC/level.dat" ]]
RUN="$GEN/forge/run"
WORLD_DST="$RUN/saves/$QA_WORLD"
rm -rf "$WORLD_DST"
mkdir -p "$RUN/saves" "$RUN/mods" "$RUN/config"
cp -a "$WORLD_SRC" "$WORLD_DST"
rm -f "$RUN/mods"/*.jar
cp "$QA_JAR" "$RUN/mods/"
printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"
if [[ -f "$RUN/options.txt" ]] && grep -q '^onboardAccessibility:' "$RUN/options.txt"; then
  sed -i 's/^onboardAccessibility:.*/onboardAccessibility:false/' "$RUN/options.txt"
else
  printf 'onboardAccessibility:false\n' >> "$RUN/options.txt"
fi
[[ "$(grep -Ec '^onboardAccessibility:false$' "$RUN/options.txt")" -eq 1 ]]

rm -rf "$RUN/logs"
: > "$QA_LOG"
set +e
timeout --signal=TERM --kill-after=10s 240s env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$GEN" :forge:runClient --args "--quickPlaySingleplayer $QA_WORLD" \
  </dev/null > "$QA_LOG" 2>&1
CLIENT_STATUS=$?
set -e
CLIENT_LATEST=$(find "$RUN" -type f -path '*/logs/latest.log' | head -n1 || true)
QA_FILES=("$QA_LOG"); [[ -n "$CLIENT_LATEST" ]] && QA_FILES+=("$CLIENT_LATEST")

if grep -Fq '[Wizards QA] ARCANE_BLAST_GAMEPLAY_FAIL' "${QA_FILES[@]}"; then
  cat "${QA_FILES[@]}"
  exit 1
fi
if grep -Fq '[Wizards QA] ARCANE_BLAST_GAMEPLAY_PASS' "${QA_FILES[@]}"; then
  echo '[Wizards 3.1.1] ARCANE_BLAST_NATIVE_GAMEPLAY_PASS spell=wizards:arcane_blast damage=true caster_arcane_charge=true spell_engine=1.10.4'
  exit 0
fi

FATAL='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|The game crashed|Exception in thread "Render thread"|Missing or unsupported mandatory dependencies'
if grep -Eiq "$FATAL" "${QA_FILES[@]}"; then
  cat "${QA_FILES[@]}"
  exit 1
fi
cat "${QA_FILES[@]}"
echo "[Wizards QA] native gameplay proof not reached; client_status=$CLIENT_STATUS" >&2
exit 1
