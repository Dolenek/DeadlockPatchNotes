# Deadlock Patch Notes

UI-first implementation of a Deadlock patch notes site inspired by League of Legends and Dota patch pages.

## Monorepo Layout

- `web/`: Next.js App Router frontend (SSR React)
- `api/`: Go API (`net/http` + `chi`) with typed patch contracts

## Local Run

### API

```bash
cd api
go mod tidy
go run ./cmd/server
```

Server listens on `http://localhost:8080` by default.

### Web

```bash
cd web
npm install
npm run dev
```

Frontend runs on `http://localhost:3000` and calls `API_BASE_URL` (default `http://localhost:8080`).

## API Endpoints

- `GET /api/healthz`
- `GET /api/v1/patches?page=<int>&limit=<int>`
- `GET /api/v1/patches/{slug}`

## V1 Notes

- Uses one hardcoded patch dataset (`Gameplay Update - 03-06-2026`) from:
  `https://store.steampowered.com/news/app/1422450/view/519740319207522795`
- Ingestion/sync pipeline from forum + Steam is intentionally deferred.
