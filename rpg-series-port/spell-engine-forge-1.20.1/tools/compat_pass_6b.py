#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6b.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
java = root / 'forge/src/main/java/net/spell_engine/forge'

def replace(path, old, new):
    p = java / path
    s = p.read_text()
    if old not in s:
        raise SystemExit(f'expected Forge 47 transform missing in {path}: {old}')
    p.write_text(s.replace(old, new))

# Yarn 1.20.1 names this vanilla connection method sendPacket; 1.21 renamed it to send.
replace('PlatformImpl.java', 'player.networkHandler.send(packet);', 'player.networkHandler.sendPacket(packet);')
replace('PlatformClientImpl.java', 'player.networkHandler.send(packet);', 'player.networkHandler.sendPacket(packet);')

# Advancement criteria were not a RegistryKeys registry in 1.20.1. The common target-compatible
# registerCriteria() already uses Criteria.register(...) directly, so invoke it during mod setup rather
# than trying to route it through Forge RegisterEvent.
p = java / 'ForgeMod.java'
s = p.read_text()
s = s.replace('        SpellEngineMod.init();\n        ForgeNetwork.init();',
              '        SpellEngineMod.init();\n        SpellEngineMod.registerCriteria();\n        ForgeNetwork.init();')
s = s.replace('        event.register(RegistryKeys.CRITERION, helper -> SpellEngineMod.registerCriteria());\n', '')
p.write_text(s)

assert 'sendPacket(packet)' in (java / 'PlatformImpl.java').read_text()
assert 'sendPacket(packet)' in (java / 'PlatformClientImpl.java').read_text()
assert 'SpellEngineMod.registerCriteria();' in p.read_text()
assert 'RegistryKeys.CRITERION' not in p.read_text()
print('Spell Engine compatibility pass 6b applied: Forge 47/Yarn 1.20.1 packet + criterion signatures')
