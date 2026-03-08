-- PostgreSQL draft schema for patch notes ingestion and read-optimized rendering.
-- Status: draft for upcoming DB-backed backend implementation.

CREATE TABLE IF NOT EXISTS patches (
  id BIGSERIAL PRIMARY KEY,
  source_gid TEXT UNIQUE,
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  category TEXT NOT NULL,
  intro TEXT NOT NULL,
  hero_image_url TEXT,
  published_at TIMESTAMPTZ NOT NULL,
  source_type TEXT NOT NULL,
  source_url TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS patch_sections (
  id BIGSERIAL PRIMARY KEY,
  patch_id BIGINT NOT NULL REFERENCES patches(id) ON DELETE CASCADE,
  section_key TEXT NOT NULL,
  title TEXT NOT NULL,
  kind TEXT NOT NULL,
  sort_order INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (patch_id, section_key)
);

CREATE TABLE IF NOT EXISTS patch_entries (
  id BIGSERIAL PRIMARY KEY,
  section_id BIGINT NOT NULL REFERENCES patch_sections(id) ON DELETE CASCADE,
  entry_key TEXT NOT NULL,
  entity_name TEXT NOT NULL,
  entity_icon_url TEXT,
  entity_icon_fallback_url TEXT,
  summary TEXT,
  sort_order INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (section_id, entry_key)
);

CREATE TABLE IF NOT EXISTS patch_entry_groups (
  id BIGSERIAL PRIMARY KEY,
  entry_id BIGINT NOT NULL REFERENCES patch_entries(id) ON DELETE CASCADE,
  group_key TEXT NOT NULL,
  title TEXT NOT NULL,
  icon_url TEXT,
  icon_fallback_url TEXT,
  sort_order INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (entry_id, group_key)
);

CREATE TABLE IF NOT EXISTS patch_changes (
  id BIGSERIAL PRIMARY KEY,
  entry_id BIGINT REFERENCES patch_entries(id) ON DELETE CASCADE,
  group_id BIGINT REFERENCES patch_entry_groups(id) ON DELETE CASCADE,
  change_key TEXT NOT NULL,
  body TEXT NOT NULL,
  sort_order INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK ((entry_id IS NOT NULL) <> (group_id IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS patch_assets (
  id BIGSERIAL PRIMARY KEY,
  patch_id BIGINT NOT NULL REFERENCES patches(id) ON DELETE CASCADE,
  external_url TEXT NOT NULL,
  local_path TEXT,
  media_type TEXT,
  content_hash TEXT,
  byte_size BIGINT,
  synced_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (patch_id, external_url)
);

-- Optional read cache for rendered patch JSON payload.
CREATE TABLE IF NOT EXISTS patch_render_cache (
  patch_slug TEXT PRIMARY KEY,
  payload JSONB NOT NULL,
  etag TEXT,
  generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_patches_published_at ON patches (published_at DESC);
CREATE INDEX IF NOT EXISTS idx_patch_sections_patch_sort ON patch_sections (patch_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_patch_entries_section_sort ON patch_entries (section_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_patch_entry_groups_entry_sort ON patch_entry_groups (entry_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_patch_changes_entry_sort ON patch_changes (entry_id, sort_order) WHERE entry_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_patch_changes_group_sort ON patch_changes (group_id, sort_order) WHERE group_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_patch_assets_patch_synced ON patch_assets (patch_id, synced_at DESC);
CREATE INDEX IF NOT EXISTS idx_patch_render_cache_expires ON patch_render_cache (expires_at);
