package net.fabric_extras.structure_pool.mixin;

import net.fabric_extras.structure_pool.internal.StructurePoolExtension;
import net.minecraft.structure.pool.StructurePool;
import net.minecraft.structure.pool.StructurePoolElement;
import net.minecraft.util.Identifier;
import org.jetbrains.annotations.Nullable;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Unique;

import java.util.HashMap;
import java.util.Map;

@Mixin(StructurePool.class)
public class StructurePoolMixin implements StructurePoolExtension {
    @Unique
    private final Map<StructurePoolElement, Identifier> structurePoolApi$identifiedElements = new HashMap<>();

    @Override
    public void remember(StructurePoolElement element, Identifier identifier) {
        structurePoolApi$identifiedElements.put(element, identifier);
    }

    @Override
    public @Nullable Identifier identify(StructurePoolElement element) {
        return structurePoolApi$identifiedElements.get(element);
    }
}
