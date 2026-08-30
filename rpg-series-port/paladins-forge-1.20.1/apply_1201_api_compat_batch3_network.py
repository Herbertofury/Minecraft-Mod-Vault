#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

root = pathlib.Path(sys.argv[1]).resolve()
path = root / "net/paladins/entity/BarrierEntity.java"
text = path.read_text(encoding="utf-8")
old = '''                                serverPlayer.networkHandler.send(
                                        new EntityVelocityUpdateS2CPacket(serverPlayer.getId(), serverPlayer.getVelocity())
                                );'''
new = '''                                serverPlayer.networkHandler.sendPacket(
                                        new EntityVelocityUpdateS2CPacket(serverPlayer.getId(), serverPlayer.getVelocity())
                                );'''
count = text.count(old)
if count != 1:
    raise SystemExit(f"[Barrier sendPacket] expected exactly one batch2 packet-send shape, found {count}")
path.write_text(text.replace(old, new, 1), encoding="utf-8")
if ".networkHandler.send(" in path.read_text(encoding="utf-8"):
    raise SystemExit("[Barrier sendPacket] later ServerPlayNetworkHandler.send API survived target translation")
print("[Paladins 1.20.1 API batch3] Barrier ServerPlayNetworkHandler.send -> sendPacket (Yarn 1.20.1 exact API)")
