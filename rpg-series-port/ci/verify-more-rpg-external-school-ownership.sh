#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
JAR="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more_rpg_library-forge-2.7.2+1.20.1.jar"
test -f "$JAR"; unzip -tq "$JAR" >/dev/null
unzip -p "$JAR" META-INF/MANIFEST.MF | tr -d '\r' | grep -Fx 'MixinConfigs: more-rpg-classes.mixins.json' >/dev/null

# Lock the exact modern 2.7.2 JSON authority that the 1.20.1 compatibility class must represent.
python3 - "$JAR" <<'PY'
import json, sys, zipfile
jar = sys.argv[1]
expected = {
    'typhoon': (['spell_power:air', 'spell_power:water'], '#more_rpg_classes:enchantable/typhoon'),
    'stonebloom': (['spell_power:earth', 'spell_power:nature'], '#more_rpg_classes:enchantable/stonebloom'),
}
with zipfile.ZipFile(jar) as zf:
    names = set(zf.namelist())
    compat = 'net/more_rpg_classes/compat/MoreRpg1201Enchantments.class'
    if compat not in names:
        raise SystemExit('[More RPG 2.7.2] packaged 1.20.1 enchantment compatibility class missing')
    for name, (attrs, supported) in expected.items():
        p = f'data/more_rpg_classes/enchantment/{name}.json'
        if p not in names:
            raise SystemExit(f'[More RPG 2.7.2] authoritative enchantment JSON missing: {p}')
        data = json.loads(zf.read(p))
        if data.get('max_level') != 5 or data.get('weight') != 2 or data.get('slots') != ['armor']:
            raise SystemExit(f'[More RPG 2.7.2] {name} level/weight/slot parity drift: {data}')
        if data.get('exclusive_set') != '#spell_power:multi_school' or data.get('supported_items') != supported:
            raise SystemExit(f'[More RPG 2.7.2] {name} exclusivity/tag parity drift: {data}')
        effects = data.get('effects', {}).get('minecraft:attributes', [])
        got = [e.get('attribute') for e in effects]
        if got != attrs:
            raise SystemExit(f'[More RPG 2.7.2] {name} school parity drift: {got} != {attrs}')
        for effect in effects:
            amount = effect.get('amount', {})
            if effect.get('operation') != 'add_multiplied_base' or amount.get('type') != 'minecraft:linear' \
                    or amount.get('base') != 0.03 or amount.get('per_level_above_first') != 0.03:
                raise SystemExit(f'[More RPG 2.7.2] {name} +3% power parity drift: {effect}')
print('[More RPG 2.7.2] PACKAGED_ENCHANTMENT_JSON_AUTHORITY_PASS typhoon=air+water stonebloom=earth+nature bonus=0.03 max_level=5 weight=2 armor_only=true exclusive=multi_school')
PY

DUMP="$(mktemp)"; STRINGS="$(mktemp)"; HEART="$(mktemp)"; STEALTH="$(mktemp)"; REFMAP="$(mktemp)"; INDY="$(mktemp)"; ENCHANT="$(mktemp)"; BYTE_DIR="$(mktemp -d)"
trap 'rm -f "$DUMP" "$STRINGS" "$HEART" "$STEALTH" "$REFMAP" "$INDY" "$ENCHANT"; rm -rf "$BYTE_DIR"' EXIT
javap -classpath "$JAR" -c -p net.more_rpg_classes.compat.MoreRpg1201Enchantments > "$ENCHANT"
for needle in 'typhoon' 'stonebloom' 'SchoolFilteredEnchantment' 'PowerEnchantmentConfig' 'armorEnchantmentLevel'; do
  grep -Fq "$needle" "$ENCHANT" || { echo "[More RPG 2.7.2] packaged enchantment compatibility bytecode missing $needle" >&2; cat "$ENCHANT" >&2; exit 1; }
done
echo '[More RPG 2.7.2] PACKAGED_ENCHANTMENT_1201_BRIDGE_PASS class=MoreRpg1201Enchantments spell_power_specialized=true'

javap -classpath "$JAR" -c -p net.more_rpg_classes.custom.MoreSpellSchools > "$DUMP"
external_count="$(grep -c 'SpellSchool\$Manage.EXTERNAL' "$DUMP" || true)"; write_count="$(grep -c 'SpellSchool.attributeManagement' "$DUMP" || true)"
[[ "$external_count" -eq 3 ]] || { echo "[More RPG 2.7.2] packaged external school ownership count wrong: external=$external_count expected=3" >&2; cat "$DUMP" >&2; exit 1; }
[[ "$write_count" -eq 3 ]] || { echo "[More RPG 2.7.2] packaged attributeManagement write count wrong: writes=$write_count expected=3" >&2; cat "$DUMP" >&2; exit 1; }
for cls in net.more_rpg_classes.MRPGCMod net.more_rpg_classes.forge.ForgeMod; do javap -classpath "$JAR" -c -p "$cls" >> "$STRINGS"; done
if grep -Fq '[More RPG Runtime Trace]' "$STRINGS"; then echo '[More RPG 2.7.2] diagnostic runtime trace leaked into clean production package' >&2; cat "$STRINGS" >&2; exit 1; fi
javap -classpath "$JAR" -s -p net.more_rpg_classes.mixin.DrawHeartsMixin > "$HEART"
grep -F 'IIIZZLorg/spongepowered/asm/mixin/injection/callback/CallbackInfo;)V' "$HEART" >/dev/null || { echo '[More RPG 2.7.2] packaged DrawHearts callback is not target 1.20.1 IIIZZ' >&2; cat "$HEART" >&2; exit 1; }
if grep -Fq 'IIZZZLorg/spongepowered/asm/mixin/injection/callback/CallbackInfo;)V' "$HEART"; then echo '[More RPG 2.7.2] modern DrawHearts IIZZZ callback survived package remap' >&2; exit 1; fi
javap -classpath "$JAR" -s -p net.more_rpg_classes.mixin.LivingEntityRenderStealth > "$STEALTH"
grep -F 'IIFFFFLcom/llamalad7/mixinextras/injector/wrapoperation/Operation;' "$STEALTH" >/dev/null || { echo '[More RPG 2.7.2] packaged stealth renderer handler is not target 1.20.1 RGBA shape' >&2; cat "$STEALTH" >&2; exit 1; }
unzip -p "$JAR" more_rpg_library-common-common-refmap.json > "$REFMAP"
grep -F 'VertexConsumer;IIFFFF)V' "$REFMAP" >/dev/null || { echo '[More RPG 2.7.2] 1.20.1 RGBA model render target missing from refmap' >&2; cat "$REFMAP" >&2; exit 1; }
if grep -Fq 'VertexConsumer;III)V' "$REFMAP"; then echo '[More RPG 2.7.2] modern packed-color model render target survived refmap' >&2; exit 1; fi
grep -F 'Gui$HeartType;IIIZZ)V' "$REFMAP" >/dev/null || { echo '[More RPG 2.7.2] DrawHearts IIIZZ production mapping missing from refmap' >&2; exit 1; }
(cd "$BYTE_DIR" && unzip -qq "$JAR" 'net/more_rpg_classes/*.class' 'net/more_rpg_classes/**/*.class' 2>/dev/null || true)
if grep -aRIlE 'getDeclared(Field|Method)|get(Field|Method)|Class\.forName|find(Virtual|Static|Special)' "$BYTE_DIR/net/more_rpg_classes" --include='*.class' | grep -q .; then echo '[More RPG 2.7.2] production reflection/member-name literal survived packaged bytecode' >&2; exit 1; fi
mapfile -t LAMBDA_CLASSES < <(grep -aRl 'java/lang/invoke/LambdaMetafactory' "$BYTE_DIR/net/more_rpg_classes" --include='*.class' | sed "s#^$BYTE_DIR/##;s#/#.#g;s#\.class\$##" | sort -u)
if ((${#LAMBDA_CLASSES[@]})); then javap -classpath "$JAR" -v -p "${LAMBDA_CLASSES[@]}" > "$INDY"; fi
mapfile -t MC_INDY < <(grep -E '// InvokeDynamic #[0-9]+:.*\)Lnet/minecraft/' "$INDY" | sort -u || true)
for line in "${MC_INDY[@]}"; do if ! grep -Eq '// InvokeDynamic #[0-9]+:m_[0-9]+_:' <<<"$line"; then echo "[More RPG 2.7.2] non-SRG Minecraft LambdaMetafactory SAM survived: $line" >&2; exit 1; fi; done
echo '[More RPG 2.7.2] PACKAGED_EXTERNAL_SCHOOL_ATTRIBUTE_OWNERSHIP_PASS schools=frost_ranged,fire_ranged,rage_melee external_writes=3'
echo '[More RPG 2.7.2] PACKAGED_CLIENT_MIXIN_DESCRIPTOR_PASS draw_hearts=IIIZZ stealth_model=IIFFFF'
printf '[More RPG 2.7.2] PRODUCTION_SYMBOLIC_LINKAGE_AUDIT_PASS reflection_literals=0 minecraft_lambda_sams=%s\n' "${#MC_INDY[@]}"
echo '[More RPG 2.7.2] CLEAN_PRODUCTION_RUNTIME_TRACE_ABSENT_PASS'
echo '[More RPG 2.7.2] PACKAGED_MIXIN_MANIFEST_REGISTRATION_PASS config=more-rpg-classes.mixins.json'
