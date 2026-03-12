# API Contracts

## Base Behavior

- Router root: `api/internal/httpapi/router.go`.
- CORS middleware sets:
  - `Access-Control-Allow-Origin: *`
  - `Access-Control-Allow-Methods: GET, OPTIONS`
  - `Access-Control-Allow-Headers: Content-Type`
- `OPTIONS` requests return `204 No Content` from middleware.
- Handler success responses use `Content-Type: application/json`.
- Handled errors now use a structured payload:

```json
{
  "error": {
    "code": "invalid_query_param",
    "message": "invalid from query value",
    "requestId": "..."
  }
}
```

- Error code values emitted by handlers:
  - `missing_path_param`
  - `invalid_query_param`
  - `resource_not_found`
  - `internal_error`

## Time Query Parsing (`from`, `to`)

Accepted formats:

- RFC3339 timestamp (example: `2026-03-06T22:36:00Z`)
- Date-only `YYYY-MM-DD`

Date-only behavior:

- `from=YYYY-MM-DD` becomes `00:00:00.000000000Z`
- `to=YYYY-MM-DD` becomes `23:59:59.999999999Z`

## Meta Endpoints

### `GET /api/healthz`

```json
{ "status": "ok" }
```

### `GET /api/scalar`

- `200` with `Content-Type: text/html; charset=utf-8`
- Serves Scalar docs that read schema from `/api/openapi.json`

### `GET /api/openapi.json`

- `200` with `Content-Type: application/json`
- Returns committed OpenAPI 3.1 schema

## `GET /api/v1/patches`

Query:

- `page` default `1`
- `limit` default `12`, max `50`

Response shape:

```json
{
  "patches": [PatchSummary],
  "pagination": {
    "page": 1,
    "pageSize": 12,
    "totalItems": 1,
    "totalPages": 1
  }
}
```

`PatchSummary` contract highlights:

- `imageUrl` (was legacy `coverImageUrl`)
- `source` object (`type`, `url`)
- `releaseTimeline` for compact timeline summary blocks (`id`, `releaseType`, `title`, `releasedAt`)

## `GET /api/v1/patches/{slug}`

Success payload is `PatchDetail` with:

- `imageUrl` (was legacy `heroImageUrl`)
- `releaseTimeline` (was legacy `timeline`)
- timeline block field `releaseType` (was legacy `kind`)

Errors:

- `404` with `resource_not_found` for unknown slug
- `500` with `internal_error` for read/decode failures

## `GET /api/v1/heroes`

Response:

```json
{
  "heroes": [HeroSummary]
}
```

## `GET /api/v1/heroes/{heroSlug}/changes`

Query:

- `skill` optional (`general` => general-only block payload)
- `from` optional
- `to` optional

Response:

```json
{
  "hero": HeroSummary,
  "timeline": [HeroTimelineBlock]
}
```

Hero timeline block naming:

- `releaseType` (was `kind`)
- `displayLabel` (was `label`)
- `patchRef` (was `patch`)

## `GET /api/v1/items`

Response:

```json
{
  "items": [ItemSummary]
}
```

## `GET /api/v1/items/{itemSlug}/changes`

Response:

```json
{
  "item": ItemSummary,
  "timeline": [ItemTimelineBlock]
}
```

Timeline block naming follows hero pattern (`releaseType`, `displayLabel`, `patchRef`).

## `GET /api/v1/spells`

Response:

```json
{
  "spells": [SpellSummary]
}
```

## `GET /api/v1/spells/{spellSlug}/changes`

Response:

```json
{
  "spell": SpellSummary,
  "timeline": [SpellTimelineBlock]
}
```

Timeline block naming follows hero/item pattern (`releaseType`, `displayLabel`, `patchRef`).

## Read Optimization Behavior

- DB-backed reads use an in-memory snapshot cache in `PostgresStore`.
- Cache refresh is TTL-based (default `10m`, env: `API_READ_CACHE_TTL`).
- Snapshot refresh preloads and hydrates patch payloads once per cache window.
- List/detail/entity endpoints read from the cached snapshot in steady state.
