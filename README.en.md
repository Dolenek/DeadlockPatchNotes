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
  <a href="https://www.deadlockpatchnotes.com/api/scalar">API docs</a>
  &middot;
  <a href="./docs/index.md">Project docs</a>
  &middot;
  <a href="./docs/architecture.md">Architecture</a>
</p>

<p align="center">
  <img alt="Frontend: Next.js + TypeScript" src="https://img.shields.io/badge/frontend-Next.js%20%2B%20TypeScript-111827?style=flat-square&logo=nextdotjs" />
  <img alt="Backend: Go API" src="https://img.shields.io/badge/backend-Go%20API-00ADD8?style=flat-square&logo=go&logoColor=white" />
  <img alt="Database: PostgreSQL" src="https://img.shields.io/badge/database-PostgreSQL-4169E1?style=flat-square&logo=postgresql&logoColor=white" />
  <img alt="Deployment: Docker" src="https://img.shields.io/badge/deployment-Docker-2496ED?style=flat-square&logo=docker&logoColor=white" />
</p>

![Deadlock Patch Notes homepage](./web/public/readme-home.PNG)

## What It Does

Deadlock Patch Notes turns official Deadlock update posts into a structured archive that is easy to browse, search, and consume programmatically.

- Browse patch history with release timelines and follow-up hotfixes.
- Track hero, item, and spell changes across updates.
- Read normalized JSON payloads through a public API.
- Subscribe to patch and hero-specific RSS feeds.

## Why It Is Interesting

This is a full-stack monorepo built around a real data pipeline, not just a static frontend.

- **Automated ingestion:** sync process crawls official changelog sources, parses patch content, normalizes sections, and persists structured payloads.
- **Backend read model:** Go API serves cached PostgreSQL snapshots for patch details, entity timelines, OpenAPI docs, and RSS feeds.
- **Frontend experience:** Next.js App Router UI renders patch pages, hero/item/spell history, SEO metadata, and responsive archive views.
- **Operational workflow:** Docker Compose runs the database, API, web app, and one-shot sync process locally or on a server.

## Architecture

```text
api/       Go HTTP API, ingestion sync process, PostgreSQL persistence
web/       Next.js frontend, typed API client, archive UI
scripts/   fixture generation, asset mirroring, maintenance checks
docs/      runtime behavior, API contracts, parser rules, operations
```

The runtime flow is:

```text
Deadlock changelog sources
        |
        v
Go sync pipeline
        |
        v
PostgreSQL read model
        |
        v
Go HTTP API + RSS
        |
        v
Next.js frontend
```

Read more in the [runtime overview](./docs/runtime-overview.md), [API contracts](./docs/api-contracts.md), and [architecture notes](./docs/architecture.md).

## Run Locally

### Docker

```bash
cp .env.example .env
docker-compose up -d --build db api web
docker-compose run --rm sync
```

### API

```bash
cd api
go mod tidy
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/server
```

### Web

```bash
cd web
npm install
npm run dev
```

### One Sync Pass

```bash
cd api
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

## Documentation

- [Documentation index](./docs/index.md)
- [Runtime overview](./docs/runtime-overview.md)
- [API contracts](./docs/api-contracts.md)
- [Development workflow](./docs/development.md)
- [Ops and scripts](./docs/ops-and-scripts.md)

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
