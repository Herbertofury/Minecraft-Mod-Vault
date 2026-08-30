#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
JAR="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more_rpg_library-forge-2.7.2+1.20.1.jar"
test -f "$JAR"; unzip -tq "$JAR" >/dev/null
unzip -p "$JAR" META-INF/MANIFEST.MF | tr -d '\r' | grep -Fx 'MixinConfigs: more-rpg-classes.mixins.json' >/dev/null
DUMP="$(mktemp)"; STRINGS="$(mktemp)"; HEART="$(mktemp)"; STEALTH="$(mktemp)"; REFMAP="$(mktemp)"; INDY="$(mktemp)"; BYTE_DIR="$(mktemp -d)"
trap 'rm -f "$DUMP" "$STRINGS" "$HEART" "$STEALTH" "$REFMAP" "$INDY"; rm -rf "$BYTE_DIR"' EXIT

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
if grep -aRIlE 'getDeclared(Field|Method)|get(Field|Method)|Class\.forName|find(Virtual|Static|Special)' "$BYTE_DIR/net/more_rpg_classes" --include='*.class' | grep -q .; then
  echo '[More RPG 2.7.2] production reflection/member-name literal survived packaged bytecode' >&2; exit 1
fi
mapfile -t LAMBDA_CLASSES < <(grep -aRl 'java/lang/invoke/LambdaMetafactory' "$BYTE_DIR/net/more_rpg_classes" --include='*.class' | sed "s#^$BYTE_DIR/##;s#/#.#g;s#\.class\$##" | sort -u)
if ((${#LAMBDA_CLASSES[@]})); then javap -classpath "$JAR" -v -p "${LAMBDA_CLASSES[@]}" > "$INDY"; fi
mapfile -t MC_INDY < <(grep -E '// InvokeDynamic #[0-9]+:.*\)Lnet/minecraft/' "$INDY" | sort -u || true)
for line in "${MC_INDY[@]}"; do
  if ! grep -Eq '// InvokeDynamic #[0-9]+:m_[0-9]+_:' <<<"$line"; then
    echo "[More RPG 2.7.2] non-SRG Minecraft LambdaMetafactory SAM survived: $line" >&2; exit 1
  fi
done

echo '[More RPG 2.7.2] PACKAGED_EXTERNAL_SCHOOL_ATTRIBUTE_OWNERSHIP_PASS schools=frost_ranged,fire_ranged,rage_melee external_writes=3'
echo '[More RPG 2.7.2] PACKAGED_CLIENT_MIXIN_DESCRIPTOR_PASS draw_hearts=IIIZZ stealth_model=IIFFFF'
printf '[More RPG 2.7.2] PRODUCTION_SYMBOLIC_LINKAGE_AUDIT_PASS reflection_literals=0 minecraft_lambda_sams=%s\n' "${#MC_INDY[@]}"
echo '[More RPG 2.7.2] CLEAN_PRODUCTION_RUNTIME_TRACE_ABSENT_PASS'
echo '[More RPG 2.7.2] PACKAGED_MIXIN_MANIFEST_REGISTRATION_PASS config=more-rpg-classes.mixins.json'
