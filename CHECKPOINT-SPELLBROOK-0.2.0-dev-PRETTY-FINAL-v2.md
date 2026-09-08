# Spellbrook Forge 1.20.1 — 0.2.0-dev PRETTY FINAL v2

Status: **PASS — Glimmer Store alignment repair + native broom showcase checkpoint**

## Final release

- JAR: `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL-v2.jar`
- Size: **6,930,800 bytes**
- SHA-256: `6282f05de62545838fe2798d364fcf28f52fbb3c3900b2e9168f352e009bf493`
- Drive file ID: `1pmKgVMdkVuSwkv56K7Li07jXVIMqHCGc`

Source:
- `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL-v2-SOURCE.zip`
- SHA-256: `f11f48a3ac676370a068736e520d19222200a8daa53ef48b0aa42ab5086c1fe6`
- Drive file ID: `1Ik9u40vjehfPQVRuG7unXFVHUyfDKW9I`

Evidence:
- `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL-v2-EVIDENCE.zip`
- SHA-256: `815caac26dcdb1b6fe2a832868d58b2902e6f21b1b3cea89ddcf19c1e4f6260a`
- Drive file ID: `1hw5Xn8s4icA2l4WbOFjHH_p-e4XVtjXM`

Verification: Drive file ID `1gKB3gg5XtVu-8qU1RY4kqoGYcfdDxkub`
SHA manifest: Drive file ID `1osWC7hfTPVO4bAWOcBBbsH416M-0swGO`

Canonical Drive folder: `Spellbrook Forge 1.20.1` / `1xA5xcdiMVD3kDo0UD5l4ouI_goWMsZFg`

## Root-cause repair

The Glimmer Store artwork has a literal **5-column x 6-row** item grid with an **18 px cell pitch**. The prior renderer treated the same 30 items as **6 columns x 5 rows at 15 px pitch**, so the icon positions progressively drifted away from the painted boxes.

Final geometry:
- texture-relative grid origin `(105, 50)`
- 5 columns x 6 rows
- 18 px cell pitch
- 14 px visual item size centered in each painted cell
- hover fill confined to each painted cell
- page dots moved to the footer strip

## Native QA

Real Forge client screenshots:
- Featured fixed — Drive file ID `1_Je_EKJZw5BjJ2My-IFfNKmxj8AEydW1`
- Keys & Mysteries fixed — Drive file ID `1PJTx7rv70I_hK3sRcxMJFzUOR6HMxT-Z`

Both pages render icons box-for-box in the actual 5x6 painted slots.

Real Minecraft broom GIFs — **no generated imagery**:
- Sakura Broom Flight — Drive file ID `1kdLkc44mLPF2aD9Tg-oA5aoWvCvA4qSL`
- Dragon Broom Boost/Climb/Turn — Drive file ID `1KiTzM8Fclb2J0PrwfowXVEjI3zRR2rlJ`

The native client log records real `spellbrook:broom` summon and mount events for both recording passes.

## Gates

- changed-path `compileJava`: PASS
- real Forge integrated client: PASS
- Featured Store native visual gate: PASS
- Keys & Mysteries native visual gate: PASS
- Sakura broom native flight capture: PASS
- Dragon broom boost/climb/turn native capture: PASS
- clean Save & Quit: PASS; all dimensions saved
- task-owned runtime scan: 0 Spellbrook errors, 0 missing Spellbrook models/textures, 0 DecoderException, 0 EncoderException, 0 wrong-side/invalid-message failures
- normal product `build`: PASS, including `reobfJar`; QA-only dependency alias omitted
- final JAR archive integrity: PASS

The only remaining warnings in the restored visual harness are the already-known missing external Mojang sound asset objects; they are not part of the shipped Spellbrook artifact.
