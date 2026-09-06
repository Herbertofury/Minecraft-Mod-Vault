package com.herbertofury.spellbrook.client;

import com.mojang.blaze3d.vertex.PoseStack;
import com.mojang.math.Axis;
import net.minecraft.client.Minecraft;
import net.minecraft.client.renderer.MultiBufferSource;
import net.minecraft.client.renderer.texture.OverlayTexture;
import net.minecraft.util.Mth;
import net.minecraft.world.entity.Entity;
import net.minecraft.world.item.ItemDisplayContext;
import net.minecraft.world.item.ItemStack;

public final class BroomRenderUtil {
    private BroomRenderUtil(){}
    public static void render(Entity entity,ItemStack stack,float partialTick,float bank,float pitch,PoseStack pose,MultiBufferSource buffers,int light){float yaw=Mth.rotLerp(partialTick,entity.yRotO,entity.getYRot());pose.pushPose();pose.translate(0,0.55,0);pose.mulPose(Axis.YP.rotationDegrees(180f-yaw));pose.mulPose(Axis.ZP.rotationDegrees(bank));pose.mulPose(Axis.XP.rotationDegrees(pitch));Minecraft.getInstance().getItemRenderer().renderStatic(stack, ItemDisplayContext.FIXED,light, OverlayTexture.NO_OVERLAY,pose,buffers,entity.level(),entity.getId());pose.popPose();}
}
