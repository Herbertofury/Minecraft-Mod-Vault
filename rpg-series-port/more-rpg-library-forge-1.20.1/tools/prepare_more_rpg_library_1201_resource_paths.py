#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_resource_paths.py <prepared-port-root>')

root = Path(sys.argv[1]).resolve()
resource_roots = [
    root / 'common/src/main/resources',
    root / 'common/src/main/generated',
]
if not resource_roots[0].is_dir():
    raise SystemExit(f'More RPG resource-path migration missing resources root: {resource_roots[0]}')
if not resource_roots[1].is_dir():
    raise SystemExit(f'More RPG resource-path migration missing generated root: {resource_roots[1]}')

counts = {
    'items': 0,
    'entity_types': 0,
    'recipes': 0,
    'advancements': 0,
    'wizard_runes': 0,
}


def migrate_file(src: Path, dst: Path, family: str) -> None:
    data = src.read_bytes()
    dst.parent.mkdir(parents=True, exist_ok=True)
    if dst.exists():
        if not dst.is_file():
            raise SystemExit(f'More RPG resource-path destination is not a file: {dst}')
        if dst.read_bytes() != data:
            raise SystemExit(
                f'More RPG resource-path collision has different bytes: source={src} destination={dst}'
            )
        src.unlink()
    else:
        src.replace(dst)
    counts[family] += 1


def remove_empty_dirs(base: Path) -> None:
    for path in sorted((p for p in base.rglob('*') if p.is_dir()), key=lambda p: len(p.parts), reverse=True):
        try:
            path.rmdir()
        except OSError:
            pass


for base in resource_roots:
    # Work from a stable snapshot because files move while iterating.
    files = sorted(p for p in base.rglob('*') if p.is_file())
    for src in files:
        if not src.exists():
            continue
        rel = src.relative_to(base)
        parts = list(rel.parts)
        if len(parts) < 3 or parts[0] != 'data':
            continue

        # Standard registry/datapack directory names changed from plural in 1.20.1 to singular
        # in 1.21. Preserve namespace and payload bytes; translate only historically proven
        # vanilla registry directory names.
        if len(parts) >= 5 and parts[2] == 'tags' and parts[3] == 'item':
            target = parts.copy()
            target[3] = 'items'
            migrate_file(src, base.joinpath(*target), 'items')
            continue
        if len(parts) >= 5 and parts[2] == 'tags' and parts[3] == 'entity_type':
            target = parts.copy()
            target[3] = 'entity_types'
            migrate_file(src, base.joinpath(*target), 'entity_types')
            continue
        if len(parts) >= 4 and parts[2] == 'recipe':
            target = parts.copy()
            target[2] = 'recipes'
            migrate_file(src, base.joinpath(*target), 'recipes')
            continue
        if len(parts) >= 4 and parts[2] == 'advancement':
            target = parts.copy()
            target[2] = 'advancements'
            migrate_file(src, base.joinpath(*target), 'advancements')
            continue

        # Wizards' rune-compat data is a special case, not a generic registry pluralization.
        # Genuine 1.20.1 Wizards stores this exact file directly under data/wizards/.
        if rel.as_posix() == 'data/wizards/item/wizard_runes.json':
            migrate_file(src, base / 'data/wizards/wizard_runes.json', 'wizard_runes')

    remove_empty_dirs(base)

# Fail closed if any standard 1.21 singular resource path remains after migration.
survivors = []
for base in resource_roots:
    for p in sorted(x for x in base.rglob('*') if x.is_file()):
        rel = p.relative_to(base).as_posix()
        parts = rel.split('/')
        if len(parts) >= 5 and parts[0] == 'data' and parts[2] == 'tags' and parts[3] in {'item', 'entity_type'}:
            survivors.append(rel)
        elif len(parts) >= 4 and parts[0] == 'data' and parts[2] in {'recipe', 'advancement'}:
            survivors.append(rel)
        elif rel == 'data/wizards/item/wizard_runes.json':
            survivors.append(rel)
if survivors:
    raise SystemExit('More RPG 1.21 resource-path survivors remain:\n' + '\n'.join(survivors))

# Current 2.7.2 contains these families. Requiring nonzero counts catches source-layout drift before
# Gradle can silently package data Forge 1.20.1 will not read.
minimums = {
    'items': 17,
    'entity_types': 6,
    'recipes': 20,
    'advancements': 1,
    'wizard_runes': 1,
}
for family, minimum in minimums.items():
    if counts[family] < minimum:
        raise SystemExit(
            f'More RPG resource-path migration count too small for {family}: '
            f'found={counts[family]} expected_at_least={minimum}'
        )

# Representative target-native files must exist somewhere in the merged resource inputs.
representatives = [
    'data/more_rpg_classes/tags/items/enchantable/typhoon.json',
    'data/more_rpg_classes/tags/items/enchantable/stonebloom.json',
    'data/runes/tags/items/wizard_stones.json',
    'data/berserker_rpg/tags/entity_types/hatred_of_undead.json',
    'data/more_rpg_classes/recipes/aqua_rune_medium_altar.json',
    'data/more_rpg_content/advancements/recipes/equipment/arcane_alley.json',
    'data/wizards/wizard_runes.json',
]
for rel in representatives:
    if not any((base / rel).is_file() for base in resource_roots):
        raise SystemExit(f'More RPG target-native resource missing after migration: {rel}')

# The modern enchantment JSON authority remains intentionally packaged for parity/audit; the
# compatibility layer adds 1.20.1 registry entries rather than deleting those source definitions.
for rel in (
    'data/more_rpg_classes/enchantment/typhoon.json',
    'data/more_rpg_classes/enchantment/stonebloom.json',
):
    if not any((base / rel).is_file() for base in resource_roots):
        raise SystemExit(f'More RPG modern enchantment authority lost during path migration: {rel}')

print(
    '[More RPG 2.7.2] RESOURCE_PATH_1201_MIGRATION_PASS '
    f"items={counts['items']} entity_types={counts['entity_types']} "
    f"recipes={counts['recipes']} advancements={counts['advancements']} "
    f"wizard_runes={counts['wizard_runes']}"
)
