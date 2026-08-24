# Creator Catalogs — hot-drop database contract

Minecraft Mod Vault creator catalogs are **data**, not creator-specific application code. A catalog can be embedded with a release or dropped at runtime under the app config folder's `creator-catalogs/` directory. Nested subfolders are supported. The archive checks for changes automatically; the Creator Archive also exposes a manual **Reload catalogs** action.

## Guarantees

- Schema version `1` is additive and evidence-first.
- A bundle can describe zero, one, or many videos and zero, one, or many recommendations per video.
- `projectType` supports at least `mod`, `modpack`, `resourcepack`, `shader`, and `datapack`.
- A catalog can seed incomplete research (`coverage.complete=false`) without pretending the creator's history is complete.
- `coverage.expectedVideos` records the known channel-size target when verified.
- Live/provider-resolved metadata wins over weaker catalog metadata. Catalog reloads fill missing fields and union recommendations rather than downgrading a verified provider URL/project ID.
- Malformed/oversized bundles are surfaced as errors and do not wipe the last-known-good creator database.
- Local bundles are capped at 64 MiB each. Rename a bundle to `*.disabled.json` to disable it without deleting it.
- Catalogs are rescanned while the app is running, so updating a JSON file does not require a rebuild or restart.

## Minimal bundle

```json
{
  "schemaVersion": 1,
  "id": "creator-slug",
  "title": "Creator archive seed",
  "updatedAt": "2026-08-23T20:00:00Z",
  "source": "Public creator research",
  "sourceUrl": "https://www.youtube.com/@Creator",
  "creator": {
    "platform": "youtube",
    "handle": "@Creator",
    "url": "https://www.youtube.com/@Creator",
    "channelId": "UC...",
    "title": "Creator",
    "required": false
  },
  "coverage": {
    "expectedVideos": 100,
    "complete": false,
    "notes": "Live full-history sync fills remaining uploads."
  },
  "videos": [
    {
      "id": "VIDEO_ID",
      "platform": "youtube",
      "url": "https://www.youtube.com/watch?v=VIDEO_ID",
      "title": "10 Mods You Need",
      "publishedAt": "2026-08-01T00:00:00Z",
      "mods": [
        {
          "name": "Example Mod",
          "projectType": "mod",
          "url": "",
          "timestamp": "1:23",
          "timestampSeconds": 83,
          "videoLink": "",
          "evidence": "Listed in the creator's public video description.",
          "sourceKinds": ["description", "catalog"],
          "confidence": 0.94
        }
      ]
    }
  ]
}
```

Leave `provider`, `projectId`, or `url` empty when they are not actually verified. The normal Creator Vault resolver can enrich them later. Do not manufacture plausible URLs.

## AsianHalfSquat bootstrap

The first shipped catalog is `catalogs/creators/asianhalfsquat.json`. Its current snapshot records the verified channel target of **349 videos**, exact metadata for the newest 11 videos available to this ingestion run, and 93 evidence-backed recommendations across mods, shaders, and resource packs. It is intentionally marked incomplete. A network-enabled Creator Vault full-history sync continues the same creator record rather than creating a parallel database.
