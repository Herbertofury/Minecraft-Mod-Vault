package net.runes.crafting;
import com.google.gson.JsonObject;
import net.minecraft.advancement.criterion.AbstractCriterion;
import net.minecraft.advancement.criterion.AbstractCriterionConditions;
import net.minecraft.predicate.entity.AdvancementEntityPredicateDeserializer;
import net.minecraft.predicate.entity.LootContextPredicate;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.util.Identifier;
public class RuneCraftingCriteria extends AbstractCriterion<RuneCraftingCriteria.Condition> {
    public static final Identifier ID=RuneCrafting.ID;
    public static final RuneCraftingCriteria INSTANCE=new RuneCraftingCriteria();
    protected Condition conditionsFromJson(JsonObject obj,LootContextPredicate playerPredicate,AdvancementEntityPredicateDeserializer d){ return new Condition(); }
    public Identifier getId(){ return ID; }
    public void trigger(ServerPlayerEntity player){ trigger(player,c->true); }
    public static class Condition extends AbstractCriterionConditions { public Condition(){ super(ID,LootContextPredicate.EMPTY); } }
}
