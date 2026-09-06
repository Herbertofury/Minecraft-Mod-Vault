var Opcodes = Java.type('org.objectweb.asm.Opcodes');
var MethodInsnNode = Java.type('org.objectweb.asm.tree.MethodInsnNode');

var MP_OWNER = 'net/solocraft/network/SololevelingModVariables$PlayerVariables';
var BRIDGE = 'com/bert/slrsharedmana/SlrBorrowBridge';

var FULL_TARGETS = [
    'net.solocraft.procedures.DashResetProcedure',
    'net.solocraft.procedures.WillPowerGiveProcedure',
    'net.solocraft.procedures.SwordDanceGiveProcedure',
    'net.solocraft.procedures.FireChargeIncreaseProcedure',
    'net.solocraft.procedures.DetectEyeSpawnProcedure',
    'net.solocraft.procedures.GoliathArmorTickProcedure',
    'net.solocraft.procedures.ARISEProcedure',
    'net.solocraft.procedures.TauntCastProcedure',
    'net.solocraft.procedures.UseSkillOnKeyPressedProcedure',
    'net.solocraft.procedures.ShieldBashProcedure',
    'net.solocraft.procedures.ShadowDeathReviveProcedure',
    'net.solocraft.procedures.UpforceSlashProcedure',
    'net.solocraft.procedures.LllRightclickedProcedure',
    'net.solocraft.procedures.DaggerRushActProcedure',
    'net.solocraft.procedures.SwordOfLightGiveProcedure',
    'net.solocraft.procedures.PhantomLeapAttackProcedure',
    'net.solocraft.util.ColdBloodSkillManager',
    'net.solocraft.util.DaggerThrowManager',
    'net.solocraft.procedures.AriseSkillProcedure',
    'net.solocraft.util.FrostMonarchManager',
    'net.solocraft.util.TankerSkillManager',
    'net.solocraft.util.RangerCombatManager',
    'net.solocraft.util.GoliathCombatManager',
    'net.solocraft.util.FrostArchitectureManager',
    'net.solocraft.util.AssassinSkillManager',
    'net.solocraft.util.BeastMonarchManager',
    'net.solocraft.util.FireMageSpellManager',
    'net.solocraft.util.ShadowMonarchManager',
    'net.solocraft.procedures.ShadowARMORHelmetTickEventProcedure',
    'net.solocraft.util.StormMageSpellManager',
    'net.solocraft.util.RulersAuthorityManager',
    'net.solocraft.util.WhiteFlameMonarchManager',
    'net.solocraft.util.ArcaneMageSpellManager',
    'net.solocraft.procedures.ConsecutiveSlashesOnEffectActiveTickProcedure',
    'net.solocraft.util.BarrierMageSpellManager',
    'net.solocraft.util.LiuZhigangCombatManager',
    'net.solocraft.procedures.DemonKingsLongSwordEntitySwingsItemProcedure',
    'net.solocraft.procedures.HasteBuffCastProcedure',
    'net.solocraft.procedures.PhysicalBuffCastProcedure',
    'net.solocraft.procedures.ManaGunRightclickedProcedure',
    'net.solocraft.procedures.SpiritBowRangedItemUsedProcedure',
    'net.solocraft.procedures.RulersHandProjectileHitsLivingEntityProcedure',
    'net.solocraft.procedures.StormBreatheTickProcedure'
];

var READ_ONLY_TARGETS = [
    'net.solocraft.procedures.SpiritBowCanUseRangedItemProcedure'
];

function transformMpFields(classNode, rewriteWrites) {
    var methods = classNode.methods.iterator();
    while (methods.hasNext()) {
        var method = methods.next();
        var insns = method.instructions.toArray();
        for (var i = 0; i < insns.length; i++) {
            var insn = insns[i];
            var opcode = insn.getOpcode();
            if ((opcode === Opcodes.GETFIELD || opcode === Opcodes.PUTFIELD)
                    && insn.owner === MP_OWNER && insn.name === 'MP' && insn.desc === 'D') {
                if (opcode === Opcodes.GETFIELD) {
                    method.instructions.set(insn, new MethodInsnNode(
                            Opcodes.INVOKESTATIC, BRIDGE, 'readMp', '(Ljava/lang/Object;)D', false));
                } else if (rewriteWrites) {
                    method.instructions.set(insn, new MethodInsnNode(
                            Opcodes.INVOKESTATIC, BRIDGE, 'writeMp', '(Ljava/lang/Object;D)V', false));
                }
            }
        }
    }
    return classNode;
}

function entry(className, rewriteWrites) {
    return {
        'target': {
            'type': 'CLASS',
            'name': className
        },
        'transformer': function(classNode) {
            return transformMpFields(classNode, rewriteWrites);
        }
    };
}

function initializeCoreMod() {
    var result = {};
    for (var i = 0; i < FULL_TARGETS.length; i++) {
        result['slr_mp_spend_' + i] = entry(FULL_TARGETS[i], true);
    }
    for (var j = 0; j < READ_ONLY_TARGETS.length; j++) {
        result['slr_mp_afford_' + j] = entry(READ_ONLY_TARGETS[j], false);
    }
    return result;
}
