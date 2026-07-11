package ingest

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type steamFallbackResult struct {
	DiscoveredNews int
	AddedBlocks    int
	DeferredNews   int
}

const maxSteamMinorUpdateGap = 14 * 24 * time.Hour

type storedPatchSeed struct {
	patchID     int64
	thread      ForumThread
	coverImage  string
	publishedAt time.Time
	blocks      []timelineCandidate
}

func syncLatestPatchFromSteamNews(ctx context.Context, db *sql.DB, client *http.Client, catalog *AssetCatalog, sourceURL string) (steamFallbackResult, error) {
	updates, err := FetchSteamMinorUpdates(ctx, client, sourceURL)
	if err != nil {
		return steamFallbackResult{}, err
	}
	result := steamFallbackResult{DiscoveredNews: len(updates)}
	seed, err := loadLatestPatchSeed(ctx, db)
	if err != nil {
		return result, err
	}
	blocks, added, deferred := appendSteamMinorUpdates(seed.blocks, updates, seed.publishedAt)
	result.AddedBlocks = added
	result.DeferredNews = deferred
	if deferred > 0 {
		return result, fmt.Errorf("%d Steam minor updates are outside the %s follow-up window; forum discovery is required", deferred, maxSteamMinorUpdateGap)
	}
	if added == 0 {
		return result, nil
	}

	payload := buildDetailPayload(seed.thread, blocks, seed.coverImage, catalog)
	detail := patchDetailRecord{Payload: payload, Excerpt: buildIntro(payload.Sections[0].Entries)}
	_, err = upsertPatch(ctx, db, seed.thread, detail, blocks, blocks[0].ReleasedAt, blocks[len(blocks)-1].ReleasedAt)
	if err != nil {
		return result, fmt.Errorf("upsert Steam fallback patch: %w", err)
	}
	return result, nil
}

func loadLatestPatchSeed(ctx context.Context, db *sql.DB) (storedPatchSeed, error) {
	var seed storedPatchSeed
	err := db.QueryRowContext(ctx, `
		SELECT id, thread_id, slug, title, hero_image_url, published_at
		FROM patches
		ORDER BY published_at DESC
		LIMIT 1
	`).Scan(
		&seed.patchID,
		&seed.thread.ThreadID,
		&seed.thread.Slug,
		&seed.thread.Title,
		&seed.coverImage,
		&seed.publishedAt,
	)
	if err != nil {
		return seed, fmt.Errorf("load latest patch for Steam fallback: %w", err)
	}
	seed.thread.URL = fmt.Sprintf("https://forums.playdeadlock.com/threads/%s.%d/", seed.thread.Slug, seed.thread.ThreadID)
	seed.blocks, err = loadStoredTimelineBlocks(ctx, db, seed.patchID)
	if err != nil {
		return seed, err
	}
	if len(seed.blocks) == 0 {
		return seed, fmt.Errorf("latest patch %s has no stored release blocks", seed.thread.Slug)
	}
	return seed, nil
}

func loadStoredTimelineBlocks(ctx context.Context, db *sql.DB, patchID int64) ([]timelineCandidate, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT block_key, kind, title, source_type, source_url, post_id, released_at, body_text
		FROM patch_release_blocks
		WHERE patch_id = $1
		ORDER BY sort_order
	`, patchID)
	if err != nil {
		return nil, fmt.Errorf("load stored release blocks: %w", err)
	}
	defer rows.Close()

	blocks := make([]timelineCandidate, 0, 8)
	for rows.Next() {
		block, err := scanStoredTimelineBlock(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stored release blocks: %w", err)
	}
	return blocks, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanStoredTimelineBlock(row rowScanner) (timelineCandidate, error) {
	var block timelineCandidate
	var sourceURL sql.NullString
	var postID sql.NullString
	err := row.Scan(&block.Key, &block.Kind, &block.Title, &block.SourceType, &sourceURL, &postID, &block.ReleasedAt, &block.BodyText)
	if err != nil {
		return block, fmt.Errorf("scan stored release block: %w", err)
	}
	block.SourceURL = sourceURL.String
	block.PostID = postID.String
	return block, nil
}

func appendSteamMinorUpdates(existing []timelineCandidate, updates []SteamMinorUpdate, patchPublishedAt time.Time) ([]timelineCandidate, int, int) {
	blocks := append([]timelineCandidate(nil), existing...)
	seenBodies := make(map[string]bool, len(existing)+len(updates))
	latestRelease := patchPublishedAt
	for _, block := range existing {
		seenBodies[hashText(normalizeBodyForHash(block.BodyText))] = true
		if block.ReleasedAt.After(latestRelease) {
			latestRelease = block.ReleasedAt
		}
	}
	added := 0
	deferred := 0
	for _, update := range updates {
		body := normalizeBodyForHash(update.BodyText)
		bodyHash := hashText(body)
		if !update.PublishedAt.After(patchPublishedAt) || body == "" || seenBodies[bodyHash] {
			continue
		}
		if update.PublishedAt.After(latestRelease.Add(maxSteamMinorUpdateGap)) {
			deferred++
			continue
		}
		seenBodies[bodyHash] = true
		blocks = append(blocks, steamMinorUpdateBlock(update, body))
		added++
		if update.PublishedAt.After(latestRelease) {
			latestRelease = update.PublishedAt
		}
	}
	sort.SliceStable(blocks, func(i, j int) bool {
		return blocks[i].ReleasedAt.Before(blocks[j].ReleasedAt)
	})
	normalizeStoredBlockKinds(blocks)
	return blocks, added, deferred
}

func steamMinorUpdateBlock(update SteamMinorUpdate, body string) timelineCandidate {
	return timelineCandidate{
		Key:        "steam-announcement-" + update.GID,
		Kind:       "hotfix",
		Title:      "Hotfix " + update.PublishedAt.UTC().Format("2006-01-02"),
		SourceType: "steam-news",
		SourceURL:  update.SourceURL,
		PostID:     "steam-" + update.GID,
		ReleasedAt: update.PublishedAt.UTC(),
		BodyText:   body,
	}
}

func normalizeStoredBlockKinds(blocks []timelineCandidate) {
	for index := range blocks {
		if index == 0 {
			blocks[index].Kind = "initial"
			continue
		}
		blocks[index].Kind = "hotfix"
		if strings.EqualFold(strings.TrimSpace(blocks[index].Title), "Initial Update") {
			blocks[index].Title = "Hotfix " + blocks[index].ReleasedAt.UTC().Format("2006-01-02")
		}
	}
}
