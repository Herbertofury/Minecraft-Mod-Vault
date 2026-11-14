package mrpg.qa;

import com.sun.tools.attach.VirtualMachine;

/** Loads StateAgent into the already-running mapped Minecraft client JVM. */
public final class AttachMain {
    private AttachMain() {}

    public static void main(String[] args) throws Exception {
        if (args.length != 3) {
            throw new IllegalArgumentException("usage: AttachMain <pid> <agent-jar> <output-path>");
        }
        VirtualMachine vm = VirtualMachine.attach(args[0]);
        try {
            vm.loadAgent(args[1], args[2]);
        } finally {
            vm.detach();
        }
    }
}
