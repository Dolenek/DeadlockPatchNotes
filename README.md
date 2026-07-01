<p align="center">
  <img src="./web/public/deadlock_logo.webp" alt="Deadlock Patch Notes" width="96" />
</p>

<h1 align="center">Deadlock Patch Notes</h1>

<p align="center">
  Prohledávatelný komunitní archiv aktualizací hry Deadlock, změn hrdinů, historie předmětů a vývoje schopností.
</p>

<p align="center">
  <a href="./README.en.md">English README</a>
  &middot;
  <a href="https://www.deadlockpatchnotes.com">Živý web</a>
  &middot;
  <a href="https://www.deadlockpatchnotes.com/api/scalar">API dokumentace</a>
  &middot;
  <a href="./docs/index.md">Projektová dokumentace</a>
  &middot;
  <a href="./docs/architecture.md">Architektura</a>
</p>

<p align="center">
  <img alt="Frontend: Next.js + TypeScript" src="https://img.shields.io/badge/frontend-Next.js%20%2B%20TypeScript-111827?style=flat-square&logo=nextdotjs" />
  <img alt="Backend: Go API" src="https://img.shields.io/badge/backend-Go%20API-00ADD8?style=flat-square&logo=go&logoColor=white" />
  <img alt="Databáze: PostgreSQL" src="https://img.shields.io/badge/database-PostgreSQL-4169E1?style=flat-square&logo=postgresql&logoColor=white" />
  <img alt="Deployment: Docker" src="https://img.shields.io/badge/deployment-Docker-2496ED?style=flat-square&logo=docker&logoColor=white" />
</p>

![Domovská stránka Deadlock Patch Notes](./web/public/readme-home.PNG)

## Co projekt dělá

Deadlock Patch Notes převádí oficiální update posty ke hře Deadlock do strukturovaného archivu, který se dá pohodlně procházet, vyhledávat a používat přes API.

- Zobrazuje historii patchů včetně release timeline a navazujících hotfixů.
- Sleduje změny hrdinů, předmětů a schopností napříč aktualizacemi.
- Vystavuje normalizovaná JSON data přes veřejné API.
- Nabízí RSS feedy pro patche i konkrétní hrdiny.

## Proč je to technicky zajímavé

Jde o full-stack monorepo postavené kolem reálné datové pipeline, ne jen o statický frontend.

- **Automatizovaný ingestion:** sync proces prochází oficiální changelog zdroje, parsuje obsah patchů, normalizuje sekce a ukládá strukturovaná data.
- **Backend read model:** Go API servíruje cachované PostgreSQL snapshoty pro detail patchů, entity timelines, OpenAPI dokumentaci a RSS feedy.
- **Frontend aplikace:** Next.js App Router renderuje patch stránky, historii hrdinů/předmětů/schopností, SEO metadata a responzivní archiv.
- **Provozní workflow:** Docker Compose lokálně nebo na serveru spustí databázi, API, web i jednorázový sync proces.

## Architektura

```text
api/       Go HTTP API, ingestion sync proces, PostgreSQL persistence
web/       Next.js frontend, typovaný API klient, archivní UI
scripts/   generování fixtures, zrcadlení assetů, maintenance kontroly
docs/      runtime chování, API kontrakty, parser pravidla, operations
```

Runtime tok:

```text
Deadlock changelog zdroje
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

Více detailů je v [runtime overview](./docs/runtime-overview.md), [API contracts](./docs/api-contracts.md) a [architecture notes](./docs/architecture.md).

## Lokální spuštění

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

### Jeden sync běh

```bash
cd api
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

## Dokumentace

- [Documentation index](./docs/index.md)
- [Runtime overview](./docs/runtime-overview.md)
- [API contracts](./docs/api-contracts.md)
- [Development workflow](./docs/development.md)
- [Ops and scripts](./docs/ops-and-scripts.md)

## Kontroly kvality

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

Kontrola doporučených limitů zdrojových souborů:

```bash
node scripts/check_source_limits.mjs
```
