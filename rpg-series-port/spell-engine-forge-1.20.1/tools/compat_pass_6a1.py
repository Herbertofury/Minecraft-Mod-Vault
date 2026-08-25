#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6a1.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()

common_build = root / 'common/build.gradle'
cs = common_build.read_text()
aw_block = "loom { accessWidenerPath = file('src/main/resources/spell_engine.accesswidener') }"
if aw_block not in cs:
    cs = cs.rstrip() + "\n\n" + aw_block + "\n"
common_build.write_text(cs)

forge = root / 'forge'
forge.joinpath('gradle.properties').write_text('loom.platform=forge\n')
fb = forge.joinpath('build.gradle')
s = fb.read_text()
s = s.replace("accessWidenerPath = project(':common').loom.accessWidenerPath",
              "accessWidenerPath = project(':common').file('src/main/resources/spell_engine.accesswidener')")
s = s.replace("extraAccessWideners.add loom.accessWidenerPath.get().asFile.name",
              "extraAccessWideners.add 'spell_engine.accesswidener'")
fb.write_text(s)

assert (root / 'common/src/main/resources/spell_engine.accesswidener').is_file()
assert aw_block in common_build.read_text()
assert forge.joinpath('gradle.properties').read_text().strip() == 'loom.platform=forge'
assert "project(':common').file('src/main/resources/spell_engine.accesswidener')" in fb.read_text()
assert "extraAccessWideners.add 'spell_engine.accesswidener'" in fb.read_text()
print('Spell Engine compatibility pass 6a1 applied: explicit common/Forge Loom access-widener binding')
