# Domain Model Rules

## Canonical Data Shape

`patches.PatchDetail` is the canonical read model. It contains:

- Patch metadata (`id`, `slug`, `title`, `publishedAt`, `category`)
- Source (`type`, `url`)
- Hero image and intro text
- Top-level sections (`general`, `items`, `heroes`)
- Timeline blocks (`initial` + `hotfix` sequence)

Frontend mirrors this shape in `web/lib/types.ts`.

## Section and Entry Model

- `PatchSection`
  - `kind` is `general`, `items`, or `heroes`
  - Contains ordered `entries`
- `PatchEntry`
  - Entity-level container (`entityName`, optional icon URLs)
  - May contain direct `changes`
  - May contain grouped changes (`groups`)
- `PatchEntryGroup`
  - Nested group under an entry
  - Used heavily for hero abilities, talents, and card-type groups

## Timeline Hydration Invariants

Hydration is applied on patch detail reads (`hydratePatchDetail`).

Guaranteed behaviors:

1. Missing timeline becomes one synthesized `initial` block from top-level sections.
2. Timeline blocks missing `sections` are reconstructed from block `changes` using parse templates derived from merged sections.
3. Timeline blocks missing `changes` are flattened from block sections.
4. Empty block titles are normalized:
   - `initial` -> `Initial Update`
   - non-initial with valid date -> `Hotfix YYYY-MM-DD`
   - non-initial with invalid date -> `Hotfix`
5. Duplicate blocks (same normalized change-body signature) are dropped.
6. Hydrated timeline is sorted ascending by release time for canonicalization.
7. After sort, first block is forced to `kind=initial`.
8. Empty reconstruction paths emit fallback general content (`Core Gameplay` with one fallback line).

## Fallback Line Rule

Fallback text used when no concrete line-item changes can be derived:

- `No line-item changes listed.`

Used by:

- Ingestion fallback section generation
- Hydration flattening and empty block reconstruction

## Hero Timeline Rules

Hero index and detail payloads are built from hydrated timeline `heroes` sections only.

Hero entry validity rules:

- Valid hero timeline entry if at least one of:
  - has groups
  - has `entityIconUrl`
  - has `entityIconFallbackUrl`
- Entries failing this check are excluded from hero timelines.

Hero detail payload shape:

- `generalChanges`: direct entry-level changes
- `skills`: grouped changes

Skill filtering:

- `skill=general` returns only general entry-level changes.
- Other `skill` values use normalized exact-title matching against groups.
- Unknown skill for an existing hero produces empty `items` (not an error).

## Item Timeline Rules

- Item index is built from timeline `items` sections.
- Item timeline blocks contain only direct entry changes.
- Nested groups are not used in item timeline output.
- Date filtering is inclusive and UTC-normalized.

## Spell Timeline Rules

Spell timeline data is inferred from hero timeline data.

1. Primary mode:
   - Hero groups become spell candidates.
   - Groups named `Talents` and `Card Types` are excluded.
2. Fallback mode:
   - Standalone hero entry may be treated as spell only when icon metadata exists.
   - Standalone entries matching known item names are excluded.
3. Name collisions:
   - Same spell slug across heroes is merged into one spell timeline with multiple hero entries.

## Date, Label, and Ordering Rules

- Internal timeline timestamps are RFC3339 UTC where valid.
- Invalid/empty `releasedAt` parses as zero-time.
- Date filters do not exclude zero-time entries.
- Timeline labels use kind-based UI format:
  - `initial` -> `Update MM-DD-YYYY`
  - others -> `Patch MM-DD-YYYY`
  - invalid/zero release date -> `Unknown Date` in label output
- Hero/item/spell timeline outputs are sorted newest-first (`releasedAt DESC`) with tie-breakers:
  - `patch.slug DESC`
  - `id ASC` when same patch slug
