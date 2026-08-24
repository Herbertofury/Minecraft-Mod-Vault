# Bedrock management contract

OmniManager treats Bedrock packages as structured content, not opaque downloads. It supports package inspection, native installation, installed-root scans, world activation and reversible removal across Stable, Preview/Beta and custom roots.

All ZIP-derived formats are validated before extraction. UUID/version pairs identify packs; localized manifest text and original art are retained. Installing or activating content produces a receipt containing source hash, destination, prior bytes when applicable, result hash and undo state.

No Bedrock action silently rewrites an unrelated world or pack, and no Java-style `.disabled` suffix is relied on inside active `com.mojang` content roots.
