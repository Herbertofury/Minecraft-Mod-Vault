#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit("usage: compat_pass_4.py <generated-port-root>")

root = Path(sys.argv[1]).resolve()
meta = root / "forge/src/main/resources/META-INF/mods.toml"
forge_build = root / "forge/build.gradle"
forge_mod = root / "forge/src/main/java/net/jewelry/forge/ForgeMod.java"

# Upstream 2.4.0 makes Curios optional. The generated Forge shell initially made it mandatory only
# to simplify early compile work; restore the real dependency contract before runtime/release gates.
text = meta.read_text()
curios_required = '''[[dependencies.jewelry]]
modId="curios"
mandatory=true
versionRange="[5.14,6)"
ordering="AFTER"
side="BOTH"
'''
curios_optional = '''[[dependencies.jewelry]]
modId="curios"
mandatory=false
versionRange="[5.14,6)"
ordering="AFTER"
side="BOTH"
'''
if curios_required not in text:
    raise SystemExit("compat pass 4 expected temporary mandatory Curios metadata")
meta.write_text(text.replace(curios_required, curios_optional, 1))

# NeoForge 2.4.0 declares jewelry.mixins.json in its metadata. Forge 47 discovers mod mixin configs
# from the packaged JAR manifest, so carry that declaration through every Jar/shadow/remap path.
build = forge_build.read_text()
if "attributes 'MixinConfigs': 'jewelry.mixins.json'" not in build:
    build += r'''

// Forge 47 packaged-runtime mixin discovery equivalent of NeoForge's [[mixins]] metadata.
tasks.withType(Jar).configureEach {
    manifest {
        attributes 'MixinConfigs': 'jewelry.mixins.json'
    }
}
'''
forge_build.write_text(build)

# Runtime-only acceptance hooks. They are inert for players and execute only when CI explicitly sets
# -Djewelry.ci.selftest=true. This proves registration/content and both optional-Curios branches from
# the packaged JAR, rather than merely proving Loom can compile the source tree.
fm = forge_mod.read_text()
if "import net.jewelry.Platform;" not in fm:
    fm = fm.replace("import net.jewelry.JewelryMod;\n", "import net.jewelry.JewelryMod;\nimport net.jewelry.Platform;\n", 1)
if "import net.minecraft.util.Identifier;" not in fm:
    fm = fm.replace("import net.minecraft.registry.RegistryKeys;\n", "import net.minecraft.registry.RegistryKeys;\nimport net.minecraft.util.Identifier;\n", 1)
if "import net.minecraftforge.event.server.ServerStartedEvent;" not in fm:
    fm = fm.replace("import net.minecraftforge.event.village.VillagerTradesEvent;\n", "import net.minecraftforge.event.village.VillagerTradesEvent;\nimport net.minecraftforge.event.server.ServerStartedEvent;\n", 1)
listener_anchor = "        MinecraftForge.EVENT_BUS.addListener(ForgeMod::onVillagerTrades);\n"
if "ForgeMod::onServerStarted" not in fm:
    if listener_anchor not in fm:
        raise SystemExit("compat pass 4 server self-test listener anchor missing")
    fm = fm.replace(listener_anchor, listener_anchor + "        MinecraftForge.EVENT_BUS.addListener(ForgeMod::onServerStarted);\n", 1)

if "private static void onServerStarted(ServerStartedEvent event)" not in fm:
    close = fm.rfind("}\n")
    if close < 0:
        raise SystemExit("compat pass 4 could not find ForgeMod class close")
    method = r'''

    private static void ciAssert(boolean condition, String message) {
        if (!condition) {
            throw new IllegalStateException("Jewelry CI self-test failed: " + message);
        }
    }

    private static void onServerStarted(ServerStartedEvent event) {
        if (!Boolean.getBoolean("jewelry.ci.selftest")) {
            return;
        }

        String[] itemIds = {
                "diamond_ring", "unique_attack_ring", "unique_crit_ring", "unique_dex_ring",
                "unique_arcane_ring", "unique_fire_ring", "unique_frost_ring",
                "unique_healing_ring", "unique_spell_ring", "unique_tank_ring"
        };
        for (String path : itemIds) {
            ciAssert(Registries.ITEM.containsId(new Identifier(JewelryMod.ID, path)),
                    "missing current item " + path);
        }
        ciAssert(Registries.BLOCK.containsId(new Identifier(JewelryMod.ID, "jewelers_kit")),
                "missing jewelers_kit block");
        ciAssert(Registries.VILLAGER_PROFESSION.containsId(new Identifier(JewelryMod.ID, JewelryVillagers.JEWELER)),
                "missing jeweler profession");
        ciAssert(Registries.POINT_OF_INTEREST_TYPE.containsId(JewelryVillagers.POI_ID),
                "missing jeweler POI");

        var trades = JewelryVillagers.createTrades();
        for (int tier = 1; tier <= 5; tier++) {
            ciAssert(trades.containsKey(tier) && !trades.get(tier).isEmpty(),
                    "missing villager trade tier " + tier);
        }

        boolean expectCurios = Boolean.parseBoolean(System.getProperty("jewelry.ci.expectCurios", "false"));
        boolean curiosLoaded = Platform.util().isModLoaded("curios");
        ciAssert(curiosLoaded == expectCurios,
                "Curios loaded=" + curiosLoaded + " expected=" + expectCurios);
        var diamondRing = Registries.ITEM.get(new Identifier(JewelryMod.ID, "diamond_ring"));
        boolean curioFactoryActive = diamondRing.getClass().getName().endsWith("JewelryCurioItem");
        ciAssert(curioFactoryActive == expectCurios,
                "Curios item factory active=" + curioFactoryActive + " expected=" + expectCurios);

        System.out.println("[Jewelry CI] Packaged runtime self-test passed: items=" + itemIds.length
                + ", villagerTiers=" + trades.size() + ", curios=" + curiosLoaded);
    }
'''
    fm = fm[:close] + method + fm[close:]
forge_mod.write_text(fm)

final_meta = meta.read_text()
if 'modId="curios"\nmandatory=false' not in final_meta:
    raise SystemExit("compat pass 4 did not restore optional Curios metadata")
final_build = forge_build.read_text()
if "attributes 'MixinConfigs': 'jewelry.mixins.json'" not in final_build:
    raise SystemExit("compat pass 4 missing packaged Forge mixin manifest")
final_mod = forge_mod.read_text()
for required in (
    "ForgeMod::onServerStarted",
    "Boolean.getBoolean(\"jewelry.ci.selftest\")",
    "Registries.POINT_OF_INTEREST_TYPE.containsId",
    "JewelryVillagers.createTrades()",
    "JewelryCurioItem",
    "[Jewelry CI] Packaged runtime self-test passed",
):
    if required not in final_mod:
        raise SystemExit(f"compat pass 4 missing packaged runtime invariant: {required}")

print("Jewelry compatibility pass 4 applied: optional Curios + packaged mixin manifest + runtime self-test")
