# Ops and Scripts

## Docker Deployment Model

`docker-compose.yml` defines four services:

- `db`: PostgreSQL 16 with persisted `pgdata` volume.
- `api`: Go API process (`server`) on internal port `8080`.
- `web`: Next.js app on internal port `3000`.
- `sync`: one-shot ingestion process (`sync` binary).

Networking and publish defaults:

- API publish: `${API_HOST_BIND:-0.0.0.0}:${API_PORT:-18081}:8080`
- Web publish: `${WEB_HOST_BIND:-127.0.0.1}:${WEB_PORT:-3000}:3000`

## Environment Variables

Compose `.env`:

- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `API_HOST_BIND`
- `API_PORT`
- `API_READ_CACHE_TTL`
- `WEB_HOST_BIND`
- `WEB_PORT`
- `PATCH_FORUM_URL`
- `PATCH_SYNC_MAX_PAGES`
- `PATCH_SYNC_TIMEOUT_SECONDS`

Web local env (`web/.env.example`):

- `API_BASE_URL`

Runtime defaults in code:

- API address default `:8080`.
- Web API client default base URL: `https://deadlockpatchnotes.com/api`.
- `API_BASE_URL` exact `/api` suffix is normalized in web client config parsing.
- Sync defaults:
  - changelog URL `https://forums.playdeadlock.com/forums/changelog.10/`
  - max pages `20`
  - HTTP timeout `30` seconds
- Invalid/non-positive sync numeric env values fall back to defaults.

## Build and Run Commands

### API

```bash
cd api
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/server
```

### Sync

```bash
cd api
DATABASE_URL='postgres://deadlock:deadlock@localhost:5432/deadlock_patchnotes?sslmode=disable' go run ./cmd/sync
```

### Web

```bash
cd web
npm run dev
```

### Compose

```bash
docker-compose up -d --build db api web
docker-compose run --rm sync
```

## Server Automation Scripts

- `scripts/server/run_sync.sh`
  - requires repository `.env`
  - runs one `docker-compose run --rm sync`
  - calls `/usr/bin/docker-compose` explicitly
- `scripts/server/backup_postgres.sh`
  - requires repository `.env`
  - performs `pg_dump -Fc` from compose `db`
  - writes `backups/patchnotes-<UTC timestamp>.dump`
  - deletes backups older than 14 days
  - calls `/usr/bin/docker-compose` explicitly
- `scripts/server/install_cron.sh`
  - installs two cron entries:
    - sync every 6 hours
    - backup daily at 03:30
  - logs to `/var/log/deadlockpatchnotes-*.log`

## Fixture Pipeline Scripts

### `scripts/generate_patch_fixture.mjs`

Purpose:

- Generates fixture JSON at `api/internal/patches/data/<slug>.json`.
- Mirrors referenced assets into `web/public/assets/...`.
- Writes patch asset manifest JSON.

Behavior details:

- Clears patch asset output directory before re-downloading assets.
- Asset download failures are logged as warnings and do not fail the whole run.

### `scripts/patch_fixture/config.mjs`

- Defines fixed target patch identifiers (slug/GID/title/source URLs).
- Defines output paths for fixture, manifest, and mirrored assets.

### `scripts/patch_fixture/build_patch_detail.mjs`

- Finds target Steam news item.
- Parses general/items/heroes sections.
- Resolves hero/item/ability icon metadata.
- Builds canonical fixture payload + summary stats.

### `scripts/patch_fixture/lookups.mjs`

- Hero/item lookup tables and aliases.
- Group helpers for hero abilities, talents, and card types.

### `scripts/patch_fixture/text_parse.mjs`

- Steam text cleanup.
- Section splitting and bullet parsing helpers.

### `scripts/patch_fixture/assets.mjs`

- HTTP fetch wrappers with user-agent.
- Asset registry and local file download writes.

### `scripts/patch_fixture/utils.mjs`

- String normalization/slug/hash helpers.
- Ability-prefix matching helpers.

### `scripts/audit_api_icons.mjs`

Purpose:

- Crawls production-compatible API endpoints and inventories icon/media URL fields used by web pages.
- Reports unique URLs and classifies each as local existing, local missing, remote allowed-host, remote disallowed-host, or other.

Behavior details:

- Scans list + detail endpoints for patches/heroes/items/spells.
- Writes JSON and CSV reports (defaults to `/tmp/deadlock-icon-audit-<timestamp>.{json,csv}`).
- Accepts API base URL override and detail fetch concurrency.

### `scripts/mirror_api_icons.mjs`

Purpose:

- Downloads allowed-host remote icon URLs from an audit report into deduped local paths.
- Writes mirror manifest used by `web/lib/api.ts` URL normalization.

Behavior details:

- Input is required via `--audit <path>` using output from `audit_api_icons.mjs`.
- Stores assets under `web/public/assets/mirror/icons/<sha1>.<ext>`.
- Writes/updates `web/public/assets/mirror/manifest.json` with `url -> localPath` mappings.
- Keeps failed URLs in manifest `failed` metadata and logs warnings.

## Source Guidance Script

`scripts/check_source_limits.mjs`:

- Walks `api`, `web`, `scripts`, `docs`.
- Excludes generated/vendor-style paths.
- Reports:
  - file-length warnings/errors (`>400`, `>500` lines)
  - function-length violations (`>40` lines via heuristic pattern scanning)

## Active vs Draft SQL

- Active runtime migration:
  - `api/internal/db/migrations/001_patchnotes.sql`
- Draft/legacy reference:
  - `api/sql/drafts/001_patchnotes_schema_postgres.sql`
