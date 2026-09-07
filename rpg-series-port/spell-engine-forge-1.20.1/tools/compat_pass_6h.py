#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6h.py <generated-port-root> <spell-engine-1.20.1-baseline>')

root = Path(sys.argv[1]).resolve()
forge_build = root / 'forge/build.gradle'
s = forge_build.read_text()

# Spell Power and Ranged Weapon API are separately installed RPG Series mods. Compile the common
# module against their common JARs (pass 1), and put their real Forge JARs on this loader module's
# compile/runtime classpath. Do NOT `include` or shadow either dependency into Spell Engine.
if 'SPELL_POWER_FORGE_JAR' not in s:
    s += r'''

def requireExternalModJar = { String envName ->
 def raw = System.getenv(envName)
 if (raw == null || raw.isBlank()) {
  throw new GradleException("Missing required external Forge mod JAR environment variable: ${envName}")
 }
 def jarFile = file(raw)
 if (!jarFile.isFile()) {
  throw new GradleException("External Forge mod JAR does not exist for ${envName}: ${jarFile}")
 }
 return jarFile
}

dependencies {
 modImplementation files(requireExternalModJar('SPELL_POWER_FORGE_JAR'))
 modImplementation files(requireExternalModJar('RANGED_FORGE_JAR'))
}
'''
forge_build.write_text(s)

# Spell Power 1.6.0's modern API represents these schools with RegistryEntry<EntityAttribute>, which
# means Spell Power does not own their backing attributes/effects. The 1.20.1 compatibility API uses
# SpellSchool.Manage for the same ownership distinction. Preserve it explicitly: without this bridge,
# Forge sees vanilla GENERIC_ATTACK_DAMAGE as INTERNAL and attempts to register that same object again
# as spell_power:physical_melee during the attribute RegisterEvent.
external_schools = root / 'common/src/main/java/net/spell_engine/api/spell/ExternalSpellSchools.java'
es = external_schools.read_text()
anchor = '    private static boolean initialized = false;\n'
ownership_bridge = '''    static {\n        // Backport Spell Power 1.6.0 RegistryEntry ownership semantics. These schools borrow\n        // attributes owned by Minecraft or Ranged Weapon API and must never register them anew.\n        for (var school : new SpellSchool[]{PHYSICAL_MELEE, PHYSICAL_MELEE_DUAL, PHYSICAL_RANGED, DEFENSE, HEALTH}) {\n            school.attributeManagement = SpellSchool.Manage.EXTERNAL;\n            school.powerEffectManagement = SpellSchool.Manage.EXTERNAL;\n        }\n    }\n\n'''
if 'school.attributeManagement = SpellSchool.Manage.EXTERNAL;' not in es:
    if anchor not in es:
        raise SystemExit('ExternalSpellSchools initialization anchor missing')
    es = es.replace(anchor, ownership_bridge + anchor, 1)
external_schools.write_text(es)

# 1.21 ModelPart passes a packed color through renderCuboids, but 1.20.1 passes separate RGBA floats.
# The previous packed-int ModifyVariable therefore found zero arguments at runtime. Tint the four
# arguments at the leaf Cuboid.renderCuboid invocation instead: this preserves the modern once-per-part
# semantics without compounding through ModelPart.render's child recursion.
entity_tints = root / 'common/src/main/java/net/spell_engine/api/effect/EntityTints.java'
et = entity_tints.read_text()
color_anchor = '''        public static int apply(int color) {\n            return argb == NEUTRAL ? color : multiply(color, argb);\n        }'''
color_helpers = '''        public static int apply(int color) {\n            return argb == NEUTRAL ? color : multiply(color, argb);\n        }\n\n        public static float applyRed(float red) {\n            return argb == NEUTRAL ? red : red * (((argb >> 16) & 0xFF) / 255.0F);\n        }\n\n        public static float applyGreen(float green) {\n            return argb == NEUTRAL ? green : green * (((argb >> 8) & 0xFF) / 255.0F);\n        }\n\n        public static float applyBlue(float blue) {\n            return argb == NEUTRAL ? blue : blue * ((argb & 0xFF) / 255.0F);\n        }\n\n        public static float applyAlpha(float alpha) {\n            return argb == NEUTRAL ? alpha : alpha * ((argb >>> 24) / 255.0F);\n        }'''
if 'public static float applyRed(float red)' not in et:
    if color_anchor not in et:
        raise SystemExit('EntityTints.Current packed-color helper anchor missing')
    et = et.replace(color_anchor, color_helpers, 1)
entity_tints.write_text(et)

model_part = root / 'common/src/main/java/net/spell_engine/mixin/client/render/tint/ModelPartMixin.java'
model_part.write_text(r'''package net.spell_engine.mixin.client.render.tint;

import net.minecraft.client.model.ModelPart;
import net.spell_engine.api.effect.EntityTints;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.ModifyArg;

@Mixin(ModelPart.class)
public class ModelPartMixin {
    private static final String CUBOID_RENDER = "Lnet/minecraft/client/model/ModelPart$Cuboid;renderCuboid(Lnet/minecraft/client/util/math/MatrixStack$Entry;Lnet/minecraft/client/render/VertexConsumer;IIFFFF)V";

    // 1.20.1 passes RGBA as four floats. Modify only the leaf Cuboid call so child ModelParts receive
    // the original color and the entity tint is applied exactly once per rendered cuboid hierarchy leaf.
    @ModifyArg(method = "renderCuboids", at = @At(value = "INVOKE", target = CUBOID_RENDER), index = 4)
    private float spellEngine_applyEntityTintRed(float red) {
        return EntityTints.Current.applyRed(red);
    }

    @ModifyArg(method = "renderCuboids", at = @At(value = "INVOKE", target = CUBOID_RENDER), index = 5)
    private float spellEngine_applyEntityTintGreen(float green) {
        return EntityTints.Current.applyGreen(green);
    }

    @ModifyArg(method = "renderCuboids", at = @At(value = "INVOKE", target = CUBOID_RENDER), index = 6)
    private float spellEngine_applyEntityTintBlue(float blue) {
        return EntityTints.Current.applyBlue(blue);
    }

    @ModifyArg(method = "renderCuboids", at = @At(value = "INVOKE", target = CUBOID_RENDER), index = 7)
    private float spellEngine_applyEntityTintAlpha(float alpha) {
        return EntityTints.Current.applyAlpha(alpha);
    }
}
''')

final = forge_build.read_text()
for required in ('SPELL_POWER_FORGE_JAR', 'RANGED_FORGE_JAR', 'modImplementation files(requireExternalModJar'):
    if required not in final:
        raise SystemExit(f'pass6h missing external Forge dependency wiring: {required}')
for forbidden in (
    "include files(requireExternalModJar('SPELL_POWER_FORGE_JAR'))",
    "include files(requireExternalModJar('RANGED_FORGE_JAR'))",
    'SPELL_POWER_SOURCE_DIRS',
    'RANGED_SOURCE_DIRS',
):
    if forbidden in final:
        raise SystemExit(f'pass6h would embed or source-inject a separate dependency: {forbidden}')

final_schools = external_schools.read_text()
for school in ('PHYSICAL_MELEE', 'PHYSICAL_MELEE_DUAL', 'PHYSICAL_RANGED', 'DEFENSE', 'HEALTH'):
    if school not in final_schools:
        raise SystemExit(f'pass6h external spell school missing: {school}')
for required in (
    'new SpellSchool[]{PHYSICAL_MELEE, PHYSICAL_MELEE_DUAL, PHYSICAL_RANGED, DEFENSE, HEALTH}',
    'school.attributeManagement = SpellSchool.Manage.EXTERNAL;',
    'school.powerEffectManagement = SpellSchool.Manage.EXTERNAL;',
):
    if required not in final_schools:
        raise SystemExit(f'pass6h missing external Spell Power ownership bridge: {required}')
if final_schools.count('school.attributeManagement = SpellSchool.Manage.EXTERNAL;') != 1:
    raise SystemExit('pass6h produced ambiguous duplicate external-attribute ownership bridge')

final_tints = entity_tints.read_text()
for required in ('applyRed(float red)', 'applyGreen(float green)', 'applyBlue(float blue)', 'applyAlpha(float alpha)'):
    if required not in final_tints:
        raise SystemExit(f'pass6h missing 1.20.1 float entity-tint helper: {required}')
final_model = model_part.read_text()
for index, name in ((4, 'Red'), (5, 'Green'), (6, 'Blue'), (7, 'Alpha')):
    if f'index = {index}' not in final_model or f'spellEngine_applyEntityTint{name}' not in final_model:
        raise SystemExit(f'pass6h missing ModelPart RGBA tint hook index {index}')
for stale in ('@ModifyVariable', 'private int spellEngine_applyEntityTint(int color)'):
    if stale in final_model:
        raise SystemExit(f'pass6h left stale 1.21 packed-color ModelPart hook: {stale}')

print('Spell Engine compatibility pass 6h applied: external RPG deps/ownership + exact 1.20.1 ModelPart RGBA tint hook')
