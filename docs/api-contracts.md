# API Contracts

## Base Behavior

- Router root: `api/internal/httpapi/router.go`.
- CORS middleware sets:
  - `Access-Control-Allow-Origin: *`
  - `Access-Control-Allow-Methods: GET, OPTIONS`
  - `Access-Control-Allow-Headers: Content-Type`
- `OPTIONS` requests return `204 No Content` directly from middleware.
- Handler responses are JSON with `Content-Type: application/json`.
- Error payload for handled endpoint errors:

```json
{ "error": "message" }
```

- Unmatched routes/methods (for example router-level `404`/`405`) are not produced by `writeError`; they follow chi default behavior.

## Time Query Parsing (`from`, `to`)

Supported formats:

- RFC3339 timestamp (example: `2026-03-06T22:36:00Z`)
- Date-only `YYYY-MM-DD`

Date-only behavior:

- `from=YYYY-MM-DD` becomes `00:00:00.000000000Z`
- `to=YYYY-MM-DD` becomes `23:59:59.999999999Z`

Notes:

- Invalid `from`/`to` returns `400`.
- `from > to` is accepted (no order validation); this typically yields empty results.

## `GET /api/healthz`

Response:

```json
{ "status": "ok" }
```

## `GET /api/v1/patches`

Query:

- `page` (default `1`)
- `limit` (default `12`, max clamp `50`)

Normalization and clamps:

- Non-integer `page`/`limit` values fall back to defaults.
- `page <= 0` becomes `1`.
- `limit <= 0` becomes `12`.
- `page > totalPages` is clamped to `totalPages`.
- `totalPages` is always at least `1`, including when `total = 0`.

Response shape:

```json
{
  "items": [PatchSummary],
  "page": 1,
  "limit": 12,
  "total": 1,
  "totalPages": 1
}
```

Ordering:

- Backed by `patches` table ordered by `updated_at DESC`.

Failure/partial-data behavior:

- If total-count query fails, response falls back to an empty payload (`items=[]`, `page=1`, `total=0`, `totalPages=1`).
- If list query fails, response returns empty `items` with computed pagination metadata.
- Row scan failures are skipped, so `items` can be partial while status remains `200`.

## `GET /api/v1/patches/{slug}`

Success:

- Loads `patches.detail_payload` JSON by slug.
- Applies timeline hydration before response.

Hydration guarantees include:

- Synthesize one `initial` block if timeline is missing.
- Rebuild block sections from change lines when block sections are missing.
- Rebuild flat changes from sections when block changes are missing.
- Fill missing block title (`Initial Update` / `Hotfix YYYY-MM-DD` / `Hotfix`).
- De-duplicate timeline blocks by normalized change-body signature.
- Sort hydrated timeline ascending by release time for canonicalization.
- Force first hydrated block kind to `initial`.

Errors:

- `404` if slug not found.
- `500` for storage/decode failures.

## `GET /api/v1/heroes`

Success:

- Returns `HeroListResponse` generated from hydrated timeline `heroes` sections.
- Sorted alphabetically by hero name (case-insensitive).

Hero entry inclusion rules:

- Entry is treated as hero timeline data only if it has groups, `entityIconUrl`, or `entityIconFallbackUrl`.
- Entries without these signals are excluded.

Failure behavior:

- If details cannot be loaded from DB, returns `200` with `items=[]`.

## `GET /api/v1/heroes/{heroSlug}/changes`

Query:

- `skill` (optional)
- `from` (optional)
- `to` (optional)

`skill` behavior:

- Empty: include hero general changes and all skill groups.
- `general` (case-insensitive): include only hero general changes.
- Any other value: normalized exact-title match against hero skill groups.
- Unknown skill for an existing hero returns `200` with empty `items`.

Date behavior:

- Inclusive UTC filtering on block `releasedAt`.
- Blocks with invalid/empty `releasedAt` parse to zero-time and are not excluded by date filtering.

Ordering:

- Timeline blocks sorted by `releasedAt DESC`.
- Tie-breaker: `patch.slug DESC`.
- Secondary tie-breaker: `id ASC`.

Errors:

- `400` invalid `from`/`to`.
- `404` unknown hero slug.
- `500` internal load/build failure.

## `GET /api/v1/items`

Success:

- Returns `ItemListResponse` from hydrated timeline `items` sections.
- Sorted alphabetically by item name.

Failure behavior:

- If details cannot be loaded from DB, returns `200` with `items=[]`.

## `GET /api/v1/items/{itemSlug}/changes`

Query:

- `from` (optional)
- `to` (optional)

Behavior:

- Inclusive date filtering (same zero-time behavior as hero changes).
- Uses direct item entry changes only (no nested groups in payload).
- Sorting and tie-breakers match hero timeline ordering rules.

Errors:

- `400` invalid date query.
- `404` unknown item slug.
- `500` internal load/build failure.

## `GET /api/v1/spells`

Success:

- Returns `SpellListResponse` inferred from hero sections.
- Sorted alphabetically by spell name.

Inference rules:

- Primary source: hero entry groups.
- Excludes groups titled `Talents` and `Card Types`.
- Standalone hero entries are fallback spell candidates only if icon metadata exists.
- Standalone entries matching known item names are excluded.

Failure behavior:

- If details cannot be loaded from DB, returns `200` with `items=[]`.

## `GET /api/v1/spells/{spellSlug}/changes`

Query:

- `from` (optional)
- `to` (optional)

Behavior:

- Inclusive date filtering (same zero-time behavior as hero/item changes).
- Same spell slug across multiple heroes is merged into one spell timeline block with multiple hero entries.
- Sorting and tie-breakers match hero/item timeline ordering rules.

Errors:

- `400` invalid date query.
- `404` unknown spell slug.
- `500` internal load/build failure.

## Key Payload Types

Core response payload types are mirrored between backend and frontend:

- Patch:
  - `PatchSummary`
  - `PatchDetail`
  - `PatchSection`
  - `PatchEntry`
  - `PatchEntryGroup`
  - `PatchTimelineBlock`
- Timelines:
  - `HeroListResponse`, `HeroChangesResponse`
  - `ItemListResponse`, `ItemChangesResponse`
  - `SpellListResponse`, `SpellChangesResponse`
