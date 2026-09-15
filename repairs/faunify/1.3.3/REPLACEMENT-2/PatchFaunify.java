import java.nio.file.Files;
import java.nio.file.Path;

import jdk.internal.org.objectweb.asm.ClassReader;
import jdk.internal.org.objectweb.asm.ClassVisitor;
import jdk.internal.org.objectweb.asm.ClassWriter;
import jdk.internal.org.objectweb.asm.Label;
import jdk.internal.org.objectweb.asm.MethodVisitor;
import jdk.internal.org.objectweb.asm.Opcodes;

public final class PatchFaunify {
    private static final String TARGET_METHOD = "findHighestPriorityLightSource";
    private static final String TARGET_DESC = "()Lnet/minecraft/core/BlockPos;";

    private static final class SafeClassWriter extends ClassWriter {
        SafeClassWriter(ClassReader reader, int flags) {
            super(reader, flags);
        }

        @Override
        protected String getCommonSuperClass(String type1, String type2) {
            if (type1.equals(type2)) return type1;
            return "java/lang/Object";
        }
    }

    public static void main(String[] args) throws Exception {
        if (args.length != 2) {
            throw new IllegalArgumentException("usage: PatchFaunify <input-class> <output-class>");
        }

        byte[] input = Files.readAllBytes(Path.of(args[0]));
        ClassReader cr = new ClassReader(input);
        SafeClassWriter cw = new SafeClassWriter(cr, ClassWriter.COMPUTE_FRAMES | ClassWriter.COMPUTE_MAXS);

        final boolean[] replaced = {false};
        ClassVisitor cv = new ClassVisitor(Opcodes.ASM8, cw) {
            @Override
            public MethodVisitor visitMethod(int access, String name, String descriptor, String signature, String[] exceptions) {
                if (name.equals(TARGET_METHOD) && descriptor.equals(TARGET_DESC)) {
                    if (replaced[0]) throw new IllegalStateException("target method occurs more than once");
                    replaced[0] = true;
                    emitReplacement(cw, access, name, descriptor, signature, exceptions);
                    return null;
                }
                return super.visitMethod(access, name, descriptor, signature, exceptions);
            }
        };

        cr.accept(cv, 0);
        if (!replaced[0]) throw new IllegalStateException("target method not found");

        byte[] output = cw.toByteArray();
        Files.write(Path.of(args[1]), output);
        System.out.println("patched class bytes: " + input.length + " -> " + output.length);
    }

    private static void emitReplacement(ClassWriter cw, int access, String name, String desc, String signature, String[] exceptions) {
        MethodVisitor mv = cw.visitMethod(access, name, desc, signature, exceptions);
        mv.visitCode();

        // Level level = moth.level();
        mv.visitVarInsn(Opcodes.ALOAD, 0);
        mv.visitFieldInsn(Opcodes.GETFIELD, "com/pepper/faunify/entity/SilkMothEntity$LightAttractionGoal", "moth", "Lcom/pepper/faunify/entity/SilkMothEntity;");
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "com/pepper/faunify/entity/SilkMothEntity", "m_9236_", "()Lnet/minecraft/world/level/Level;", false);
        mv.visitVarInsn(Opcodes.ASTORE, 1);

        // BlockPos mothPos = moth.blockPosition();
        mv.visitVarInsn(Opcodes.ALOAD, 0);
        mv.visitFieldInsn(Opcodes.GETFIELD, "com/pepper/faunify/entity/SilkMothEntity$LightAttractionGoal", "moth", "Lcom/pepper/faunify/entity/SilkMothEntity;");
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "com/pepper/faunify/entity/SilkMothEntity", "m_20183_", "()Lnet/minecraft/core/BlockPos;", false);
        mv.visitVarInsn(Opcodes.ASTORE, 2);

        // BlockPos best = null;
        mv.visitInsn(Opcodes.ACONST_NULL);
        mv.visitVarInsn(Opcodes.ASTORE, 3);

        // int bestBrightness = -1;
        mv.visitInsn(Opcodes.ICONST_M1);
        mv.visitVarInsn(Opcodes.ISTORE, 4);

        // double bestDistance = +infinity;
        mv.visitLdcInsn(Double.POSITIVE_INFINITY);
        mv.visitVarInsn(Opcodes.DSTORE, 5);

        // Iterator<BlockPos> iterator = BlockPos.betweenClosed(mothPos.offset(-15...), mothPos.offset(15...)).iterator();
        mv.visitVarInsn(Opcodes.ALOAD, 2);
        mv.visitIntInsn(Opcodes.BIPUSH, -15);
        mv.visitIntInsn(Opcodes.BIPUSH, -15);
        mv.visitIntInsn(Opcodes.BIPUSH, -15);
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "net/minecraft/core/BlockPos", "m_7918_", "(III)Lnet/minecraft/core/BlockPos;", false);
        mv.visitVarInsn(Opcodes.ALOAD, 2);
        mv.visitIntInsn(Opcodes.BIPUSH, 15);
        mv.visitIntInsn(Opcodes.BIPUSH, 15);
        mv.visitIntInsn(Opcodes.BIPUSH, 15);
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "net/minecraft/core/BlockPos", "m_7918_", "(III)Lnet/minecraft/core/BlockPos;", false);
        mv.visitMethodInsn(Opcodes.INVOKESTATIC, "net/minecraft/core/BlockPos", "m_121940_", "(Lnet/minecraft/core/BlockPos;Lnet/minecraft/core/BlockPos;)Ljava/lang/Iterable;", false);
        mv.visitMethodInsn(Opcodes.INVOKEINTERFACE, "java/lang/Iterable", "iterator", "()Ljava/util/Iterator;", true);
        mv.visitVarInsn(Opcodes.ASTORE, 7);

        Label loopCheck = new Label();
        Label loopBody = new Label();
        Label continueLoop = new Label();
        Label compareDistance = new Label();
        Label updateBest = new Label();
        Label loopDone = new Label();

        mv.visitJumpInsn(Opcodes.GOTO, loopCheck);
        mv.visitLabel(loopBody);

        // BlockPos pos = (BlockPos) iterator.next();
        mv.visitVarInsn(Opcodes.ALOAD, 7);
        mv.visitMethodInsn(Opcodes.INVOKEINTERFACE, "java/util/Iterator", "next", "()Ljava/lang/Object;", true);
        mv.visitTypeInsn(Opcodes.CHECKCAST, "net/minecraft/core/BlockPos");
        mv.visitVarInsn(Opcodes.ASTORE, 8);

        // BlockState state = level.getBlockState(pos);
        mv.visitVarInsn(Opcodes.ALOAD, 1);
        mv.visitVarInsn(Opcodes.ALOAD, 8);
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "net/minecraft/world/level/Level", "m_8055_", "(Lnet/minecraft/core/BlockPos;)Lnet/minecraft/world/level/block/state/BlockState;", false);
        mv.visitVarInsn(Opcodes.ASTORE, 9);

        // if (state.is(SILKMOTH_BLACKLIST)) continue;
        mv.visitVarInsn(Opcodes.ALOAD, 9);
        mv.visitFieldInsn(Opcodes.GETSTATIC, "com/pepper/faunify/registry/FaunifyBlocks", "SILKMOTH_BLACKLIST", "Lnet/minecraft/tags/TagKey;");
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "net/minecraft/world/level/block/state/BlockState", "m_204336_", "(Lnet/minecraft/tags/TagKey;)Z", false);
        mv.visitJumpInsn(Opcodes.IFNE, continueLoop);

        // if (state.getLightEmission() < 8) continue;
        mv.visitVarInsn(Opcodes.ALOAD, 9);
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "net/minecraft/world/level/block/state/BlockState", "m_60791_", "()I", false);
        mv.visitIntInsn(Opcodes.BIPUSH, 8);
        mv.visitJumpInsn(Opcodes.IF_ICMPLT, continueLoop);

        // int brightness = level.getBrightness(LightLayer.BLOCK, pos);
        mv.visitVarInsn(Opcodes.ALOAD, 1);
        mv.visitFieldInsn(Opcodes.GETSTATIC, "net/minecraft/world/level/LightLayer", "BLOCK", "Lnet/minecraft/world/level/LightLayer;");
        mv.visitVarInsn(Opcodes.ALOAD, 8);
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "net/minecraft/world/level/Level", "m_45517_", "(Lnet/minecraft/world/level/LightLayer;Lnet/minecraft/core/BlockPos;)I", false);
        mv.visitVarInsn(Opcodes.ISTORE, 10);

        // double distance = pos.distSqr(mothPos);
        mv.visitVarInsn(Opcodes.ALOAD, 8);
        mv.visitVarInsn(Opcodes.ALOAD, 2);
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "net/minecraft/core/BlockPos", "m_123331_", "(Lnet/minecraft/core/Vec3i;)D", false);
        mv.visitVarInsn(Opcodes.DSTORE, 11);

        // if (brightness > bestBrightness) update;
        mv.visitVarInsn(Opcodes.ILOAD, 10);
        mv.visitVarInsn(Opcodes.ILOAD, 4);
        mv.visitJumpInsn(Opcodes.IF_ICMPGT, updateBest);

        // if (brightness == bestBrightness && distance < bestDistance) update;
        mv.visitVarInsn(Opcodes.ILOAD, 10);
        mv.visitVarInsn(Opcodes.ILOAD, 4);
        mv.visitJumpInsn(Opcodes.IF_ICMPEQ, compareDistance);
        mv.visitJumpInsn(Opcodes.GOTO, continueLoop);

        mv.visitLabel(compareDistance);
        mv.visitVarInsn(Opcodes.DLOAD, 11);
        mv.visitVarInsn(Opcodes.DLOAD, 5);
        mv.visitInsn(Opcodes.DCMPG);
        mv.visitJumpInsn(Opcodes.IFLT, updateBest);
        mv.visitJumpInsn(Opcodes.GOTO, continueLoop);

        mv.visitLabel(updateBest);
        // best = pos.immutable();
        mv.visitVarInsn(Opcodes.ALOAD, 8);
        mv.visitMethodInsn(Opcodes.INVOKEVIRTUAL, "net/minecraft/core/BlockPos", "m_7949_", "()Lnet/minecraft/core/BlockPos;", false);
        mv.visitVarInsn(Opcodes.ASTORE, 3);
        mv.visitVarInsn(Opcodes.ILOAD, 10);
        mv.visitVarInsn(Opcodes.ISTORE, 4);
        mv.visitVarInsn(Opcodes.DLOAD, 11);
        mv.visitVarInsn(Opcodes.DSTORE, 5);

        mv.visitLabel(continueLoop);
        mv.visitLabel(loopCheck);
        mv.visitVarInsn(Opcodes.ALOAD, 7);
        mv.visitMethodInsn(Opcodes.INVOKEINTERFACE, "java/util/Iterator", "hasNext", "()Z", true);
        mv.visitJumpInsn(Opcodes.IFNE, loopBody);

        mv.visitLabel(loopDone);
        mv.visitVarInsn(Opcodes.ALOAD, 3);
        mv.visitInsn(Opcodes.ARETURN);

        mv.visitMaxs(0, 0);
        mv.visitEnd();
    }
}
