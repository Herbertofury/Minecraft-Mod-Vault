from pathlib import Path

path = Path("sable/buildSrc/src/main/java/RapierProductionJarMapper.java")
text = path.read_text(encoding="utf-8")
old = '''        private boolean resolvesNamedMember(String owner, String name, String descriptor, boolean method) {
            if (method ? declaresMethod(owner, name, descriptor) : declaresField(owner, name, descriptor)) {
                return true;
            }
            for (String supertype : supertypes(owner)) {
'''
new = '''        private boolean resolvesNamedMember(String owner, String name, String descriptor, boolean method) {
            if (method ? declaresMethod(owner, name, descriptor) : declaresField(owner, name, descriptor)) {
                return true;
            }
            ClassInfo direct = classes.get(owner);
            if (method && direct != null && "java/lang/Enum".equals(direct.superName)
                    && "ordinal".equals(name) && "()I".equals(descriptor)) {
                // JVM method resolution walks into java.lang.Enum. The verifier deliberately indexes
                // only Minecraft/Forge/Sable bytecode, so prove this one inherited JDK enum method
                // structurally instead of treating it as a missing packaged Sable method.
                return true;
            }
            for (String supertype : supertypes(owner)) {
'''
if old not in text:
    raise SystemExit("resolvesNamedMember anchor not found")
path.write_text(text.replace(old, new, 1), encoding="utf-8")
print("Patched cross-artifact ABI verifier for structurally proven java.lang.Enum.ordinal() inheritance")
