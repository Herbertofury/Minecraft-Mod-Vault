package net.fabric_extras.ranged_weapon.mixin.item;

import com.google.common.collect.ImmutableMultimap;
import com.google.common.collect.Multimap;
import net.fabric_extras.ranged_weapon.api.*;
import net.fabric_extras.ranged_weapon.internal.RangedItemSettings;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.item.Item;
import net.minecraft.item.RangedWeaponItem;
import net.minecraft.registry.Registries;
import net.minecraft.util.Identifier;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Unique;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;
import java.nio.charset.StandardCharsets;
import java.util.UUID;

@Mixin(RangedWeaponItem.class)
abstract class RangedWeaponItemMixin extends Item implements CustomRangedWeapon {
    @Unique private RangedConfig rwa$config = RangedConfig.EMPTY;
    @Unique private RangedConfig rwa$baseline = RangedConfig.BOW;
    RangedWeaponItemMixin(Settings settings) { super(settings); }

    @Inject(method="<init>", at=@At("TAIL"))
    private void rwa$init(Settings settings, CallbackInfo ci) {
        if ((Object)settings instanceof RangedItemSettings ranged && ranged.getRangedAttributes()!=null) rwa$config=ranged.getRangedAttributes();
    }
    @Override public void setTypeBaseline(RangedConfig config) { rwa$baseline=config; }
    @Override public RangedConfig getTypeBaseline() { return rwa$baseline; }
    @Override public RangedConfig getRangedWeaponConfig() { return rwa$config; }
    @Override public void setRangedWeaponConfig(RangedConfig config) { rwa$config=config == null ? RangedConfig.EMPTY : config; }

    // RangedWeaponItem inherits this method from Item on 1.20.1, so the mixin adds an override
    // to the target class rather than using @Overwrite (which requires a method declared by target).
    public Multimap<EntityAttribute, EntityAttributeModifier> getAttributeModifiers(EquipmentSlot slot) {
        var builder=ImmutableMultimap.<EntityAttribute,EntityAttributeModifier>builder();
        builder.putAll(super.getAttributeModifiers(slot));
        if (slot!=EquipmentSlot.MAINHAND && slot!=EquipmentSlot.OFFHAND) return builder.build();
        if (rwa$config.damage()!=0) builder.put(EntityAttributes_RangedWeapon.DAMAGE.attribute, modifier(AttributeModifierIDs.WEAPON_DAMAGE_ID,"Ranged Weapon Damage",rwa$config.damage(),EntityAttributeModifier.Operation.ADDITION));
        if (rwa$config.pull_time_bonus()!=0) builder.put(EntityAttributes_RangedWeapon.PULL_TIME.attribute, modifier(AttributeModifierIDs.WEAPON_PULL_TIME_ID,"Ranged Weapon Pull Time",rwa$config.pull_time_bonus(),EntityAttributeModifier.Operation.ADDITION));
        if (rwa$config.velocity_bonus()!=0) builder.put(EntityAttributes_RangedWeapon.VELOCITY.attribute, modifier(AttributeModifierIDs.WEAPON_VELOCITY_ID,"Ranged Weapon Velocity",rwa$config.velocity_bonus(),EntityAttributeModifier.Operation.ADDITION));
        if (rwa$config.attributes()!=null) for (var entry:rwa$config.attributes()) {
            if (entry==null || entry.modifier()==null) continue;
            Identifier attrId=Identifier.tryParse(entry.attributeId()); Identifier modifierId=Identifier.tryParse(entry.modifier().modifierId());
            if (attrId==null || modifierId==null || !Registries.ATTRIBUTE.containsId(attrId)) continue;
            builder.put(Registries.ATTRIBUTE.get(attrId), modifier(modifierId,"Ranged Weapon Bonus",entry.modifier().value(),entry.modifier().operation()));
        }
        return builder.build();
    }
    @Unique private static EntityAttributeModifier modifier(Identifier id,String name,double value,EntityAttributeModifier.Operation op) { return new EntityAttributeModifier(uuid(id),name,value,op); }
    @Unique private static UUID uuid(Identifier id) { return UUID.nameUUIDFromBytes(id.toString().getBytes(StandardCharsets.UTF_8)); }
}
