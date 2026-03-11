# Ingestion and Parser Rules

## Sync Entry Point

`api/cmd/sync/main.go` runs one ingestion pass:

1. Read env config.
2. Open DB + apply migrations.
3. Run `ingest.RunPatchSync`.
4. Print discovered/processed/inserted/updated counters.

Config defaults and validation:

- `PATCH_FORUM_URL` default: `https://forums.playdeadlock.com/forums/changelog.10/`
- `PATCH_SYNC_MAX_PAGES` default: `20`
- `PATCH_SYNC_TIMEOUT_SECONDS` default: `30`
- Non-integer or non-positive values for `PATCH_SYNC_MAX_PAGES` and `PATCH_SYNC_TIMEOUT_SECONDS` fall back to defaults.

## Sync Run Semantics

`RunPatchSync` is best-effort per discovered thread.

- Thread fetch/parse/build/upsert failures are skipped and sync continues.
- Run status can still finish as `success` when some discovered threads fail processing.
- `processed_threads` counts only successfully upserted threads.

Asset catalog behavior:

- Sync attempts to load heroes/items catalog.
- If catalog load fails, sync continues with `catalog=nil` (degraded structured parsing/icon enrichment).

## Thread Discovery

`CrawlChangelogThreads` behavior:

- Crawls paginated listing using `rel="next"` until:
  - max pages reached, or
  - page loop detected, or
  - no next page.
- Keeps unique thread URLs.
- Includes only URLs where:
  - path starts with `/threads/`
  - path contains `-update`
- Excludes URLs containing `changelog-feedback-process`.

## Thread Ordering

After discovery, thread refs are sorted lexicographically by URL before processing.

## Thread Parsing

`FetchThread` behavior:

- Parses thread slug and numeric thread ID from URL path.
- Keeps posts authored by `Yoshi` only.
- For each accepted post:
  - extracts `postID`
  - parses publish time (`YYYY-MM-DDTHH:MM:SS-0700`)
  - extracts normalized forum body text
  - captures Steam news URL when present
  - captures preview image when present
  - builds canonical forum post URL `<thread>/post-<postID>`

Post filtering:

- If both forum body and Steam URL are empty, post is dropped.
- Forum body that looks like a Steam unfurl card is intentionally collapsed to empty text.

## Steam Event Parsing

When a post links Steam news:

1. Read `data-partnereventstore` payload.
2. Decode JSON and use the first event entry.
3. Parse announcement headline/body/post time.
4. Normalize BBCode-like markup to plain text.
5. Split body into timeline blocks:
   - first block defaults to `initial`
   - headings `MM-DD-YYYY Patch:` start `hotfix` blocks
   - parsed heading dates are normalized to `12:00:00Z`
6. Resolve hero image from:
   - page `og:image`, else
   - Steam capsule image metadata

If Steam parsing fails for a post, forum fallback is used only when forum body text is non-empty; otherwise that post yields no block.

## Block Assembly and Precedence

`buildPatchFromThread` rules:

- For each post:
  - prefer Steam body blocks when available
  - otherwise use one forum block from forum body
- Block dedupe is SHA1 over normalized lowercase body text.
- Empty normalized bodies are dropped.
- Blocks are sorted ascending by release time (key tie-break).
- First effective block is forced to `kind=initial` and title `Initial Update` when needed.
- Patch `published_at` = first block release time.
- Patch `updated_at` = last block release time.

## Structured Section Parsing

`buildStructuredSections` converts block text into canonical sections.

Supported patterns:

- Explicit section headers:
  - `[ General ]`, `[ Items ]`, `[ Heroes ]`
  - plain `General`, `Items`, `Heroes`
- Prefixed lines: `Prefix: change text`
- Non-prefixed continuation lines

Inference and matching behavior:

- Hero/item/ability matching uses loaded asset catalog.
- Hero aliases:
  - `doorman` -> `the doorman`
  - `vindcita` -> `vindicta`
- Item aliases:
  - `backstabber` -> `stalker`
- Item rename extraction:
  - `Renamed to <name>` can map to destination item asset.

Hero grouping behavior:

- Ability-prefixed lines are bound to current hero ability groups.
- `Talents ...` lines populate `Talents` group.
- `Card Types` heading opens `Card Types` group.
- `Spades|Diamond|Hearts|Clubs|Joker` prefixed lines remain under `Card Types`.
- Ability/card-type prefixes do not become standalone hero entries.

Item behavior:

- Item-prefixed line with empty change text becomes `Updated.`.

Skipped lines:

- empty lines
- `Read more`
- boilerplate Steam news heading lines
- date heading lines used for block splitting

## Fallback Section Strategy

If structured parse yields no sections:

- Build one `general` section from timeline blocks.
- If still empty, emit one `Core Gameplay` entry with fallback line.

## Persistence Rules

`upsertPatch` transaction:

1. Insert or update `patches` row by slug.
2. Replace all `patch_release_blocks` rows for that patch ID.
3. Insert timeline block metadata with stable sort order.
4. Commit transaction.

Slug/thread identity contract:

- Upsert lookup key is `slug`.
- Database uniqueness is enforced on both `slug` and `thread_id`.

`patch_release_blocks` notes:

- `source_url` and `post_id` are nullable.
- `kind` and `source_type` are free-text fields (no DB enum/check constraint).

## Sync Observability

- `sync_runs` row is inserted with `status='running'` at start.
- Final row updates status, counters, finish timestamp, and message.
- `sync_runs.status` is free-text in schema.

## External Dependencies Used by Ingest

- Forum changelog pages:
  - `https://forums.playdeadlock.com/forums/changelog.10/`
- Steam news pages:
  - `https://store.steampowered.com/news/app/1422450/view/...`
- Asset catalog:
  - `https://assets.deadlock-api.com/v2/heroes`
  - `https://assets.deadlock-api.com/v2/items`
