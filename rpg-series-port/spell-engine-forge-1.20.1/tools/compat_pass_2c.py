#!/usr/bin/env python3
from pathlib import Path
import json, re, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_2.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
base = Path(sys.argv[2]).resolve()
java = root / 'common/src/main/java'
resources = root / 'common/src/main/resources'


def path(rel): return java / rel
def read(rel): return path(rel).read_text()
def write(rel, text):
    f = path(rel); f.parent.mkdir(parents=True, exist_ok=True); f.write_text(text)
def patch(rel, fn):
    f = path(rel)
    if f.exists(): f.write_text(fn(f.read_text()))

# 1.21 data-tracker builder -> 1.20.1 startTracking.
for rel in ['net/spell_engine/entity/SpellCloud.java','net/spell_engine/entity/SpellProjectile.java','net/spell_engine/entity/SpellModelEffect.java','net/spell_engine/entity/SummonedEntity.java']:
    def entity_tracker(s):
        return s.replace('protected void initDataTracker(DataTracker.Builder builder) {', 'protected void initDataTracker() {').replace('super.initDataTracker(builder);', 'super.initDataTracker();').replace('builder.add(', 'this.dataTracker.startTracking(')
    patch(rel, entity_tracker)
for rel in ['net/spell_engine/mixin/entity/PlayerEntityMixin.java','net/spell_engine/mixin/effect/LivingEntityStatusEffectSync.java','net/spell_engine/mixin/arrow/PersistentProjectileEntityMixin.java']:
    def mixin_tracker(s, rel=rel):
        s = s.replace('(DataTracker.Builder builder, CallbackInfo ci)', '(CallbackInfo ci)')
        if rel.endswith('PlayerEntityMixin.java'):
            s = s.replace('builder.add(', 'player().getDataTracker().startTracking(')
        elif rel.endswith('PersistentProjectileEntityMixin.java'):
            s = s.replace('builder.add(', 'arrow().getDataTracker().startTracking(')
        else:
            s = s.replace('builder.add(', 'this.dataTracker.startTracking(')
        return s
    patch(rel, mixin_tracker)

# Java 17 collection conveniences.
repls = {
 'castTicks >= tickHolder.ticks.getFirst()':'castTicks >= tickHolder.ticks.get(0)',
 'targets.getFirst()':'targets.get(0)', 'action.spawns.getFirst()':'action.spawns.get(0)', 'effects.getFirst()':'effects.get(0)',
 'removalSelection.getFirst()':'removalSelection.get(0)', 'result.items().getFirst()':'result.items().get(0)',
 'selectedSpells.getFirst()':'selectedSpells.get(0)', 'cooldownsRemoved.getFirst()':'cooldownsRemoved.get(0)', 'spells.getFirst()':'spells.get(0)',
 'hitTicks.getLast()':'hitTicks.get(hitTicks.size() - 1)', 'this.entries.getLast()':'this.entries.get(this.entries.size() - 1)'
}
for f in java.rglob('*.java'):
    s = f.read_text()
    for a,b in repls.items(): s = s.replace(a,b)
    f.write_text(s)

# 1.20.1 has no vanilla generic entity scale attribute. Preserve observed physical scaling from dimensions.
write('net/spell_engine/compat/EntityScaleCompat.java', 'package net.spell_engine.compat; import net.minecraft.entity.LivingEntity; public final class EntityScaleCompat { private EntityScaleCompat(){} public static float scale(LivingEntity e){float base=e.getType().getDimensions().width;return base<=0F?1F:Math.max(0.01F,e.getWidth()/base);} }\n')
for f in java.rglob('*.java'):
    s = f.read_text()
    for name in ['livingEntity','player','caster','target']:
        s = s.replace(name + '.getScale()', 'net.spell_engine.compat.EntityScaleCompat.scale(' + name + ')')
    f.write_text(s)
patch('net/spell_engine/entity/SummonedEntity.java', lambda s: s.replace('\n    // MARK:', '\n    public float getScale() { return net.spell_engine.compat.EntityScaleCompat.scale(this); }\n\n    // MARK:', 1) if 'public float getScale()' not in s else s)

# Sprite particle facing modes on the 1.20.1 quad API.
def spell_particle(s):
    s = s.replace('? livingEntity.getScale() : 1F;', '? net.spell_engine.compat.EntityScaleCompat.scale(livingEntity) : 1F;')
    start = s.find('    /// -90 and not +90:')
    end = s.find('    // MARK: Factories', start)
    if start != -1 and end != -1:
        replacement = r'''    private Quaternionf resolvedRotation(Camera camera) {
        return switch (facing) {
            case CAMERA -> new Quaternionf(camera.getRotation());
            case UPRIGHT -> { var q=camera.getRotation(); yield new Quaternionf(0F,q.y,0F,q.w).normalize(); }
            case GROUND -> new Quaternionf().rotationX((float)Math.toRadians(-90));
            case VELOCITY -> { var d=new Vector3f((float)velocityX,(float)velocityY,(float)velocityZ); if(d.lengthSquared()<1.0E-6F) yield new Quaternionf(camera.getRotation()); d.normalize(); yield new Quaternionf().rotationTo(0F,1F,0F,d.x,d.y,d.z); }
        };
    }
    @Override public void buildGeometry(VertexConsumer vc, Camera camera, float tickDelta) {
        if(skipRender)return; var cam=camera.getPos(); float px=(float)(MathHelper.lerp(tickDelta,prevPosX,x)-cam.x); float py=(float)(MathHelper.lerp(tickDelta,prevPosY,y)-cam.y)+pivot*getSize(tickDelta); float pz=(float)(MathHelper.lerp(tickDelta,prevPosZ,z)-cam.z); var q=resolvedRotation(camera); if(angle!=0F)q.rotateZ(MathHelper.lerp(tickDelta,prevAngle,angle)); drawQuad(vc,q,px,py,pz,tickDelta); if(facing==ParticleGroup.Facing.GROUND)drawQuad(vc,new Quaternionf(q).rotateX((float)Math.PI),px,py,pz,tickDelta);
    }
    private void drawQuad(VertexConsumer vc,Quaternionf q,float x,float y,float z,float tickDelta){float size=getSize(tickDelta);var c=new Vector3f[]{new Vector3f(-1F,-1F,0F),new Vector3f(-1F,1F,0F),new Vector3f(1F,1F,0F),new Vector3f(1F,-1F,0F)};for(var v:c)v.rotate(q).mul(size).add(x,y,z);float minU=getMinU(),maxU=getMaxU(),minV=getMinV(),maxV=getMaxV();int light=getBrightness(tickDelta);vc.vertex(c[0].x(),c[0].y(),c[0].z()).texture(maxU,maxV).color(red,green,blue,alpha).light(light).next();vc.vertex(c[1].x(),c[1].y(),c[1].z()).texture(maxU,minV).color(red,green,blue,alpha).light(light).next();vc.vertex(c[2].x(),c[2].y(),c[2].z()).texture(minU,minV).color(red,green,blue,alpha).light(light).next();vc.vertex(c[3].x(),c[3].y(),c[3].z()).texture(minU,maxV).color(red,green,blue,alpha).light(light).next();}

'''
        s = s[:start] + replacement + s[end:]
    return s
patch('net/spell_engine/client/particle/SpellParticle.java', spell_particle)

# 1.20.1 attributes have no 1.21 display category metadata.
patch('net/spell_engine/api/entity/SpellEngineAttributes.java', lambda s: re.sub(r'\n\s*public Entry category\(EntityAttribute\.Category category\) \{.*?\n\s*\}', '', s, flags=re.S))
for f in java.rglob('*.java'):
    s=f.read_text(); s=re.sub(r'\.category\(EntityAttribute\.Category\.[A-Z_]+\)', '', s); f.write_text(s)

# Commands use IdentifierArgumentType in 1.20.1.
def commands(s):
    s=s.replace('import net.minecraft.command.argument.RegistryEntryReferenceArgumentType;', 'import net.minecraft.command.argument.IdentifierArgumentType;')
    s=s.replace('RegistryEntryReferenceArgumentType.registryEntry(registryAccess, SpellRegistry.KEY)', 'IdentifierArgumentType.identifier()')
    s=s.replace('var spell = RegistryEntryReferenceArgumentType.getRegistryEntry(context, "spell", SpellRegistry.KEY);', 'var spellId = IdentifierArgumentType.getIdentifier(context, "spell");\n                                                var spell = SpellRegistry.from(context.getSource().getWorld()).getEntry(spellId).orElseThrow();')
    return s
patch('net/spell_engine/misc/SpellEngineCommands.java', commands)
