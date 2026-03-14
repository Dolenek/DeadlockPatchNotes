# Deadlock Patch Notes

Deadlock Patch Notes is a searchable archive of Deadlock updates across heroes, items, and spells.
This monorepo powers deadlockpatchnotes.com with a Go API, Next.js frontend, and an automated forum/Steam ingestion pipeline.

## Live Site

- https://www.deadlockpatchnotes.com

## What You Can Do

- Browse patch history with release timelines.
- Track hero balance changes over time.
- Explore item and spell change timelines.
- Consume JSON and RSS data from the API.

## Quick Start

### Local Development

Run API:

```bash
cd api
go mod tidy
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/server
```

Run web:

```bash
cd web
npm install
npm run dev
```

Run one sync pass:

```bash
cd api
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

### Docker

```bash
cp .env.example .env
docker-compose up -d --build db api web
docker-compose run --rm sync
```

## Monorepo Layout

- `api/`: Go HTTP API, sync process, and PostgreSQL persistence.
- `web/`: Next.js App Router frontend.
- `scripts/`: fixture generation, asset mirroring, and maintenance scripts.

## Documentation

- Docs index: [`docs/index.md`](./docs/index.md)
- Runtime overview: [`docs/runtime-overview.md`](./docs/runtime-overview.md)
- API contracts: [`docs/api-contracts.md`](./docs/api-contracts.md)
- Development workflow: [`docs/development.md`](./docs/development.md)
- Ops and scripts: [`docs/ops-and-scripts.md`](./docs/ops-and-scripts.md)

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

Source guidance report:

```bash
node scripts/check_source_limits.mjs
```
