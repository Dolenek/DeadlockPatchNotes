# Runtime Overview

## Purpose

Deadlock Patch Notes ingests official Deadlock changelog content, stores normalized patch payloads in PostgreSQL, and serves those payloads through a Go API consumed by a Next.js frontend.

## Repository Topology

- `api/`
  - `cmd/server`: HTTP API process
  - `cmd/sync`: one-shot ingestion/sync process
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

1. Reads `DATABASE_URL` and optional `API_ADDR` (default `:8080`).
2. Opens PostgreSQL via `db.OpenPostgres`.
3. Applies embedded migrations from `api/internal/db/migrations`.
4. Constructs `patches.NewPostgresStore`.
5. Serves routes from `httpapi.NewRouter`.

### Sync Process (`api/cmd/sync`)

1. Reads `DATABASE_URL`, `PATCH_FORUM_URL`, `PATCH_SYNC_MAX_PAGES`, `PATCH_SYNC_TIMEOUT_SECONDS`.
2. Applies defaults for missing/invalid sync numeric env values.
3. Opens DB and applies migrations.
4. Crawls changelog thread listing, fetches threads/posts, optionally parses Steam links.
5. Builds patch payload + timeline blocks and upserts DB rows.
6. Writes run observability row to `sync_runs`.

Run semantics:

- Sync is best-effort per thread (thread-level failures are skipped).
- Sync can finish with `status=success` even when some discovered threads failed to process.

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
  - first effective block forced to `initial`
5. Structured payload construction:
  - parse/infer sections (`general`, `items`, `heroes`)
  - fallback general shape when structured parsing fails
6. Persistence:
  - full JSON payload in `patches.detail_payload`
  - timeline metadata/hash rows in `patch_release_blocks`
  - run counters/status in `sync_runs`
7. Read path:
  - API list endpoint reads summary columns from `patches`
  - API detail/timeline endpoints read `detail_payload` and run hydration/query builders
  - frontend fetches API responses and renders SSR output

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
- `PATCH_FORUM_URL` (sync only; default changelog URL)
- `PATCH_SYNC_MAX_PAGES` (sync only; default `20`, invalid/non-positive -> default)
- `PATCH_SYNC_TIMEOUT_SECONDS` (sync only; default `30`, invalid/non-positive -> default)

### Web

- `API_BASE_URL`
  - default: `https://deadlock.jakubdolenek.xyz/api`
  - exact `/api` suffix is normalized, so host-only and `/api`-suffixed values are accepted
  - invalid non-empty value throws during API client initialization

### Docker Compose

- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`
- `API_PORT`
- `WEB_HOST_BIND`, `WEB_PORT`
- Sync vars listed above
