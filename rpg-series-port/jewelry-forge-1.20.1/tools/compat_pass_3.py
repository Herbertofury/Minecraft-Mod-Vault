#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit("usage: compat_pass_3.py <generated-port-root>")

root = Path(sys.argv[1]).resolve()
client = root / "forge/src/main/java/net/jewelry/forge/client"
client.mkdir(parents=True, exist_ok=True)

# NeoForge's client bootstrap mixes a lifecycle event and a gameplay event through its newer event
# registration APIs. Forge 47.4.x has the same two responsibilities, but they belong to different
# buses. Keep that topology explicit so packaged clients work without constructor-time listener magic.
stale = client / "NeoForgeClientMod.java"
if stale.exists():
    stale.unlink()

(client / "ForgeClientMod.java").write_text(r'''package net.jewelry.forge.client;

import net.jewelry.JewelryMod;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLClientSetupEvent;

/** Forge MOD-bus client lifecycle bootstrap. */
@Mod.EventBusSubscriber(modid = JewelryMod.ID, value = Dist.CLIENT, bus = Mod.EventBusSubscriber.Bus.MOD)
public final class ForgeClientMod {
    private ForgeClientMod() { }

    @SubscribeEvent
    public static void onClientSetup(FMLClientSetupEvent event) {
        JewelryMod.init();
    }
}
''')

(client / "ForgeClientEvents.java").write_text(r'''package net.jewelry.forge.client;

import net.jewelry.JewelryMod;
import net.jewelry.client.JewelryModClient;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.event.entity.player.ItemTooltipEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;

/** Forge gameplay-bus client hooks. */
@Mod.EventBusSubscriber(modid = JewelryMod.ID, value = Dist.CLIENT, bus = Mod.EventBusSubscriber.Bus.FORGE)
public final class ForgeClientEvents {
    private ForgeClientEvents() { }

    @SubscribeEvent
    public static void onItemTooltip(ItemTooltipEvent event) {
        JewelryModClient.removeTooltipDuplicates(event.getItemStack(), event.getToolTip());
    }
}
''')

for path in (client / "ForgeClientMod.java", client / "ForgeClientEvents.java"):
    text = path.read_text()
    if "net.neoforged" in text or "NeoForge" in text:
        raise SystemExit(f"compat pass 3 left NeoForge client API in {path.name}")

setup = (client / "ForgeClientMod.java").read_text()
forge_events = (client / "ForgeClientEvents.java").read_text()
for required in (
    "bus = Mod.EventBusSubscriber.Bus.MOD",
    "FMLClientSetupEvent",
    "JewelryMod.init();",
):
    if required not in setup:
        raise SystemExit(f"compat pass 3 missing Forge client lifecycle invariant: {required}")
for required in (
    "bus = Mod.EventBusSubscriber.Bus.FORGE",
    "ItemTooltipEvent",
    "JewelryModClient.removeTooltipDuplicates",
):
    if required not in forge_events:
        raise SystemExit(f"compat pass 3 missing Forge tooltip invariant: {required}")

print("Jewelry compatibility pass 3 applied: native Forge client MOD/FORGE bus split")
