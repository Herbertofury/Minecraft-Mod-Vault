#!/usr/bin/env python3
from pathlib import Path
import re, sys
if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_5a.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
base = Path(sys.argv[2]).resolve()
J = root / 'common/src/main/java'

def p(rel): return J / rel
def ed(rel, fn):
    q = p(rel)
    if not q.exists():
        raise SystemExit(f'missing expected source: {rel}')
    old = q.read_text()
    new = fn(old)
    if old == new:
        raise SystemExit(f'pass5a expected transform did not match: {rel}')
    q.write_text(new)
def rp(rel, *pairs):
    def x(s):
        for a,b in pairs:
            s = s.replace(a,b)
        return s
    ed(rel,x)

# Pure Java/source compatibility: enum case labels are unqualified in Java 17.
rp('net/spell_engine/client/compatibility/FirstPersonAnimationCompatibility.java',
   ('case TriStateAuto.YES:', 'case YES:'), ('case TriStateAuto.NO:', 'case NO:'))

# 1.20.1 VertexConsumer takes Matrix4f and double vertex coordinates, and exposes fixed-color state.
rp('net/spell_engine/client/render/BeamRenderer.java',
   ('vertices.vertex(matrix, x, y, z)', 'vertices.vertex(matrix.getPositionMatrix(), x, y, z)'))

def item_glow(s):
    s = s.replace('public VertexConsumer vertex(float x, float y, float z) {',
                  'public VertexConsumer vertex(double x, double y, double z) {')
    insert = '''\n    @Override\n    public void next() {\n        delegate.next();\n    }\n\n    @Override\n    public void fixedColor(int red, int green, int blue, int alpha) {\n        delegate.fixedColor(this.red, this.green, this.blue, this.alpha);\n    }\n\n    @Override\n    public void unfixColor() {\n        delegate.unfixColor();\n    }\n'''
    marker = '\n    @Override\n    public VertexConsumer normal(float x, float y, float z) {'
    if marker not in s: raise SystemExit('ItemGlow normal marker missing')
    return s.replace(marker, insert + marker)
ed('net/spell_engine/client/util/ItemGlowVertexConsumer.java', item_glow)

# Proven genuine 1.20.1 widget/screen/model signatures.
rp('net/spell_engine/client/render/SpellBindingBlockEntityRenderer.java',
   ('this.book.renderBook(matrixStack, vertexConsumer, i, j, -1);',
    'this.book.renderBook(matrixStack, vertexConsumer, i, j, 1.0F, 1.0F, 1.0F, 1.0F);'))
rp('net/spell_engine/client/gui/CustomButton.java',
   ('protected void renderWidget(DrawContext context, int mouseX, int mouseY, float delta)',
    'protected void renderButton(DrawContext context, int mouseX, int mouseY, float delta)'))

def hud(s):
    s = re.sub(r'''var checkBox = CheckboxWidget\.builder\(Text\.of\(""\), textRenderer\)\s*\.pos\(x, y\)\s*\.checked\(checked\)\s*\.build\(\);''',
               'var checkBox = new CheckboxWidget(x, y, buttonSize, buttonSize, Text.of(""), checked);', s)
    return s.replace('renderBackground(context, mouseX, mouseY, delta);', 'renderBackground(context);')
ed('net/spell_engine/client/gui/HudConfigScreen.java', hud)
for rel in ('net/spell_engine/client/gui/ConfigMenuScreen.java','net/spell_engine/spellbinding/SpellBindingScreen.java'):
    rp(rel, ('renderBackground(context, mouseX, mouseY, delta);','renderBackground(context);'),
            ('this.renderBackground(context, mouseX, mouseY, delta);','this.renderBackground(context);'))

# 1.20.1 has no TooltipType parameter in this hook. Use the same GameOptions advanced-tooltip flag
# that vanilla ItemStack uses to decide where the dark-gray advanced lines live.
def tooltip(s):
    s = s.replace('if (tooltipType.isAdvanced()) {', 'if (MinecraftClient.getInstance().options.advancedItemTooltips) {')
    s = s.replace('attribute.getKey().ifPresent(key -> modifiers.put(key.getValue().toString(), modifier))',
                  'modifiers.put(Registries.ATTRIBUTE.getId(attribute).toString(), modifier)')
    return s
ed('net/spell_engine/client/gui/SpellTooltip.java', tooltip)

# BakedModelManager in this target accepts ModelIdentifier. The Forge target already uses the
# standalone variant for the exact same custom model IDs, so use that target-native representation
# in the common compile path too.
for rel in ('net/spell_engine/api/render/CustomModels.java','net/spell_engine/mixin/client/ItemRendererMixin.java'):
    rp(rel, ('getModel(modelId)', 'getModel(new ModelIdentifier(modelId, "standalone"))'))

# Residual holder/direct-value conversions left deliberately visible by pass 4b.
for rel in ('net/spell_engine/api/effect/KnockbackImmunity.java',
            'net/spell_engine/mixin/effect/LivingEntityStatusEffectSync.java',
            'net/spell_engine/internals/delivery/SpellStashHelper.java',
            'net/spell_engine/SpellEngineMod.java'):
    q=p(rel); s=q.read_text(); old=s
    s=s.replace('((KnockbackImmunity) effect.value())','((KnockbackImmunity) effect)')
    s=s.replace('var effect = entry.getKey().value();','var effect = entry.getKey();')
    s=s.replace('RemoveOnHit.removeCount(entity.getWorld(), effect.value(), source)','RemoveOnHit.removeCount(entity.getWorld(), effect, source)')
    if s != old: q.write_text(s)

# 1.20.1 EntityType.Builder / advancement criterion signatures.
def mod(s):
    s=s.replace('.dimensions(0.25F, 0.25F)', '.setDimensions(0.25F, 0.25F)')
    s=s.replace('.dimensions(6F, 0.5F)', '.setDimensions(6F, 0.5F)')
    s=s.replace('.dimensions(0.5F, 0.5F)', '.setDimensions(0.5F, 0.5F)')
    s=re.sub(r'Criteria\.register\([^,\n]+,\s*([^\)]+)\);', r'Criteria.register(\1);', s)
    return s
ed('net/spell_engine/SpellEngineMod.java', mod)

# 1.20.1 ItemStack durability API uses a break-status callback rather than an EquipmentSlot argument.
rp('net/spell_engine/internals/cost/SpellCost.java',
   ('stackToDamage.damage(spell.cost.durability, player, EquipmentSlot.MAINHAND);',
    'stackToDamage.damage(spell.cost.durability, player, e -> e.sendEquipmentBreakStatus(EquipmentSlot.MAINHAND));'))

# Static ownership guards for the exact cluster this pass claims.
guards = {
 'case TriStateAuto.': 'qualified enum switch labels',
 'renderWidget(DrawContext': '1.21 CustomButton hook',
 'CheckboxWidget.builder(': '1.21 checkbox builder',
 'renderBackground(context, mouseX, mouseY, delta)': '1.21 Screen background signature',
 'tooltipType.isAdvanced()': '1.21 TooltipType',
 'effect.value()': 'status-effect holder residue',
 '.dimensions(0.25F, 0.25F)': '1.21 entity builder dimensions',
 'Criteria.register(EnchantmentSpecificCriteria.ID.toString()': '1.21 criterion registration',
}
for needle,label in guards.items():
    hits=[str(q.relative_to(J)) for q in J.rglob('*.java') if needle in q.read_text()]
    if hits: raise SystemExit(f'pass5a incomplete ({label}): {hits[:20]}')
print('Spell Engine compatibility pass 5a applied: Java17 + client/UI/render + direct effects + 1.20.1 builder/criteria/durability signatures')
