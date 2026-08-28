# Rogues 3.1.1 Native Forge 1.20.1 - Runtime QA Contract

This contract is derived from the pinned current 3.1.1 authority at `d4a7af565559dcff4384eabb2481f63eb5f97d55`. It is intentionally stricter than compile/package success and must be executed against the exact packaged release before Rogues can graduate.

## Baseline exact-head gates

One exact head must prove all of the following before behavior assertions are accepted:

- immutable current/historical source pins and deterministic materialization;
- current-source common + native Forge compilation;
- remapped Forge 1.20.1 release packaging;
- owned Java classes major 61 and no newer bytecode leakage;
- separate foundation JARs with anti-shading/dependency-leakage checks;
- deterministic/certified release identity and deterministic source archive;
- dedicated Forge dev-server readiness;
- native Forge client LWJGL/resource/render bootstrap;
- fresh official Forge 47.4.23 packaged-server readiness with the exact Rogues release and separate runtime dependencies.

## Real-player action-impairment matrix

Use an exact certified packaged Forge server plus a genuine native Forge client/player. Drive movement, jump, attack and item-use input through a real client input path, with positive control phases after each impairment is removed.

### SHOCK = full STUN

Current source configures `rogues:shock` with `EntityActionsAllowed.STUN`.

Required evidence:

- movement blocked while SHOCK is active;
- jump blocked while SHOCK is active;
- melee attack blocked while SHOCK is active;
- item use blocked while SHOCK is active;
- after SHOCK is removed, identical input must move the player and damage/use normally.

Candidate markers:

- `ROGUES_SHOCK_STUN_MOVE_BLOCK_PASS`
- `ROGUES_SHOCK_STUN_ATTACK_BLOCK_PASS`
- `ROGUES_SHOCK_STUN_ITEM_USE_BLOCK_PASS`
- `ROGUES_SHOCK_CONTROL_MOVE_PASS`
- `ROGUES_SHOCK_CONTROL_ATTACK_PASS`

### BEAR_TRAP = ROOT, not STUN

Current source defines ROOT as `canMove=false`, `canJump=false`, while player attack, item use and spell casting remain allowed.

Required evidence:

- movement blocked under `rogues:bear_trap`;
- jump blocked under `rogues:bear_trap`;
- melee attack remains functional against a deterministic target;
- item use remains functional;
- if a deterministic cast fixture is available, casting remains functional;
- removal restores normal movement.

A test that only proves immobility is insufficient: it must positively prove at least attack and item-use remain allowed so ROOT cannot silently regress into STUN.

Candidate markers:

- `ROGUES_BEAR_ROOT_MOVE_BLOCK_PASS`
- `ROGUES_BEAR_ROOT_ATTACK_ALLOWED_PASS`
- `ROGUES_BEAR_ROOT_ITEM_USE_ALLOWED_PASS`
- `ROGUES_BEAR_ROOT_CONTROL_MOVE_PASS`

### NET_TRAP = ROOT + knockback immunity

Repeat the BEAR_TRAP ROOT matrix for `rogues:net_trap`, then apply deterministic knockback while trapped and after removal.

Required evidence:

- movement/jump blocked;
- attack and item use still allowed;
- external knockback does not displace the trapped player beyond a tight tolerance;
- identical knockback after removal measurably displaces the player.

Candidate markers:

- `ROGUES_NET_ROOT_MOVE_BLOCK_PASS`
- `ROGUES_NET_ROOT_ATTACK_ALLOWED_PASS`
- `ROGUES_NET_ROOT_ITEM_USE_ALLOWED_PASS`
- `ROGUES_NET_KNOCKBACK_IMMUNITY_PASS`
- `ROGUES_NET_KNOCKBACK_CONTROL_PASS`

## STEALTH lifecycle matrix

Current source configures `rogues:stealth` as synchronized, tinted, reduced enemy detection, `RemoveOnHit.ANY_HIT`, and additionally removes it when the stealth user attacks, uses an item, or casts any spell other than `rogues:vanish`. Removal triggers `StealthEffect.onRemove` and removes `stealth_speed` if present.

Required server/player evidence must independently prove these branches rather than conflating them:

1. **Incoming hit:** apply STEALTH, damage the player, require STEALTH removed.
2. **Outgoing attack:** apply STEALTH, have the player attack a deterministic target, require STEALTH removed.
3. **Item use:** apply STEALTH, perform a deterministic usable-item action, require STEALTH removed.
4. **Non-Vanish spell cast:** when a deterministic spell-cast fixture is available, apply STEALTH, cast a non-`rogues:vanish` spell, require STEALTH removed.
5. **Vanish exemption:** where the exact Vanish cast fixture is available, casting `rogues:vanish` must not immediately remove STEALTH through the non-Vanish hook.
6. **Cleanup coupling:** apply STEALTH + `stealth_speed`; remove STEALTH through one supported path; require `stealth_speed` also removed.
7. **Visibility/detection:** prove the target-acquisition distance is reduced according to current tweak configuration, using deterministic hostile-mob positions and positive control without STEALTH if this can be made stable on the packaged server.

Candidate markers:

- `ROGUES_STEALTH_REMOVE_ON_INCOMING_HIT_PASS`
- `ROGUES_STEALTH_REMOVE_ON_ATTACK_PASS`
- `ROGUES_STEALTH_REMOVE_ON_ITEM_USE_PASS`
- `ROGUES_STEALTH_REMOVE_ON_NONVANISH_CAST_PASS`
- `ROGUES_STEALTH_VANISH_EXEMPT_PASS`
- `ROGUES_STEALTH_SPEED_CLEANUP_PASS`
- `ROGUES_STEALTH_DETECTION_SCALE_PASS`

## CHARGE cadence

Current `rogues:charge` is a `TickingStatusEffect` with interval `5` so downstream spell/tree logic can react every quarter second.

Require either a direct game-thread instrumentation/self-test or deterministic runtime trigger evidence proving the effect callback cadence is exactly every 5 ticks, not every tick and not an older interval.

Candidate marker: `ROGUES_CHARGE_5TICK_CADENCE_PASS`.

## LAST_STAND current fortification

Current source documents per-stack semantics up to five stacks: +20% max health per stack and current damage-taken fortification supplied by the modern tree/spell path. The 1.20.1 port must preserve the current 3.1.1 behavior, not silently fall back to historical 1.2.0 behavior.

Require deterministic health/damage measurements for at least one stack and full five-stack state, with removal returning to baseline. Exact expected values must be derived from the materialized current spell/tree configuration at implementation time rather than guessed from historical code.

Candidate markers:

- `ROGUES_LAST_STAND_STACK1_PASS`
- `ROGUES_LAST_STAND_STACK5_PASS`
- `ROGUES_LAST_STAND_REMOVE_CONTROL_PASS`

## Arms merchant and trade parity

On a real game thread, require current POI/profession registration and all five current trade tiers/economics to be present. Do not accept merely finding JSON or Java source strings.

Candidate marker: `ROGUES_ARMS_MERCHANT_CURRENT_TRADES_PASS`.

## Strength rebalance

When the current tweak is enabled, prove the vanilla Strength attack modifier is altered exactly as current 3.1.1 intends, then prove disabling the tweak preserves vanilla behavior. This must be a reversible runtime/config test, not a source grep.

Candidate markers:

- `ROGUES_STRENGTH_REBALANCE_ENABLED_PASS`
- `ROGUES_STRENGTH_REBALANCE_DISABLED_CONTROL_PASS`

## Graduation marker

Only emit `ROGUES_FULL_DEEP_BEHAVIOR_PASS` after the exact-head baseline runtime matrix and every applicable current-semantics assertion above have passed on the certified packaged release. Unsupported or unstable fixtures must be improved; assertions must not be weakened merely to obtain green CI.
