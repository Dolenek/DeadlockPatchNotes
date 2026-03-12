package patches

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// PostgresStore reads patch data from PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) List(page, limit int) ListResponse {
	if limit <= 0 {
		limit = 12
	}
	if page <= 0 {
		page = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM patches`).Scan(&total); err != nil {
		return ListResponse{Items: []PatchSummary{}, Page: 1, Limit: limit, Total: 0, TotalPages: 1}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	offset := (page - 1) * limit

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			thread_id,
			slug,
			title,
			published_at,
			category,
			hero_image_url,
			source_url,
			detail_payload
		FROM patches
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return ListResponse{Items: []PatchSummary{}, Page: page, Limit: limit, Total: total, TotalPages: totalPages}
	}
	defer rows.Close()

	items := make([]PatchSummary, 0, limit)
	for rows.Next() {
		var threadID int64
		var summary PatchSummary
		var publishedAt time.Time
		var rawDetail []byte
		if err := rows.Scan(
			&threadID,
			&summary.Slug,
			&summary.Title,
			&publishedAt,
			&summary.Category,
			&summary.CoverImageURL,
			&summary.SourceURL,
			&rawDetail,
		); err != nil {
			continue
		}
		summary.ID = fmt.Sprintf("%d", threadID)
		summary.PublishedAt = publishedAt.UTC().Format(time.RFC3339)
		if len(rawDetail) > 0 {
			var detail PatchDetail
			if err := json.Unmarshal(rawDetail, &detail); err == nil {
				summary.Timeline = buildSummaryTimeline(detail)
			}
		}
		items = append(items, summary)
	}

	return ListResponse{
		Items:      items,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

func (s *PostgresStore) GetBySlug(slug string) (PatchDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var raw []byte
	if err := s.db.QueryRowContext(ctx, `SELECT detail_payload FROM patches WHERE slug = $1`, slug).Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return PatchDetail{}, ErrPatchNotFound
		}
		return PatchDetail{}, fmt.Errorf("load patch detail: %w", err)
	}

	var detail PatchDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		return PatchDetail{}, fmt.Errorf("decode patch detail: %w", err)
	}

	return hydratePatchDetail(detail), nil
}

func (s *PostgresStore) ListHeroes() HeroListResponse {
	details, err := s.loadAllDetails()
	if err != nil {
		return HeroListResponse{Items: []HeroSummary{}}
	}
	return buildHeroList(details)
}

func (s *PostgresStore) GetHeroChanges(query HeroChangesQuery) (HeroChangesResponse, error) {
	details, err := s.loadAllDetails()
	if err != nil {
		return HeroChangesResponse{}, fmt.Errorf("load details: %w", err)
	}
	return buildHeroChanges(details, query)
}

func (s *PostgresStore) ListItems() ItemListResponse {
	details, err := s.loadAllDetails()
	if err != nil {
		return ItemListResponse{Items: []ItemSummary{}}
	}
	return buildItemList(details)
}

func (s *PostgresStore) GetItemChanges(query ItemChangesQuery) (ItemChangesResponse, error) {
	details, err := s.loadAllDetails()
	if err != nil {
		return ItemChangesResponse{}, fmt.Errorf("load details: %w", err)
	}
	return buildItemChanges(details, query)
}

func (s *PostgresStore) ListSpells() SpellListResponse {
	details, err := s.loadAllDetails()
	if err != nil {
		return SpellListResponse{Items: []SpellSummary{}}
	}
	return buildSpellList(details)
}

func (s *PostgresStore) GetSpellChanges(query SpellChangesQuery) (SpellChangesResponse, error) {
	details, err := s.loadAllDetails()
	if err != nil {
		return SpellChangesResponse{}, fmt.Errorf("load details: %w", err)
	}
	return buildSpellChanges(details, query)
}

func (s *PostgresStore) loadAllDetails() ([]PatchDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `
		SELECT detail_payload
		FROM patches
		ORDER BY published_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	details := make([]PatchDetail, 0, 128)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var detail PatchDetail
		if err := json.Unmarshal(raw, &detail); err != nil {
			continue
		}
		details = append(details, detail)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return details, nil
}
