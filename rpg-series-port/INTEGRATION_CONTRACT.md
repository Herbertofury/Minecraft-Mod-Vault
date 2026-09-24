# RPG Series native Forge 1.20.1 integration contract

Status: **ACTIVE / NOT PROMOTED** until the exact integration branch head passes the full combined native runtime lane and its evidence is persisted and byte-verified.

## Frozen leaf authorities

Integration is not permission to rebuild the graduated leaves into new identities.

- Paladins 3.1.1 runtime authority: `257171c71a363285bb5bdbb58083121f3a7456d3`
  - release SHA-256: `95e8f9e074dd1c432f9486c922173b76a3b3bc57667b28fe154e47dba5c374ee`
  - deterministic source SHA-256: `fb0e5812857a2fd46de488cd17a80011ef5d18795ff96fa1a3ebed5fd19a4377`
- Rogues 3.1.1 product/runtime authority: `77af5b416ab0889e770028ad3b93fbf2034d311c`
  - release SHA-256: `8845d27451fac83666b4ed4dd0ea5bac96c6a38b39c5d3585b8c9492274f3f23`
  - deterministic source SHA-256: `44f1ee45b7f50bfd5ee4c370ae03a8eaad51d0dd1aa2dfa11acb8b09c2f7c8e8`

Any integration candidate that cannot reproduce/certify those exact identities fails closed.

## Runtime topology

The combined packaged Forge 47.4.23 runtime contains exactly twelve separate mod JARs:

1. Paladins
2. Rogues
3. Shield API
4. Runes
5. Armor Model API
6. Structure Pool API
7. Spell Power
8. Spell Engine
9. TinyConfig
10. Cloth Config
11. Player Animator
12. Curios

No dependency is shaded into either leaf to make coexistence pass.

## Acceptance gates

1. Reconstruct both leaves from their pinned authorities and certify the exact graduated release/source hashes.
2. Detect the known Paladins intermediate-byte seam: Rogues may reuse Paladins' foundation reconstruction, but only the post-certification Paladins JAR may enter the combined runtime.
3. Require exactly twelve separate packaged mod JARs and exact installed Paladins/Rogues hashes.
4. Fail on duplicate `modId` ownership across different JARs.
5. Fail if Paladins packages Rogues-owned classes or Rogues packages Paladins-owned classes.
6. Require production mixin activation metadata for each leaf where its production runtime contract requires it.
7. Start the exact combined packaged Forge dedicated server and require the real `Done (...)!` readiness marker with no loader/mixin/registry/dependency fatal signature.
8. Exercise both namespaces on the same game thread/world: Paladins Priest Absorption semantics, Rogues Stealth registration/removal, and Rogues Bear Trap entity registration must succeed while both mods are loaded.
9. Start a real native LWJGL Forge client with both leaves present and join that exact twelve-mod packaged server through the normal modded handshake/network stack.
10. Reject leaf-owned missing texture/model failures in the combined native client.
11. Require server acknowledgement of the joined real player and a clean packaged-server shutdown.
12. Preserve exact-run GitHub evidence and publish the evidence ZIP/checkpoint to the canonical Google Drive project folder with exact SHA-256/size round-trip verification before promotion.

## Promotion rule

The integration branch may fast-forward the canonical workbench only after `FULL_RPG_SERIES_INTEGRATION_PASS` is proven on the exact proposed head and the evidence publication gate is complete. Promotion is guarded and non-forced. The integration bookkeeping head does not supersede either frozen leaf runtime authority.
