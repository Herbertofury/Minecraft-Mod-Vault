package net.runes.mixin;
import net.minecraft.entity.player.PlayerEntity;
import net.runes.crafting.RuneCrafter;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Unique;
@Mixin(PlayerEntity.class)
public class PlayerEntityMixin implements RuneCrafter {
    @Unique private int runes$lastRuneCrafted;
    public void setLastRuneCrafted(int time) { runes$lastRuneCrafted=time; }
    public int getLastRuneCrafted() { return runes$lastRuneCrafted; }
}
