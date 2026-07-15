<p align="center">
  <img src="./web/public/deadlock_logo.webp" alt="Deadlock Patch Notes" width="96" />
</p>

<h1 align="center">Deadlock Patch Notes</h1>

<p align="center">
  Prohledávatelný komunitní archiv aktualizací hry Deadlock, změn hrdinů, historie předmětů a vývoje schopností.
</p>

<p align="center">
  <a href="https://www.deadlockpatchnotes.com">Živý web</a>
  &middot;
  <a href="./docs/index.md">Dokumentace</a>
  &middot;
  <a href="https://www.deadlockpatchnotes.com/api/scalar">API</a>
  &middot;
  <a href="./README.en.md">English</a>
</p>

![Domovská stránka Deadlock Patch Notes](./web/public/readme-home.PNG)

## O projektu

Deadlock Patch Notes převádí oficiální oznámení aktualizací hry Deadlock do strukturovaného archivu, který lze pohodlně procházet, vyhledávat a používat programově. Projekt spojuje automatizovaný sběr dat, veřejné API a responzivní webovou aplikaci v jednom monorepu.

## Hlavní funkce

- Historie patchů s časovou osou vydání a navazujících hotfixů.
- Přehled změn hrdinů, předmětů a schopností napříč aktualizacemi.
- Normalizovaná JSON data dostupná přes veřejné API.
- RSS feedy pro nové patche, konkrétní hrdiny a dobu od jejich poslední změny.
- Serverově renderované stránky, responzivní rozhraní a SEO metadata.

## Jak projekt funguje

1. Jednorázový sync proces načte oficiální changelogy z Deadlock fóra a Steam Web API.
2. Go ingestion pipeline obsah parsuje, deduplikuje a převádí do normalizovaných patchů a časových os.
3. Strukturovaná data, release bloky a informace o sync bězích se uloží do PostgreSQL.
4. Go API nad databází vytváří cachovaný read model a poskytuje JSON endpointy, OpenAPI dokumentaci a RSS feedy.
5. Next.js frontend čte API, serverově renderuje archiv a nabízí samostatné historie patchů, hrdinů, předmětů a schopností.

```text
Deadlock fórum + Steam Web API
              ↓
       Go sync pipeline
              ↓
          PostgreSQL
              ↓
       Go HTTP API + RSS
              ↓
       Next.js frontend
```

## Architektura

| Adresář | Odpovědnost |
| --- | --- |
| `api/` | Go HTTP API, ingestion a sync proces, databázové migrace a PostgreSQL persistence. |
| `web/` | Next.js App Router frontend, typovaný API klient a uživatelské rozhraní archivu. |
| `scripts/` | Generování fixtures, zrcadlení assetů, serverová automatizace a údržbové kontroly. |
| `docs/` | Kanonická dokumentace runtime chování, API kontraktů, parseru, vývoje a provozu. |

API, web a sync jsou oddělené procesy. Migrace vytváří samostatnou read-only roli pro API a omezenou zapisovací roli pro sync, takže běžící služby nepotřebují vlastnická oprávnění k databázi. Podrobnosti popisuje [dokumentace architektury](./docs/architecture.md).

## Použité technologie

<p>
  <img alt="Frontend: Next.js + TypeScript" src="https://img.shields.io/badge/frontend-Next.js%20%2B%20TypeScript-111827?style=flat-square&logo=nextdotjs" />
  <img alt="Backend: Go API" src="https://img.shields.io/badge/backend-Go%20API-00ADD8?style=flat-square&logo=go&logoColor=white" />
  <img alt="Databáze: PostgreSQL" src="https://img.shields.io/badge/database-PostgreSQL-4169E1?style=flat-square&logo=postgresql&logoColor=white" />
  <img alt="Deployment: Docker" src="https://img.shields.io/badge/deployment-Docker-2496ED?style=flat-square&logo=docker&logoColor=white" />
</p>

- **Frontend:** Next.js 16, React 19 a TypeScript 5.8.
- **Backend:** Go 1.25, Chi router a pgx.
- **Databáze:** PostgreSQL 16.
- **Testování:** Vitest, TypeScript compiler a standardní Go testy.
- **Provoz:** Docker a Docker Compose.

## Lokální spuštění

Nejjednodušší je spustit celý projekt přes Docker Compose. Zkopírujte výchozí konfiguraci a v `.env` nastavte vlastní hodnoty `POSTGRES_PASSWORD`, `API_DB_PASSWORD` a `SYNC_DB_PASSWORD`:

```bash
cp .env.example .env
docker-compose up -d --build db migrate api web
docker-compose run --rm sync
```

Web poběží na `http://localhost:3000` a API ve výchozím nastavení Compose na `http://localhost:18081`.

### Ruční spuštění

Pro ruční běh potřebujete Node.js 24+, npm, Go 1.25+ a PostgreSQL 16+. Nejprve aplikujte migrace a vytvořte oddělené runtime role:

```bash
cd api
API_DB_PASSWORD='replace-with-a-distinct-api-password' \
SYNC_DB_PASSWORD='replace-with-a-distinct-sync-password' \
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/migrate
```

Spusťte API:

```bash
cd api
DATABASE_URL='postgres://deadlock_api:replace-with-a-distinct-api-password@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/server
```

V samostatném terminálu spusťte frontend:

```bash
cd web
npm install
API_BASE_URL=http://localhost:8080 npm run dev
```

Volitelně spusťte jeden sync běh:

```bash
cd api
DATABASE_URL='postgres://deadlock_sync:replace-with-a-distinct-sync-password@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

## Testy a kontrola kvality

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

Kontrola doporučených limitů velikosti zdrojových souborů a funkcí:

```bash
node scripts/check_source_limits.mjs
```

## Dokumentace

- [Přehled dokumentace](./docs/index.md)
- [Runtime overview](./docs/runtime-overview.md)
- [API contracts](./docs/api-contracts.md)
- [Development workflow](./docs/development.md)
- [Ops and scripts](./docs/ops-and-scripts.md)
- [Architecture](./docs/architecture.md)
