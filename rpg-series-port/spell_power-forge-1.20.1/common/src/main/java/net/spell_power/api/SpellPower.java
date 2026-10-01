package net.spell_power.api;

import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.attribute.EntityAttributeInstance;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.spell_power.SpellPowerMod;
import net.spell_power.api.statuseffects.VulnerabilityEffect;
import java.util.*;
import java.util.function.Function;

public class SpellPower {
    public record Result(SpellSchool school, double baseValue, double criticalChance, double criticalDamage) {
        public static Result empty(SpellSchool school) { return new Result(school, 0, 0, 0); }
        private static final Random RNG = new Random();
        private enum CriticalStrikeMode { DISABLED, ALLOWED, FORCED }
        public record Value(double amount, boolean isCritical) {}
        public Value random() { return value(CriticalStrikeMode.ALLOWED, Vulnerability.none); }
        public double randomValue() { return random().amount(); }
        public Value random(Vulnerability v) { return value(CriticalStrikeMode.ALLOWED, v); }
        public double randomValue(Vulnerability v) { return random(v).amount(); }
        public Value nonCritical() { return value(CriticalStrikeMode.DISABLED, Vulnerability.none); }
        public double nonCriticalValue() { return nonCritical().amount(); }
        public Value forcedCritical() { return value(CriticalStrikeMode.FORCED, Vulnerability.none); }
        public double forcedCriticalValue() { return forcedCritical().amount(); }
        private Value value(CriticalStrikeMode mode, Vulnerability v) {
            double out = baseValue * (1F + v.powerBaseMultiplier); boolean crit = false;
            if (mode != CriticalStrikeMode.DISABLED) {
                crit = mode == CriticalStrikeMode.FORCED || RNG.nextFloat() < (criticalChance + v.criticalChanceBonus);
                if (crit) out *= criticalDamage + v.criticalDamageBonus;
            }
            return new Value(out, crit);
        }
    }
    public record VulnerabilityQuery(LivingEntity entity, SpellSchool school) {}
    public static final ArrayList<Function<VulnerabilityQuery, List<Vulnerability>>> vulnerabilitySources = new ArrayList<>(List.of(query -> {
        var list = new ArrayList<Vulnerability>();
        for (var effect : query.entity().getStatusEffects()) if (effect.getEffectType() instanceof VulnerabilityEffect v) list.add(v.getVulnerability(query.school(), effect.getAmplifier()));
        return list;
    }));
    public static Vulnerability getVulnerability(LivingEntity entity, SpellSchool school) {
        var query = new VulnerabilityQuery(entity, school); var all = new ArrayList<Vulnerability>();
        for (var source : vulnerabilitySources) all.addAll(source.apply(query));
        return Vulnerability.sum(all);
    }
    public record Vulnerability(float powerBaseMultiplier, float criticalChanceBonus, float criticalDamageBonus) {
        public static final Vulnerability none = new Vulnerability(0, 0, 0);
        public static Vulnerability sum(List<Vulnerability> list) {
            var value = none;
            for (var entry : list) value = new Vulnerability(value.powerBaseMultiplier + entry.powerBaseMultiplier,
                    value.criticalChanceBonus + entry.criticalChanceBonus, value.criticalDamageBonus + entry.criticalDamageBonus);
            return value;
        }
        public Vulnerability multiply(float n) { return new Vulnerability(powerBaseMultiplier * n, criticalChanceBonus * n, criticalDamageBonus * n); }
    }
    public static Result getSpellPower(SpellSchool school, LivingEntity entity) {
        var args = new SpellSchool.QueryArgs(entity); double power = school.getValue(SpellSchool.Trait.POWER, args);
        if (school.archetype == SpellSchool.Archetype.MAGIC) {
            EntityAttributeInstance instance = entity.getAttributeInstance(school.attribute);
            if (instance != null) {
                double flatPower = getAttributeFlatValue(instance);
                double genericBase = SpellSchools.GENERIC.attribute.getDefaultValue();
                double genericMultiplier = entity.getAttributeValue(SpellSchools.GENERIC.attribute) / genericBase;
                genericMultiplier += SpellPowerMod.genericSpellPowerEnchantBonus(entity);
                power += flatPower * (genericMultiplier - 1);
            }
        }
        return new Result(school, power, school.getValue(SpellSchool.Trait.CRIT_CHANCE, args), school.getValue(SpellSchool.Trait.CRIT_DAMAGE, args));
    }
    private static double getAttributeFlatValue(EntityAttributeInstance instance) {
        double out = 0;
        for (var modifier : instance.getModifiers()) if (modifier.getOperation() == EntityAttributeModifier.Operation.ADDITION) out += modifier.getValue();
        return out;
    }
    @Deprecated public static float getHaste(LivingEntity entity) { return getHaste(entity, SpellSchools.ARCANE); }
    public static float getHaste(LivingEntity entity, SpellSchool school) { return (float) school.getValue(SpellSchool.Trait.HASTE, new SpellSchool.QueryArgs(entity)); }
}
