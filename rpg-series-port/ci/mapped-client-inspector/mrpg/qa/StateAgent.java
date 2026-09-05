package mrpg.qa;

import java.lang.instrument.Instrumentation;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;

/** CI-only mapped-userdev state probe. This class is never packaged into a mod JAR. */
public final class StateAgent {
    private StateAgent() {}

    public static void agentmain(String outputPath, Instrumentation instrumentation) {
        String state;
        try {
            Class<?> minecraftClass = null;
            for (Class<?> loaded : instrumentation.getAllLoadedClasses()) {
                if ("net.minecraft.client.MinecraftClient".equals(loaded.getName())) {
                    minecraftClass = loaded;
                    break;
                }
            }
            if (minecraftClass == null) {
                state = "[More RPG QA] MAPPED_CLIENT_STATE error=MinecraftClient_not_loaded";
            } else {
                Method getInstance = minecraftClass.getMethod("getInstance");
                Object client = getInstance.invoke(null);
                if (client == null) {
                    state = "[More RPG QA] MAPPED_CLIENT_STATE error=MinecraftClient_instance_null";
                } else {
                    Object screen = firstFieldValue(minecraftClass, client, "currentScreen", "screen");
                    Object world = firstFieldValue(minecraftClass, client, "world", "level");
                    Object player = firstFieldValue(minecraftClass, client, "player");
                    String title = screen == null ? "<none>" : screenTitle(screen);
                    boolean integrated = invokeBoolean(client, minecraftClass,
                            "isIntegratedServerRunning", "hasSingleplayerServer", "isInSingleplayer");
                    state = "[More RPG QA] MAPPED_CLIENT_STATE"
                            + " screen=" + typeName(screen)
                            + " title=" + sanitize(title)
                            + " world=" + typeName(world)
                            + " player=" + typeName(player)
                            + " integratedServer=" + integrated;
                }
            }
        } catch (Throwable t) {
            state = "[More RPG QA] MAPPED_CLIENT_STATE error="
                    + sanitize(t.getClass().getName() + ":" + t.getMessage());
        }

        try {
            Path out = Path.of(outputPath).toAbsolutePath();
            Files.createDirectories(out.getParent());
            Files.writeString(out, state + System.lineSeparator(), StandardCharsets.UTF_8);
        } catch (Throwable writeFailure) {
            System.err.println(state);
            writeFailure.printStackTrace(System.err);
        }
    }

    private static Object firstFieldValue(Class<?> type, Object instance, String... names) {
        for (String name : names) {
            try {
                Field field = type.getField(name);
                return field.get(instance);
            } catch (ReflectiveOperationException ignored) {
                try {
                    Field field = type.getDeclaredField(name);
                    if (field.trySetAccessible()) {
                        return field.get(instance);
                    }
                } catch (ReflectiveOperationException ignoredAgain) {
                    // Try the next mapped alias.
                }
            }
        }
        return null;
    }

    private static boolean invokeBoolean(Object instance, Class<?> type, String... names) {
        for (String name : names) {
            try {
                Method method = type.getMethod(name);
                Object value = method.invoke(instance);
                if (value instanceof Boolean) return (Boolean) value;
            } catch (ReflectiveOperationException ignored) {
                // Try the next mapped alias.
            }
        }
        return false;
    }

    private static String screenTitle(Object screen) {
        for (String name : new String[]{"getTitle", "getNarratedTitle"}) {
            try {
                Method method = screen.getClass().getMethod(name);
                Object value = method.invoke(screen);
                if (value != null) return value.toString();
            } catch (ReflectiveOperationException ignored) {
                // Best-effort title only; the concrete screen class remains authoritative.
            }
        }
        return "<unavailable>";
    }

    private static String typeName(Object value) {
        return value == null ? "<null>" : value.getClass().getName();
    }

    private static String sanitize(String value) {
        if (value == null) return "<null>";
        return value.replace('\n', ' ').replace('\r', ' ').replace('\t', ' ');
    }
}
