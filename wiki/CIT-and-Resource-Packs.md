# 🪑 CIT & Resource-Pack Conversion

> **Status:** 📋 OmniBridge roadmap.

## CIT → functional content

The target is not merely to preserve appearance. A furniture/item conversion should be able to recreate, when supported or inferable:

- placement and rotation;
- collision;
- recipes and drops;
- sounds and lighting;
- sitting/lying interactions;
- storage/inventory;
- open/close or animated parts;
- connected/variant models;
- localization;
- survival/creative acquisition.

## Mob replacement → standalone entity

A resource pack that visually replaces a cow/zombie/etc. can become an independent entity rather than permanently hijacking the vanilla mob. The converted entity should receive its own ID, spawn rules, model, animations, sounds, particles, attributes, AI, interactions, loot and config where evidence permits.

## Fidelity rule

The UI must distinguish:

- behavior explicitly present in the source;
- behavior borrowed from a vanilla donor/template;
- behavior inferred by OmniBridge;
- user-authored overrides.
