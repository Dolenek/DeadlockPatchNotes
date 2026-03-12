# Frontend Behavior

## Runtime Model

- Framework: Next.js App Router (`web/app`).
- Rendering: route pages are server components.
- Client components currently in tree:
  - `FallbackImage`
  - `TableOfContents` (implemented but not mounted in routes)
- Data source: API via `web/lib/api.ts`.

API base URL resolution:

- Uses `API_BASE_URL` env when set.
- Falls back to `https://api.deadlock.jakubdolenek.xyz` when env is missing/blank.
- Invalid non-empty `API_BASE_URL` throws `Invalid API_BASE_URL: ...`.

Fetch caching:

- All API fetches use `next: { revalidate: 30 }`.

## Route Map

### Root

- `/` redirects to `/patches`.

### Patch Pages

- `/patches`
  - reads `page` query
  - normalizes with `clampPage` (invalid or `<1` -> `1`)
  - calls `getPatches(page, 12)`
  - renders patch card grid + pagination
  - masthead hero image uses local asset `/Oldgods_header.png`
- `/patches/[slug]`
  - calls `getPatchBySlug(slug)`
  - API `404` triggers `notFound()`
  - renders timeline blocks + section renderer
  - fallback behavior:
    - if patch has no timeline, synthesizes one display block
    - if timeline block has no sections, uses top-level patch sections for that block

### Entity Index Pages

- `/heroes` -> `getHeroes()`
- `/items` -> `getItems()`
- `/spells` -> `getSpells()`

Index-page API behavior:

- API `404` is treated as empty list.
- Other API errors are rethrown.

### Entity Detail Pages

- `/heroes/[slug]` -> `getHeroChanges(slug)`
- `/items/[slug]` -> `getItemChanges(slug)`
- `/spells/[slug]` -> `getSpellChanges(slug)`

Detail-page behavior:

- API `404` triggers `notFound()`.
- Non-404 API errors are rethrown.
- Timeline rows include link back to corresponding patch detail page.

## Shared UI and Rendering Rules

### Global Layout + Navigation

- `TopNav` is rendered globally in `app/layout.tsx` for all routes.
- Global texture stack is rendered behind all routes in layout:
  - base: `/bg_texture.jpg`
  - mid layer: `/bg_texture_dark.jpg`
  - deep layer: `/bg_texture_darkest.jpg`
  - seams use `scratch_mask_*` overlays between layer transitions
- Top nav includes:
  - internal links: patches, heroes, spells, items
  - external links: changelog forum, assets API docs, Steam store page

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
- Exists in component library but is not mounted in current pages.

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

## Utility Semantics

- `formatDisplayDate`: long `en-US` date.
- `formatForumDate`: UTC `MM-DD-YYYY`.
- `formatUpdateLabel`:
  - `initial` -> `Update MM-DD-YYYY`
  - other kinds -> `Patch MM-DD-YYYY`
