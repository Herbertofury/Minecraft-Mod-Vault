package mrpg.qa;

import java.lang.instrument.Instrumentation;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;

/** CI-only fallback that re-enters vanilla QuickPlay.startSingleplayer on the Minecraft client thread. */
public final class QuickPlayDriverAgent {
    private QuickPlayDriverAgent() {}

    public static void agentmain(String rawArgs, Instrumentation instrumentation) {
        String[] args = rawArgs == null ? new String[0] : rawArgs.split("\\|", 2);
        if (args.length != 2) {
            throw new IllegalArgumentException("agent args must be <output-path>|<level-name>");
        }
        Path output = Path.of(args[0]).toAbsolutePath();
        String levelName = args[1];
        try {
            Class<?> minecraftClass = findLoaded(instrumentation, "net.minecraft.client.MinecraftClient");
            if (minecraftClass == null) throw new IllegalStateException("MinecraftClient not loaded");
            Object client = minecraftClass.getMethod("getInstance").invoke(null);
            if (client == null) throw new IllegalStateException("MinecraftClient instance is null");

            ClassLoader gameLoader = minecraftClass.getClassLoader();
            Class<?> quickPlayClass = Class.forName("net.minecraft.client.QuickPlay", true, gameLoader);
            Method startSingleplayer = quickPlayClass.getDeclaredMethod("startSingleplayer", minecraftClass, String.class);
            if (!startSingleplayer.trySetAccessible()) {
                throw new IllegalStateException("QuickPlay.startSingleplayer is not reflectively accessible");
            }
            Method execute = minecraftClass.getMethod("execute", Runnable.class);
            Runnable action = () -> {
                try {
                    startSingleplayer.invoke(null, client, levelName);
                    write(output, "[More RPG QA] VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE invoked=true level=" + levelName);
                } catch (Throwable t) {
                    write(output, "[More RPG QA] VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE error=" + sanitize(t.toString()));
                }
            };
            execute.invoke(client, action);
            write(output, "[More RPG QA] VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE scheduled=true level=" + levelName);
        } catch (Throwable t) {
            write(output, "[More RPG QA] VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE setup_error=" + sanitize(t.toString()));
        }
    }

    private static Class<?> findLoaded(Instrumentation instrumentation, String name) {
        for (Class<?> loaded : instrumentation.getAllLoadedClasses()) {
            if (name.equals(loaded.getName())) return loaded;
        }
        return null;
    }

    private static synchronized void write(Path output, String line) {
        try {
            Files.createDirectories(output.getParent());
            Files.writeString(output, line + System.lineSeparator(), StandardCharsets.UTF_8,
                    java.nio.file.StandardOpenOption.CREATE, java.nio.file.StandardOpenOption.APPEND);
        } catch (Throwable failure) {
            System.err.println(line);
            failure.printStackTrace(System.err);
        }
    }

    private static String sanitize(String value) {
        return value.replace('\n', ' ').replace('\r', ' ').replace('\t', ' ');
    }
}
