import java.nio.file.*;
import java.util.*;
import jdk.internal.org.objectweb.asm.*;
import jdk.internal.org.objectweb.asm.tree.*;

public final class PatchProtectionPixel {
    private static final String TARGET = "net/mcreator/protectionpixel/procedures/ArmorrenderProcedure";
    private static final String HELPER = "net/mcreator/protectionpixel/repair/ArmorRenderTargetCache";
    private static final String EXEC_DESC = "(Lnet/minecraftforge/eventbus/api/Event;Lnet/minecraft/world/level/LevelAccessor;)V";

    public static void main(String[] args) throws Exception {
        if (args.length != 2) throw new IllegalArgumentException("usage: PatchProtectionPixel in.class out.class");
        byte[] in = Files.readAllBytes(Path.of(args[0]));
        ClassNode cn = new ClassNode();
        new ClassReader(in).accept(cn, ClassReader.EXPAND_FRAMES);
        if (!TARGET.equals(cn.name)) throw new IllegalStateException("wrong class: " + cn.name);

        MethodNode target = null;
        for (MethodNode m : cn.methods) {
            if ("execute".equals(m.name) && EXEC_DESC.equals(m.desc) && (m.access & Opcodes.ACC_STATIC) != 0) {
                target = m;
                break;
            }
        }
        if (target == null) throw new IllegalStateException("private execute(Event,LevelAccessor) not found");

        AbstractInsnNode start = null;
        AbstractInsnNode prepEnd = null;
        JumpInsnNode iteratorEmptyJump = null;
        JumpInsnNode initialNotClientJump = null;

        // Initial type guard target is the method's final return label.
        for (AbstractInsnNode n = target.instructions.getFirst(); n != null; n = n.getNext()) {
            if (n instanceof TypeInsnNode t && t.getOpcode() == Opcodes.INSTANCEOF &&
                    "net/minecraft/client/multiplayer/ClientLevel".equals(t.desc)) {
                AbstractInsnNode p = nextReal(n.getNext());
                if (p instanceof JumpInsnNode j && j.getOpcode() == Opcodes.IFEQ) {
                    initialNotClientJump = j;
                    break;
                }
            }
        }
        if (initialNotClientJump == null) throw new IllegalStateException("client-level guard not found");
        LabelNode finalReturnLabel = initialNotClientJump.label;

        // Scan begins with: iload 4; ineg; istore 11, after the player BlockPos was stored in local 5.
        for (AbstractInsnNode n = target.instructions.getFirst(); n != null; n = n.getNext()) {
            if (n instanceof VarInsnNode v && v.getOpcode() == Opcodes.ILOAD && v.var == 4) {
                AbstractInsnNode a = nextReal(n.getNext());
                AbstractInsnNode b = a == null ? null : nextReal(a.getNext());
                if (a != null && a.getOpcode() == Opcodes.INEG && b instanceof VarInsnNode s && s.getOpcode() == Opcodes.ISTORE && s.var == 11) {
                    start = n;
                    break;
                }
            }
        }
        if (start == null) throw new IllegalStateException("scan start not found");

        // Upstream scan setup ends at the iterator stored into local 13.
        for (AbstractInsnNode n = start; n != null; n = n.getNext()) {
            if (n instanceof VarInsnNode v && v.getOpcode() == Opcodes.ASTORE && v.var == 13) {
                prepEnd = n;
                break;
            }
        }
        if (prepEnd == null) throw new IllegalStateException("scan iterator store not found");

        // Iterator-empty branch marks the end of the per-entry render loop.
        for (AbstractInsnNode n = prepEnd.getNext(); n != null && n != finalReturnLabel; n = n.getNext()) {
            if (n instanceof MethodInsnNode m && m.getOpcode() == Opcodes.INVOKEINTERFACE &&
                    "java/util/Iterator".equals(m.owner) && "hasNext".equals(m.name) && "()Z".equals(m.desc)) {
                AbstractInsnNode jn = nextReal(n.getNext());
                if (jn instanceof JumpInsnNode j && j.getOpcode() == Opcodes.IFEQ) {
                    iteratorEmptyJump = j;
                    break;
                }
            }
        }
        if (iteratorEmptyJump == null) throw new IllegalStateException("iterator-empty branch not found");
        LabelNode loopEndLabel = iteratorEmptyJump.label;

        // Replace the every-frame render-distance chunk scan with a cached, same-order target map.
        InsnList repl = new InsnList();
        // Preserve locals required by the existing StackMap frame at the iterator loop header.
        repl.add(new InsnNode(Opcodes.ACONST_NULL));
        repl.add(new VarInsnNode(Opcodes.ASTORE, 6));
        repl.add(new InsnNode(Opcodes.ICONST_0));
        repl.add(new VarInsnNode(Opcodes.ISTORE, 11));
        repl.add(new InsnNode(Opcodes.ICONST_0));
        repl.add(new VarInsnNode(Opcodes.ISTORE, 12));
        repl.add(new VarInsnNode(Opcodes.ALOAD, 3));
        repl.add(new VarInsnNode(Opcodes.ALOAD, 5));
        repl.add(new VarInsnNode(Opcodes.ILOAD, 4));
        repl.add(new MethodInsnNode(Opcodes.INVOKESTATIC, HELPER, "targets",
                "(Lnet/minecraft/client/multiplayer/ClientLevel;Lnet/minecraft/core/BlockPos;I)Ljava/util/Map;", false));
        repl.add(new MethodInsnNode(Opcodes.INVOKEINTERFACE, "java/util/Map", "entrySet", "()Ljava/util/Set;", true));
        repl.add(new MethodInsnNode(Opcodes.INVOKEINTERFACE, "java/util/Set", "iterator", "()Ljava/util/Iterator;", true));
        repl.add(new VarInsnNode(Opcodes.ASTORE, 13));
        target.instructions.insertBefore(start, repl);
        removeInclusive(target.instructions, start, prepEnd);

        // The original code incremented inner/outer chunk-loop locals after iterator exhaustion.
        // Those loops no longer exist; preserve the loop-end frame then jump to the original return.
        AbstractInsnNode firstRealAfterLoopEnd = nextReal(loopEndLabel.getNext());
        if (firstRealAfterLoopEnd == null) throw new IllegalStateException("loop end has no instructions");
        AbstractInsnNode cur = firstRealAfterLoopEnd;
        while (cur != null && cur != finalReturnLabel) {
            AbstractInsnNode next = cur.getNext();
            target.instructions.remove(cur);
            cur = next;
        }
        target.instructions.insertBefore(finalReturnLabel, new JumpInsnNode(Opcodes.GOTO, finalReturnLabel));

        ClassWriter cw = new ClassWriter(ClassWriter.COMPUTE_MAXS);
        cn.accept(cw);
        byte[] out = cw.toByteArray();
        Files.write(Path.of(args[1]), out);
        System.out.println("patched bytes=" + out.length);
    }

    private static AbstractInsnNode nextReal(AbstractInsnNode n) {
        while (n != null && n.getOpcode() < 0) n = n.getNext();
        return n;
    }

    private static void removeInclusive(InsnList insns, AbstractInsnNode first, AbstractInsnNode last) {
        AbstractInsnNode n = first;
        while (n != null) {
            AbstractInsnNode next = n.getNext();
            insns.remove(n);
            if (n == last) return;
            n = next;
        }
        throw new IllegalStateException("range end not reached");
    }
}
