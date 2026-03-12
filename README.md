# Deadlock Patch Notes

Deadlock patch notes monorepo with:
- DB-backed Go API
- Next.js frontend
- Automated forum/Steam ingestion pipeline

## Documentation

- Canonical docs entrypoint: [`docs/index.md`](./docs/index.md)
- Current-state technical references are linked from that index:
  - runtime overview
  - API contracts
  - domain model rules
  - ingestion/parser rules
  - frontend behavior
  - ops/scripts
  - maintenance rules

## Monorepo Layout

- `web/`: Next.js App Router frontend (SSR React)
- `api/`: Go API (`net/http` + `chi`) with PostgreSQL storage
- `scripts/`: fixture pipeline, source checks, and server automation scripts

## Local Run

### API

```bash
cd api
go mod tidy
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/server
```

Server listens on `http://localhost:8080` by default.

### Web

```bash
cd web
npm install
npm run dev
```

Frontend runs on `http://localhost:3000`.
Default API base is `https://deadlockpatchnotes.com/api`.
To override, set `API_BASE_URL` (for example `API_BASE_URL=http://localhost:8080 npm run dev`).
If `API_BASE_URL` is set to an invalid URL, web startup fails fast.

### Sync patch notes into DB

```bash
cd api
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

Default source is `https://forums.playdeadlock.com/forums/changelog.10/`.
Default max pages is `20`; default timeout is `30` seconds.

### Windows Helper Startup

```bat
start-site.bat
```

This launches API and web in separate Windows shells and opens `http://localhost:3000/patches`.

## API Endpoints

- `GET /api/healthz`
- `GET /api/scalar`
- `GET /api/openapi.json`
- `GET /api/v1/patches?page=<int>&limit=<int>`
- `GET /api/v1/patches/{slug}`
- `GET /api/v1/heroes`
- `GET /api/v1/heroes/{heroSlug}/changes?skill=<name>&from=<date|rfc3339>&to=<date|rfc3339>`
- `GET /api/v1/spells`
- `GET /api/v1/spells/{spellSlug}/changes?from=<date|rfc3339>&to=<date|rfc3339>`
- `GET /api/v1/items`
- `GET /api/v1/items/{itemSlug}/changes?from=<date|rfc3339>&to=<date|rfc3339>`

## Docker Deployment

```bash
cp .env.example .env
# edit .env (especially POSTGRES_PASSWORD)
docker-compose up -d --build db api web
```

Useful `.env` knobs:
- `WEB_HOST_BIND=127.0.0.1` keeps web private to the server; set `WEB_HOST_BIND=0.0.0.0` for LAN access.
- `API_HOST_BIND=0.0.0.0` makes API reachable for path-based cloudflared/tunnel routing.
- Set `API_HOST_BIND=127.0.0.1` to keep API private and access it via SSH tunneling.
- `API_PORT=18081` controls published API host port (`${API_HOST_BIND}:${API_PORT}`).
- `API_READ_CACHE_TTL=10m` controls API read snapshot cache freshness window.

Run one ingestion pass:

```bash
docker-compose run --rm sync
```

Install sync + backup cron jobs:

```bash
./scripts/server/install_cron.sh
```

## Fixture Generation

Generate/update fixture JSON + mirrored assets for the configured target patch:

```bash
node scripts/generate_patch_fixture.mjs
```

Related parser modules live in `scripts/patch_fixture/`.

Sync hero page media for heroes marked in-game from the assets API (`background_image*` + `name_image`) into local web assets:

```bash
node scripts/sync_hero_media.mjs
```

Audit icon asset URLs currently returned by the API and classify which ones are still remote (allowed hosts, local misses, etc):

```bash
node scripts/audit_api_icons.mjs
```

Mirror remote API icon URLs to local deduped files and write a URL mapping manifest consumed by the web app (`web/public/assets/mirror/manifest.json`):

```bash
node scripts/mirror_api_icons.mjs --audit /tmp/deadlock-icon-audit-<timestamp>.json
```

### Local Dev Against Server Data

Your local Next.js dev server uses the public API hostname by default:

```bash
cd web
npm run dev
```

If you want a private tunnel instead, forward server API to localhost:

```bash
ssh -L 8080:127.0.0.1:18081 root@10.0.0.169
API_BASE_URL=http://127.0.0.1:8080 npm run dev
```

Optional DB tunnel (for direct SQL access from your local tools):

```bash
ssh -L 5433:127.0.0.1:5432 root@10.0.0.169
```

## SQL Drafts

- Legacy draft schema is in `api/sql/drafts/001_patchnotes_schema_postgres.sql`
- Active runtime migration is in `api/internal/db/migrations/001_patchnotes.sql`
