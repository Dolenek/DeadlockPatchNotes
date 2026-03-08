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

## Run Web
```bash
cd web
npm install
npm run dev
```

Default web URL: `http://localhost:3000`.
Default API URL used by the web app: `https://api.deadlock.jakubdolenek.xyz`.
Override with `API_BASE_URL` when needed (for example `API_BASE_URL=http://localhost:8080 npm run dev`).

## Run Patch Sync
```bash
cd api
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

Optional env vars:
- `PATCH_FORUM_URL` (default `https://forums.playdeadlock.com/forums/changelog.10/`)
- `PATCH_SYNC_MAX_PAGES` (default `20`)
- `PATCH_SYNC_TIMEOUT_SECONDS` (default `30`)

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
