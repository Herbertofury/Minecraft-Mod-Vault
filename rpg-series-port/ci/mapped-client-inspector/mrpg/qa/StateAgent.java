package mrpg.qa;

import java.lang.instrument.Instrumentation;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Collections;
import java.util.List;

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
                    StringBuilder out = new StringBuilder("[More RPG QA] MAPPED_CLIENT_STATE")
                            .append(" screen=").append(typeName(screen))
                            .append(" title=").append(sanitize(title))
                            .append(" world=").append(typeName(world))
                            .append(" player=").append(typeName(player))
                            .append(" integratedServer=").append(integrated);
                    if (screen != null && "net.minecraftforge.client.gui.LoadingErrorScreen".equals(screen.getClass().getName())) {
                        appendForgeLoadingDetails(out, screen);
                    }
                    state = out.toString();
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

    private static void appendForgeLoadingDetails(StringBuilder out, Object screen) {
        List<?> errors = listFieldValue(screen, "modLoadErrors");
        List<?> warnings = listFieldValue(screen, "modLoadWarnings");
        Object dumpedLocation = firstFieldValue(screen.getClass(), screen, "dumpedLocation");
        out.append(" loadingErrors=").append(errors.size())
                .append(" loadingWarnings=").append(warnings.size())
                .append(" dumpedLocation=").append(sanitize(String.valueOf(dumpedLocation)));
        appendFormattedEntries(out, "error", errors);
        appendFormattedEntries(out, "warning", warnings);
    }

    private static List<?> listFieldValue(Object instance, String name) {
        Object value = firstFieldValue(instance.getClass(), instance, name);
        return value instanceof List<?> ? (List<?>) value : Collections.emptyList();
    }

    private static void appendFormattedEntries(StringBuilder out, String kind, List<?> values) {
        int limit = Math.min(values.size(), 16);
        for (int i = 0; i < limit; i++) {
            Object value = values.get(i);
            out.append(" ").append(kind).append("[").append(i).append("]=")
                    .append(sanitize(formatLoadingEntry(value)));
        }
        if (values.size() > limit) {
            out.append(" ").append(kind).append("Truncated=").append(values.size() - limit);
        }
    }

    private static String formatLoadingEntry(Object value) {
        if (value == null) return "<null>";
        try {
            Method method = value.getClass().getMethod("formatToString");
            Object formatted = method.invoke(value);
            return String.valueOf(formatted);
        } catch (ReflectiveOperationException ignored) {
            return String.valueOf(value);
        }
    }

    private static Object firstFieldValue(Class<?> type, Object instance, String... names) {
        for (String name : names) {
            Class<?> cursor = type;
            while (cursor != null) {
                try {
                    Field field = cursor.getDeclaredField(name);
                    if (field.trySetAccessible()) {
                        return field.get(instance);
                    }
                } catch (ReflectiveOperationException ignored) {
                    // Try the superclass, then the next mapped alias.
                }
                cursor = cursor.getSuperclass();
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
        return value.replace('\n', ' ').replace('\r', ' ').replace('\t', ' ').replace('|', '/');
    }
}
