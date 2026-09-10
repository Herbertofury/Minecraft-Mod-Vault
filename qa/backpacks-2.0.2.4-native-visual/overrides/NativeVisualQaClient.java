package dev.scaiquest.backpacks20.nativeqa;

import com.mojang.blaze3d.systems.RenderSystem;
import com.mojang.blaze3d.vertex.PoseStack;
import com.mojang.math.Axis;
import dev.scaiquest.backpacks20.client.BedrockModelRenderer;
import net.minecraft.client.Minecraft;
import net.minecraft.client.gui.GuiGraphics;
import net.minecraft.client.gui.screens.Screen;
import net.minecraft.client.renderer.LightTexture;
import net.minecraft.client.renderer.MultiBufferSource;
import net.minecraft.client.renderer.texture.OverlayTexture;
import net.minecraft.network.chat.Component;
import net.minecraft.resources.ResourceLocation;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.event.TickEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;

@Mod.EventBusSubscriber(modid = NativeVisualQaMod.MOD_ID, bus = Mod.EventBusSubscriber.Bus.FORGE, value = Dist.CLIENT)
public final class NativeVisualQaClient {
    private static boolean opened;
    private static int startupTicks;

    private NativeVisualQaClient() {}

    @SubscribeEvent
    public static void onClientTick(TickEvent.ClientTickEvent event) {
        if (event.phase != TickEvent.Phase.END || opened) return;
        Minecraft mc = Minecraft.getInstance();
        startupTicks++;
        if (startupTicks >= 40 && mc.level == null && mc.screen != null) {
            opened = true;
            System.out.println("QA_NATIVE_REPLACING_SCREEN=" + mc.screen.getClass().getName());
            System.out.println("QA_NATIVE_GUI=" + mc.getWindow().getGuiScaledWidth() + "x" + mc.getWindow().getGuiScaledHeight() + " scale=" + mc.getWindow().getGuiScale());
            BedrockModelRenderer.clearCache();
            mc.setScreen(new ProofScreen());
            System.out.println("QA_NATIVE_SCREEN_OPEN");
        }
    }

    private static final class ProofScreen extends Screen {
        private int frames;
        private boolean readyLogged;
        private boolean doneLogged;

        private ProofScreen() {
            super(Component.literal("Backpacks 2.0.2.4 Native Forge Proof"));
        }

        @Override
        public void render(GuiGraphics graphics, int mouseX, int mouseY, float partialTick) {
            graphics.fill(0, 0, width, height, 0xFF070A10);
            graphics.fill(20, 18, width - 20, 82, 0xFF111827);
            graphics.drawCenteredString(font, "BACKPACKS 2.0.2.4 - REAL FORGE 1.20.1 RUNTIME", width / 2, 30, 0xFFFFFFFF);
            graphics.drawCenteredString(font, "Exact patched Bedrock renderer + exact source geometry/textures", width / 2, 48, 0xFFB8C4D8);
            graphics.drawCenteredString(font, "Top: full base light | Bottom: base light = 0 (alpha emissive stays FULL BRIGHT)", width / 2, 64, 0xFF8FA9C7);

            int[] designs = {16, 23, 26};
            String[] names = {"HONEY BEE", "WARDEN", "CREAKING"};
            int[] xs = {width / 6, width / 2, width * 5 / 6};

            for (int i = 0; i < designs.length; i++) {
                int x = xs[i];
                graphics.drawCenteredString(font, names[i], x, 105, 0xFFFFFFFF);
                graphics.drawCenteredString(font, "geometry " + designs[i], x, 119, 0xFF7F8FA6);
                renderDesign(graphics, designs[i], x, 255, LightTexture.FULL_BRIGHT, 100.0f);
                graphics.drawCenteredString(font, "FULL LIGHT", x, 365, 0xFFB8C4D8);
                renderDesign(graphics, designs[i], x, 515, 0, 100.0f);
                graphics.drawCenteredString(font, "DARK / EMISSIVE", x, 635, 0xFFE5F4FF);
            }

            graphics.drawCenteredString(font, "NATIVE QA: source basis (-X pivot/origin, -X/-Y rotation) + entity_emissive_alpha bridge", width / 2, 690, 0xFF7FC8FF);

            frames++;
            if (!readyLogged && frames >= 100) {
                readyLogged = true;
                System.out.println("QA_NATIVE_VISUAL_READY");
            }
            if (!doneLogged && frames >= 420) {
                doneLogged = true;
                System.out.println("QA_NATIVE_VISUAL_DONE");
                Minecraft.getInstance().stop();
            }
        }

        private void renderDesign(GuiGraphics graphics, int design, int x, int y, int packedLight, float scale) {
            ResourceLocation model = new ResourceLocation("sqst_bkpk", "bedrock/models/backpack_" + design + ".geo.json");
            ResourceLocation texture = new ResourceLocation("sqst_bkpk", "textures/backpacks/entity/" + design + "/0.png");
            ResourceLocation emissive = new ResourceLocation("sqst_bkpk", "textures/backpacks/emissive/" + design + "/0.png");

            PoseStack pose = graphics.pose();
            pose.pushPose();
            pose.translate(x, y, 260.0f);
            pose.scale(scale, -scale, scale);
            pose.mulPose(Axis.XP.rotationDegrees(18.0f));
            pose.mulPose(Axis.YP.rotationDegrees(180.0f));
            pose.translate(0.0f, -modelCenterY(design), 0.0f);

            RenderSystem.enableDepthTest();
            MultiBufferSource.BufferSource buffers = Minecraft.getInstance().renderBuffers().bufferSource();
            BedrockModelRenderer.renderEmissiveAlpha(
                    model,
                    "geometry.sqst_bkpk.backpack_" + design,
                    texture,
                    emissive,
                    pose,
                    buffers,
                    packedLight,
                    OverlayTexture.NO_OVERLAY
            );
            buffers.endBatch();
            pose.popPose();
        }

        private float modelCenterY(int design) {
            return switch (design) {
                case 23 -> 10.055f / 16.0f;
                case 16, 26 -> 6.0f / 16.0f;
                default -> 0.5f;
            };
        }

        @Override
        public boolean isPauseScreen() {
            return false;
        }
    }
}
