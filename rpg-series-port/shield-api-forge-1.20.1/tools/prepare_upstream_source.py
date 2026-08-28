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

# Mixin classes are merged into their target and are deliberately invalid for direct runtime
# classloading. Keep the 1.20.1 ItemStack.damage callback in an ordinary runtime class so javac
# does not emit an invokedynamic bootstrap handle owned by PlayerEntityMixin.
callback=gj/'net/fabric_extras/shield_api/internal/ShieldDamageCallbacks.java'
callback.parent.mkdir(parents=True, exist_ok=True)
callback.write_text('''package net.fabric_extras.shield_api.internal;

import net.minecraft.entity.LivingEntity;
import net.minecraft.util.Hand;

import java.util.function.Consumer;

public final class ShieldDamageCallbacks {
    private ShieldDamageCallbacks() { }

    public static Consumer<LivingEntity> sendToolBreakStatus(Hand hand) {
        return entity -> entity.sendToolBreakStatus(hand);
    }
}
''', encoding='utf-8')

# Descriptor-only backport of PlayerEntity hooks. Behavioral branches remain exactly 2.1.0:
# server-side USED stat; >=3 durability path; correct-hand break handling; unconditional 100t cooldown.
player=gj/'net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.java'
player.write_text('''package net.fabric_extras.shield_api.mixin.entity.player;

import net.fabric_extras.shield_api.internal.ShieldDamageCallbacks;
import net.fabric_extras.shield_api.item.CustomShieldItem;
import net.minecraft.entity.EntityType;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.player.ItemCooldownManager;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.item.ItemStack;
import net.minecraft.sound.SoundEvents;
import net.minecraft.stat.Stat;
import net.minecraft.stat.Stats;
import net.minecraft.util.Hand;
import net.minecraft.util.math.MathHelper;
import net.minecraft.world.World;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Shadow;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;

@Mixin(PlayerEntity.class)
public abstract class PlayerEntityMixin extends LivingEntity {
    @Shadow public abstract void incrementStat(Stat<?> stat);
    @Shadow public abstract ItemCooldownManager getItemCooldownManager();

    protected PlayerEntityMixin(EntityType<? extends LivingEntity> type, World world) { super(type, world); }

    @Inject(method = "damageShield", at = @At("HEAD"))
    protected void shield_api$damageShield(float amount, CallbackInfo ci) {
        if (this.activeItemStack.getItem() instanceof CustomShieldItem customShieldItem) {
            if (!this.getWorld().isClient) this.incrementStat(Stats.USED.getOrCreateStat(customShieldItem));
            if (amount >= 3.0F) {
                int damage = 1 + MathHelper.floor(amount);
                Hand hand = this.getActiveHand();
                this.activeItemStack.damage(damage, this, ShieldDamageCallbacks.sendToolBreakStatus(hand));
                if (this.activeItemStack.isEmpty()) {
                    this.equipStack(hand == Hand.MAIN_HAND ? EquipmentSlot.MAINHAND : EquipmentSlot.OFFHAND, ItemStack.EMPTY);
                    this.activeItemStack = ItemStack.EMPTY;
                    this.playSound(SoundEvents.ITEM_SHIELD_BREAK, 0.8F,
                            0.8F + this.getWorld().random.nextFloat() * 0.4F);
                }
            }
        }
    }

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

# Assert anti-regression behavior and the ModLauncher-safe mixin bytecode shape directly.
pt=player.read_text(encoding='utf-8')
assert '@At("HEAD")' in pt and 'set(customShieldItem, 100)' in pt
assert 'EnchantmentHelper' not in pt and 'BREAK_SHIELD' not in pt and 'nextFloat() <' not in pt
assert '->' not in pt and '::' not in pt
assert 'ShieldDamageCallbacks.sendToolBreakStatus(hand)' in pt
assert callback.exists() and 'sendToolBreakStatus(hand)' in callback.read_text(encoding='utf-8')
ct=custom.read_text(encoding='utf-8')
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
print(f'[Shield API] exact 2.1.0 current behavior materialized on 1.20.1 signatures; validated {count} JSON resources; mixin lambda moved to runtime helper; post-1.20.1 Axe hook omitted; client mixins dist-safe')
