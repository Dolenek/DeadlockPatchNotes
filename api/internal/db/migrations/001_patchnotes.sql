CREATE TABLE IF NOT EXISTS patches (
  id BIGSERIAL PRIMARY KEY,
  thread_id BIGINT NOT NULL UNIQUE,
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  category TEXT NOT NULL,
  intro TEXT NOT NULL,
  excerpt TEXT NOT NULL,
  hero_image_url TEXT NOT NULL,
  published_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  source_type TEXT NOT NULL,
  source_url TEXT NOT NULL,
  detail_payload JSONB NOT NULL,
  last_synced_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_record_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS patch_release_blocks (
  id BIGSERIAL PRIMARY KEY,
  patch_id BIGINT NOT NULL REFERENCES patches(id) ON DELETE CASCADE,
  block_key TEXT NOT NULL,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  source_type TEXT NOT NULL,
  source_url TEXT,
  post_id TEXT,
  released_at TIMESTAMPTZ NOT NULL,
  body_text TEXT NOT NULL,
  body_hash TEXT NOT NULL,
  sort_order INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (patch_id, block_key),
  UNIQUE (patch_id, body_hash)
);

CREATE TABLE IF NOT EXISTS sync_runs (
  id BIGSERIAL PRIMARY KEY,
  status TEXT NOT NULL,
  run_started_at TIMESTAMPTZ NOT NULL,
  run_finished_at TIMESTAMPTZ,
  discovered_threads INT NOT NULL DEFAULT 0,
  processed_threads INT NOT NULL DEFAULT 0,
  inserted_patches INT NOT NULL DEFAULT 0,
  updated_patches INT NOT NULL DEFAULT 0,
  message TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_patches_updated_at_desc ON patches (updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_patches_published_at_desc ON patches (published_at DESC);
CREATE INDEX IF NOT EXISTS idx_patch_release_blocks_patch_sort ON patch_release_blocks (patch_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_sync_runs_started_at_desc ON sync_runs (run_started_at DESC);
