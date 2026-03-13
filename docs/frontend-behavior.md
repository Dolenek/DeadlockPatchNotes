# Frontend Behavior

## Runtime Model

- Framework: Next.js App Router (`web/app`).
- Rendering: route pages are server components.
- Client components currently in tree:
  - `FallbackImage`
  - `TableOfContents`
  - `PatchHeroesRail`
- Data source: API via `web/lib/api.ts`.
- Icon URL normalization in `web/lib/api.ts`:
  - checks `web/public/assets/mirror/manifest.json` for `remote-url -> local-path` mappings
  - when found, promotes local path to primary icon URL field
  - preserves remote URL as fallback field when available

API base URL resolution:

- Uses `API_BASE_URL` env when set.
- Falls back to `https://deadlockpatchnotes.com/api` when env is missing/blank.
- Normalizes exact `/api` suffix so both `https://deadlockpatchnotes.com` and `https://deadlockpatchnotes.com/api` are valid inputs.
- Invalid non-empty `API_BASE_URL` throws `Invalid API_BASE_URL: ...`.

SEO URL resolution:

- Uses `SITE_URL` env when set.
- Falls back to `https://www.deadlockpatchnotes.com` when env is missing/blank.
- Invalid non-empty `SITE_URL` throws `Invalid SITE_URL: ...`.

Fetch caching:

- All API fetches use `next: { revalidate: 30 }`.

## Route Map

### Root

- `/` is an indexable landing page.
  - fetches latest patch cards via `getPatches(1, 6)`
  - links to patch archive and entity timelines
  - emits `WebSite` + `CollectionPage` JSON-LD

### Patch Pages

- `/patches`
  - reads `page` query
  - normalizes with `clampPage` (invalid or `<1` -> `1`)
  - calls `getPatches(page, 12)`
  - metadata behavior:
    - page 1: canonical `/patches`, indexable
    - page > 1: canonical self (`/patches?page=N`), `noindex,follow`
  - emits `CollectionPage` JSON-LD
  - does not fetch per-card patch detail payloads
  - renders patch card grid + pagination
  - each card shows title/date on one row
  - card body renders follow-up timeline links (when present) that jump to timeline block anchors in `/patches/[slug]`
  - card images use Next image optimization with responsive `sizes`
  - masthead hero image uses local asset `/Oldgods_header.png`
- `/patches/[slug]`
  - calls `getPatchBySlug(slug)`
  - API `404` triggers `notFound()`
  - emits article metadata (canonical, OpenGraph, Twitter) and `Article` JSON-LD
  - renders timeline blocks + section renderer
  - patch timeline cross-links:
    - hero entry names link to `/heroes/[slug]`
    - item entry names link to `/items/[slug]`
    - hero ability group names link to `/spells/[slug]` (except `Talents` and `Card Types`)
  - desktop uses 3 rails:
    - left rail: table of contents (`TableOfContents`)
    - center rail: timeline content
    - right rail: hero quick-nav (`PatchHeroesRail`)
  - right rail behavior:
    - tracks active timeline block while scrolling
    - hero icon click jumps to that hero entry in active block
    - hidden when active block has no hero entries
    - hidden on tablet/mobile breakpoints
  - fallback behavior:
    - if patch has no timeline, synthesizes one display block
    - if timeline block has no sections, uses top-level patch sections for that block

### Entity Index Pages

- `/heroes` -> `getHeroes()`
- `/items` -> `getItems()`
- `/spells` -> `getSpells()`
- all emit index metadata and `ItemList` JSON-LD

Index-page API behavior:

- API `404` is treated as empty list.
- Other API errors are rethrown.

### Entity Detail Pages

- `/heroes/[slug]` -> `getHeroChanges(slug)`
- `/items/[slug]` -> `getItemChanges(slug)`
- `/spells/[slug]` -> `getSpellChanges(slug)`
- all emit detail metadata and entity-focused `WebPage` JSON-LD

Detail-page behavior:

- API `404` triggers `notFound()`.
- Non-404 API errors are rethrown.
- Timeline rows include link back to corresponding patch detail page.
- Timeline row titles link to the exact source timeline block in `/patches/[slug]` when block IDs can be mapped; otherwise they fall back to patch root.
- Hero timeline skill titles link to `/spells/[slug]` (except meta groups `Talents` and `Card Types`).
- Spell timeline hero names link to `/heroes/[slug]` when hero slug metadata is present.

## Shared UI and Rendering Rules

### Global Layout + Navigation

- `TopNav` is rendered globally in `app/layout.tsx` for all routes.
- global metadata defaults are defined in `app/layout.tsx` (`metadataBase`, title template, robots, OpenGraph, Twitter).
- Global texture stack is rendered behind all routes in layout:
  - base: `/bg_texture.jpg`
  - mid layer: `/bg_texture_dark.jpg`
  - deep layer: `/bg_texture_darkest.jpg`
  - seams use `scratch_mask_*` overlays between layer transitions
- Top nav includes:
  - brand link points to `/`
  - internal links: patches, heroes, spells, items
  - docs/links: changelog forum, PatchNotes API (`/api/scalar`), assets API docs, Steam store page

### SEO Utility Routes

- `GET /robots.txt`
  - allows crawl on content routes
  - disallows `/api`, `/api/`, `/image-proxy`
  - includes sitemap pointer
- `GET /sitemap.xml`
  - includes core content pages (`/`, `/patches`, `/heroes`, `/items`, `/spells`)
  - includes dynamic detail URLs from patches/heroes/items/spells APIs
  - intentionally excludes paginated `/patches?page=N` URLs
  - revalidated using the shared API fetch cache window (`30s`)

### Image Proxy Route

- Web utility route for remote image sampling is `GET /image-proxy?url=<https-url>`.
- This route is intentionally outside `/api/*` to avoid overlap with backend API reverse-proxy routing.

### `FallbackImage` (`use client`)

- Starts with primary `src`.
- On load error, swaps once to `fallbackSrc` when available.
- Returns `null` when both image sources are absent.

### `PatchSectionRenderer`

- Renders section entries by section `kind`.
- Hero/item sections use portrait-style entry image class.
- Entry-level changes and grouped changes are rendered independently.
- Section anchor IDs are built via `sectionAnchor(section.id)`.

### `Pagination`

- Hidden when `totalPages <= 1`.
- Previous/next links are always rendered with disabled styling at bounds.

### `TableOfContents` (`use client`)

- Collapsible section anchor list component.
- Mounted on patch detail page left rail.
- Tracks active timeline/section anchors using `IntersectionObserver`.

### `PatchHeroesRail` (`use client`)

- Mounted on patch detail page right rail (desktop only).
- Tracks active timeline block with `IntersectionObserver`.
- Displays hero icon links for active block only.
- Links target hero entry anchors in the patch timeline content.

## Error Surfaces

`APIError` from `web/lib/api.ts`:

- `status=0`: network/fetch failure.
- `status>0`: non-2xx API response.

Route-specific notes:

- Entity index pages (heroes/items/spells): `404` -> empty list.
- Patch list page (`/patches`): no special `404` handling, errors propagate.
- Detail pages: `404` -> `notFound()`.

`not-found.tsx` is generic and reused for all `notFound()` calls:

- headline: `Oops.`
- includes centered `/lil_troopers.png` artwork
- primary action links to `/patches`
- metadata sets `noindex,nofollow`

## Utility Semantics

- `formatDisplayDate`: long `en-US` date.
- `formatForumDate`: UTC `MM-DD-YYYY`.
- `formatUpdateLabel`:
  - `initial` -> `Update MM-DD-YYYY`
  - other kinds -> `Patch MM-DD-YYYY`
