package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type SyncConfig struct {
	ForumURL string
	MaxPages int
}

type SyncStats struct {
	DiscoveredThreads int
	ProcessedThreads  int
	InsertedPatches   int
	UpdatedPatches    int
}

type timelineCandidate struct {
	Key        string
	Kind       string
	Title      string
	SourceType string
	SourceURL  string
	PostID     string
	ReleasedAt time.Time
	BodyText   string
}

func RunPatchSync(ctx context.Context, db *sql.DB, client *http.Client, cfg SyncConfig) (SyncStats, error) {
	stats := SyncStats{}

	runID, err := startSyncRun(ctx, db)
	if err != nil {
		return stats, err
	}
	finalize := func(status, message string, runErr error) (SyncStats, error) {
		_ = finishSyncRun(ctx, db, runID, status, message, stats)
		if runErr != nil {
			return stats, runErr
		}
		return stats, nil
	}

	refs, err := CrawlChangelogThreads(ctx, client, cfg.ForumURL, cfg.MaxPages)
	if err != nil {
		return finalize("failed", err.Error(), err)
	}
	stats.DiscoveredThreads = len(refs)

	for _, ref := range refs {
		thread, err := FetchThread(ctx, client, ref.URL)
		if err != nil || len(thread.Posts) == 0 {
			continue
		}

		detail, blocks, publishedAt, updatedAt := buildPatchFromThread(ctx, client, thread)
		if len(blocks) == 0 {
			continue
		}

		inserted, err := upsertPatch(ctx, db, thread, detail, blocks, publishedAt, updatedAt)
		if err != nil {
			continue
		}
		stats.ProcessedThreads++
		if inserted {
			stats.InsertedPatches++
		} else {
			stats.UpdatedPatches++
		}
	}

	return finalize("success", "sync complete", nil)
}

func upsertPatch(ctx context.Context, db *sql.DB, thread ForumThread, detail patchDetailRecord, blocks []timelineCandidate, publishedAt, updatedAt time.Time) (bool, error) {
	detailRaw, err := json.Marshal(detail.Payload)
	if err != nil {
		return false, fmt.Errorf("encode detail payload: %w", err)
	}

	excerpt := detail.Excerpt
	if len(excerpt) > 160 {
		excerpt = strings.TrimSpace(excerpt[:157]) + "..."
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var patchID int64
	inserted := false
	err = tx.QueryRowContext(ctx, `SELECT id FROM patches WHERE slug = $1`, thread.Slug).Scan(&patchID)
	if err == sql.ErrNoRows {
		inserted = true
		err = tx.QueryRowContext(ctx, `
			INSERT INTO patches (
				thread_id,
				slug,
				title,
				category,
				intro,
				excerpt,
				hero_image_url,
				published_at,
				updated_at,
				source_type,
				source_url,
				detail_payload,
				last_synced_at,
				updated_record_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,now(),now())
			RETURNING id
		`,
			thread.ThreadID,
			thread.Slug,
			detail.Payload.Title,
			detail.Payload.Category,
			detail.Payload.Intro,
			excerpt,
			detail.Payload.HeroImageURL,
			publishedAt,
			updatedAt,
			detail.Payload.Source.Type,
			detail.Payload.Source.URL,
			detailRaw,
		).Scan(&patchID)
		if err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE patches
			SET
				thread_id = $2,
				title = $3,
				category = $4,
				intro = $5,
				excerpt = $6,
				hero_image_url = $7,
				published_at = $8,
				updated_at = $9,
				source_type = $10,
				source_url = $11,
				detail_payload = $12,
				last_synced_at = now(),
				updated_record_at = now()
			WHERE id = $1
		`,
			patchID,
			thread.ThreadID,
			detail.Payload.Title,
			detail.Payload.Category,
			detail.Payload.Intro,
			excerpt,
			detail.Payload.HeroImageURL,
			publishedAt,
			updatedAt,
			detail.Payload.Source.Type,
			detail.Payload.Source.URL,
			detailRaw,
		)
		if err != nil {
			return false, err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM patch_release_blocks WHERE patch_id = $1`, patchID); err != nil {
		return false, err
	}

	for index, block := range blocks {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO patch_release_blocks (
				patch_id,
				block_key,
				kind,
				title,
				source_type,
				source_url,
				post_id,
				released_at,
				body_text,
				body_hash,
				sort_order,
				updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now())
		`,
			patchID,
			block.Key,
			block.Kind,
			block.Title,
			block.SourceType,
			block.SourceURL,
			block.PostID,
			block.ReleasedAt,
			block.BodyText,
			hashText(block.BodyText),
			index+1,
		)
		if err != nil {
			return false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return inserted, nil
}

func startSyncRun(ctx context.Context, db *sql.DB) (int64, error) {
	var runID int64
	err := db.QueryRowContext(ctx, `
		INSERT INTO sync_runs (status, run_started_at)
		VALUES ('running', now())
		RETURNING id
	`).Scan(&runID)
	if err != nil {
		return 0, err
	}
	return runID, nil
}

func finishSyncRun(ctx context.Context, db *sql.DB, runID int64, status, message string, stats SyncStats) error {
	_, err := db.ExecContext(ctx, `
		UPDATE sync_runs
		SET
			status = $2,
			run_finished_at = now(),
			discovered_threads = $3,
			processed_threads = $4,
			inserted_patches = $5,
			updated_patches = $6,
			message = $7
		WHERE id = $1
	`, runID, status, stats.DiscoveredThreads, stats.ProcessedThreads, stats.InsertedPatches, stats.UpdatedPatches, message)
	return err
}
