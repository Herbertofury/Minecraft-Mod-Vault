package com.herbertofury.spellbrook.compat.hexerei;

import com.herbertofury.spellbrook.client.BroomRenderUtil;
import com.mojang.blaze3d.vertex.PoseStack;
import net.minecraft.client.renderer.MultiBufferSource;
import net.minecraft.client.renderer.entity.EntityRenderer;
import net.minecraft.client.renderer.entity.EntityRendererProvider;
import net.minecraft.client.renderer.texture.TextureAtlas;
import net.minecraft.resources.ResourceLocation;
import net.minecraft.util.Mth;

public class HexereiSpellbrookBroomRenderer extends EntityRenderer<HexereiSpellbrookBroomEntity> {
    public HexereiSpellbrookBroomRenderer(EntityRendererProvider.Context context){super(context);shadowRadius=0.5f;}
    @Override public void render(HexereiSpellbrookBroomEntity entity,float yaw,float partialTick,PoseStack pose,MultiBufferSource buffers,int light){float bank=Mth.lerp(partialTick,entity.deltaRotationOld,entity.deltaRotation)*4.0f;float pitch=Mth.clamp((float)-entity.getDeltaMovement().y*18f,-12f,12f);BroomRenderUtil.render(entity,entity.getRenderStack(),partialTick,bank,pitch,pose,buffers,light);super.render(entity,yaw,partialTick,pose,buffers,light);}
    @Override public ResourceLocation getTextureLocation(HexereiSpellbrookBroomEntity entity){return TextureAtlas.LOCATION_BLOCKS;}
}
