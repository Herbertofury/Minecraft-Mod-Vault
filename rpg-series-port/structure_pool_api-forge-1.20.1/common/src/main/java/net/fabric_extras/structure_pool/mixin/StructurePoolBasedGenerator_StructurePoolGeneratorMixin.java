package net.fabric_extras.structure_pool.mixin;

import com.llamalad7.mixinextras.injector.wrapoperation.Operation;
import com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;
import net.fabric_extras.structure_pool.api.StructurePoolAPI;
import net.fabric_extras.structure_pool.internal.StructurePoolExtension;
import net.minecraft.registry.Registry;
import net.minecraft.registry.RegistryKey;
import net.minecraft.structure.PoolStructurePiece;
import net.minecraft.structure.StructureTemplate;
import net.minecraft.structure.pool.StructurePool;
import net.minecraft.structure.pool.StructurePoolBasedGenerator;
import net.minecraft.structure.pool.StructurePoolElement;
import net.minecraft.util.Identifier;
import net.minecraft.util.math.random.Random;
import org.jetbrains.annotations.Nullable;
import org.spongepowered.asm.mixin.Final;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Shadow;
import org.spongepowered.asm.mixin.Unique;
import org.spongepowered.asm.mixin.injection.At;

import java.util.HashMap;
import java.util.List;

@Mixin(StructurePoolBasedGenerator.StructurePoolGenerator.class)
public class StructurePoolBasedGenerator_StructurePoolGeneratorMixin {
    @Shadow @Final private Registry<StructurePool> registry;
    @Unique private final HashMap<Identifier, HashMap<Identifier, Integer>> structurePoolApi$limitedSpawns = new HashMap<>();
    @Unique @Nullable private Identifier structurePoolApi$currentPoolId;

    @WrapOperation(
            method = "generatePiece",
            at = @At(
                    value = "INVOKE",
                    target = "Lnet/minecraft/structure/pool/StructurePoolBasedGenerator$StructurePoolGenerator;getPoolKey(Lnet/minecraft/structure/StructureTemplate$StructureBlockInfo;)Lnet/minecraft/registry/RegistryKey;"
            )
    )
    private RegistryKey<StructurePool> structurePoolApi$capturePoolKey(
            StructureTemplate.StructureBlockInfo blockInfo,
            Operation<RegistryKey<StructurePool>> original
    ) {
        var key = original.call(blockInfo);
        var poolId = key.getValue();

        if (!structurePoolApi$limitedSpawns.containsKey(poolId) && StructurePoolAPI.spawnLimitations.containsKey(poolId)) {
            var freshLimitations = new HashMap<Identifier, Integer>();
            for (var entry : StructurePoolAPI.spawnLimitations.get(poolId).entrySet()) {
                freshLimitations.put(entry.getKey(), entry.getValue().limit());
            }
            structurePoolApi$limitedSpawns.put(poolId, freshLimitations);
        }

        structurePoolApi$currentPoolId = poolId;
        return key;
    }

    @WrapOperation(
            method = "generatePiece",
            at = @At(
                    value = "INVOKE",
                    target = "Lnet/minecraft/structure/pool/StructurePool;getElementIndicesInRandomOrder(Lnet/minecraft/util/math/random/Random;)Ljava/util/List;"
            )
    )
    private List<StructurePoolElement> structurePoolApi$filterLimitedElements(
            StructurePool pool,
            Random random,
            Operation<List<StructurePoolElement>> original
    ) {
        var result = original.call(pool, random);
        result.removeIf(element -> !structurePoolApi$limitAllowsSpawn(element));
        return result;
    }

    @WrapOperation(
            method = "generatePiece",
            at = @At(value = "INVOKE", target = "Ljava/util/List;add(Ljava/lang/Object;)Z")
    )
    private boolean structurePoolApi$consumeLimitOnSuccessfulChild(
            List<Object> list,
            Object object,
            Operation<Boolean> original
    ) {
        if (object instanceof PoolStructurePiece piece) {
            structurePoolApi$consumeLimit(piece.getPoolElement());
        }
        return original.call(list, object);
    }

    @Unique
    private @Nullable Identifier structurePoolApi$resolveStructureElementId(StructurePoolElement element) {
        if (structurePoolApi$currentPoolId == null) {
            return null;
        }
        var pool = registry.get(structurePoolApi$currentPoolId);
        if (pool == null) {
            return null;
        }
        return ((StructurePoolExtension) pool).identify(element);
    }

    @Unique
    private boolean structurePoolApi$limitAllowsSpawn(StructurePoolElement element) {
        var structureId = structurePoolApi$resolveStructureElementId(element);
        if (structureId == null || structurePoolApi$currentPoolId == null) {
            return true;
        }
        var poolLimits = structurePoolApi$limitedSpawns.get(structurePoolApi$currentPoolId);
        if (poolLimits == null) {
            return true;
        }
        var remaining = poolLimits.get(structureId);
        return remaining == null || remaining > 0;
    }

    @Unique
    private void structurePoolApi$consumeLimit(StructurePoolElement element) {
        var structureId = structurePoolApi$resolveStructureElementId(element);
        if (structureId == null || structurePoolApi$currentPoolId == null) {
            return;
        }
        var poolLimits = structurePoolApi$limitedSpawns.get(structurePoolApi$currentPoolId);
        if (poolLimits == null) {
            return;
        }
        var remaining = poolLimits.get(structureId);
        if (remaining != null && remaining > 0) {
            poolLimits.put(structureId, remaining - 1);
        }
    }
}
