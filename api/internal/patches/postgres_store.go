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

const defaultReadCacheTTL = 10 * time.Minute

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

	cacheTTL time.Duration

	snapshotMu       sync.RWMutex
	refreshMu        sync.Mutex
	snapshot         *patchReadSnapshot
	snapshotExpiresAt time.Time
}

func NewPostgresStore(db *sql.DB, cacheTTL time.Duration) *PostgresStore {
	if cacheTTL <= 0 {
		cacheTTL = defaultReadCacheTTL
	}

	return &PostgresStore{
		db:       db,
		cacheTTL: cacheTTL,
	}
}

func (s *PostgresStore) List(page, limit int) (PatchListResponse, error) {
	if limit <= 0 {
		limit = 12
	}
	if page <= 0 {
		page = 1
	}

	snapshot, err := s.getSnapshot()
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
	for _, patch := range snapshot.patchSummaries[start:end] {
		patches = append(patches, patch)
	}

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

func (s *PostgresStore) GetBySlug(slug string) (PatchDetail, error) {
	snapshot, err := s.getSnapshot()
	if err != nil {
		return PatchDetail{}, fmt.Errorf("load read snapshot: %w", err)
	}

	detail, ok := snapshot.detailBySlug[slug]
	if !ok {
		return PatchDetail{}, ErrPatchNotFound
	}

	return detail, nil
}

func (s *PostgresStore) ListHeroes() (HeroListResponse, error) {
	snapshot, err := s.getSnapshot()
	if err != nil {
		return HeroListResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return snapshot.heroList, nil
}

func (s *PostgresStore) GetHeroChanges(query HeroChangesQuery) (HeroChangesResponse, error) {
	snapshot, err := s.getSnapshot()
	if err != nil {
		return HeroChangesResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return buildHeroChanges(snapshot.details, query)
}

func (s *PostgresStore) ListItems() (ItemListResponse, error) {
	snapshot, err := s.getSnapshot()
	if err != nil {
		return ItemListResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return snapshot.itemList, nil
}

func (s *PostgresStore) GetItemChanges(query ItemChangesQuery) (ItemChangesResponse, error) {
	snapshot, err := s.getSnapshot()
	if err != nil {
		return ItemChangesResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return buildItemChanges(snapshot.details, query)
}

func (s *PostgresStore) ListSpells() (SpellListResponse, error) {
	snapshot, err := s.getSnapshot()
	if err != nil {
		return SpellListResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return snapshot.spellList, nil
}

func (s *PostgresStore) GetSpellChanges(query SpellChangesQuery) (SpellChangesResponse, error) {
	snapshot, err := s.getSnapshot()
	if err != nil {
		return SpellChangesResponse{}, fmt.Errorf("load read snapshot: %w", err)
	}
	return buildSpellChanges(snapshot.details, query)
}

func (s *PostgresStore) getSnapshot() (*patchReadSnapshot, error) {
	now := time.Now()
	s.snapshotMu.RLock()
	if s.snapshot != nil && now.Before(s.snapshotExpiresAt) {
		snapshot := s.snapshot
		s.snapshotMu.RUnlock()
		return snapshot, nil
	}
	s.snapshotMu.RUnlock()

	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	now = time.Now()
	s.snapshotMu.RLock()
	if s.snapshot != nil && now.Before(s.snapshotExpiresAt) {
		snapshot := s.snapshot
		s.snapshotMu.RUnlock()
		return snapshot, nil
	}
	s.snapshotMu.RUnlock()

	snapshot, err := s.buildSnapshot()
	if err != nil {
		return nil, err
	}

	s.snapshotMu.Lock()
	s.snapshot = snapshot
	s.snapshotExpiresAt = time.Now().Add(s.cacheTTL)
	s.snapshotMu.Unlock()

	return snapshot, nil
}

func (s *PostgresStore) buildSnapshot() (*patchReadSnapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
		ORDER BY updated_at DESC
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

	return &patchReadSnapshot{
		details:        details,
		detailBySlug:   detailBySlug,
		patchSummaries: patchSummaries,
		heroList:       buildHeroList(details),
		itemList:       buildItemList(details),
		spellList:      buildSpellList(details),
	}, nil
}
