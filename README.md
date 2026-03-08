# Deadlock Patch Notes

Deadlock patch notes monorepo with:
- DB-backed Go API
- Next.js frontend
- Automated forum/Steam ingestion pipeline

## Documentation

- Main docs index: `docs/index.md`
- Architecture overview: `docs/architecture.md`
- Development workflow: `docs/development.md`

## Monorepo Layout

- `web/`: Next.js App Router frontend (SSR React)
- `api/`: Go API (`net/http` + `chi`) with PostgreSQL storage
- `scripts/`: helper scripts for checks and server automation

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
API_BASE_URL=http://localhost:8080 npm run dev
```

Frontend runs on `http://localhost:3000`.

### Sync patch notes into DB

```bash
cd api
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

Default source is `https://forums.playdeadlock.com/forums/changelog.10/`.

## API Endpoints

- `GET /api/healthz`
- `GET /api/v1/patches?page=<int>&limit=<int>`
- `GET /api/v1/patches/{slug}`

## Docker Deployment

```bash
cp .env.example .env
# edit .env (especially POSTGRES_PASSWORD)
docker-compose up -d --build db api web
```

Run one ingestion pass:

```bash
docker-compose run --rm sync
```

Install sync + backup cron jobs:

```bash
./scripts/server/install_cron.sh
```

## SQL Drafts

- Legacy draft schema is in `api/sql/drafts/001_patchnotes_schema_postgres.sql`
- Active runtime migration is in `api/internal/db/migrations/001_patchnotes.sql`
