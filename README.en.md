<p align="center">
  <img src="./web/public/deadlock_logo.webp" alt="Deadlock Patch Notes" width="96" />
</p>

<h1 align="center">Deadlock Patch Notes</h1>

<p align="center">
  A searchable community archive for Deadlock updates, hero balance changes, item history, and spell timelines.
</p>

<p align="center">
  <a href="https://www.deadlockpatchnotes.com">Live site</a>
  &middot;
  <a href="./docs/index.md">Documentation</a>
  &middot;
  <a href="https://www.deadlockpatchnotes.com/api/scalar">API</a>
  &middot;
  <a href="./README.md">Čeština</a>
</p>

![Deadlock Patch Notes homepage](./web/public/readme-home.PNG)

## About

Deadlock Patch Notes turns official Deadlock update announcements into a structured archive that is easy to browse, search, and consume programmatically. The project combines automated data ingestion, a public API, and a responsive web application in a single monorepo.

## Key Features

- Patch history with release timelines and follow-up hotfixes.
- Hero, item, and spell changes tracked across updates.
- Normalized JSON data available through a public API.
- RSS feeds for new patches, individual heroes, and time since their last change.
- Server-rendered pages, a responsive interface, and SEO metadata.

## How It Works

1. A one-shot sync process reads official changelogs from the Deadlock forum and Steam Web API.
2. The Go ingestion pipeline parses and deduplicates the content, then converts it into normalized patches and timelines.
3. Structured payloads, release blocks, and sync run information are stored in PostgreSQL.
4. The Go API builds a cached read model over the database and exposes JSON endpoints, OpenAPI documentation, and RSS feeds.
5. The Next.js frontend consumes the API, server-renders the archive, and provides dedicated patch, hero, item, and spell histories.

```text
Deadlock forum + Steam Web API
              ↓
       Go sync pipeline
              ↓
          PostgreSQL
              ↓
       Go HTTP API + RSS
              ↓
       Next.js frontend
```

## Architecture

| Directory | Responsibility |
| --- | --- |
| `api/` | Go HTTP API, ingestion and sync process, database migrations, and PostgreSQL persistence. |
| `web/` | Next.js App Router frontend, typed API client, and archive user interface. |
| `scripts/` | Fixture generation, asset mirroring, server automation, and maintenance checks. |
| `docs/` | Canonical documentation for runtime behavior, API contracts, parsing, development, and operations. |

The API, web app, and sync job run as separate processes. Migrations provision a read-only API role and a scoped write role for sync, so runtime services do not need database-owner privileges. See the [architecture documentation](./docs/architecture.md) for details.

## Technology Stack

<p>
  <img alt="Frontend: Next.js + TypeScript" src="https://img.shields.io/badge/frontend-Next.js%20%2B%20TypeScript-111827?style=flat-square&logo=nextdotjs" />
  <img alt="Backend: Go API" src="https://img.shields.io/badge/backend-Go%20API-00ADD8?style=flat-square&logo=go&logoColor=white" />
  <img alt="Database: PostgreSQL" src="https://img.shields.io/badge/database-PostgreSQL-4169E1?style=flat-square&logo=postgresql&logoColor=white" />
  <img alt="Deployment: Docker" src="https://img.shields.io/badge/deployment-Docker-2496ED?style=flat-square&logo=docker&logoColor=white" />
</p>

- **Frontend:** Next.js 16, React 19, and TypeScript 5.8.
- **Backend:** Go 1.25, the Chi router, and pgx.
- **Database:** PostgreSQL 16.
- **Testing:** Vitest, the TypeScript compiler, and standard Go tests.
- **Operations:** Docker and Docker Compose.

## Running Locally

Docker Compose is the simplest way to start the complete project. Copy the default configuration and set custom `POSTGRES_PASSWORD`, `API_DB_PASSWORD`, and `SYNC_DB_PASSWORD` values in `.env`:

```bash
cp .env.example .env
docker-compose up -d --build db migrate api web
docker-compose run --rm sync
```

The web app will be available at `http://localhost:3000`; the Compose API defaults to `http://localhost:18081`.

### Manual Setup

A manual setup requires Node.js 24+, npm, Go 1.25+, and PostgreSQL 16+. Apply migrations and provision the separate runtime roles first:

```bash
cd api
API_DB_PASSWORD='replace-with-a-distinct-api-password' \
SYNC_DB_PASSWORD='replace-with-a-distinct-sync-password' \
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/migrate
```

Start the API:

```bash
cd api
DATABASE_URL='postgres://deadlock_api:replace-with-a-distinct-api-password@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/server
```

Start the frontend in a separate terminal:

```bash
cd web
npm install
API_BASE_URL=http://localhost:8080 npm run dev
```

Optionally, run one sync pass:

```bash
cd api
DATABASE_URL='postgres://deadlock_sync:replace-with-a-distinct-sync-password@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

## Tests and Quality Checks

Frontend:

```bash
cd web
npm run lint
npm run test
npm run build
npm run test:runtime
```

Backend:

```bash
cd api
go test ./...
```

Check the recommended source-file and function-size limits:

```bash
node scripts/check_source_limits.mjs
```

## Documentation

- [Documentation index](./docs/index.md)
- [Runtime overview](./docs/runtime-overview.md)
- [API contracts](./docs/api-contracts.md)
- [Development workflow](./docs/development.md)
- [Ops and scripts](./docs/ops-and-scripts.md)
- [Architecture](./docs/architecture.md)
