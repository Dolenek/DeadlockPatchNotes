# Deadlock Patch Notes

UI-first implementation of a Deadlock patch notes site inspired by League of Legends and Dota patch pages.

## Documentation

- Main docs index: `docs/index.md`
- Architecture overview: `docs/architecture.md`
- Development workflow: `docs/development.md`

## Monorepo Layout

- `web/`: Next.js App Router frontend (SSR React)
- `api/`: Go API (`net/http` + `chi`) with typed patch contracts
- `scripts/`: data generation scripts (Steam patch parsing + asset mirroring)

## Local Run

### API

```bash
cd api
go mod tidy
go run ./cmd/server
```

Server listens on `http://localhost:8080` by default.

### Web

```bash
cd web
npm install
npm run dev
```

Frontend runs on `http://localhost:3000` and calls `API_BASE_URL` (default `http://localhost:8080`).

## API Endpoints

- `GET /api/healthz`
- `GET /api/v1/patches?page=<int>&limit=<int>`
- `GET /api/v1/patches/{slug}`

## Fixture + Assets

- Patch data is loaded from JSON fixtures in `api/internal/patches/data`.
- The current fixture includes the full Steam patch `Gameplay Update - 03-06-2026`.
- Patch images/icons are mirrored under `web/public/assets/patches/2026-03-06-update`.
- Runtime rendering is local-first with remote URL fallback for missing assets.

Regenerate fixture and mirrored assets:

```bash
node scripts/generate_patch_fixture.mjs
```

## SQL Drafts

- PostgreSQL draft schema for DB-backed ingestion is in:
  `api/sql/drafts/001_patchnotes_schema_postgres.sql`
