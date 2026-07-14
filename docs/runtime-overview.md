# Runtime Overview

## Purpose

Deadlock Patch Notes ingests official Deadlock changelog content, stores normalized patch payloads in PostgreSQL, and serves those payloads through a Go API consumed by a Next.js frontend.

## Repository Topology

- `api/`
  - `cmd/server`: HTTP API process
  - `cmd/sync`: one-shot ingestion/sync process
  - `cmd/migrate`: one-shot schema migration and runtime-role provisioning process
  - `internal/httpapi`: routes, query parsing, response shaping
  - `internal/patches`: domain models, hydration, hero/item/spell timeline builders, DB-backed repository
  - `internal/ingest`: forum/Steam crawling, parsing, payload assembly, persistence orchestration
  - `internal/db`: PostgreSQL connection + embedded migration bootstrap
- `web/`
  - Next.js App Router UI for patch list/detail and hero/item/spell timelines
  - Server-side API client in `web/lib/api.ts`
- `scripts/`
  - Fixture generation and asset mirroring pipeline
  - Source-size/function-length guidance check
  - Server automation scripts (sync + backup cron helpers)

## Runtime Processes

### API Server (`api/cmd/server`)

1. Reads `DATABASE_URL`, optional `API_ADDR` (default `:8080`), optional `API_READ_CACHE_TTL` (default `10m`), and optional `SITE_URL` (for canonical RSS item links).
2. Opens PostgreSQL via `db.OpenPostgres`.
3. Constructs `patches.NewPostgresStore`.
4. Serves routes from `httpapi.NewRouter` with a read-only database role.

### Migration Process (`api/cmd/migrate`)

1. Connects with the database-owner `DATABASE_URL`.
2. Applies embedded migrations from `api/internal/db/migrations`.
3. Creates or updates `deadlock_api` and `deadlock_sync` with separate passwords.
4. Grants the API role read-only table access and the sync role only the table/sequence write access needed by ingestion.

### Sync Process (`api/cmd/sync`)

1. Reads `DATABASE_URL`, `PATCH_FORUM_URL`, `PATCH_SYNC_MAX_PAGES`, `PATCH_SYNC_TIMEOUT_SECONDS`.
2. Applies defaults for missing/invalid sync numeric env values.
3. Opens DB with the `deadlock_sync` runtime role; migrations must already be complete.
4. Crawls changelog thread listing, fetches threads/posts, and resolves every referenced Steam event. If the forum listing is blocked or empty, it discovers official minor updates through the Steam Web API.
5. Builds patch payload + timeline blocks and upserts DB rows.
6. Writes run observability row to `sync_runs`.

Run semantics:

- Empty/challenge discovery, catalog failure, or no successfully processed threads fails the run.
- A successful Steam fallback updates the latest patch with previously unseen minor-update blocks. A gap above 14 days fails loudly because it may represent a new top-level patch.
- A mix of successful and failed threads is persisted as `partial` and exits non-zero.
- A patch is not overwritten when one of its referenced Steam events fails.

## Data Lifecycle

1. Source discovery:
  - forum listing crawl from changelog URL
  - URL filtering to patch-update threads
2. Thread parsing:
  - keep `Yoshi` posts only
  - parse forum body, Steam URL, post metadata
3. Content resolution:
  - prefer Steam body blocks when parseable
  - forum body fallback when available
4. Block normalization:
  - dedupe by normalized body hash
  - sort by release time
  - exactly the first effective block is `initial`; all later post/event starts are `hotfix`
5. Structured payload construction:
  - parse/infer sections (`general`, `items`, `heroes`)
  - fallback general shape when structured parsing fails
6. Persistence:
  - full JSON payload in `patches.detail_payload`
  - timeline metadata/hash rows in `patch_release_blocks`
  - run counters/status in `sync_runs`
7. Read path:
  - API builds a cached in-memory snapshot from `patches.detail_payload` and summary columns
  - list/detail/timeline + RSS endpoints serve from the snapshot until TTL expiry
  - stale data remains available during transient refresh failures, while request cancellation propagates to PostgreSQL
  - duplicate cross-patch events are removed from aggregate entity histories but retained on patch detail pages
  - frontend fetches API responses without a Next.js filesystem cache and renders SSR output
  - the Go snapshot cache remains the shared cache for frontend reads, allowing the web container filesystem to stay read-only

## Persistence Model

Primary active tables:

- `patches`: one row per patch slug with summary columns + canonical detail JSON payload
- `patch_release_blocks`: normalized timeline metadata/body hashes per patch
- `sync_runs`: ingestion run observability

Schema notes:

- `patches` uniqueness is enforced on `slug` and `thread_id`.
- `patch_release_blocks.source_url` and `post_id` are nullable.
- `patch_release_blocks.kind`, `patch_release_blocks.source_type`, and `sync_runs.status` are free-text fields in schema.

## Configuration Surface

### API + Sync

- `DATABASE_URL` (required)
- `API_ADDR` (server only; default `:8080`)
- `API_READ_CACHE_TTL` (server only; Go duration, default `10m`)
- `SITE_URL` (server + web; optional canonical site URL used for SEO and RSS item links)
- `PATCH_FORUM_URL` (sync only; default changelog URL)
- `PATCH_STEAM_NEWS_URL` (sync only; default official Steam Web API query for the latest 100 app news items)
- `PATCH_SYNC_MAX_PAGES` (sync only; default `20`, invalid/non-positive -> default)
- `PATCH_SYNC_TIMEOUT_SECONDS` (sync only; default `30`, invalid/non-positive -> default)

### Migration

- `DATABASE_URL` (required database-owner connection)
- `API_DB_PASSWORD` (required, at least 16 characters)
- `SYNC_DB_PASSWORD` (required, at least 16 characters and distinct in deployment configuration)

### Web

- `API_BASE_URL`
  - default: `https://deadlockpatchnotes.com/api`
  - exact `/api` suffix is normalized, so host-only and `/api`-suffixed values are accepted
  - invalid non-empty value throws during API client initialization
- `SITE_URL`
  - default: `https://www.deadlockpatchnotes.com`
  - controls canonical host used for metadata/sitemap and API RSS item links

### Docker Compose

- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`
- `API_DB_PASSWORD`, `SYNC_DB_PASSWORD` (use URL-safe values because Compose places them in runtime connection URLs)
- `API_HOST_BIND`, `API_PORT`
- `WEB_HOST_BIND`, `WEB_PORT`
- `SITE_URL`
- Sync vars listed above
