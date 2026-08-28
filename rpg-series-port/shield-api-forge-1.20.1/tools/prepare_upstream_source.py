#!/usr/bin/env python3
from pathlib import Path
import json, shutil, sys

if len(sys.argv) != 4:
    raise SystemExit('usage: prepare_upstream_source.py <upstream-1.20.1> <upstream-2.1.0> <common-dir>')
old=Path(sys.argv[1]); modern=Path(sys.argv[2]); common=Path(sys.argv[3])
gj=common/'src/generatedUpstream/java'; gr=common/'src/generatedUpstream/resources'
for p in (gj,gr):
    shutil.rmtree(p,ignore_errors=True); p.mkdir(parents=True,exist_ok=True)

# 2.1.0 is behavior/content authority. Copy its complete common source/resources first.
shutil.copytree(modern/'common/src/main/java', gj, dirs_exist_ok=True)
shutil.copytree(modern/'common/src/main/resources', gr, dirs_exist_ok=True)

# 1.20.1 translation of the 2.1.0 public CustomShieldItem contract.
custom=gj/'net/fabric_extras/shield_api/item/CustomShieldItem.java'
custom.write_text('''package net.fabric_extras.shield_api.item;

import com.google.common.collect.ImmutableMultimap;
import com.google.common.collect.Multimap;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.item.ItemStack;
import net.minecraft.item.ShieldItem;
import net.minecraft.recipe.Ingredient;
import net.minecraft.sound.SoundEvent;
import net.minecraft.util.Pair;
import org.jetbrains.annotations.Nullable;

import java.util.HashSet;
import java.util.List;
import java.util.function.Supplier;

public class CustomShieldItem extends ShieldItem {
    private Multimap<EntityAttribute, EntityAttributeModifier> attributeModifiers;
    public static final HashSet<CustomShieldItem> instances = new HashSet<>();
    @Nullable private final SoundEvent equipSound;
    private final Supplier<Ingredient> repairIngredientSupplier;

    public CustomShieldItem(@Nullable SoundEvent equipSound,
                            Supplier<Ingredient> repairIngredientSupplier,
                            List<Pair<EntityAttribute, EntityAttributeModifier>> attributeModifierList,
                            Settings settings) {
        super(settings);
        this.attributeModifiers = buildModifiers(attributeModifierList);
        this.equipSound = equipSound;
        this.repairIngredientSupplier = repairIngredientSupplier;
        instances.add(this);
    }

    @Override
    public boolean canRepair(ItemStack stack, ItemStack ingredient) {
        return this.repairIngredientSupplier.get().test(ingredient) || super.canRepair(stack, ingredient);
    }

    public void setAttributeModifiers(List<Pair<EntityAttribute, EntityAttributeModifier>> attributeModifierList) {
        this.attributeModifiers = buildModifiers(attributeModifierList);
    }

    @Override
    public Multimap<EntityAttribute, EntityAttributeModifier> getAttributeModifiers(EquipmentSlot slot) {
        return slot == EquipmentSlot.OFFHAND ? this.attributeModifiers : super.getAttributeModifiers(slot);
    }

    protected Multimap<EntityAttribute, EntityAttributeModifier> buildModifiers(
            List<Pair<EntityAttribute, EntityAttributeModifier>> attributeModifierList) {
        ImmutableMultimap.Builder<EntityAttribute, EntityAttributeModifier> builder = ImmutableMultimap.builder();
        for (Pair<EntityAttribute, EntityAttributeModifier> pair : attributeModifierList) {
            builder.put(pair.getLeft(), pair.getRight());
        }
        return builder.build();
    }

    @Override
    public @Nullable SoundEvent getEquipSound() {
        return this.equipSound != null ? this.equipSound : super.getEquipSound();
    }
}
''', encoding='utf-8')

# Forge 47.4.x already widens vanilla Player#damageShield/hurtCurrentlyUsedShield from
# Items.SHIELD to any ItemStack that can perform ToolActions.SHIELD_BLOCK. ShieldItem supplies
# DEFAULT_SHIELD_ACTIONS, so every CustomShieldItem naturally receives the vanilla stat,
# durability, break callback, hand clearing, and break sound path. Keeping the Fabric-origin
# damageShield mixin on Forge would double-award ITEM_USED and double-damage the shield.
# Preserve only the current Shield API behavior Forge does not provide itself: unconditional
# 100-tick cooldown application to every CustomShieldItem when a shield is disabled.
player=gj/'net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.java'
player.write_text('''package net.fabric_extras.shield_api.mixin.entity.player;

import net.fabric_extras.shield_api.item.CustomShieldItem;
import net.minecraft.entity.EntityType;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.player.ItemCooldownManager;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.world.World;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Shadow;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;

@Mixin(PlayerEntity.class)
public abstract class PlayerEntityMixin extends LivingEntity {
    @Shadow public abstract ItemCooldownManager getItemCooldownManager();

    protected PlayerEntityMixin(EntityType<? extends LivingEntity> type, World world) { super(type, world); }

    @Inject(method = "disableShield", at = @At("HEAD"))
    public void shield_api$disableShield(boolean sprinting, CallbackInfo ci) {
        for (CustomShieldItem customShieldItem : CustomShieldItem.instances) {
            this.getItemCooldownManager().set(customShieldItem, 100);
        }
    }
}
''', encoding='utf-8')

client=gj/'net/fabric_extras/shield_api/client/ShieldAPIClient.java'
text=client.read_text(encoding='utf-8').replace('Identifier.of("blocking")', 'new Identifier("blocking")')
client.write_text(text, encoding='utf-8')

# AxeItem.shouldCancelStripAttempt was introduced after 1.20.1. In 1.20.1 even the
# vanilla shield has no equivalent strip-cancel hook, so retaining the 1.21 mixin would
# be both an invalid target and behaviorally wrong for the target vanilla baseline.
axe=gj/'net/fabric_extras/shield_api/mixin/item/AxeItemMixin.java'
axe.unlink(missing_ok=True)

# The modern Architectury mixin config assumes Java 21 and places a MinecraftClient
# target in the common mixin list. Translate those loader/runtime details without
# changing client behavior: Java 17 target, and client-only mixins stay client-only.
mixins=gr/'shield_api.mixins.json'
config=json.loads(mixins.read_text(encoding='utf-8'))
config['compatibilityLevel']='JAVA_17'
common_mixins=[m for m in config.get('mixins',[]) if m not in ('item.AxeItemMixin','client.MinecraftClientMixin')]
client_mixins=list(config.get('client',[]))
if 'client.MinecraftClientMixin' not in client_mixins:
    client_mixins.insert(0,'client.MinecraftClientMixin')
config['mixins']=common_mixins
config['client']=client_mixins
mixins.write_text(json.dumps(config,indent=2)+"\n",encoding='utf-8')

# No loader metadata from upstream is allowed to bleed into the common ABI.
for name in ('fabric.mod.json','neoforge.mods.toml'):
    (gr/name).unlink(missing_ok=True)

# Assert target-native ownership and current Shield API behavior directly.
pt=player.read_text(encoding='utf-8')
assert 'shield_api$damageShield' not in pt
assert 'Stats.USED' not in pt and 'activeItemStack' not in pt
assert '@Inject(method = "disableShield", at = @At("HEAD"))' in pt
assert 'set(customShieldItem, 100)' in pt
assert 'EnchantmentHelper' not in pt and 'BREAK_SHIELD' not in pt and 'nextFloat() <' not in pt
assert '->' not in pt and '::' not in pt
ct=custom.read_text(encoding='utf-8')
assert 'extends ShieldItem' in ct
assert 'EquipmentSlot.OFFHAND' in ct and 'setAttributeModifiers' in ct and 'instances.add(this)' in ct
assert not axe.exists()
mc=json.loads(mixins.read_text(encoding='utf-8'))
assert mc['compatibilityLevel']=='JAVA_17'
assert 'item.AxeItemMixin' not in mc.get('mixins',[])
assert 'client.MinecraftClientMixin' not in mc.get('mixins',[])
assert 'client.MinecraftClientMixin' in mc.get('client',[])

count=0
for p in gr.rglob('*.json'):
    with p.open('r',encoding='utf-8') as fh: json.load(fh)
    count += 1
print(f'[Shield API] exact 2.1.0 contract materialized on Forge 1.20.1: native SHIELD_BLOCK owns stat/durability once; 100t all-custom cooldown preserved; validated {count} JSON resources; post-1.20.1 Axe hook omitted; client mixins dist-safe')
