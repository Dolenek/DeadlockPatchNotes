# API Contracts

## Base Behavior

- Router root: `api/internal/httpapi/router.go`.
- CORS middleware sets:
  - `Access-Control-Allow-Origin: *`
  - `Access-Control-Allow-Methods: GET, OPTIONS`
  - `Access-Control-Allow-Headers: Content-Type`
- `OPTIONS` requests return `204 No Content` from middleware.
- JSON endpoints use `Content-Type: application/json` on success.
- RSS endpoints use `Content-Type: application/rss+xml; charset=utf-8` on success.
- `GET /api/v1/days-since-last-update` uses `Content-Type: text/plain; charset=utf-8` on success.
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

## `GET /api/v1/days-since-last-update`

Query:

- `hero` optional hero slug
  - when present, computes days from that hero's latest recorded change
  - when absent, computes days from latest patch `publishedAt`

Success response:

- `200` with `Content-Type: text/plain; charset=utf-8`
- body format: `Days since last update: X`

Errors:

- `404` with `resource_not_found` when hero slug is unknown (or no patch baseline exists)
- `500` with `internal_error` on read/timezone failures

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

## `GET /api/v1/patches/rss.xml`

- RSS 2.0 feed of recent patches.
- One RSS item per patch slug.
- Item ordering is `publishedAt DESC` (newest first).
- Item links target `/patches/{slug}` on `SITE_URL` host when configured (fallback: request host).
- Feed size is capped to `50` items.

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

## `GET /api/v1/heroes/{heroSlug}/rss.xml`

- RSS 2.0 per-hero feed.
- One item per patch where the hero has timeline changes.
- Item descriptions aggregate all hero changes from that patch (general + skill group lines).
- Item links target `/heroes/{heroSlug}` on `SITE_URL` host when configured (fallback: request host).
- Unknown hero slug returns `404` with `resource_not_found` JSON error payload.
- Feed size is capped to `50` patch items.

## `GET /api/v1/heroes/{heroSlug}/days-without-update/rss.xml`

- RSS 2.0 per-hero live streak feed.
- Returns a single item titled `Days since last update: N`.
- `N` uses Europe/Berlin noon checkpoints (`12:00`) with immediate reset to `0` when a hero update lands.
- Item links target `/heroes/{heroSlug}` on `SITE_URL` host when configured (fallback: request host).
- Unknown hero slug returns `404` with `resource_not_found` JSON error payload.

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
- List/detail/entity + RSS endpoints read from the cached snapshot in steady state.
