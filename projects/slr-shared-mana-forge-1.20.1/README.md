# SLR Shared Mana — Forge 1.20.1

Clean-room mana compatibility bridge for **Solo Leveling: Reawakening (SLR)** and **Iron's Spells 'n Spellbooks**, designed to coexist with **Iron's Botany 2.0.1**.

## Modes

The common config is `config/slr-shared-mana.toml`.

### Classic shared pool — default

```toml
[shared_mana]
enabled = true
separatePoolsBorrowFromIron = false
```

This is the original v1.0 behavior and remains the default:

- SLR `MP` is authoritative.
- Iron mana reads/spends resolve from SLR MP on demand; there is **no player-tick mirror**.
- Iron mana writes reuse ISS's own cancellable `ChangeManaEvent`, then apply the accepted delta to SLR exactly once.
- Native ISS passive mana regeneration is disabled while sharing is active; SLR regeneration remains authoritative.
- Intentional positive ISS mana changes can feed SLR MP, preserving Iron's Botany `ISS_PRIMARY`/bidirectional behavior.
- Iron's mana HUD is hidden by default so SLR is the single visible pool.

### Separate pools + emergency SLR borrowing — opt in

```toml
[shared_mana]
enabled = true
separatePoolsBorrowFromIron = true
```

In this mode:

- SLR keeps its own native MP, regen, rewards, potions, resets and HUD.
- Iron keeps its own native mana, regen, `ChangeManaEvent`, max-mana rules and HUD.
- An SLR skill spends SLR MP first.
- Only if that real SLR spend is short does the missing amount come from Iron, using `slrMpPerIronMana`.
- No mana moves while idle. There is no tick poll, mirror map, nearby scan or passive drain.
- Positive SLR gains stay SLR-only; they never credit Iron.
- The bytecode bridge is intentionally restricted to verified SLR spend paths plus the Spirit Bow's pre-use affordability gate. SLR regen/reward/reset/HUD code is not virtualized.

Set `enabled = false` to disable both bridge modes.

## Iron's Botany

This mod does **not** add a second competing Botania router. In classic shared mode Iron's Botany retains ownership of its normal `manaUnificationMode`. In separate-pools mode Iron's side stays native, and SLR only borrows an actual spending deficit.

## Command

`/slrmana status` reports the active mode, SLR MP, Iron mana, ratio and detected Iron's Botany mode. In separate mode the Iron number is the actual native Iron pool.

## Performance design

The bridge is event/call driven. It does not run a new per-player tick loop. Owner binding occurs on login, respawn and dimension changes and is held with weak references; selected SLR mana field accesses are transformed once at class load.
