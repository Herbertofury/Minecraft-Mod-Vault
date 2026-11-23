package net.archers.forge;

import net.archers.ArchersMod;
import net.archers.block.ArcherBlocks;
import net.archers.effect.ArcherEffects;
import net.archers.entity.ArcherEntities;
import net.archers.item.ArcherArmors;
import net.archers.item.ArcherWeapons;
import net.archers.item.Quivers;
import net.archers.item.misc.AutoFireHook;
import net.archers.village.ArcherVillagers;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.registry.Registries;
import net.minecraft.util.Identifier;

/** CI-only semantic smoke checks. Inert unless ARCHERS_SELF_TEST=1 is present. */
public final class ArchersCiSelfTest {
    private ArchersCiSelfTest() { }

    public static void runIfRequested() {
        if (!"1".equals(System.getenv("ARCHERS_SELF_TEST"))) return;

        requireItem(ArcherWeapons.composite_longbow.item(), "composite_longbow");
        requireItem(ArcherWeapons.rapid_crossbow.item(), "rapid_crossbow");
        requireItem(ArcherArmors.archerArmorSet_T1.head, "archer_armor_head");
        requireItem(AutoFireHook.item, "auto_fire_hook");

        var workbenchId = Registries.BLOCK.getId(ArcherBlocks.WORKBENCH.block());
        requireId(workbenchId, ArcherBlocks.WORKBENCH.name(), "workbench block");

        requireId(Registries.STATUS_EFFECT.getId(ArcherEffects.HUNTERS_MARK.effect), "hunters_mark", "Hunter's Mark effect");
        requireId(Registries.STATUS_EFFECT.getId(ArcherEffects.ENTANGLING_ROOTS.effect), "entangling_roots", "Entangling Roots effect");
        requireId(Registries.ENTITY_TYPE.getId(ArcherEntities.SPIRIT_WOLF.type), "spirit_wolf", "Spirit Wolf entity");
        requireId(Registries.VILLAGER_PROFESSION.getId(ArcherVillagers.PROFESSION), ArcherVillagers.ARCHERY_ARTISAN, "archery artisan profession");
        requireId(Registries.POINT_OF_INTEREST_TYPE.getId(Registries.POINT_OF_INTEREST_TYPE.get(ArcherVillagers.POI_ID)), ArcherVillagers.ARCHERY_ARTISAN, "archery artisan POI");

        if (ArcherVillagers.TRADES.size() != 5) {
            fail("expected 5 villager trade tiers, found " + ArcherVillagers.TRADES.size());
        }
        for (int tier = 1; tier <= 5; tier++) {
            if (!ArcherVillagers.TRADES.containsKey(tier) || ArcherVillagers.TRADES.get(tier).isEmpty()) {
                fail("missing villager trade tier " + tier);
            }
        }

        if (Quivers.entries.size() != 3) {
            fail("expected 3 current quivers, found " + Quivers.entries.size());
        }
        int[] capacities = {4, 8, 12};
        String[] ids = {"small_quiver", "medium_quiver", "large_quiver"};
        for (int i = 0; i < capacities.length; i++) {
            var entry = Quivers.entries.get(i);
            if (entry.capacity() != capacities[i]) {
                fail(ids[i] + " capacity expected " + capacities[i] + ", found " + entry.capacity());
            }
            requireItem(entry.item(), ids[i]);
            if (entry.item().getMaxCount() != 1) {
                fail(ids[i] + " max stack must remain 1");
            }
        }

        var autoFireStack = new ItemStack(ArcherWeapons.rapid_crossbow.item());
        if (AutoFireHook.isApplied(autoFireStack)) fail("Auto Fire unexpectedly pre-applied");
        AutoFireHook.apply(autoFireStack);
        if (!AutoFireHook.isApplied(autoFireStack)) fail("Auto Fire NBT apply did not persist");
        AutoFireHook.remove(autoFireStack);
        if (AutoFireHook.isApplied(autoFireStack)) fail("Auto Fire NBT removal failed");

        System.out.println("ARCHERS_SELF_TEST_PASS");
    }

    private static void requireItem(Item item, String path) {
        requireId(Registries.ITEM.getId(item), path, "item " + path);
    }

    private static void requireId(Identifier id, String path, String label) {
        var expected = new Identifier(ArchersMod.ID, path);
        if (!expected.equals(id)) {
            fail(label + " expected id " + expected + ", found " + id);
        }
    }

    private static void fail(String message) {
        throw new IllegalStateException("ARCHERS_SELF_TEST_FAILED: " + message);
    }
}
