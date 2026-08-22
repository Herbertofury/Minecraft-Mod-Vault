# 🤖 Agent Test Driver

> **Status:** 📋 0.11.0 roadmap.

The Agent Test Driver gives an authorized test agent the practical action and observation surface of a real Minecraft player inside controlled test environments.

## Player actions

- move, strafe, sprint, sneak, jump, swim, climb, crawl and fly;
- camera/yaw/pitch/look-at;
- break/place/use/interact;
- combat and damage/death/respawn;
- inventory/equipment/crafting/smelting/trading;
- ride boats, minecarts, horses and supported modded vehicles;
- chat/commands;
- keybinds, held keys, mouse buttons and scrolling.

## GUI automation

The driver should manipulate real Minecraft/mod screens: buttons, tabs, slots, text fields, sliders, toggles, dropdowns, scroll views, keybind controls and custom-rendered UIs (with screenshot/vision + real input fallback where semantic widgets are unavailable).

## Telemetry

Expose player/world state, nearby entities/blocks, screen hierarchy, logs, crashes, screenshots/video and performance counters where supported.

## Safety scope

Full automation is for local/disposable/explicitly authorized environments. It is **not** an anti-cheat bypass system for unrelated public servers.
