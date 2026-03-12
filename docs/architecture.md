# Architecture

## Overview

Monorepo structure:

- `api/`: Go HTTP API + ingestion sync process + PostgreSQL persistence.
- `web/`: Next.js App Router frontend.
- `scripts/`: fixture pipeline, source-limit checks, and server automation.

## API Layer (`api/`)

- Server entrypoint: `api/cmd/server/main.go`.
- Router: `api/internal/httpapi/router.go` (`chi`).
- Repository: `api/internal/patches/postgres_store.go`.
- API surface:
  - `/api/healthz`
  - `/api/scalar`, `/api/openapi.json`
  - `/api/v1/patches`, `/api/v1/patches/{slug}`
  - `/api/v1/heroes`, `/api/v1/heroes/{heroSlug}/changes`
  - `/api/v1/items`, `/api/v1/items/{itemSlug}/changes`
  - `/api/v1/spells`, `/api/v1/spells/{spellSlug}/changes`

Read path split:

- Patch list reads summary columns directly from `patches` table.
- Patch detail/entity timelines load `patches.detail_payload` JSON and apply hydration/query builders.

Migrations in `api/internal/db/migrations/*.sql` are applied at startup.

## Ingestion Pipeline (`api/cmd/sync` + `api/internal/ingest`)

- Crawls changelog forum listing pages.
- Fetches patch threads and keeps official `Yoshi` posts.
- Parses Steam event payload when available, with forum fallback behavior.
- Builds canonical patch detail payload and timeline blocks.
- Upserts:
  - `patches` (summary + detail JSON)
  - `patch_release_blocks` (timeline block metadata)
  - `sync_runs` (run observability)

## Web Layer (`web/`)

- App routes:
  - `web/app/patches/...`
  - `web/app/heroes/...`
  - `web/app/items/...`
  - `web/app/spells/...`
- API client: `web/lib/api.ts`.
- Domain types: `web/lib/types.ts`.
- Shared components: `web/components/`.

Frontend API base behavior:

- Default API base is `https://deadlock.jakubdolenek.xyz/api`.
- `API_BASE_URL` may override it.
- Exact `/api` suffix is normalized, so both host-only and `/api`-suffixed values are accepted.
- Invalid non-empty `API_BASE_URL` throws during client initialization.

## Deployment Model

- `docker-compose.yml` runs `db`, `api`, `web`, and one-shot `sync`.
- PostgreSQL persistence uses named volume `pgdata`.
- API is loopback-published by default (`127.0.0.1:${API_PORT}`)
- Web bind host/port is configurable via `WEB_HOST_BIND` and `WEB_PORT`.
