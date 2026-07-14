# Development Workflow

## Prerequisites

- Node.js 24+
- npm
- Go 1.22+
- PostgreSQL 16+

## Run API

```bash
cd api
go mod tidy
API_DB_PASSWORD='replace-with-a-distinct-api-password' \
SYNC_DB_PASSWORD='replace-with-a-distinct-sync-password' \
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/migrate
DATABASE_URL='postgres://deadlock_api:replace-with-a-distinct-api-password@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/server
```

Default API URL: `http://localhost:8080`.

Optional API env var:

- `API_READ_CACHE_TTL` (Go duration, default `10m`)
- `SITE_URL` (optional canonical web host used for RSS item links; fallback is `https://www.deadlockpatchnotes.com`)

## Run Web

```bash
cd web
npm install
npm run dev
```

Default web URL: `http://localhost:3000`.
Default API base in web code: `https://deadlockpatchnotes.com/api`.
Override with `API_BASE_URL` when needed, for example:

```bash
API_BASE_URL=http://localhost:8080 npm run dev
```

`API_BASE_URL` accepts HTTP(S) host-only and `/api`-suffixed values without embedded credentials.

SEO env vars for web:

- `SITE_URL` (default `https://www.deadlockpatchnotes.com`) accepts HTTP(S) URLs without embedded credentials and controls canonical URL + sitemap host and API RSS item-link host.

## Run Patch Sync

```bash
cd api
DATABASE_URL='postgres://deadlock_sync:replace-with-a-distinct-sync-password@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

Optional env vars:

- `PATCH_FORUM_URL` (default `https://forums.playdeadlock.com/forums/changelog.10/`)
- `PATCH_SYNC_MAX_PAGES` (default `20`; invalid/non-positive -> `20`)
- `PATCH_SYNC_TIMEOUT_SECONDS` (default `30`; invalid/non-positive -> `30`)

## Windows Helper Startup

`start-site.bat` starts local API + web in separate Windows shells and opens `/patches`.

Behavior:

- Verifies `go` and `npm` are present in PATH.
- Expects `DATABASE_URL` to point to an already provisioned API role; run `cmd/migrate` first for a fresh database.
- Runs API via `go run ./cmd/server` in `api/`.
- Runs web dev server in `web/` and sets `API_BASE_URL=http://localhost:8080` for that shell.

## Quality Checks

Frontend:

```bash
cd web
npm run lint
npm run test
npm run build
npm run test:runtime
```

`test:runtime` stages the standalone build like the production image and verifies HTTPS routing, Next image optimization, same-origin fonts, and dynamic response caching.

Backend:

```bash
cd api
go test ./...
```

Source guidance check:

```bash
node scripts/check_source_limits.mjs
```

## Docker Stack

```bash
cp .env.example .env
# set POSTGRES_PASSWORD, API_DB_PASSWORD, and SYNC_DB_PASSWORD
docker-compose up -d --build db migrate api web
```

Run one sync pass:

```bash
docker-compose run --rm sync
```

## Server Cron Helpers

- `scripts/server/run_sync.sh`
- `scripts/server/backup_postgres.sh`
- `scripts/server/install_cron.sh`
