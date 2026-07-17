package patches

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

const (
	defaultReadCacheTTL     = 10 * time.Minute
	failedRefreshRetryDelay = 5 * time.Second
)

type patchReadSnapshot struct {
	details        []PatchDetail
	detailBySlug   map[string]PatchDetail
	patchSummaries []PatchSummary
	heroList       HeroListResponse
	itemList       ItemListResponse
	spellList      SpellListResponse
}

// PostgresStore reads patch data from PostgreSQL.
type PostgresStore struct {
	db *sql.DB

	cacheTTL        time.Duration
	buildSnapshotFn func(context.Context) (*patchReadSnapshot, error)

	snapshotMu        sync.RWMutex
	refreshPermit     chan struct{}
	snapshot          *patchReadSnapshot
	snapshotExpiresAt time.Time
}

func NewPostgresStore(db *sql.DB, cacheTTL time.Duration) *PostgresStore {
	if cacheTTL <= 0 {
		cacheTTL = defaultReadCacheTTL
	}

	refreshPermit := make(chan struct{}, 1)
	refreshPermit <- struct{}{}
	return &PostgresStore{
		db:            db,
		cacheTTL:      cacheTTL,
		refreshPermit: refreshPermit,
	}
}

func (s *PostgresStore) List(ctx context.Context, page, limit int) (PatchListResponse, error) {
	if limit <= 0 {
		limit = 12
	}
	if page <= 0 {
		page = 1
	}

	snapshot, err := s.getSnapshot(ctx)
	if err != nil {
		return PatchListResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}

	total := len(snapshot.patchSummaries)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	patches := make([]PatchSummary, 0, end-start)
	patches = append(patches, snapshot.patchSummaries[start:end]...)

	return PatchListResponse{
		Patches: patches,
		Pagination: Pagination{
			Page:       page,
			PageSize:   limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *PostgresStore) GetBySlug(ctx context.Context, slug string) (PatchDetail, error) {
	snapshot, err := s.getSnapshot(ctx)
	if err != nil {
		return PatchDetail{}, fmt.Errorf("load read snapshot: %w", err)
	}

	detail, ok := snapshot.detailBySlug[slug]
	if !ok {
		return PatchDetail{}, ErrPatchNotFound
	}

	return detail, nil
}

func (s *PostgresStore) ListHeroes(ctx context.Context) (HeroListResponse, error) {
	snapshot, err := s.getSnapshot(ctx)
	if err != nil {
		return HeroListResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return snapshot.heroList, nil
}

func (s *PostgresStore) GetHeroChanges(ctx context.Context, query HeroChangesQuery) (HeroChangesResponse, error) {
	snapshot, err := s.getSnapshot(ctx)
	if err != nil {
		return HeroChangesResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return buildHeroChanges(snapshot.details, query)
}

func (s *PostgresStore) ListItems(ctx context.Context) (ItemListResponse, error) {
	snapshot, err := s.getSnapshot(ctx)
	if err != nil {
		return ItemListResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return snapshot.itemList, nil
}

func (s *PostgresStore) GetItemChanges(ctx context.Context, query ItemChangesQuery) (ItemChangesResponse, error) {
	snapshot, err := s.getSnapshot(ctx)
	if err != nil {
		return ItemChangesResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return buildItemChanges(snapshot.details, query)
}

func (s *PostgresStore) ListSpells(ctx context.Context) (SpellListResponse, error) {
	snapshot, err := s.getSnapshot(ctx)
	if err != nil {
		return SpellListResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return snapshot.spellList, nil
}

func (s *PostgresStore) GetSpellChanges(ctx context.Context, query SpellChangesQuery) (SpellChangesResponse, error) {
	snapshot, err := s.getSnapshot(ctx)
	if err != nil {
		return SpellChangesResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return buildSpellChanges(snapshot.details, query)
}

func (s *PostgresStore) getSnapshot(ctx context.Context) (*patchReadSnapshot, error) {
	now := time.Now()
	s.snapshotMu.RLock()
	if s.snapshot != nil && now.Before(s.snapshotExpiresAt) {
		snapshot := s.snapshot
		s.snapshotMu.RUnlock()
		return snapshot, nil
	}
	s.snapshotMu.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.refreshPermit:
	}
	defer func() { s.refreshPermit <- struct{}{} }()

	now = time.Now()
	s.snapshotMu.RLock()
	if s.snapshot != nil && now.Before(s.snapshotExpiresAt) {
		snapshot := s.snapshot
		s.snapshotMu.RUnlock()
		return snapshot, nil
	}
	s.snapshotMu.RUnlock()

	s.snapshotMu.RLock()
	staleSnapshot := s.snapshot
	s.snapshotMu.RUnlock()

	snapshot, err := s.refreshSnapshot(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if staleSnapshot != nil {
			s.snapshotMu.Lock()
			s.snapshotExpiresAt = time.Now().Add(s.refreshRetryDelay())
			s.snapshotMu.Unlock()
			return staleSnapshot, nil
		}
		return nil, err
	}

	s.snapshotMu.Lock()
	s.snapshot = snapshot
	s.snapshotExpiresAt = time.Now().Add(s.cacheTTL)
	s.snapshotMu.Unlock()

	return snapshot, nil
}

func (s *PostgresStore) refreshRetryDelay() time.Duration {
	if s.cacheTTL < failedRefreshRetryDelay {
		return s.cacheTTL
	}
	return failedRefreshRetryDelay
}

func (s *PostgresStore) refreshSnapshot(ctx context.Context) (*patchReadSnapshot, error) {
	if s.buildSnapshotFn != nil {
		return s.buildSnapshotFn(ctx)
	}
	return s.buildSnapshot(ctx)
}

func (s *PostgresStore) buildSnapshot(parent context.Context) (*patchReadSnapshot, error) {
	ctx, cancel := context.WithTimeout(parent, 15*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			thread_id,
			slug,
			title,
			published_at,
			category,
			hero_image_url,
			source_type,
			source_url,
			detail_payload
		FROM patches
		ORDER BY published_at DESC, slug DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	details := make([]PatchDetail, 0, 128)
	detailBySlug := make(map[string]PatchDetail, 128)
	patchSummaries := make([]PatchSummary, 0, 128)

	for rows.Next() {
		var threadID int64
		var slug string
		var title string
		var publishedAt time.Time
		var category string
		var heroImageURL string
		var sourceType string
		var sourceURL string
		var rawDetail []byte

		if err := rows.Scan(
			&threadID,
			&slug,
			&title,
			&publishedAt,
			&category,
			&heroImageURL,
			&sourceType,
			&sourceURL,
			&rawDetail,
		); err != nil {
			return nil, fmt.Errorf("scan patch row: %w", err)
		}

		var detail PatchDetail
		if err := json.Unmarshal(rawDetail, &detail); err != nil {
			return nil, fmt.Errorf("decode patch detail %s: %w", slug, err)
		}
		detail = hydratePatchDetail(detail)

		if detail.ID == "" {
			detail.ID = fmt.Sprintf("%d", threadID)
		}
		if detail.Slug == "" {
			detail.Slug = slug
		}
		if detail.Title == "" {
			detail.Title = title
		}
		if detail.PublishedAt == "" {
			detail.PublishedAt = publishedAt.UTC().Format(time.RFC3339)
		}
		if detail.Category == "" {
			detail.Category = category
		}
		if detail.Source.Type == "" {
			detail.Source.Type = sourceType
		}
		if detail.Source.URL == "" {
			detail.Source.URL = sourceURL
		}
		if detail.HeroImageURL == "" {
			detail.HeroImageURL = heroImageURL
		}

		details = append(details, detail)
		detailBySlug[slug] = detail
		patchSummaries = append(patchSummaries, PatchSummary{
			ID:            fmt.Sprintf("%d", threadID),
			Slug:          slug,
			Title:         title,
			PublishedAt:   publishedAt.UTC().Format(time.RFC3339),
			Category:      category,
			CoverImageURL: heroImageURL,
			Source: PatchSource{
				Type: sourceType,
				URL:  sourceURL,
			},
			Timeline: buildSummaryTimeline(detail),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.SliceStable(details, func(i, j int) bool {
		left := parseRFC3339(details[i].PublishedAt)
		right := parseRFC3339(details[j].PublishedAt)
		if left.Equal(right) {
			return details[i].Slug < details[j].Slug
		}
		return left.Before(right)
	})
	aggregateDetails := deduplicateTimelineEvents(details)

	return &patchReadSnapshot{
		details:        aggregateDetails,
		detailBySlug:   detailBySlug,
		patchSummaries: patchSummaries,
		heroList:       buildHeroList(aggregateDetails),
		itemList:       buildItemList(aggregateDetails),
		spellList:      buildSpellList(aggregateDetails),
	}, nil
}
