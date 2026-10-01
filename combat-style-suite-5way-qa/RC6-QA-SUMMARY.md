# Combat Style Suite 5.0.0 RC6 QA Summary

RC6 fixes the production-only logout linkage failure discovered by native gameplay QA in RC5. The compile stub now mirrors Forge 47.4.23's `PlayerLoggedOutEvent.getEntity(): Player` descriptor instead of the invalid `Object` descriptor, and verification contains a regression guard.

## Exact candidate
- Director SHA-256: d555543489e6aaae4ee397cec8357aa233b734b7bc61aae28d3568fb889b303d
- Compat SHA-256: 63a7f7ac99a23326edd9872bd23cf6d60af18a92233b489d100f549f903e4db8
- Minecraft: 1.20.1
- Forge: 47.4.23
- YSM exact publisher SHA-256: 25b5e902b96f4c298690208f8b433cbc31737c23f87590354dbd86f00207bc8f

## Native packaged-client gameplay
A disposable QA driver exercised the real packaged Forge client/integrated server with exact Epic Fight, Better Combat, Punchy, TaCZ and the YSM configuration/routing contract shim. Native YSM itself was not modified or bypassed; its protected native core refuses virtual machines.

Final restart/stress run: `CSS5_QA_COMPLETE failures=0 assertions=77 finalRoute=better_combat_ysm_punchy health=HEALTHY: Better Combat + YSM + Punchy`

Covered:
- startup/restart persistence before mutation;
- Vanilla + YSM routing;
- Better Combat + Punchy + YSM routing;
- real Better Combat sword attack damage;
- Epic Fight + YSM ownership and real melee damage;
- Smart Hybrid fusion;
- TaCZ real M4A1 detection and automatic fire/ammo consumption;
- gun passthrough preserving YSM and queued melee profile;
- gun -> melee owner restoration;
- 75 rapid combat-profile changes;
- actual registered F12 quick-profile key path;
- YSM AUTO disable/re-enable transitions;
- perspective/render ownership churn;
- self-check/repair;
- automatic TaCZ fire while requested combat profiles churn;
- TaCZ retaining attack animation/camera/crosshair ownership during active trigger;
- final queued Epic restoration after stressed gun release;
- 80 YSM AUTO config transitions under live routing;
- clean logout and integrated-server shutdown after all assertions.

## Exact upstream strict gate
`verify-release.sh` passed against exact publisher hashes/APIs for Epic Fight 20.14.17, Better Combat 1.9.0, Punchy 2.7d, YSM 2.6.5, and TaCZ 1.1.8 hotfix, plus all prior 5.0/4.x/3.x regressions.

## Real dedicated server with actual YSM
Exact RC6 + exact five real mods + dependencies reached `Done (5.003s)!`, then stopped cleanly and saved Overworld, Nether and End. No task-owned `NoSuchMethodError`, mod-loading failure, mixin failure, or unexpected server exception occurred.

## Remaining native-YSM client boundary
Actual YSM 2.6.5 client rendering/model behavior remains a physical/non-VM client gate because YSM's own protected native core refuses virtual machines. This QA did not patch or defeat that protection and does not mislabel the contract shim as real YSM rendering proof.
