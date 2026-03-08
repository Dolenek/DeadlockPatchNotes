# Architecture

## Overview
The repository is a monorepo for a Deadlock patch notes product:
- `api/`: serves patch list/detail payloads for the frontend.
- `web/`: Next.js App Router UI rendering patch list and patch detail views.
- `scripts/`: helper scripts for source checks and server automation.

## API Layer (`api/`)
- Server entrypoint: `api/cmd/server/main.go`.
- Router: `api/internal/httpapi/router.go` using `chi`.
- Storage: `api/internal/patches/postgres_store.go` reads `patches.detail_payload` JSON from PostgreSQL.
- Migrations: `api/internal/db/migrations/*.sql` applied on startup.
- Sync command: `api/cmd/sync/main.go` crawls forum/Steam sources and upserts DB rows.

Primary runtime source is PostgreSQL (`DATABASE_URL` required).

## Web Layer (`web/`)
- App routes:
  - `web/app/patches/page.tsx`
  - `web/app/patches/[slug]/page.tsx`
- API client: `web/lib/api.ts`.
- Domain types: `web/lib/types.ts`.
- Main presentation components under `web/components/`.

The frontend defaults to `API_BASE_URL=http://localhost:8080`.

## Ingestion Pipeline
- Source list crawl: `https://forums.playdeadlock.com/forums/changelog.10/`.
- Thread parser: extracts official posts from `Yoshi` and their timestamps/content.
- Steam click-through parser: resolves linked `store.steampowered.com/news/.../view/<id>` pages and parses embedded partner event payloads.
- Output persistence:
  - `patches` table (summary + detail JSON)
  - `patch_release_blocks` table (initial/hotfix timeline blocks)
  - `sync_runs` table (sync observability)

At runtime, API list/detail reads from DB only.

## Deployment Model
- `docker-compose.yml` runs `db`, `api`, `web`, and one-shot `sync` service.
- PostgreSQL persistence uses named volume `pgdata`.
- Web binds to loopback-only host port (`127.0.0.1:${WEB_PORT}`).
- Recommended external exposure is via Cloudflare Tunnel to the web port.
