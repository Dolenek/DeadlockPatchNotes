package patches

import (
	"errors"
	"math"
	"sort"
	"time"
)

var ErrPatchNotFound = errors.New("patch not found")

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
	return item.detail, nil
}

func seedPatchData() []listItem {
	published := time.Date(2026, 3, 6, 22, 36, 0, 0, time.UTC)

	detail := PatchDetail{
		ID:          "519740319207522795",
		Slug:        "2026-03-06-update",
		Title:       "Gameplay Update - 03-06-2026",
		PublishedAt: published.Format(time.RFC3339),
		Category:    "Regular Update",
		Source: PatchSource{
			Type: "steam-news",
			URL:  "https://store.steampowered.com/news/app/1422450/view/519740319207522795",
		},
		HeroImageURL: "https://clan.akamai.steamstatic.com/images/45164767/1a200778c94a048c5b2580a1e1a36071679ff19e.png",
		Intro: "Large gameplay update touching core map systems, economy pacing, items, and hero balance across the roster.",
		Sections: []PatchSection{
			{
				ID:    "general",
				Title: "General",
				Kind:  "general",
				Entries: []PatchEntry{
					{
						ID:         "movement-and-flow",
						EntityName: "Movement and Flow",
						Summary:    "Core movement and lane pacing were adjusted to speed up mid-game pressure windows.",
						Changes: []PatchChange{
							{ID: "g-1", Text: "Can now jump during slide."},
							{ID: "g-2", Text: "Dash jump grants a brief period of increased air control (30% for 0.25s)."},
							{ID: "g-3", Text: "Trooper wave interval increased from every 25s to every 20s starting at 35 minutes."},
							{ID: "g-4", Text: "Subsequent CC reduction increased from 8%/24% to 10%/30% (window from 7s to 8s)."},
						},
					},
					{
						ID:         "objectives-and-bounties",
						EntityName: "Objectives and Bounties",
						Summary:    "Objective durability and reward values were rebalanced to push clearer comeback and snowball states.",
						Changes: []PatchChange{
							{ID: "g-5", Text: "Shrine HP changed from 8100 to 5000/10000, with stage swap after the first shrine dies."},
							{ID: "g-6", Text: "Guardians bounty increased from 1000 to 1500; Walkers bounty increased from 3500 to 4000."},
							{ID: "g-7", Text: "Mid Boss base bounty increased from 2000 to 3000 and base HP increased from 11900 to 13000."},
							{ID: "g-8", Text: "Rejuv duration no longer refreshes when hitting a crystal later."},
						},
					},
				},
			},
			{
				ID:    "items",
				Title: "Items",
				Kind:  "items",
				Entries: []PatchEntry{
					{
						ID:         "new-item",
						EntityName: "Golden Goose Egg",
						EntityIconURL: "https://assets-bucket.deadlock-api.com/assets-api-res/icons/spirit.svg",
						Summary:    "New Tier 1 Spirit item added.",
						Changes: []PatchChange{
							{ID: "i-1", Text: "Added new T1 Spirit item: Golden Goose Egg."},
						},
					},
					{
						ID:         "weapon-items",
						EntityName: "Weapon Items",
						Summary:    "Several early and mid-tier weapon options were shifted toward stronger damage identity.",
						Changes: []PatchChange{
							{ID: "i-2", Text: "Extended Magazine weapon damage increased from +6% to +8%."},
							{ID: "i-3", Text: "High-Velocity Rounds no longer grants fire rate and now grants +8% weapon damage."},
							{ID: "i-4", Text: "Weighted Shots now builds from Slowing Bullets and weapon damage increased from 35% to 40%."},
							{ID: "i-5", Text: "Inhibitor weapon damage bonus increased from +22% to +25%."},
						},
					},
					{
						ID:         "survivability-items",
						EntityName: "Survivability and Utility",
						Summary:    "Defensive thresholds and utility cooldowns were tuned for clearer burst-vs-sustain choices.",
						Changes: []PatchChange{
							{ID: "i-6", Text: "Bullet Resilience and Spirit Resilience low-health threshold increased from 40% to 50%."},
							{ID: "i-7", Text: "Dispel Magic cooldown reduced from 50s to 40s."},
							{ID: "i-8", Text: "Diviner's Kevlar cooldown reduced from 64s to 40s."},
							{ID: "i-9", Text: "Magic Carpet duration increased from 8s to 12s."},
						},
					},
				},
			},
			{
				ID:    "heroes",
				Title: "Heroes",
				Kind:  "heroes",
				Entries: []PatchEntry{
					{
						ID:            "abrams",
						EntityName:    "Abrams",
						EntityIconURL: "https://assets-bucket.deadlock-api.com/assets-api-res/icons/infernus.svg",
						Summary:       "Damage and control profile adjusted to reduce oppressive reliability while preserving engage role.",
						Changes: []PatchChange{
							{ID: "h-1", Text: "Bullet damage reduced from 3.86+0.13/boon to 3.6+0.1/boon."},
							{ID: "h-2", Text: "Siphon Life range reduced from 10m to 7.5m."},
							{ID: "h-3", Text: "Seismic Impact damage increased from 75 to 100; radius reduced from 10.5m to 9m."},
						},
					},
					{
						ID:            "apollo",
						EntityName:    "Apollo",
						Summary:       "Targeted reliability nerfs to defensive triggers and ultimate hit profile.",
						Changes: []PatchChange{
							{ID: "h-4", Text: "Riposte no longer triggers from trooper or neutral damage."},
							{ID: "h-5", Text: "Flawless Advance now gets interrupted by stun and sleep."},
							{ID: "h-6", Text: "Itani Lo Sahn base damage reduced from 225 to 190, spirit scaling increased from 1.6 to 2.3."},
						},
					},
					{
						ID:            "victor",
						EntityName:    "Victor",
						Summary:       "Damage curves and cooldown structure were redistributed across kit uptime windows.",
						Changes: []PatchChange{
							{ID: "h-7", Text: "Pain Battery range increased from 20m to 28m and bolt count from 5 to 7."},
							{ID: "h-8", Text: "Jumpstart cooldown increased from 23s to 30s with adjusted scaling talents."},
							{ID: "h-9", Text: "Aura of Suffering max DPS reduced from 70 to 54 and radius increased from 7.7m to 10m."},
						},
					},
					{
						ID:            "warden",
						EntityName:    "Warden",
						Summary:       "Crowd-control and defensive toolchain re-tuned for longer combat windows and clearer upgrade choices.",
						Changes: []PatchChange{
							{ID: "h-10", Text: "Bullet growth per boon reduced from 0.44 to 0.38."},
							{ID: "h-11", Text: "Binding Word cast range reduced from 20m to 15m and cooldown decreased from 37s to 34s."},
							{ID: "h-12", Text: "Last Stand channeling resist increased from 30% to 60%; radius reduced from 13m to 12m."},
						},
					},
				},
			},
		},
	}

	summary := PatchSummary{
		ID:            detail.ID,
		Slug:          detail.Slug,
		Title:         detail.Title,
		PublishedAt:   detail.PublishedAt,
		Category:      detail.Category,
		Excerpt:       "Movement, objectives, item balance, and broad hero tuning across the roster.",
		CoverImageURL: detail.HeroImageURL,
		SourceURL:     detail.Source.URL,
	}

	return []listItem{{
		summary:   summary,
		detail:    detail,
		published: published,
	}}
}
