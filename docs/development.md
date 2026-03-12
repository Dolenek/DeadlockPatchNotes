# Development Workflow

## Prerequisites

- Node.js 20+
- npm
- Go 1.22+
- PostgreSQL 16+

## Run API

```bash
cd api
go mod tidy
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/server
```

Default API URL: `http://localhost:8080`.

Optional API env var:

- `API_READ_CACHE_TTL` (Go duration, default `10m`)

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

`API_BASE_URL` accepts both host-only and `/api`-suffixed values for the production domain.

## Run Patch Sync

```bash
cd api
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

Optional env vars:

- `PATCH_FORUM_URL` (default `https://forums.playdeadlock.com/forums/changelog.10/`)
- `PATCH_SYNC_MAX_PAGES` (default `20`; invalid/non-positive -> `20`)
- `PATCH_SYNC_TIMEOUT_SECONDS` (default `30`; invalid/non-positive -> `30`)

## Windows Helper Startup

`start-site.bat` starts local API + web in separate Windows shells and opens `/patches`.

Behavior:

- Verifies `go` and `npm` are present in PATH.
- Runs API via `go run ./cmd/server` in `api/`.
- Runs web dev server in `web/` and sets `API_BASE_URL=http://localhost:8080` for that shell.

## Quality Checks

Frontend:

```bash
cd web
npm run lint
npm run test
npm run build
```

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
# set POSTGRES_PASSWORD
docker-compose up -d --build db api web
```

Run one sync pass:

```bash
docker-compose run --rm sync
```

## Server Cron Helpers

- `scripts/server/run_sync.sh`
- `scripts/server/backup_postgres.sh`
- `scripts/server/install_cron.sh`
