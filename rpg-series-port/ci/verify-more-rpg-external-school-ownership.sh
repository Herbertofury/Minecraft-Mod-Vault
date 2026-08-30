#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
JAR="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more_rpg_library-forge-2.7.2+1.20.1.jar"
test -f "$JAR"
unzip -tq "$JAR" >/dev/null
unzip -p "$JAR" META-INF/MANIFEST.MF | tr -d '\r' | grep -Fx 'MixinConfigs: more-rpg-classes.mixins.json' >/dev/null
DUMP="$(mktemp)"; STRINGS="$(mktemp)"; HEART="$(mktemp)"; STEALTH="$(mktemp)"; REFMAP="$(mktemp)"
trap 'rm -f "$DUMP" "$STRINGS" "$HEART" "$STEALTH" "$REFMAP"' EXIT
javap -classpath "$JAR" -c -p net.more_rpg_classes.custom.MoreSpellSchools > "$DUMP"
external_count="$(grep -c 'SpellSchool\$Manage.EXTERNAL' "$DUMP" || true)"
write_count="$(grep -c 'SpellSchool.attributeManagement' "$DUMP" || true)"
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

echo '[More RPG 2.7.2] PACKAGED_EXTERNAL_SCHOOL_ATTRIBUTE_OWNERSHIP_PASS schools=frost_ranged,fire_ranged,rage_melee external_writes=3'
echo '[More RPG 2.7.2] PACKAGED_CLIENT_MIXIN_DESCRIPTOR_PASS draw_hearts=IIIZZ stealth_model=IIFFFF'
echo '[More RPG 2.7.2] CLEAN_PRODUCTION_RUNTIME_TRACE_ABSENT_PASS'
echo '[More RPG 2.7.2] PACKAGED_MIXIN_MANIFEST_REGISTRATION_PASS config=more-rpg-classes.mixins.json'
