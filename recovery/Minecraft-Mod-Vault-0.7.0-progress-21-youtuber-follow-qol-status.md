# Minecraft Mod Vault v0.7.0 progress 21

Creator Archive follow-management and curated YouTuber catalog milestone.

- Canonical parent: progress 20 Creator Archive source, commit d5d7b6059a5eea1fcf8795a8d29844945e0ee496.
- Any YouTube creator can be followed by @handle, channel URL, legacy /user/ or /c/ URL, or UC channel ID.
- Followed custom channels use the same full-history uploads/Shorts/streams, transcript, recommendation, evidence, retry, and refresh pipeline as the core channels.
- Curated recommendations added for current showcase, current modded, deep-modded, historical, and utility channels.
- One-click Follow + Sync and Add Top Current Picks controls added.
- Pause/resume automatic refresh is persistent; explicit manual rescans still work.
- Non-core unfollow preserves indexed videos, transcripts, and recommendations.
- Bulk Sync All skips paused channels.
- Archive channel filters resolve by channel ID or handle.

Verification at this milestone:
- go test ./... PASS
- go vet ./... PASS
- node --check web/app.js PASS
- node --check web/catalog.js PASS

Drive source ZIP ID: 12VB4zJtOdMYRWQpoCkK6C75MExxmHoe9
Source ZIP SHA-256: 8824e3aab1360be5efd8eefba451cb6f7be22eb8356004b5e9e471b1fd8727ec
Source ZIP size: 229293 bytes
Remote Drive ZIP was re-downloaded and SHA-256 matched exactly.
