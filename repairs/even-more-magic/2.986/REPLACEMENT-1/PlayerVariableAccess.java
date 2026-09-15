package net.mcreator.evenmoremagic.network;
import net.minecraftforge.common.util.LazyOptional;
/** Avoids eager fallback allocation and maintains derived player-variable values at their source writes. */
public final class PlayerVariableAccess {
    private PlayerVariableAccess() {}
    public static EvenMoreMagicModVariables.PlayerVariables resolve(LazyOptional<?> optional) {
        Object value = optional.orElse(null);
        return value instanceof EvenMoreMagicModVariables.PlayerVariables variables ? variables : new EvenMoreMagicModVariables.PlayerVariables();
    }
    public static void setSpellCooldownValue(EvenMoreMagicModVariables.PlayerVariables variables, double value) {
        variables.spell_cooldown_ovelay_cooldown_value = value;
        double seconds = (value - value % 20.0D) / 20.0D;
        variables.spell_cooldown_ovelay_cooldown_value_in_seconds = seconds;
        variables.spell_cooldown_ovelay_cooldown_value_type = seconds < 10.0D ? 1.0D : (seconds < 100.0D ? 2.0D : 3.0D);
    }
}
