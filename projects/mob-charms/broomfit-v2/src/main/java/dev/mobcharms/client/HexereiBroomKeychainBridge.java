package dev.mobcharms.client;

import com.mojang.blaze3d.vertex.PoseStack;
import com.mojang.logging.LogUtils;
import dev.mobcharms.core.CharmAttachment;
import dev.mobcharms.core.CharmType;
import net.minecraft.client.Minecraft;
import net.minecraft.client.player.LocalPlayer;
import net.minecraft.client.renderer.MultiBufferSource;
import net.minecraft.world.entity.HumanoidArm;
import net.minecraft.world.item.ItemStack;
import org.slf4j.Logger;

import java.lang.reflect.Method;

/**
 * Client-only soft bridge between Hexerei's broom keychain render anchor and
 * Punchy's Charmsy geometry renderer.
 *
 * Hexerei already owns the broom-tip attachment point, its chain, and the
 * per-broom world-space swing. When the ItemStack stored inside that keychain
 * carries Mob Charms NBT, this bridge asks Punchy to render a dedicated
 * mascot-only definition at the exact pose Hexerei prepared. The stored item's
 * normal model is suppressed only when that Punchy render succeeds.
 *
 * Both Hexerei and Punchy remain optional: all access to Punchy is reflective,
 * and the Hexerei target itself is a @Pseudo mixin.
 */
public final class HexereiBroomKeychainBridge {
    private static final Logger LOGGER = LogUtils.getLogger();
    private static final ThreadLocal<Boolean> BROOM_KEYCHAIN_RENDER =
            ThreadLocal.withInitial(() -> Boolean.FALSE);

    private static volatile Access access;
    private static volatile boolean accessFailed;
    private static volatile boolean warned;

    private HexereiBroomKeychainBridge() {}

    /** True only while Punchy is resolving/rendering a Hexerei keychain mascot. */
    public static boolean isBroomKeychainRenderActive() {
        return Boolean.TRUE.equals(BROOM_KEYCHAIN_RENDER.get());
    }

    /**
     * Render the charm mascot at Hexerei's already-transformed keychain endpoint.
     *
     * @return true only when Punchy actually handled the mascot render; callers
     * should cancel Hexerei's normal stored-item render in that case.
     */
    public static boolean renderIfAttached(ItemStack stack, PoseStack poseStack,
                                           MultiBufferSource buffers, int light) {
        if (stack == null || stack.isEmpty() || CharmAttachment.get(stack).isEmpty()) {
            return false;
        }

        Minecraft minecraft = Minecraft.getInstance();
        LocalPlayer player = minecraft.player;
        if (player == null) {
            return false;
        }

        Access a = access();
        if (a == null) {
            return false;
        }

        CharmType charm = CharmAttachment.get(stack).orElse(null);
        if (charm == null) {
            return false;
        }

        HumanoidArm arm = player.getMainArm();
        poseStack.pushPose();
        try {
            /*
             * Hexerei deliberately shrinks a normal keychain payload to 0.25x
             * immediately before BroomRenderer.renderItem().  Charmsy's broom
             * mascots are already compact geometry, so inheriting that scale
             * makes them nearly invisible.  The factors below both undo that
             * payload shrink and normalize the nine authored mascots to a
             * roughly 0.30-block visual envelope.
             *
             * Crucially this transform is applied only inside the pushed pose
             * used for the replacement mascot. Hexerei's own collar/loop, chain,
             * attachment point, and broom-space swing remain untouched and thus
             * continue to wrap/dangle correctly on every Hexerei broom variant.
             */
            float presentationScale = broomPresentationScale(charm);
            poseStack.scale(presentationScale, presentationScale, presentationScale);

            BROOM_KEYCHAIN_RENDER.set(Boolean.TRUE);
            Object system = a.getSystem().invoke(null);
            Object raw = a.tryRenderItem().invoke(
                    system,
                    stack,
                    player,
                    arm,
                    poseStack,
                    buffers,
                    light,
                    minecraft.getFrameTime(),
                    false // Hexerei already supplied the authoritative pose.
            );
            return Boolean.TRUE.equals(raw);
        } catch (ReflectiveOperationException | RuntimeException ex) {
            warnOnce("Hexerei broom keychain charm render failed: "
                    + ex.getClass().getSimpleName());
            return false;
        } finally {
            BROOM_KEYCHAIN_RENDER.remove();
            poseStack.popPose();
        }
    }


    /**
     * Scale compensation for Hexerei's 0.25x generic keychain payload scale.
     * Values are normalized from the authored mascot bounds, not arbitrary
     * per-screenshot zooms, so variants stay visually consistent.
     */
    static float broomPresentationScale(CharmType charm) {
        return switch (charm) {
            case ALLAY -> 6.75F;
            case AXOLOTL -> 3.75F;
            case BEE -> 4.60F;
            case CAMEL -> 5.65F;
            case CHICKEN -> 5.25F;
            case CREEPER -> 5.90F;
            case PUFFERFISH -> 8.00F;
            case WARDEN -> 4.20F;
            case ZOMBIE -> 5.10F;
        };
    }

    private static Access access() {
        Access cached = access;
        if (cached != null || accessFailed) {
            return cached;
        }
        synchronized (HexereiBroomKeychainBridge.class) {
            if (access != null || accessFailed) {
                return access;
            }
            try {
                Class<?> type = Class.forName("punchy.client.bedrockitems.BedrockItemSystem");
                Method getSystem = type.getMethod("get");
                Method tryRenderItem = type.getMethod(
                        "tryRenderItem",
                        ItemStack.class,
                        LocalPlayer.class,
                        HumanoidArm.class,
                        PoseStack.class,
                        MultiBufferSource.class,
                        int.class,
                        float.class,
                        boolean.class
                );
                access = new Access(getSystem, tryRenderItem);
            } catch (ReflectiveOperationException | LinkageError ex) {
                accessFailed = true;
            }
            return access;
        }
    }

    private static void warnOnce(String message) {
        if (!warned) {
            warned = true;
            LOGGER.warn("Mob Charms: {}", message);
        }
    }

    private record Access(Method getSystem, Method tryRenderItem) {}
}
