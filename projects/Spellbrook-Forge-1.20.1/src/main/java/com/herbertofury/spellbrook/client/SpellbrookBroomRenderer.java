package com.herbertofury.spellbrook.client;

import com.herbertofury.spellbrook.entity.SpellbrookBroomEntity;
import com.mojang.blaze3d.vertex.PoseStack;
import net.minecraft.client.renderer.MultiBufferSource;
import net.minecraft.client.renderer.entity.EntityRenderer;
import net.minecraft.client.renderer.entity.EntityRendererProvider;
import net.minecraft.client.renderer.texture.TextureAtlas;
import net.minecraft.resources.ResourceLocation;

public class SpellbrookBroomRenderer extends EntityRenderer<SpellbrookBroomEntity> {
    public SpellbrookBroomRenderer(EntityRendererProvider.Context context){super(context);shadowRadius=0.5f;}
    @Override public void render(SpellbrookBroomEntity entity,float entityYaw,float partialTick,PoseStack pose,MultiBufferSource buffers,int light){BroomRenderUtil.render(entity,entity.getRenderStack(),partialTick,entity.getBank(partialTick),entity.getVisualPitch(partialTick),pose,buffers,light);super.render(entity,entityYaw,partialTick,pose,buffers,light);}
    @Override public ResourceLocation getTextureLocation(SpellbrookBroomEntity entity){return TextureAtlas.LOCATION_BLOCKS;}
}
