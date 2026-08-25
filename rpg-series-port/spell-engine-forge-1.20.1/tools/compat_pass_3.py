#!/usr/bin/env python3
from pathlib import Path
import json, re, sys

if len(sys.argv) != 2:
    raise SystemExit('usage: compat_pass_3.py <generated-port-root>')
root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'
resources = root / 'common/src/main/resources'

def p(rel): return java / rel
def write(rel, text):
    f=p(rel); f.parent.mkdir(parents=True,exist_ok=True); f.write_text(text)
def patch(rel, fn):
    f=p(rel)
    if f.exists(): f.write_text(fn(f.read_text()))

# Exact Yarn 1.20.1 field layouts for the 1.10.2 fallback-loot inspection feature.
write('net/spell_engine/mixin/loot/CombinedEntryAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.loot.entry.CombinedEntry;
import net.minecraft.loot.entry.LootPoolEntry;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
@Mixin(CombinedEntry.class) public interface CombinedEntryAccessor {
 @Accessor("children") LootPoolEntry[] spellEngine_getChildren();
}
''')
write('net/spell_engine/mixin/loot/LeafEntryAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.loot.entry.LeafEntry;
import net.minecraft.loot.function.LootFunction;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
@Mixin(LeafEntry.class) public interface LeafEntryAccessor {
 @Accessor("weight") int spellEngine_getWeight();
 @Accessor("functions") LootFunction[] spellEngine_getFunctions();
}
''')
write('net/spell_engine/mixin/loot/EnchantWithLevelsLootFunctionAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.loot.function.EnchantWithLevelsLootFunction;
import net.minecraft.loot.provider.number.LootNumberProvider;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
@Mixin(EnchantWithLevelsLootFunction.class) public interface EnchantWithLevelsLootFunctionAccessor {
 @Accessor("range") LootNumberProvider spellEngine_getLevels();
}
''')
write('net/spell_engine/mixin/loot/ItemEntryAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.item.Item;
import net.minecraft.loot.entry.ItemEntry;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
@Mixin(ItemEntry.class) public interface ItemEntryAccessor {
 @Accessor("item") Item spellEngine_getItem();
}
''')
write('net/spell_engine/mixin/loot/LootTableBuilderAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.loot.LootPool;
import net.minecraft.loot.LootTable;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
import java.util.List;
@Mixin(LootTable.Builder.class) public interface LootTableBuilderAccessor {
 @Accessor("pools") List<LootPool> spellEngine_getPools();
}
''')
patch('net/spell_engine/rpg_series/loot/LootHelper.java', lambda s: s.replace('spellEngine_getItem().value()', 'spellEngine_getItem()'))

# ItemStack's 1.21 private attribute-tooltip helper does not exist in 1.20.1. Mirror its public-facing
# translation semantics locally, using the exact 1.20.1 operation IDs and attribute translation key.
def equipment_tooltip(s):
    s=s.replace('import net.spell_engine.mixin.client.ItemStackTooltipAccessor;\n','')
    s=s.replace('import java.util.ArrayList;','import java.text.DecimalFormat;\nimport java.util.ArrayList;')
    old='''            var tooltipUtil = (ItemStackTooltipAccessor) (Object) ItemStack.EMPTY;\n            for (var modifier: bonus.attributes().modifiers()) {\n                tooltipUtil\n                        .spellEngine_appendAttributeModifierTooltip(\n                                bonusLines::add,\n                                player,\n                                modifier.attribute(),\n                                modifier.modifier()\n                        );\n            }'''
    new='''            for (var modifier: bonus.attributes().modifiers()) {\n                bonusLines.add(attributeLine(modifier.attribute().value(), modifier.modifier()));\n            }'''
    s=s.replace(old,new)
    marker='''    public static List<Text> bonusText(PlayerEntity player, ItemStack itemStack, EquipmentSet.Bonus bonus, boolean isActive) {'''
    helper='''    private static final DecimalFormat ATTRIBUTE_FORMAT = new DecimalFormat("#.##");\n    private static Text attributeLine(net.minecraft.entity.attribute.EntityAttribute attribute, net.minecraft.entity.attribute.EntityAttributeModifier modifier) {\n        var amount = modifier.getValue();\n        var display = modifier.getOperation() == net.minecraft.entity.attribute.EntityAttributeModifier.Operation.ADDITION ? amount : amount * 100.0D;\n        var key = amount >= 0 ? "attribute.modifier.plus." : "attribute.modifier.take.";\n        var color = amount >= 0 ? Formatting.BLUE : Formatting.RED;\n        return Text.translatable(key + modifier.getOperation().getId(), ATTRIBUTE_FORMAT.format(Math.abs(display)), Text.translatable(attribute.getTranslationKey())).formatted(color);\n    }\n\n'''
    return s.replace(marker, helper+marker)
patch('net/spell_engine/api/item/set/EquipmentSetTooltip.java', equipment_tooltip)
write('net/spell_engine/mixin/client/ItemStackTooltipAccessor.java','package net.spell_engine.mixin.client; public interface ItemStackTooltipAccessor { }\n')

# 1.20.1 CrossbowItem owns loadProjectile/shootAll (they are not RangedWeaponItem methods yet).
write('net/spell_engine/mixin/item/RangedWeaponAccessor.java', r'''package net.spell_engine.mixin.item;
import net.minecraft.entity.LivingEntity;
import net.minecraft.item.CrossbowItem;
import net.minecraft.item.ItemStack;
import net.minecraft.util.Hand;
import net.minecraft.world.World;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Invoker;
@Mixin(CrossbowItem.class) public interface RangedWeaponAccessor {
 @Invoker("loadProjectile") static boolean loadProjectile_SpellEngine(LivingEntity shooter, ItemStack crossbow, ItemStack projectile, boolean simulated, boolean creative) { throw new AssertionError(); }
 @Invoker("shootAll") static void shootAll_SpellEngine(World world, LivingEntity shooter, Hand hand, ItemStack stack, float speed, float divergence) { throw new AssertionError(); }
}
''')
write('net/spell_engine/internals/delivery/arrow/ArrowShotCompat.java', r'''package net.spell_engine.internals.delivery.arrow;
import com.google.common.base.Suppliers;
import net.minecraft.entity.Entity;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.entity.projectile.PersistentProjectileEntity;
import net.spell_engine.Platform;
import net.spell_engine.internals.SpellTriggers;
import net.spell_engine.internals.casting.SpellCaster;
public final class ArrowShotCompat {
 private ArrowShotCompat(){}
 public static void apply(Entity entity, LivingEntity shooter) {
  if (!(shooter instanceof PlayerEntity player) || !(entity instanceof ArrowExtension arrow)) return;
  var caster=(SpellCaster.Player)player; var ctx=caster.getArrowShootContext();
  if (ctx.critical && entity instanceof PersistentProjectileEntity projectile) projectile.setCritical(true);
  SpellTriggers.onArrowShot(arrow,player,ctx.firedBySpell);
  var trackers=Suppliers.memoize(() -> Platform.tracking(shooter));
  for(var spell:ctx.activeSpells) ArrowHelper.onArrowShot(arrow,shooter,spell,trackers);
  caster.setArrowShootContext(ArrowShootContext.empty());
 }
}
''')
patch('net/spell_engine/internals/delivery/arrow/ArrowShootContext.java', lambda s: s.replace('public boolean firedBySpell = false;', 'public boolean firedBySpell = false;\n    public boolean critical = false;'))

def arrow_helper(s):
    s=s.replace('import net.minecraft.server.world.ServerWorld;\n','')
    s=s.replace('import org.jetbrains.annotations.Nullable;','import org.jetbrains.annotations.Nullable;\nimport net.minecraft.enchantment.EnchantmentHelper;\nimport net.minecraft.enchantment.Enchantments;')
    s=s.replace('&& (world instanceof ServerWorld serverWorld)\n                && (weapon instanceof RangedWeaponItem rangedWeapon)', '&& !world.isClient')
    old='''            var loadedAmmo = RangedWeaponAccessor.load_SpellEngine(weaponStack, ammo, shooter);\n            if (loadedAmmo.isEmpty()) {\n                return;\n            }'''
    new='''            boolean bypassConsumption = !shoot_arrow.consume_arrow;\n            if (shooter instanceof PlayerEntity player) {\n                bypassConsumption |= player.getAbilities().creativeMode || EnchantmentHelper.getLevel(Enchantments.INFINITY, weaponStack) > 0;\n                if (!bypassConsumption) {\n                    boolean inventoryReference = false;\n                    for (int i = 0; i < player.getInventory().size(); i++) { if (player.getInventory().getStack(i) == ammo) { inventoryReference = true; break; } }\n                    if (!inventoryReference) {\n                        var predicate = new net.spell_engine.internals.cost.Ammo.Searched(null, ammo.getItem()).asPredicate();\n                        var source = net.spell_engine.internals.cost.Ammo.findContainer(player, predicate, 1);\n                        if (source != null && net.spell_engine.internals.cost.Ammo.takeFromContainer(source.itemStack(), predicate, 1) == 1) bypassConsumption = true;\n                    }\n                }\n            }\n            if (!RangedWeaponAccessor.loadProjectile_SpellEngine(shooter, weaponStack, ammo, false, bypassConsumption)) return;'''
    s=s.replace(old,new)
    s=s.replace('shotContext.firedBySpell = true;\n                shotContext.activeSpells.add(spellEntry);', 'shotContext.firedBySpell = true;\n                shotContext.critical = shoot_arrow.arrow_critical_strike;\n                shotContext.activeSpells.add(spellEntry);')
    pattern=re.compile(r'''            \(\(RangedWeaponAccessor\) rangedWeapon\)\.shootAll_SpellEngine\(\n                    serverWorld,\n                    shooter,\n                    Hand\.MAIN_HAND,\n                    weaponStack,\n                    loadedAmmo,\n                    shoot_arrow\.launch_properties\.velocity,\n                    divergence,\n                    shoot_arrow\.arrow_critical_strike,\n                    null\);''')
    s=pattern.sub('''            RangedWeaponAccessor.shootAll_SpellEngine(world, shooter, Hand.MAIN_HAND, weaponStack,\n                    shoot_arrow.launch_properties.velocity, divergence);''',s)
    s=s.replace('''\n            if (shooter instanceof SpellCaster.Player caster) {\n                caster.setArrowShootContext(ArrowShootContext.empty());\n            }\n''','\n')
    s=re.sub(r'''\n            // Fixing inconsistent Vanille code, shoot sound is played by `BOW` outside of `shootAll`\n            if \(weapon instanceof BowItem\) \{.*?\n            \}\n''','\n',s,flags=re.S)
    return s
patch('net/spell_engine/internals/delivery/arrow/ArrowHelper.java', arrow_helper)

write('net/spell_engine/mixin/arrow/BowItemSpellMixin.java', r'''package net.spell_engine.mixin.arrow;
import com.llamalad7.mixinextras.injector.wrapoperation.Operation;
import com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;
import net.minecraft.entity.Entity;import net.minecraft.entity.LivingEntity;import net.minecraft.item.BowItem;import net.minecraft.item.ItemStack;import net.minecraft.world.World;import net.spell_engine.internals.delivery.arrow.ArrowShotCompat;import org.spongepowered.asm.mixin.Mixin;import org.spongepowered.asm.mixin.injection.At;
@Mixin(BowItem.class) public class BowItemSpellMixin {
 @WrapOperation(method="onStoppedUsing",at=@At(value="INVOKE",target="Lnet/minecraft/world/World;spawnEntity(Lnet/minecraft/entity/Entity;)Z"))
 private boolean spellEngine$spawn(World instance, Entity entity, Operation<Boolean> original, ItemStack stack, World world, LivingEntity shooter, int remaining) { ArrowShotCompat.apply(entity,shooter); return original.call(instance,entity); }
}
''')
write('net/spell_engine/mixin/arrow/CrossbowItemSpellMixin.java', r'''package net.spell_engine.mixin.arrow;
import com.llamalad7.mixinextras.injector.wrapoperation.Operation;
import com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;
import net.minecraft.entity.Entity;import net.minecraft.entity.LivingEntity;import net.minecraft.item.CrossbowItem;import net.minecraft.item.ItemStack;import net.minecraft.util.Hand;import net.minecraft.world.World;import net.spell_engine.internals.delivery.arrow.ArrowShotCompat;import org.spongepowered.asm.mixin.Mixin;import org.spongepowered.asm.mixin.injection.At;
@Mixin(CrossbowItem.class) public class CrossbowItemSpellMixin {
 @WrapOperation(method="shoot",at=@At(value="INVOKE",target="Lnet/minecraft/world/World;spawnEntity(Lnet/minecraft/entity/Entity;)Z"))
 private static boolean spellEngine$spawn(World instance, Entity entity, Operation<Boolean> original, World world, LivingEntity shooter, Hand hand, ItemStack crossbow, ItemStack projectile, float soundPitch, boolean creative, float speed, float divergence, float simulated) { ArrowShotCompat.apply(entity,shooter); return original.call(instance,entity); }
}
''')
write('net/spell_engine/mixin/arrow/RangedWeaponItemMixin.java','package net.spell_engine.mixin.arrow; public class RangedWeaponItemMixin { }\n')

write('net/spell_engine/mixin/arrow/RangedWeaponQuiverMixin.java', r'''package net.spell_engine.mixin.arrow;
import com.llamalad7.mixinextras.injector.wrapoperation.Operation;import com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;import net.minecraft.entity.LivingEntity;import net.minecraft.entity.player.PlayerEntity;import net.minecraft.item.CrossbowItem;import net.minecraft.item.ItemStack;import net.spell_engine.internals.cost.Ammo;import org.spongepowered.asm.mixin.Mixin;import org.spongepowered.asm.mixin.injection.At;
@Mixin(CrossbowItem.class) public class RangedWeaponQuiverMixin {
 @WrapOperation(method="loadProjectile",at=@At(value="INVOKE",target="Lnet/minecraft/item/ItemStack;split(I)Lnet/minecraft/item/ItemStack;"))
 private static ItemStack spellEngine$split(ItemStack projectile,int amount,Operation<ItemStack> original,LivingEntity shooter,ItemStack crossbow,ItemStack projectileArg,boolean simulated,boolean creative){
  if(!simulated&&!creative&&shooter instanceof PlayerEntity player){boolean inventory=false;for(int i=0;i<player.getInventory().size();i++)if(player.getInventory().getStack(i)==projectile){inventory=true;break;}if(!inventory){var predicate=new Ammo.Searched(null,projectile.getItem()).asPredicate();var source=Ammo.findContainer(player,predicate,amount);if(source!=null&&Ammo.takeFromContainer(source.itemStack(),predicate,amount)==amount){var copy=projectile.copy();copy.setCount(amount);return copy;}}}return original.call(projectile,amount);}
}
''')
write('net/spell_engine/mixin/arrow/BowQuiverMixin.java', r'''package net.spell_engine.mixin.arrow;
import com.llamalad7.mixinextras.injector.wrapoperation.Operation;import com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;import net.minecraft.entity.LivingEntity;import net.minecraft.entity.player.PlayerEntity;import net.minecraft.item.BowItem;import net.minecraft.item.ItemStack;import net.minecraft.world.World;import net.spell_engine.internals.cost.Ammo;import org.spongepowered.asm.mixin.Mixin;import org.spongepowered.asm.mixin.injection.At;
@Mixin(BowItem.class) public class BowQuiverMixin {
 @WrapOperation(method="onStoppedUsing",at=@At(value="INVOKE",target="Lnet/minecraft/item/ItemStack;decrement(I)V"))
 private void spellEngine$decrement(ItemStack projectile,int amount,Operation<Void> original,ItemStack bow,World world,LivingEntity shooter,int remaining){if(shooter instanceof PlayerEntity player){boolean inventory=false;for(int i=0;i<player.getInventory().size();i++)if(player.getInventory().getStack(i)==projectile){inventory=true;break;}if(!inventory){var predicate=new Ammo.Searched(null,projectile.getItem()).asPredicate();var source=Ammo.findContainer(player,predicate,amount);if(source!=null&&Ammo.takeFromContainer(source.itemStack(),predicate,amount)==amount)return;}}original.call(projectile,amount);}
}
''')

def particle(s):
    start=s.find('    /// -90 and not +90:')
    end=s.find('    // MARK: Factory',start)
    if start<0 or end<0: raise SystemExit('SpellParticle rotator block not found')
    block=r'''    private Quaternionf resolvedRotation(Camera camera) {
        return switch (facing) {
            case CAMERA -> new Quaternionf(camera.getRotation());
            case UPRIGHT -> { var q = camera.getRotation(); yield new Quaternionf(0F, q.y, 0F, q.w).normalize(); }
            case GROUND -> new Quaternionf().rotationX((float)Math.toRadians(-90));
            case VELOCITY -> { var d = new Vector3f((float)velocityX,(float)velocityY,(float)velocityZ); if(d.lengthSquared()<1.0E-6F) yield new Quaternionf(camera.getRotation()); d.normalize(); yield new Quaternionf().rotationTo(0F,1F,0F,d.x,d.y,d.z); }
        };
    }
    @Override public void buildGeometry(VertexConsumer vc, Camera camera, float tickDelta) {
        if (skipRender) return;
        var cam=camera.getPos();
        float px=(float)(MathHelper.lerp((double)tickDelta,prevPosX,x)-cam.x);
        float py=(float)(MathHelper.lerp((double)tickDelta,prevPosY,y)-cam.y)+pivot*getSize(tickDelta);
        float pz=(float)(MathHelper.lerp((double)tickDelta,prevPosZ,z)-cam.z);
        var q=resolvedRotation(camera); if(angle!=0F) q.rotateZ(MathHelper.lerp(tickDelta,prevAngle,angle));
        drawQuad(vc,q,px,py,pz,tickDelta);
        if(facing==ParticleGroup.Facing.GROUND) drawQuad(vc,new Quaternionf(q).rotateX((float)Math.PI),px,py,pz,tickDelta);
    }
    private void drawQuad(VertexConsumer vc,Quaternionf q,float x,float y,float z,float tickDelta){
        float size=getSize(tickDelta); var c=new Vector3f[]{new Vector3f(-1F,-1F,0F),new Vector3f(-1F,1F,0F),new Vector3f(1F,1F,0F),new Vector3f(1F,-1F,0F)};
        for(var v:c)v.rotate(q).mul(size).add(x,y,z); float minU=getMinU(),maxU=getMaxU(),minV=getMinV(),maxV=getMaxV(); int light=getBrightness(tickDelta);
        vc.vertex(c[0].x(),c[0].y(),c[0].z()).texture(maxU,maxV).color(red,green,blue,alpha).light(light).next();
        vc.vertex(c[1].x(),c[1].y(),c[1].z()).texture(maxU,minV).color(red,green,blue,alpha).light(light).next();
        vc.vertex(c[2].x(),c[2].y(),c[2].z()).texture(minU,minV).color(red,green,blue,alpha).light(light).next();
        vc.vertex(c[3].x(),c[3].y(),c[3].z()).texture(minU,maxV).color(red,green,blue,alpha).light(light).next();
    }

'''
    return s[:start]+block+s[end:]
patch('net/spell_engine/client/particle/SpellParticle.java', particle)

mix=resources/'spell_engine.mixins.json'; data=json.loads(mix.read_text())
remove={'client.ItemStackTooltipAccessor','arrow.RangedWeaponItemMixin'}
data['mixins']=[x for x in data.get('mixins',[]) if x not in remove]
for x in ['arrow.BowItemSpellMixin','arrow.CrossbowItemSpellMixin','arrow.BowQuiverMixin']:
    if x not in data['mixins']: data['mixins'].append(x)
mix.write_text(json.dumps(data,indent=2)+'\n')

for needle in ('@Accessor("levels")','RegistryEntry<Item> spellEngine_getItem','ItemStackTooltipAccessor) (Object)','@Mixin(RangedWeaponItem.class)','private static final Rotator','public Rotator getRotator'):
    hits=[str(f.relative_to(java)) for f in java.rglob('*.java') if needle in f.read_text()]
    if hits: raise SystemExit(f'pass3 incomplete: {needle}: {hits[:10]}')
print('Spell Engine compatibility pass 3 applied: exact 1.20.1 loot internals + tooltip + ranged launch/quiver hooks + particle facing')
