import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import jdk.internal.org.objectweb.asm.ClassReader;
import jdk.internal.org.objectweb.asm.ClassWriter;
import jdk.internal.org.objectweb.asm.MethodVisitor;
import jdk.internal.org.objectweb.asm.Opcodes;
import jdk.internal.org.objectweb.asm.tree.AbstractInsnNode;
import jdk.internal.org.objectweb.asm.tree.ClassNode;
import jdk.internal.org.objectweb.asm.tree.FieldInsnNode;
import jdk.internal.org.objectweb.asm.tree.InsnList;
import jdk.internal.org.objectweb.asm.tree.InsnNode;
import jdk.internal.org.objectweb.asm.tree.MethodInsnNode;
import jdk.internal.org.objectweb.asm.tree.MethodNode;
import jdk.internal.org.objectweb.asm.tree.VarInsnNode;

public final class PatchProtectionPixelR3 {
    private static final String OLD_HELPER = "net/mcreator/protectionpixel/repair/ArmorRenderTargetCache";
    private static final String NEW_HELPER = "net/mcreator/protectionpixel/repair/ArmorRenderTargetIndex";
    private static final String BLOCK_ENTITY = "net/minecraft/world/level/block/entity/BlockEntity";
    private static final String LEVEL_DESC = "Lnet/minecraft/world/level/Level;";
    private static final String POS_DESC = "Lnet/minecraft/core/BlockPos;";
    private static final String TARGETS_OLD_DESC = "(Lnet/minecraft/client/multiplayer/ClientLevel;Lnet/minecraft/core/BlockPos;I)Ljava/util/Map;";
    private static final String TARGETS_NEW_DESC = "(Lnet/minecraft/world/level/Level;Lnet/minecraft/core/BlockPos;I)Ljava/util/Map;";

    public static void main(String[] args) throws Exception {
        if (args.length != 4) {
            throw new IllegalArgumentException("usage: PatchProtectionPixelR3 Armorrender.class Armorhanger.class Armorloadplatform.class outDir");
        }
        Path outDir = Path.of(args[3]);
        Files.createDirectories(outDir);

        byte[] render = patchRenderer(Files.readAllBytes(Path.of(args[0])));
        byte[] hanger = patchTargetBlockEntity(Files.readAllBytes(Path.of(args[1])),
                "net/mcreator/protectionpixel/block/entity/ArmorhangerBlockEntity");
        byte[] platform = patchTargetBlockEntity(Files.readAllBytes(Path.of(args[2])),
                "net/mcreator/protectionpixel/block/entity/ArmorloadplatformBlockEntity");

        Files.write(outDir.resolve("ArmorrenderProcedure.class"), render);
        Files.write(outDir.resolve("ArmorhangerBlockEntity.class"), hanger);
        Files.write(outDir.resolve("ArmorloadplatformBlockEntity.class"), platform);
        System.out.println("renderer=" + render.length + " hanger=" + hanger.length + " platform=" + platform.length);
    }

    private static byte[] patchRenderer(byte[] input) {
        ClassNode cn = read(input);
        if (!"net/mcreator/protectionpixel/procedures/ArmorrenderProcedure".equals(cn.name)) {
            throw new IllegalStateException("unexpected renderer class " + cn.name);
        }
        int replaced = 0;
        for (MethodNode method : cn.methods) {
            for (AbstractInsnNode insn = method.instructions.getFirst(); insn != null; insn = insn.getNext()) {
                if (insn instanceof MethodInsnNode call && call.getOpcode() == Opcodes.INVOKESTATIC &&
                        OLD_HELPER.equals(call.owner) && "targets".equals(call.name) && TARGETS_OLD_DESC.equals(call.desc)) {
                    call.owner = NEW_HELPER;
                    call.desc = TARGETS_NEW_DESC;
                    replaced++;
                }
            }
        }
        if (replaced != 1) {
            throw new IllegalStateException("expected exactly one R2 cache call, found " + replaced);
        }
        return write(cn);
    }

    private static byte[] patchTargetBlockEntity(byte[] input, String expectedName) {
        ClassNode cn = read(input);
        if (!expectedName.equals(cn.name)) {
            throw new IllegalStateException("unexpected block entity class " + cn.name + " expected " + expectedName);
        }
        for (MethodNode method : cn.methods) {
            if ("onLoad".equals(method.name) && "()V".equals(method.desc)) {
                throw new IllegalStateException(expectedName + " already declares onLoad()V");
            }
        }

        MethodNode setRemoved = null;
        for (MethodNode method : cn.methods) {
            if ("m_7651_".equals(method.name) && "()V".equals(method.desc)) {
                if (setRemoved != null) throw new IllegalStateException(expectedName + " has duplicate setRemoved");
                setRemoved = method;
            }
        }
        if (setRemoved == null) {
            throw new IllegalStateException(expectedName + " does not override m_7651_()V/setRemoved");
        }

        int returns = 0;
        List<AbstractInsnNode> returnInsns = new ArrayList<>();
        for (AbstractInsnNode insn = setRemoved.instructions.getFirst(); insn != null; insn = insn.getNext()) {
            if (insn.getOpcode() == Opcodes.RETURN) {
                returns++;
                returnInsns.add(insn);
            }
        }
        if (returns != 1) {
            throw new IllegalStateException(expectedName + " expected one setRemoved RETURN, found " + returns);
        }
        for (AbstractInsnNode ret : returnInsns) {
            setRemoved.instructions.insertBefore(ret, lifecycleCall("unregister"));
        }

        MethodNode onLoad = new MethodNode(Opcodes.ACC_PUBLIC, "onLoad", "()V", null, null);
        onLoad.instructions.add(lifecycleCall("register"));
        onLoad.instructions.add(new InsnNode(Opcodes.RETURN));
        cn.methods.add(onLoad);

        return write(cn);
    }

    private static InsnList lifecycleCall(String methodName) {
        InsnList list = new InsnList();
        list.add(new VarInsnNode(Opcodes.ALOAD, 0));
        list.add(new FieldInsnNode(Opcodes.GETFIELD, BLOCK_ENTITY, "f_58857_", LEVEL_DESC));
        list.add(new VarInsnNode(Opcodes.ALOAD, 0));
        list.add(new FieldInsnNode(Opcodes.GETFIELD, BLOCK_ENTITY, "f_58858_", POS_DESC));
        list.add(new MethodInsnNode(Opcodes.INVOKESTATIC, NEW_HELPER, methodName,
                "(Lnet/minecraft/world/level/Level;Lnet/minecraft/core/BlockPos;)V", false));
        return list;
    }

    private static ClassNode read(byte[] input) {
        ClassNode cn = new ClassNode();
        new ClassReader(input).accept(cn, ClassReader.EXPAND_FRAMES);
        return cn;
    }

    private static byte[] write(ClassNode cn) {
        ClassWriter cw = new ClassWriter(ClassWriter.COMPUTE_MAXS);
        cn.accept(cw);
        return cw.toByteArray();
    }
}
