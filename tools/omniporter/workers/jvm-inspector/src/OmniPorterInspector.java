import java.io.*;
import java.nio.file.*;
import java.security.*;
import java.util.*;
import java.util.zip.*;

public final class OmniPorterInspector {
    private static String sha256(Path path) throws Exception {
        MessageDigest md = MessageDigest.getInstance("SHA-256");
        try (InputStream in = Files.newInputStream(path)) {
            byte[] buf = new byte[1 << 20];
            int n;
            while ((n = in.read(buf)) > 0) md.update(buf, 0, n);
        }
        StringBuilder sb = new StringBuilder();
        for (byte b : md.digest()) sb.append(String.format("%02x", b));
        return sb.toString();
    }

    private static String json(String s) {
        return "\"" + s.replace("\\", "\\\\").replace("\"", "\\\"").replace("\n", "\\n").replace("\r", "\\r") + "\"";
    }

    private static String readEntry(ZipFile zip, String name) throws IOException {
        ZipEntry e = zip.getEntry(name);
        if (e == null) return null;
        try (InputStream in = zip.getInputStream(e)) {
            return new String(in.readAllBytes(), java.nio.charset.StandardCharsets.UTF_8);
        }
    }

    private static int classMajor(ZipFile zip, ZipEntry e) throws IOException {
        try (DataInputStream in = new DataInputStream(zip.getInputStream(e))) {
            if (in.readInt() != 0xCAFEBABE) return -1;
            in.readUnsignedShort();
            return in.readUnsignedShort();
        }
    }

    public static void main(String[] args) throws Exception {
        if (args.length != 1) {
            System.err.println("usage: java -jar omniporter-jvm-inspector.jar <mod.jar>");
            System.exit(2);
        }
        Path jar = Path.of(args[0]).toAbsolutePath().normalize();
        if (!Files.isRegularFile(jar)) throw new FileNotFoundException(jar.toString());

        List<String> loaderSignals = new ArrayList<>();
        List<String> mixinConfigs = new ArrayList<>();
        Set<Integer> classMajors = new TreeSet<>();
        int classCount = 0;
        int signedEntries = 0;
        int nestedJars = 0;
        boolean hasAT = false;
        boolean hasAW = false;
        boolean hasClassTweaker = false;

        String fabricMod, forgeMods, neoMods, quiltMod, manifest;
        try (ZipFile zip = new ZipFile(jar.toFile())) {
            fabricMod = readEntry(zip, "fabric.mod.json");
            quiltMod = readEntry(zip, "quilt.mod.json");
            forgeMods = readEntry(zip, "META-INF/mods.toml");
            neoMods = readEntry(zip, "META-INF/neoforge.mods.toml");
            manifest = readEntry(zip, "META-INF/MANIFEST.MF");
            if (fabricMod != null) loaderSignals.add("fabric");
            if (quiltMod != null) loaderSignals.add("quilt");
            if (forgeMods != null) loaderSignals.add("forge");
            if (neoMods != null) loaderSignals.add("neoforge");

            Enumeration<? extends ZipEntry> en = zip.entries();
            while (en.hasMoreElements()) {
                ZipEntry e = en.nextElement();
                String n = e.getName();
                String lower = n.toLowerCase(Locale.ROOT);
                if (n.endsWith(".class") && !e.isDirectory()) {
                    classCount++;
                    int major = classMajor(zip, e);
                    if (major > 0) classMajors.add(major);
                }
                if (lower.endsWith(".mixins.json") || (lower.contains("mixin") && lower.endsWith(".json"))) mixinConfigs.add(n);
                if (lower.endsWith("accesstransformer.cfg")) hasAT = true;
                if (lower.endsWith(".accesswidener")) hasAW = true;
                if (lower.endsWith(".classtweaker")) hasClassTweaker = true;
                if (lower.startsWith("meta-inf/") && (lower.endsWith(".sf") || lower.endsWith(".rsa") || lower.endsWith(".dsa"))) signedEntries++;
                if (lower.endsWith(".jar") && !e.isDirectory()) nestedJars++;
            }
        }

        Collections.sort(mixinConfigs);
        StringBuilder out = new StringBuilder();
        out.append("{\n");
        out.append("  \"path\": ").append(json(jar.toString())).append(",\n");
        out.append("  \"size\": ").append(Files.size(jar)).append(",\n");
        out.append("  \"sha256\": ").append(json(sha256(jar))).append(",\n");
        out.append("  \"loaderSignals\": ").append(loaderSignals.toString().replace(" ", "")).append(",\n");
        out.append("  \"classCount\": ").append(classCount).append(",\n");
        out.append("  \"classMajors\": ").append(classMajors).append(",\n");
        out.append("  \"mixinConfigs\": [");
        for (int i = 0; i < mixinConfigs.size(); i++) { if (i > 0) out.append(','); out.append(json(mixinConfigs.get(i))); }
        out.append("],\n");
        out.append("  \"hasAccessTransformer\": ").append(hasAT).append(",\n");
        out.append("  \"hasAccessWidener\": ").append(hasAW).append(",\n");
        out.append("  \"hasClassTweaker\": ").append(hasClassTweaker).append(",\n");
        out.append("  \"nestedJarCount\": ").append(nestedJars).append(",\n");
        out.append("  \"signatureEntryCount\": ").append(signedEntries).append(",\n");
        out.append("  \"metadata\": {\n");
        out.append("    \"fabric.mod.json\": ").append(fabricMod == null ? "null" : json(fabricMod)).append(",\n");
        out.append("    \"quilt.mod.json\": ").append(quiltMod == null ? "null" : json(quiltMod)).append(",\n");
        out.append("    \"META-INF/mods.toml\": ").append(forgeMods == null ? "null" : json(forgeMods)).append(",\n");
        out.append("    \"META-INF/neoforge.mods.toml\": ").append(neoMods == null ? "null" : json(neoMods)).append(",\n");
        out.append("    \"META-INF/MANIFEST.MF\": ").append(manifest == null ? "null" : json(manifest)).append("\n");
        out.append("  }\n");
        out.append("}\n");
        System.out.print(out);
    }
}
