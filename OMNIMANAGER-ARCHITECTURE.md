# OmniManager architecture

## Identity graph

One installed artifact owns one immutable local identity and zero or more provider identities. Exact hash/fingerprint evidence outranks declared URLs, embedded metadata outranks filenames, and fuzzy matches remain review suggestions. Each field retains its evidence source and confidence rather than hiding disagreement.

## Two-phase library scan

The local scan is authoritative and immediately usable. Network enrichment is bounded, cached and independently retryable; it may improve metadata but cannot make local content disappear. Provider outages therefore degrade enrichment rather than library availability.

## Transaction engine

Mutating actions prepare a receipt, verify source identity, stage the replacement or move, commit atomically where the filesystem permits, verify resulting bytes, then retain enough information for undo. Custom builds are never silently promoted to storefront releases.

## Java and server content

JAR metadata adapters cover modern and legacy loaders/platforms. Provider records are merged by exact evidence. Compatibility considers Minecraft version, loader/platform, Java, side, dependencies and release channel rather than filename ordering alone.

## Bedrock content

Packages are parsed before extraction, roots are bounded, manifests and localization are preserved, and installs target a selected Stable/Preview/custom `com.mojang` root. World activation edits only the required JSON association with a byte-preserving previous-state receipt.

## Presentation contract

Every installed item remains in the DOM and searchable. Card/list views are alternate representations of the same complete model. Visible controls call real authenticated backend actions and expose truthful progress, failure and recovery state.
