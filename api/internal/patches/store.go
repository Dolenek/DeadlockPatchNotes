package patches

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"path"
	"sort"
	"strings"
	"time"
)

var ErrPatchNotFound = errors.New("patch not found")
var ErrHeroNotFound = errors.New("hero not found")
var ErrItemNotFound = errors.New("item not found")
var ErrSpellNotFound = errors.New("spell not found")

//go:embed data/*.json
var fixtureFS embed.FS

// Store is an in-memory patch storage for v1 UI development.
type Store struct {
	items map[string]listItem
	order []listItem
}

func NewStore() *Store {
	items := make(map[string]listItem)
	order := seedPatchData()

	for _, item := range order {
		items[item.summary.Slug] = item
	}

	sort.Slice(order, func(i, j int) bool {
		return order[i].published.After(order[j].published)
	})

	return &Store{items: items, order: order}
}

func (s *Store) List(page, limit int) ListResponse {
	if limit <= 0 {
		limit = 12
	}
	if page <= 0 {
		page = 1
	}

	total := len(s.order)
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

	items := make([]PatchSummary, 0, end-start)
	for _, item := range s.order[start:end] {
		items = append(items, item.summary)
	}

	return ListResponse{
		Items:      items,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

func (s *Store) GetBySlug(slug string) (PatchDetail, error) {
	item, ok := s.items[slug]
	if !ok {
		return PatchDetail{}, ErrPatchNotFound
	}
	return hydratePatchDetail(item.detail), nil
}

func (s *Store) ListHeroes() HeroListResponse {
	details := make([]PatchDetail, 0, len(s.order))
	for _, item := range s.order {
		details = append(details, item.detail)
	}
	return buildHeroList(details)
}

func (s *Store) GetHeroChanges(query HeroChangesQuery) (HeroChangesResponse, error) {
	details := make([]PatchDetail, 0, len(s.order))
	for _, item := range s.order {
		details = append(details, item.detail)
	}
	return buildHeroChanges(details, query)
}

func (s *Store) ListItems() ItemListResponse {
	details := make([]PatchDetail, 0, len(s.order))
	for _, item := range s.order {
		details = append(details, item.detail)
	}
	return buildItemList(details)
}

func (s *Store) GetItemChanges(query ItemChangesQuery) (ItemChangesResponse, error) {
	details := make([]PatchDetail, 0, len(s.order))
	for _, item := range s.order {
		details = append(details, item.detail)
	}
	return buildItemChanges(details, query)
}

func (s *Store) ListSpells() SpellListResponse {
	details := make([]PatchDetail, 0, len(s.order))
	for _, item := range s.order {
		details = append(details, item.detail)
	}
	return buildSpellList(details)
}

func (s *Store) GetSpellChanges(query SpellChangesQuery) (SpellChangesResponse, error) {
	details := make([]PatchDetail, 0, len(s.order))
	for _, item := range s.order {
		details = append(details, item.detail)
	}
	return buildSpellChanges(details, query)
}

func seedPatchData() []listItem {
	details := loadFixtureDetails()
	seeded := make([]listItem, 0, len(details))

	for _, detail := range details {
		seeded = append(seeded, buildListItem(detail))
	}

	return seeded
}

func loadFixtureDetails() []PatchDetail {
	entries, err := fixtureFS.ReadDir("data")
	if err != nil {
		panic(fmt.Errorf("read fixtures: %w", err))
	}

	details := make([]PatchDetail, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		details = append(details, mustReadFixtureDetail(entry))
	}

	return details
}

func mustReadFixtureDetail(entry fs.DirEntry) PatchDetail {
	raw, err := fixtureFS.ReadFile(path.Join("data", entry.Name()))
	if err != nil {
		panic(fmt.Errorf("read fixture %s: %w", entry.Name(), err))
	}

	var detail PatchDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		panic(fmt.Errorf("decode fixture %s: %w", entry.Name(), err))
	}

	return detail
}

func buildListItem(detail PatchDetail) listItem {
	published, err := time.Parse(time.RFC3339, detail.PublishedAt)
	if err != nil {
		panic(fmt.Errorf("parse fixture time for %s: %w", detail.Slug, err))
	}

	return listItem{
		summary: PatchSummary{
			ID:            detail.ID,
			Slug:          detail.Slug,
			Title:         detail.Title,
			PublishedAt:   detail.PublishedAt,
			Category:      detail.Category,
			Excerpt:       buildExcerpt(detail.Intro),
			CoverImageURL: detail.HeroImageURL,
			SourceURL:     detail.Source.URL,
		},
		detail:    detail,
		published: published,
	}
}

func buildExcerpt(intro string) string {
	trimmed := strings.TrimSpace(intro)
	if trimmed == "" {
		return "Deadlock patch update."
	}
	if len(trimmed) <= 160 {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:157]) + "..."
}
