#!/usr/bin/env python3
from pathlib import Path
import json, re, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_3.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
base = Path(sys.argv[2]).resolve()
java = root / 'common/src/main/java'
resources = root / 'common/src/main/resources'

def path(rel): return java / rel
def write(rel, text):
    f = path(rel); f.parent.mkdir(parents=True, exist_ok=True); f.write_text(text)
def patch(rel, fn):
    f = path(rel)
    if f.exists(): f.write_text(fn(f.read_text()))

# Target-native 1.20.1 loot field descriptors/names. Mixin accessors must match both field name and descriptor.
write('net/spell_engine/mixin/loot/CombinedEntryAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.loot.entry.CombinedEntry;
import net.minecraft.loot.entry.LootPoolEntry;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
@Mixin(CombinedEntry.class)
public interface CombinedEntryAccessor {
    @Accessor("children") LootPoolEntry[] spellEngine_getChildren();
}
''')
write('net/spell_engine/mixin/loot/LeafEntryAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.loot.entry.LeafEntry;
import net.minecraft.loot.function.LootFunction;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
@Mixin(LeafEntry.class)
public interface LeafEntryAccessor {
    @Accessor("weight") int spellEngine_getWeight();
    @Accessor("functions") LootFunction[] spellEngine_getFunctions();
}
''')
write('net/spell_engine/mixin/loot/EnchantWithLevelsLootFunctionAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.loot.function.EnchantWithLevelsLootFunction;
import net.minecraft.loot.provider.number.LootNumberProvider;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
@Mixin(EnchantWithLevelsLootFunction.class)
public interface EnchantWithLevelsLootFunctionAccessor {
    @Accessor("range") LootNumberProvider spellEngine_getLevels();
}
''')
write('net/spell_engine/mixin/loot/LootTableBuilderAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.loot.LootPool;
import net.minecraft.loot.LootTable;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
import java.util.List;
@Mixin(LootTable.Builder.class)
public interface LootTableBuilderAccessor {
    @Accessor("pools") List<LootPool> spellEngine_getPools();
}
''')
write('net/spell_engine/mixin/loot/ItemEntryAccessor.java', r'''package net.spell_engine.mixin.loot;
import net.minecraft.item.Item;
import net.minecraft.loot.entry.ItemEntry;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;
@Mixin(ItemEntry.class)
public interface ItemEntryAccessor {
    @Accessor("item") Item spellEngine_getItem();
}
''')
patch('net/spell_engine/rpg_series/loot/LootHelper.java', lambda s: s.replace('.spellEngine_getItem().value()', '.spellEngine_getItem()'))

# 1.20.1 ItemStack has no appendAttributeModifierTooltip invoker. Equipment-set bonuses only need
# normal custom-modifier formatting, so reproduce the vanilla plus/take operation formatting directly.
write('net/spell_engine/compat/item/AttributeTooltipCompat.java', r'''package net.spell_engine.compat.item;
import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.entity.attribute.EntityAttributes;
import net.minecraft.registry.entry.RegistryEntry;
import net.minecraft.text.Text;
import net.minecraft.util.Formatting;
import java.text.DecimalFormat;
import java.text.DecimalFormatSymbols;
import java.util.Locale;
import java.util.function.Consumer;
public final class AttributeTooltipCompat {
    private static final DecimalFormat FORMAT = new DecimalFormat("#.##", DecimalFormatSymbols.getInstance(Locale.ROOT));
    private AttributeTooltipCompat() { }
    public static void append(Consumer<Text> out, RegistryEntry<EntityAttribute> entry, EntityAttributeModifier modifier) {
        var attribute = entry.value();
        double amount = modifier.getValue();
        if (modifier.getOperation() == EntityAttributeModifier.Operation.MULTIPLY_BASE
                || modifier.getOperation() == EntityAttributeModifier.Operation.MULTIPLY_TOTAL) {
            amount *= 100.0D;
        } else if (attribute == EntityAttributes.GENERIC_KNOCKBACK_RESISTANCE) {
            amount *= 10.0D;
        }
        if (amount > 0.0D) {
            out.accept(Text.translatable("attribute.modifier.plus." + modifier.getOperation().getId(),
                    FORMAT.format(amount), Text.translatable(attribute.getTranslationKey())).formatted(Formatting.BLUE));
        } else if (amount < 0.0D) {
            out.accept(Text.translatable("attribute.modifier.take." + modifier.getOperation().getId(),
                    FORMAT.format(-amount), Text.translatable(attribute.getTranslationKey())).formatted(Formatting.RED));
        }
    }
}
''')
def equipment_tip(s):
    s = s.replace('import net.spell_engine.mixin.client.ItemStackTooltipAccessor;\n', 'import net.spell_engine.compat.item.AttributeTooltipCompat;\n')
    s = s.replace('            var tooltipUtil = (ItemStackTooltipAccessor) (Object) ItemStack.EMPTY;\n', '')
    s = re.sub(r'''\s*tooltipUtil\s*\.spellEngine_appendAttributeModifierTooltip\(\s*bonusLines::add,\s*player,\s*modifier\.attribute\(\),\s*modifier\.modifier\(\)\s*\);''',
               '\n                AttributeTooltipCompat.append(bonusLines::add, modifier.attribute(), modifier.modifier());', s)
    return s
patch('net/spell_engine/api/item/set/EquipmentSetTooltip.java', equipment_tip)
f = path('net/spell_engine/mixin/client/ItemStackTooltipAccessor.java')
if f.exists(): f.unlink()

# 1.21 RangedWeaponItem gained generic load/shootAll internals. On 1.20.1 use a dedicated adapter
# based on CrossbowItem's native projectile/load semantics and keep the modern Spell Engine arrow hook ordering.
write('net/spell_engine/compat/item/RangedWeaponCompat.java', r'''package net.spell_engine.compat.item;
import com.google.common.base.Suppliers;
import net.minecraft.enchantment.EnchantmentHelper;
import net.minecraft.enchantment.Enchantments;
import net.minecraft.entity.CrossbowUser;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.entity.projectile.FireworkRocketEntity;
import net.minecraft.entity.projectile.PersistentProjectileEntity;
import net.minecraft.entity.projectile.ProjectileEntity;
import net.minecraft.item.ArrowItem;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.sound.SoundCategory;
import net.minecraft.sound.SoundEvents;
import net.minecraft.util.Hand;
import net.minecraft.util.math.Vec3d;
import net.minecraft.world.World;
import net.spell_engine.Platform;
import net.spell_engine.internals.SpellTriggers;
import net.spell_engine.internals.casting.SpellCaster;
import net.spell_engine.internals.cost.Ammo;
import net.spell_engine.internals.delivery.arrow.ArrowExtension;
import net.spell_engine.internals.delivery.arrow.ArrowHelper;
import net.spell_engine.internals.delivery.arrow.ArrowShootContext;
import org.jetbrains.annotations.Nullable;
import org.joml.Quaternionf;
import java.util.ArrayList;
import java.util.List;
public final class RangedWeaponCompat {
    private RangedWeaponCompat() { }
    public static List<ItemStack> load(ItemStack weaponStack, ItemStack projectileStack, LivingEntity shooter) {
        int projectileCount = EnchantmentHelper.getLevel(Enchantments.MULTISHOT, weaponStack) > 0 ? 3 : 1;
        boolean creative = shooter instanceof PlayerEntity player && player.getAbilities().creativeMode;
        var source = projectileStack;
        var original = projectileStack.copy();
        var loaded = new ArrayList<ItemStack>(projectileCount);
        for (int i = 0; i < projectileCount; i++) {
            if (i > 0) source = original.copy();
            if (source.isEmpty() && creative) {
                source = new ItemStack(Items.ARROW);
                original = source.copy();
            }
            var one = loadOne(source, shooter, i > 0, creative);
            if (one.isEmpty()) return List.of();
            loaded.add(one);
        }
        return loaded;
    }
    private static ItemStack loadOne(ItemStack projectileStack, LivingEntity shooter, boolean simulated, boolean creative) {
        if (projectileStack.isEmpty()) return ItemStack.EMPTY;
        boolean creativeArrow = creative && projectileStack.getItem() instanceof ArrowItem;
        if (!creativeArrow && !creative && !simulated) {
            if (shooter instanceof PlayerEntity player) {
                var predicate = new Ammo.Searched(null, projectileStack.getItem()).asPredicate();
                var source = Ammo.findContainer(player, predicate, 1);
                if (source != null && Ammo.takeFromContainer(source.itemStack(), predicate, 1) == 1) {
                    var result = projectileStack.copy(); result.setCount(1); return result;
                }
            }
            var result = projectileStack.split(1);
            if (projectileStack.isEmpty() && shooter instanceof PlayerEntity player) player.getInventory().removeOne(projectileStack);
            return result;
        }
        var result = projectileStack.copy(); result.setCount(1); return result;
    }
    public static void shootAll(World world, LivingEntity shooter, Hand hand, ItemStack weaponStack,
                                List<ItemStack> projectiles, float speed, float divergence, boolean critical,
                                @Nullable LivingEntity target) {
        for (int i = 0; i < projectiles.size(); i++) {
            var projectileStack = projectiles.get(i);
            if (projectileStack.isEmpty()) continue;
            float simulated = i == 1 ? -10.0F : (i == 2 ? 10.0F : 0.0F);
            float pitch = i == 0 ? 1.0F : randomShotPitch(i == 1, shooter);
            shoot(world, shooter, hand, weaponStack, projectileStack, pitch,
                    shooter instanceof PlayerEntity p && p.getAbilities().creativeMode,
                    speed, divergence, simulated, critical, target);
        }
    }
    private static float randomShotPitch(boolean high, LivingEntity shooter) {
        float base = high ? 0.63F : 0.43F;
        return 1.0F / (shooter.getRandom().nextFloat() * 0.5F + 1.8F) + base;
    }
    private static void shoot(World world, LivingEntity shooter, Hand hand, ItemStack weaponStack,
                              ItemStack projectileStack, float soundPitch, boolean creative, float speed,
                              float divergence, float simulated, boolean critical, @Nullable LivingEntity target) {
        boolean firework = projectileStack.isOf(Items.FIREWORK_ROCKET);
        ProjectileEntity projectile;
        if (firework) {
            projectile = new FireworkRocketEntity(world, projectileStack, shooter,
                    shooter.getX(), shooter.getEyeY() - 0.15000000596046448D, shooter.getZ(), true);
        } else {
            projectile = createArrow(world, shooter, weaponStack, projectileStack, critical);
        }
        applySpellHooks(projectile, shooter);
        if (projectile instanceof PersistentProjectileEntity arrow && (creative || simulated != 0.0F)) {
            arrow.pickupType = PersistentProjectileEntity.PickupPermission.CREATIVE_ONLY;
        }
        if (shooter instanceof CrossbowUser crossbowUser && target != null) {
            crossbowUser.shoot(target, weaponStack, projectile, simulated);
        } else {
            Vec3d axis = shooter.getOppositeRotationVector(1.0F);
            var rotation = new Quaternionf().setAngleAxis(simulated * 0.017453292F, axis.x, axis.y, axis.z);
            var direction = shooter.getRotationVec(1.0F).toVector3f().rotate(rotation);
            projectile.setVelocity(direction.x(), direction.y(), direction.z(), speed, divergence);
        }
        weaponStack.damage(firework ? 3 : 1, shooter, e -> e.sendToolBreakStatus(hand));
        world.spawnEntity(projectile);
        world.playSound(null, shooter.getX(), shooter.getY(), shooter.getZ(), SoundEvents.ITEM_CROSSBOW_SHOOT,
                SoundCategory.PLAYERS, 1.0F, soundPitch);
    }
    private static PersistentProjectileEntity createArrow(World world, LivingEntity shooter, ItemStack weaponStack,
                                                           ItemStack projectileStack, boolean critical) {
        var arrowItem = (ArrowItem) (projectileStack.getItem() instanceof ArrowItem ? projectileStack.getItem() : Items.ARROW);
        var arrow = arrowItem.createArrow(world, projectileStack, shooter);
        arrow.setCritical(critical);
        arrow.setSound(SoundEvents.ITEM_CROSSBOW_HIT);
        arrow.setShotFromCrossbow(true);
        int piercing = EnchantmentHelper.getLevel(Enchantments.PIERCING, weaponStack);
        if (piercing > 0) arrow.setPierceLevel((byte) piercing);
        return arrow;
    }
    private static void applySpellHooks(ProjectileEntity projectile, LivingEntity shooter) {
        if (!(shooter instanceof PlayerEntity player) || !(projectile instanceof ArrowExtension arrow)) return;
        var caster = (SpellCaster.Player) player;
        var shotContext = caster.getArrowShootContext();
        SpellTriggers.onArrowShot(arrow, player, shotContext.firedBySpell);
        var trackers = Suppliers.memoize(() -> Platform.tracking(shooter));
        for (var spellEntry : shotContext.activeSpells) ArrowHelper.onArrowShot(arrow, shooter, spellEntry, trackers);
        caster.setArrowShootContext(ArrowShootContext.empty());
    }
}
''')
def arrow_helper(s):
    s = s.replace('import net.spell_engine.mixin.item.RangedWeaponAccessor;', 'import net.spell_engine.compat.item.RangedWeaponCompat;')
    s = s.replace('RangedWeaponAccessor.load_SpellEngine(weaponStack, ammo, shooter)', 'RangedWeaponCompat.load(weaponStack, ammo, shooter)')
    s = re.sub(r'\(\(RangedWeaponAccessor\) rangedWeapon\)\.shootAll_SpellEngine\((.*?)\);', r'RangedWeaponCompat.shootAll(\1);', s, flags=re.S)
    return s
patch('net/spell_engine/internals/delivery/arrow/ArrowHelper.java', arrow_helper)
for rel in [
    'net/spell_engine/mixin/item/RangedWeaponAccessor.java',
    'net/spell_engine/mixin/arrow/RangedWeaponItemMixin.java',
    'net/spell_engine/mixin/arrow/RangedWeaponQuiverMixin.java',
]:
    f = path(rel)
    if f.exists(): f.unlink()

# 1.20.1 SpriteBillboardParticle predates Rotator and the 1.21 split quad helpers. Rebuild the same
# orientation/pivot/double-sided behavior on the proven 1.20.1 direct-quad contract.
def particle(s):
    start = s.find('    private static final Rotator GROUND_ROTATOR')
    end = s.find('    // MARK: Factory', start)
    if start == -1 or end == -1:
        raise SystemExit('SpellParticle Rotator block was not found; refusing an unscoped particle rewrite')
    replacement = r'''    private Quaternionf resolvedRotation(Camera camera) {
        return switch (facing) {
            case CAMERA -> new Quaternionf(camera.getRotation());
            case UPRIGHT -> {
                var cameraRotation = camera.getRotation();
                yield new Quaternionf(0F, cameraRotation.y, 0F, cameraRotation.w).normalize();
            }
            case GROUND -> new Quaternionf().rotationX((float)Math.toRadians(-90));
            case VELOCITY -> {
                var direction = new Vector3f((float)velocityX, (float)velocityY, (float)velocityZ);
                if (direction.lengthSquared() < 1.0E-6F) yield new Quaternionf(camera.getRotation());
                direction.normalize();
                yield new Quaternionf().rotationTo(0F, 1F, 0F, direction.x, direction.y, direction.z);
            }
        };
    }

    @Override
    public void buildGeometry(VertexConsumer vertexConsumer, Camera camera, float tickDelta) {
        if (skipRender) return;
        Vec3d cameraPos = camera.getPos();
        float x = (float)(MathHelper.lerp((double)tickDelta, this.prevPosX, this.x) - cameraPos.getX());
        float y = (float)(MathHelper.lerp((double)tickDelta, this.prevPosY, this.y) - cameraPos.getY());
        float z = (float)(MathHelper.lerp((double)tickDelta, this.prevPosZ, this.z) - cameraPos.getZ());
        var rotation = resolvedRotation(camera);
        if (this.angle != 0F) rotation.rotateZ(MathHelper.lerp(tickDelta, this.prevAngle, this.angle));
        drawQuad(vertexConsumer, rotation, x, y + pivot * this.getSize(tickDelta), z, tickDelta);
        if (facing == ParticleGroup.Facing.GROUND) {
            drawQuad(vertexConsumer, new Quaternionf(rotation).rotateX((float)Math.PI),
                    x, y + pivot * this.getSize(tickDelta), z, tickDelta);
        }
    }

    private void drawQuad(VertexConsumer vertexConsumer, Quaternionf rotation,
                          float x, float y, float z, float tickDelta) {
        var corners = new Vector3f[]{
                new Vector3f(-1F, -1F, 0F), new Vector3f(-1F, 1F, 0F),
                new Vector3f(1F, 1F, 0F), new Vector3f(1F, -1F, 0F)
        };
        float size = this.getSize(tickDelta);
        for (var corner : corners) corner.rotate(rotation).mul(size).add(x, y, z);
        float minU = this.getMinU(), maxU = this.getMaxU(), minV = this.getMinV(), maxV = this.getMaxV();
        int light = this.getBrightness(tickDelta);
        vertexConsumer.vertex(corners[0].x(), corners[0].y(), corners[0].z()).texture(maxU, maxV).color(this.red, this.green, this.blue, this.alpha).light(light).next();
        vertexConsumer.vertex(corners[1].x(), corners[1].y(), corners[1].z()).texture(maxU, minV).color(this.red, this.green, this.blue, this.alpha).light(light).next();
        vertexConsumer.vertex(corners[2].x(), corners[2].y(), corners[2].z()).texture(minU, minV).color(this.red, this.green, this.blue, this.alpha).light(light).next();
        vertexConsumer.vertex(corners[3].x(), corners[3].y(), corners[3].z()).texture(minU, maxV).color(this.red, this.green, this.blue, this.alpha).light(light).next();
    }

'''
    return s[:start] + replacement + s[end:]
patch('net/spell_engine/client/particle/SpellParticle.java', particle)

# Remove obsolete 1.21-only mixins now replaced by target-native compatibility code.
mix = resources / 'spell_engine.mixins.json'
data = json.loads(mix.read_text())
remove = {
    'client.ItemStackTooltipAccessor',
    'item.RangedWeaponAccessor',
    'arrow.RangedWeaponItemMixin',
    'arrow.RangedWeaponQuiverMixin',
}
data['mixins'] = [x for x in data.get('mixins', []) if x not in remove]
data['client'] = [x for x in data.get('client', []) if x not in remove]
mix.write_text(json.dumps(data, indent=2) + '\n')

# Guards for the exact failed API family this pass owns.
for needle in ('Rotator', 'method_60373', 'method_60374', 'appendAttributeModifierTooltip', 'shootAll_SpellEngine', 'load_SpellEngine', '@Accessor("levels")'):
    found = [str(f.relative_to(java)) for f in java.rglob('*.java') if needle in f.read_text()]
    if found:
        raise SystemExit(f'compat pass 3 incomplete; {needle!r} remains in {found[:20]}')
print('Spell Engine compatibility pass 3 applied: 1.20.1 loot descriptors + tooltip formatter + ranged adapter + particle renderer')
