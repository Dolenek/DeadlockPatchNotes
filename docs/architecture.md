# Architecture

## Overview
The repository is a small monorepo for a Deadlock patch notes product:
- `api/`: serves patch list/detail payloads for the frontend.
- `web/`: Next.js App Router UI rendering patch list and patch detail views.
- `scripts/`: fetches and converts source patch content into local fixtures/assets.

## API Layer (`api/`)
- Server entrypoint: `api/cmd/server/main.go`.
- Router: `api/internal/httpapi/router.go` using `chi`.
- Data store: `api/internal/patches/store.go` reads embedded JSON fixtures.
- Data contracts: `api/internal/patches/models.go`.

Current runtime is fixture-backed in-memory storage for UI iteration.

## Web Layer (`web/`)
- App routes:
  - `web/app/patches/page.tsx`
  - `web/app/patches/[slug]/page.tsx`
- API client: `web/lib/api.ts`.
- Domain types: `web/lib/types.ts`.
- Main presentation components under `web/components/`.

The frontend defaults to `API_BASE_URL=http://localhost:8080`.

## Fixture and Asset Pipeline (`scripts/`)
- Entry command: `node scripts/generate_patch_fixture.mjs`.
- Pulls source patch text and asset metadata from Steam and deadlock-api.
- Produces:
  - JSON fixture: `api/internal/patches/data/<slug>.json`
  - Mirrored images: `web/public/assets/patches/<slug>/...`
  - Asset manifest: `web/public/assets/patches/<slug>/manifest.json`

At runtime, UI prefers local mirrored assets and can fall back to remote URLs.
