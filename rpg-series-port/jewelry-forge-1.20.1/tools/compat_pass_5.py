#!/usr/bin/env python3
from pathlib import Path
import json
import shutil
import sys

if len(sys.argv) != 2:
    raise SystemExit("usage: compat_pass_5.py <generated-port-root>")

root = Path(sys.argv[1]).resolve()
common = root / "common/src/main/java"
forge_java = root / "forge/src/main/java"


def replace_once(text: str, old: str, new: str, label: str) -> str:
    if old not in text:
        raise SystemExit(f"compat pass 5 anchor missing: {label}")
    return text.replace(old, new, 1)


def replace_java_method(text: str, signature: str, replacement: str, label: str) -> str:
    start = text.find(signature)
    if start < 0:
        raise SystemExit(f"compat pass 5 method anchor missing: {label}")
    brace = text.find("{", start)
    if brace < 0:
        raise SystemExit(f"compat pass 5 opening brace missing: {label}")
    depth = 0
    i = brace
    while i < len(text):
        ch = text[i]
        if ch == "{":
            depth += 1
        elif ch == "}":
            depth -= 1
            if depth == 0:
                return text[:start] + replacement + text[i + 1:]
        i += 1
    raise SystemExit(f"compat pass 5 closing brace missing: {label}")


# Forge 47 owns writes while a RegisterEvent is active. Keep the common/Fabric direct-registry
# entrypoints intact, but add registrar overloads so Forge can route every value through its
# RegisterHelper instead of mutating vanilla registries after Forge has locked them.
blocks_path = common / "net/jewelry/blocks/JewelryBlocks.java"
blocks = blocks_path.read_text()
if "import java.util.function.BiConsumer;" not in blocks:
    blocks = replace_once(
        blocks,
        "import java.util.ArrayList;\n",
        "import java.util.ArrayList;\nimport java.util.function.BiConsumer;\n",
        "JewelryBlocks BiConsumer import",
    )
blocks = replace_java_method(
    blocks,
    "    public static void register()",
    r'''    public static void register() {
        registerBlocks((id, block) -> Registry.register(Registries.BLOCK, id, block));
        registerItems((id, item) -> Registry.register(Registries.ITEM, id, item));
    }

    public static void registerBlocks(BiConsumer<Identifier, Block> registrar) {
        for (var entry : all) {
            registrar.accept(new Identifier(JewelryMod.ID, entry.name), entry.block);
        }
    }

    public static void registerItems(BiConsumer<Identifier, Item> registrar) {
        for (var entry : all) {
            registrar.accept(new Identifier(JewelryMod.ID, entry.name), entry.item());
        }
    }''',
    "JewelryBlocks.register",
)
blocks_path.write_text(blocks)


gems_path = common / "net/jewelry/items/Gems.java"
gems = gems_path.read_text()
if "import java.util.function.BiConsumer;" not in gems:
    gems = replace_once(
        gems,
        "import java.util.ArrayList;\n",
        "import java.util.ArrayList;\nimport java.util.function.BiConsumer;\n",
        "Gems BiConsumer import",
    )
gems = replace_java_method(
    gems,
    "    public static void register()",
    r'''    public static void register() {
        register((id, item) -> Registry.register(Registries.ITEM, id, item));
    }

    public static void register(BiConsumer<Identifier, Item> registrar) {
        for (var entry : all) {
            registrar.accept(entry.id(), entry.item());
        }
    }''',
    "Gems.register",
)
gems_path.write_text(gems)


items_path = common / "net/jewelry/items/JewelryItems.java"
items = items_path.read_text()
if "import java.util.function.BiConsumer;" not in items:
    anchor = "import java.util.*;\n"
    if anchor not in items:
        raise SystemExit("compat pass 5 JewelryItems java.util import anchor missing")
    items = items.replace(anchor, anchor + "import java.util.function.BiConsumer;\n", 1)
items = replace_once(
    items,
    "    public static void register(ItemConfig allConfigs) {\n",
    "    public static void register(ItemConfig allConfigs) {\n"
    "        register(allConfigs, (id, item) -> Registry.register(Registries.ITEM, id, item));\n"
    "    }\n\n"
    "    public static void register(ItemConfig allConfigs, BiConsumer<Identifier, Item> registrar) {\n",
    "JewelryItems registrar overload",
)
items = replace_once(
    items,
    "            Registry.register(Registries.ITEM, entry.id(), item);\n",
    "            registrar.accept(entry.id(), item);\n",
    "JewelryItems registry write",
)
items_path.write_text(items)


sound_path = common / "net/jewelry/util/SoundHelper.java"
sound = sound_path.read_text()
if "import java.util.function.BiConsumer;" not in sound:
    marker = "import net.minecraft.util.Identifier;\n"
    sound = replace_once(
        sound,
        marker,
        marker + "\nimport java.util.function.BiConsumer;\n",
        "SoundHelper BiConsumer import",
    )
sound = replace_java_method(
    sound,
    "    public static void register()",
    r'''    public static void register() {
        JEWELRY_EQUIP_ENTRY = Registry.registerReference(Registries.SOUND_EVENT, JEWELRY_EQUIP_ID, JEWELRY_EQUIP);
        Registry.register(Registries.SOUND_EVENT, JEWELRY_WORKBENCH_ID, JEWELRY_WORKBENCH);
    }

    public static void register(BiConsumer<Identifier, SoundEvent> registrar) {
        // JEWELRY_EQUIP_ENTRY is retained for the vanilla/Fabric path; current Jewelry 2.4.0 does
        // not consume that holder elsewhere. Forge must own the actual registry writes here.
        registrar.accept(JEWELRY_EQUIP_ID, JEWELRY_EQUIP);
        registrar.accept(JEWELRY_WORKBENCH_ID, JEWELRY_WORKBENCH);
    }''',
    "SoundHelper.register",
)
sound_path.write_text(sound)


villagers_path = common / "net/jewelry/village/JewelryVillagers.java"
villagers = villagers_path.read_text()
if "import java.util.function.BiConsumer;" not in villagers:
    villagers = replace_once(
        villagers,
        "import java.util.Set;\n",
        "import java.util.Set;\nimport java.util.function.BiConsumer;\n",
        "JewelryVillagers BiConsumer import",
    )
villagers = replace_java_method(
    villagers,
    "    public static void registerVillagers()",
    r'''    public static void registerVillagers() {
        registerVillagers((id, profession) ->
                Registry.register(Registries.VILLAGER_PROFESSION, id, profession));
    }

    public static void registerVillagers(BiConsumer<Identifier, VillagerProfession> registrar) {
        var workStation = RegistryKey.of(Registries.POINT_OF_INTEREST_TYPE.getKey(), POI_ID);
        JEWELER_PROFESSION = createProfession(JEWELER, workStation);
        registrar.accept(new Identifier(JewelryMod.ID, JEWELER), JEWELER_PROFESSION);
    }''',
    "JewelryVillagers.registerVillagers",
)
villagers_path.write_text(villagers)


forge_mod_path = forge_java / "net/jewelry/forge/ForgeMod.java"
fm = forge_mod_path.read_text()
if "import net.jewelry.util.SoundHelper;" not in fm:
    fm = replace_once(
        fm,
        "import net.jewelry.items.JewelryItems;\n",
        "import net.jewelry.items.JewelryItems;\nimport net.jewelry.util.SoundHelper;\n",
        "ForgeMod SoundHelper import",
    )
# Direct Registry writes are deliberately removed from the Forge adapter. Registries stays because
# the packaged runtime self-test uses it after startup.
fm = fm.replace("import net.minecraft.registry.Registry;\n", "")
fm = replace_java_method(
    fm,
    "    public static void register(RegisterEvent event)",
    r'''    public static void register(RegisterEvent event) {
        event.register(RegistryKeys.SOUND_EVENT, reg ->
                SoundHelper.register((id, sound) -> reg.register(id, sound)));
        event.register(RegistryKeys.BLOCK, reg ->
                JewelryBlocks.registerBlocks((id, block) -> reg.register(id, block)));
        event.register(RegistryKeys.ITEM, reg -> {
            JewelryBlocks.registerItems((id, item) -> reg.register(id, item));
            Gems.register((id, item) -> reg.register(id, item));
            JewelryItems.register(JewelryMod.itemConfig.value, (id, item) -> reg.register(id, item));
            JewelryMod.itemConfig.save();
        });
        event.register(Registries.ITEM_GROUP.getKey(), reg ->
                reg.register(Group.ID, Group.JEWELRY));
        event.register(RegistryKeys.POINT_OF_INTEREST_TYPE, reg ->
                reg.register(JewelryVillagers.POI_ID,
                        new PointOfInterestType(JewelryVillagers.poiBlockStates(),
                                JewelryVillagers.POI_TICKET_COUNT, JewelryVillagers.POI_SEARCH_DISTANCE)));
        event.register(RegistryKeys.VILLAGER_PROFESSION, reg ->
                JewelryVillagers.registerVillagers((id, profession) -> reg.register(id, profession)));
    }''',
    "ForgeMod.register(RegisterEvent)",
)

# Strengthen the already-packaged server self-test: prove that registry families which were part of
# the locked-registry failure are all present, not just a representative item/block/profession.
selftest_anchor = '''        ciAssert(Registries.BLOCK.containsId(new Identifier(JewelryMod.ID, "jewelers_kit")),
                "missing jewelers_kit block");
'''
selftest_extra = '''        ciAssert(Registries.BLOCK.containsId(new Identifier(JewelryMod.ID, "gem_vein")),
                "missing gem_vein block");
        ciAssert(Registries.ITEM.containsId(new Identifier(JewelryMod.ID, "ruby")),
                "missing ruby gem item");
        ciAssert(Registries.ITEM.containsId(new Identifier(JewelryMod.ID, "jewelers_kit")),
                "missing jewelers_kit block item");
        ciAssert(Registries.SOUND_EVENT.containsId(SoundHelper.JEWELRY_EQUIP_ID),
                "missing jewelry_equip sound");
        ciAssert(Registries.SOUND_EVENT.containsId(SoundHelper.JEWELRY_WORKBENCH_ID),
                "missing jewelry_workbench sound");
        ciAssert(Registries.ITEM_GROUP.containsId(Group.ID),
                "missing Jewelry creative tab");
'''
if selftest_extra not in fm:
    fm = replace_once(fm, selftest_anchor, selftest_anchor + selftest_extra, "ForgeMod self-test registry anchors")
forge_mod_path.write_text(fm)


# Current Jewelry 2.4.0 stores its platform biome modifier in the NeoForge module under
# data/jewelry/neoforge/biome_modifier. Forge 47 consumes data/<namespace>/forge/biome_modifier and
# the stock add-features codec id is forge:add_features. Translate only loader identity; current
# biome tag, placed feature and generation step are preserved exactly.
resources = root / "forge/src/main/resources/data/jewelry"
neo = resources / "neoforge"
forge_data = resources / "forge"
if neo.exists():
    if forge_data.exists():
        shutil.rmtree(forge_data)
    neo.rename(forge_data)

modifier = forge_data / "biome_modifier/gem_vein.json"
if not modifier.exists():
    raise SystemExit(f"Jewelry Forge biome modifier missing after platform resource translation: {modifier}")
data = json.loads(modifier.read_text())
if data.get("type") not in ("neoforge:add_features", "forge:add_features"):
    raise SystemExit(f"Unexpected Jewelry biome modifier codec before translation: {data.get('type')}")
data["type"] = "forge:add_features"
modifier.write_text(json.dumps(data, indent=2) + "\n")

expected_biome = {
    "type": "forge:add_features",
    "biomes": "#minecraft:is_overworld",
    "features": "jewelry:gem_vein_placed",
    "step": "underground_ores",
}
final_biome = json.loads(modifier.read_text())
for key, value in expected_biome.items():
    if final_biome.get(key) != value:
        raise SystemExit(f"Jewelry biome modifier lost current 2.4.0 semantic {key}: {final_biome.get(key)!r}")
if neo.exists():
    raise SystemExit("Jewelry compatibility pass 5 left stale NeoForge biome modifier directory")


# Regression guards: Forge's adapter must own all registry writes, while direct vanilla registration
# remains available only behind the common default methods for non-Forge loaders.
final_fm = forge_mod_path.read_text()
register_start = final_fm.index("    public static void register(RegisterEvent event)")
register_end = final_fm.index("    private static void buildTabContents", register_start)
register_body = final_fm[register_start:register_end]
for forbidden in (
    "Registry.register(",
    "JewelryMod.registerSounds()",
    "JewelryMod.registerBlocks()",
    "JewelryMod.registerItems()",
    "JewelryMod.registerVillagers()",
    "catch (Exception e)",
):
    if forbidden in register_body:
        raise SystemExit(f"compat pass 5 left unsafe Forge registration path: {forbidden}")
for required in (
    "SoundHelper.register((id, sound) -> reg.register(id, sound))",
    "JewelryBlocks.registerBlocks((id, block) -> reg.register(id, block))",
    "JewelryBlocks.registerItems((id, item) -> reg.register(id, item))",
    "Gems.register((id, item) -> reg.register(id, item))",
    "JewelryItems.register(JewelryMod.itemConfig.value, (id, item) -> reg.register(id, item))",
    "event.register(Registries.ITEM_GROUP.getKey()",
    "reg.register(JewelryVillagers.POI_ID",
    "JewelryVillagers.registerVillagers((id, profession) -> reg.register(id, profession))",
):
    if required not in register_body:
        raise SystemExit(f"compat pass 5 missing Forge registry-helper invariant: {required}")

for path, required in (
    (blocks_path, "public static void registerBlocks(BiConsumer<Identifier, Block> registrar)"),
    (blocks_path, "public static void registerItems(BiConsumer<Identifier, Item> registrar)"),
    (gems_path, "public static void register(BiConsumer<Identifier, Item> registrar)"),
    (items_path, "public static void register(ItemConfig allConfigs, BiConsumer<Identifier, Item> registrar)"),
    (sound_path, "public static void register(BiConsumer<Identifier, SoundEvent> registrar)"),
    (villagers_path, "public static void registerVillagers(BiConsumer<Identifier, VillagerProfession> registrar)"),
):
    if required not in path.read_text():
        raise SystemExit(f"compat pass 5 missing common registrar bridge in {path.name}: {required}")

print("Jewelry compatibility pass 5 applied: Forge-owned registry helpers + packaged registry assertions + biome modifier parity")
