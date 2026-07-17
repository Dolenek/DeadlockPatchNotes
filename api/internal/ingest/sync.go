package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type SyncConfig struct {
	ForumURL     string
	SteamNewsURL string
	MaxPages     int
}

type syncDependencies struct {
	crawlChangelogThreads        func(context.Context, *http.Client, string, int) ([]ForumThreadRef, error)
	loadAssetCatalog             func(context.Context, *http.Client) (*AssetCatalog, error)
	syncDiscoveredThreads        func(context.Context, *sql.DB, *http.Client, *AssetCatalog, []ForumThreadRef, SyncStats) (SyncStats, []string)
	syncLatestPatchFromSteamNews func(context.Context, *sql.DB, *http.Client, *AssetCatalog, string) (steamFallbackResult, error)
}

type SyncStats struct {
	DiscoveredThreads int
	ProcessedThreads  int
	FailedThreads     int
	InsertedPatches   int
	UpdatedPatches    int
}

type patchWriteResult struct {
	Inserted bool
	Updated  bool
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
	return runPatchSync(ctx, db, client, cfg, defaultSyncDependencies())
}

func defaultSyncDependencies() syncDependencies {
	return syncDependencies{
		crawlChangelogThreads:        CrawlChangelogThreads,
		loadAssetCatalog:             LoadAssetCatalog,
		syncDiscoveredThreads:        syncDiscoveredThreads,
		syncLatestPatchFromSteamNews: syncLatestPatchFromSteamNews,
	}
}

func runPatchSync(ctx context.Context, db *sql.DB, client *http.Client, cfg SyncConfig, dependencies syncDependencies) (SyncStats, error) {
	stats := SyncStats{}

	runID, err := startSyncRun(ctx, db)
	if err != nil {
		return stats, err
	}

	refs, err := dependencies.crawlChangelogThreads(ctx, client, cfg.ForumURL, cfg.MaxPages)
	if err != nil || len(refs) == 0 {
		if err == nil {
			err = errors.New("no patch threads discovered")
		}
		return runSteamNewsFallback(ctx, db, client, cfg.SteamNewsURL, runID, stats, err, dependencies)
	}
	stats.DiscoveredThreads = len(refs)

	catalog, err := dependencies.loadAssetCatalog(ctx, client)
	if err != nil {
		err = fmt.Errorf("load asset catalog: %w", err)
		return finalizeSyncRun(ctx, db, runID, "failed", err.Error(), stats, err)
	}

	stats, failures := dependencies.syncDiscoveredThreads(ctx, db, client, catalog, refs, stats)
	if stats.FailedThreads == 0 {
		return finalizeSyncRun(ctx, db, runID, "success", "sync complete", stats, nil)
	}
	status := "partial"
	if stats.ProcessedThreads == 0 {
		status = "failed"
	}
	err = fmt.Errorf("%d of %d patch threads failed: %s", stats.FailedThreads, stats.DiscoveredThreads, strings.Join(failures, "; "))
	return finalizeSyncRun(ctx, db, runID, status, err.Error(), stats, err)
}

func runSteamNewsFallback(ctx context.Context, db *sql.DB, client *http.Client, sourceURL string, runID int64, stats SyncStats, forumErr error, dependencies syncDependencies) (SyncStats, error) {
	catalog, err := dependencies.loadAssetCatalog(ctx, client)
	if err != nil {
		err = fmt.Errorf("forum discovery unavailable (%v); load asset catalog for Steam fallback: %w", forumErr, err)
		return finalizeSyncRun(ctx, db, runID, "failed", err.Error(), stats, err)
	}
	result, err := dependencies.syncLatestPatchFromSteamNews(ctx, db, client, catalog, sourceURL)
	stats.DiscoveredThreads = result.DiscoveredNews
	if err != nil {
		err = fmt.Errorf("forum discovery unavailable (%v); Steam fallback: %w", forumErr, err)
		return finalizeSyncRun(ctx, db, runID, "failed", err.Error(), stats, err)
	}
	stats.ProcessedThreads = 1
	if result.AddedBlocks > 0 {
		stats.UpdatedPatches = 1
	}
	message := fmt.Sprintf("Steam fallback complete: discovered=%d added_blocks=%d", result.DiscoveredNews, result.AddedBlocks)
	return finalizeSyncRun(ctx, db, runID, "success", message, stats, nil)
}

func syncDiscoveredThreads(ctx context.Context, db *sql.DB, client *http.Client, catalog *AssetCatalog, refs []ForumThreadRef, stats SyncStats) (SyncStats, []string) {
	failures := make([]string, 0, 4)
	for _, ref := range refs {
		writeResult, err := syncPatchThread(ctx, db, client, catalog, ref)
		if err != nil {
			stats.FailedThreads++
			failures = appendSyncFailure(failures, fmt.Sprintf("%s: %v", ref.URL, err))
			continue
		}
		stats.ProcessedThreads++
		if writeResult.Inserted {
			stats.InsertedPatches++
		} else if writeResult.Updated {
			stats.UpdatedPatches++
		}
	}
	return stats, failures
}

func finalizeSyncRun(ctx context.Context, db *sql.DB, runID int64, status, message string, stats SyncStats, runErr error) (SyncStats, error) {
	if finishErr := finishSyncRun(ctx, db, runID, status, message, stats); finishErr != nil {
		runErr = errors.Join(runErr, fmt.Errorf("finish sync run: %w", finishErr))
	}
	return stats, runErr
}

func syncPatchThread(ctx context.Context, db *sql.DB, client *http.Client, catalog *AssetCatalog, ref ForumThreadRef) (patchWriteResult, error) {
	thread, err := FetchThread(ctx, client, ref.URL)
	if err != nil {
		return patchWriteResult{}, err
	}
	if len(thread.Posts) == 0 {
		return patchWriteResult{}, errors.New("no official posts parsed")
	}

	detail, blocks, publishedAt, updatedAt, err := buildPatchFromThread(ctx, client, thread, catalog)
	if err != nil {
		return patchWriteResult{}, err
	}
	if len(blocks) == 0 {
		return patchWriteResult{}, errors.New("no release blocks parsed")
	}
	writeResult, err := upsertPatch(ctx, db, thread, detail, blocks, publishedAt, updatedAt)
	if err != nil {
		return patchWriteResult{}, fmt.Errorf("upsert: %w", err)
	}
	return writeResult, nil
}

func appendSyncFailure(failures []string, message string) []string {
	const maxRecordedFailures = 5
	if len(failures) >= maxRecordedFailures {
		return failures
	}
	return append(failures, message)
}

func upsertPatch(ctx context.Context, db *sql.DB, thread ForumThread, detail patchDetailRecord, blocks []timelineCandidate, publishedAt, updatedAt time.Time) (patchWriteResult, error) {
	detailRaw, err := json.Marshal(detail.Payload)
	if err != nil {
		return patchWriteResult{}, fmt.Errorf("encode detail payload: %w", err)
	}

	excerpt := detail.Excerpt
	if len(excerpt) > 160 {
		excerpt = strings.TrimSpace(excerpt[:157]) + "..."
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return patchWriteResult{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, thread.ThreadID); err != nil {
		return patchWriteResult{}, fmt.Errorf("lock patch thread: %w", err)
	}

	var patchID int64
	var stored storedPatchState
	err = tx.QueryRowContext(ctx, `
		SELECT
			id, slug, title, category, intro, excerpt, hero_image_url,
			published_at, updated_at, source_type, source_url, detail_payload
		FROM patches
		WHERE thread_id = $1
	`, thread.ThreadID).Scan(
		&patchID,
		&stored.Slug,
		&stored.Title,
		&stored.Category,
		&stored.Intro,
		&stored.Excerpt,
		&stored.HeroImageURL,
		&stored.PublishedAt,
		&stored.UpdatedAt,
		&stored.SourceType,
		&stored.SourceURL,
		&stored.DetailPayload,
	)
	writeResult := patchWriteResult{}
	if err == sql.ErrNoRows {
		writeResult.Inserted = true
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
			return patchWriteResult{}, err
		}
	} else if err != nil {
		return patchWriteResult{}, err
	} else {
		desired := newStoredPatchState(thread, detail, excerpt, publishedAt, updatedAt, detailRaw)
		blocksMatch, err := storedTimelineMatches(ctx, tx, patchID, blocks)
		if err != nil {
			return patchWriteResult{}, err
		}
		writeResult.Updated = !stored.matches(desired) || !blocksMatch
		if !writeResult.Updated {
			if _, err := tx.ExecContext(ctx, `UPDATE patches SET last_synced_at = now() WHERE id = $1`, patchID); err != nil {
				return patchWriteResult{}, err
			}
			if err := tx.Commit(); err != nil {
				return patchWriteResult{}, err
			}
			return writeResult, nil
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE patches
			SET
				slug = $2,
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
		)
		if err != nil {
			return patchWriteResult{}, err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM patch_release_blocks WHERE patch_id = $1`, patchID); err != nil {
		return patchWriteResult{}, err
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
			return patchWriteResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return patchWriteResult{}, err
	}
	return writeResult, nil
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
			failed_threads = $5,
			inserted_patches = $6,
			updated_patches = $7,
			message = $8
		WHERE id = $1
	`, runID, status, stats.DiscoveredThreads, stats.ProcessedThreads, stats.FailedThreads, stats.InsertedPatches, stats.UpdatedPatches, message)
	return err
}
