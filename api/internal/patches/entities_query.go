package patches

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type itemAggregate struct {
	slug            string
	name            string
	iconURL         string
	iconFallbackURL string
	lastChanged     time.Time
}

type spellAggregate struct {
	slug            string
	name            string
	iconURL         string
	iconFallbackURL string
	lastChanged     time.Time
}

type spellCandidate struct {
	id              string
	name            string
	iconURL         string
	iconFallbackURL string
	heroSlug        string
	heroName        string
	heroIconURL     string
	heroIconFallbackURL string
	changes         []PatchChange
}

func buildItemList(details []PatchDetail) ItemListResponse {
	aggregates := map[string]*itemAggregate{}

	for _, detail := range details {
		hydrated := hydratePatchDetail(detail)
		for _, block := range hydrated.Timeline {
			releasedAt := parseRFC3339(block.ReleasedAt)
			for _, section := range block.Sections {
				if section.Kind != "items" {
					continue
				}
				for _, entry := range section.Entries {
					if len(entry.Changes) == 0 {
						continue
					}
					name := strings.TrimSpace(entry.EntityName)
					if name == "" {
						continue
					}
					slug := slugifyLookup(name)
					aggregate, ok := aggregates[slug]
					if !ok {
						aggregates[slug] = &itemAggregate{
							slug:            slug,
							name:            name,
							iconURL:         strings.TrimSpace(entry.EntityIconURL),
							iconFallbackURL: strings.TrimSpace(entry.EntityIconFallbackURL),
							lastChanged:     releasedAt,
						}
						continue
					}

					if aggregate.name == "" {
						aggregate.name = name
					}
					if aggregate.iconURL == "" {
						aggregate.iconURL = strings.TrimSpace(entry.EntityIconURL)
					}
					if aggregate.iconFallbackURL == "" {
						aggregate.iconFallbackURL = strings.TrimSpace(entry.EntityIconFallbackURL)
					}
					if releasedAt.After(aggregate.lastChanged) {
						aggregate.lastChanged = releasedAt
					}
				}
			}
		}
	}

	items := make([]ItemSummary, 0, len(aggregates))
	for _, aggregate := range aggregates {
		lastChanged := ""
		if !aggregate.lastChanged.IsZero() {
			lastChanged = aggregate.lastChanged.UTC().Format(time.RFC3339)
		}
		items = append(items, ItemSummary{
			Slug:            aggregate.slug,
			Name:            aggregate.name,
			IconURL:         aggregate.iconURL,
			IconFallbackURL: aggregate.iconFallbackURL,
			LastChangedAt:   lastChanged,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return ItemListResponse{Items: items}
}

func buildItemChanges(details []PatchDetail, query ItemChangesQuery) (ItemChangesResponse, error) {
	targetSlug := strings.TrimSpace(strings.ToLower(query.ItemSlug))
	if targetSlug == "" {
		return ItemChangesResponse{}, ErrItemNotFound
	}

	meta := ItemSummary{Slug: targetSlug}
	timeline := make([]ItemTimelineBlock, 0, 64)

	for _, detail := range details {
		hydrated := hydratePatchDetail(detail)
		for _, block := range hydrated.Timeline {
			releasedAt := parseRFC3339(block.ReleasedAt)
			if !withinDateRange(releasedAt, query.From, query.To) {
				continue
			}

			for _, section := range block.Sections {
				if section.Kind != "items" {
					continue
				}
				for _, entry := range section.Entries {
					if slugifyLookup(entry.EntityName) != targetSlug {
						continue
					}
					changes := cloneChanges(entry.Changes)
					if len(changes) == 0 {
						continue
					}

					if meta.Name == "" {
						meta.Name = strings.TrimSpace(entry.EntityName)
					}
					if meta.IconURL == "" {
						meta.IconURL = strings.TrimSpace(entry.EntityIconURL)
					}
					if meta.IconFallbackURL == "" {
						meta.IconFallbackURL = strings.TrimSpace(entry.EntityIconFallbackURL)
					}
					if releasedAt.After(parseRFC3339(meta.LastChangedAt)) {
						meta.LastChangedAt = releasedAt.UTC().Format(time.RFC3339)
					}

					timeline = append(timeline, ItemTimelineBlock{
						ID:         block.ID + "-" + targetSlug,
						Kind:       block.Kind,
						Label:      formatTimelineLabel(block.Kind, block.ReleasedAt),
						ReleasedAt: block.ReleasedAt,
						Patch:      TimelinePatchRef{Slug: detail.Slug, Title: detail.Title},
						Source:     block.Source,
						Changes:    changes,
					})
				}
			}
		}
	}

	if meta.Name == "" {
		return ItemChangesResponse{}, ErrItemNotFound
	}

	sort.SliceStable(timeline, func(i, j int) bool {
		left := parseRFC3339(timeline[i].ReleasedAt)
		right := parseRFC3339(timeline[j].ReleasedAt)
		if left.Equal(right) {
			if timeline[i].Patch.Slug == timeline[j].Patch.Slug {
				return timeline[i].ID < timeline[j].ID
			}
			return timeline[i].Patch.Slug > timeline[j].Patch.Slug
		}
		return left.After(right)
	})

	return ItemChangesResponse{
		Item:  meta,
		Items: timeline,
	}, nil
}

func buildSpellList(details []PatchDetail) SpellListResponse {
	aggregates := map[string]*spellAggregate{}
	knownItems := collectKnownItemNames(details)

	for _, detail := range details {
		hydrated := hydratePatchDetail(detail)
		for _, block := range hydrated.Timeline {
			releasedAt := parseRFC3339(block.ReleasedAt)
			for _, section := range block.Sections {
				if section.Kind != "heroes" {
					continue
				}
				for _, entry := range section.Entries {
					for _, candidate := range spellCandidatesFromHeroEntry(entry, knownItems) {
						aggregateSpell(aggregates, candidate, releasedAt)
					}
				}
			}
		}
	}

	items := make([]SpellSummary, 0, len(aggregates))
	for _, aggregate := range aggregates {
		lastChanged := ""
		if !aggregate.lastChanged.IsZero() {
			lastChanged = aggregate.lastChanged.UTC().Format(time.RFC3339)
		}
		items = append(items, SpellSummary{
			Slug:            aggregate.slug,
			Name:            aggregate.name,
			IconURL:         aggregate.iconURL,
			IconFallbackURL: aggregate.iconFallbackURL,
			LastChangedAt:   lastChanged,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return SpellListResponse{Items: items}
}

func buildSpellChanges(details []PatchDetail, query SpellChangesQuery) (SpellChangesResponse, error) {
	targetSlug := strings.TrimSpace(strings.ToLower(query.SpellSlug))
	if targetSlug == "" {
		return SpellChangesResponse{}, ErrSpellNotFound
	}

	knownItems := collectKnownItemNames(details)
	meta := SpellSummary{Slug: targetSlug}
	timeline := make([]SpellTimelineBlock, 0, 64)

	for _, detail := range details {
		hydrated := hydratePatchDetail(detail)
		for _, block := range hydrated.Timeline {
			releasedAt := parseRFC3339(block.ReleasedAt)
			if !withinDateRange(releasedAt, query.From, query.To) {
				continue
			}

			blockEntries := make([]SpellTimelineEntry, 0, 8)
			for _, section := range block.Sections {
				if section.Kind != "heroes" {
					continue
				}
				for _, entry := range section.Entries {
					for _, candidate := range spellCandidatesFromHeroEntry(entry, knownItems) {
						if candidate.id != targetSlug {
							continue
						}
						if meta.Name == "" {
							meta.Name = candidate.name
						}
						if meta.IconURL == "" {
							meta.IconURL = candidate.iconURL
						}
						if meta.IconFallbackURL == "" {
							meta.IconFallbackURL = candidate.iconFallbackURL
						}
						if releasedAt.After(parseRFC3339(meta.LastChangedAt)) {
							meta.LastChangedAt = releasedAt.UTC().Format(time.RFC3339)
						}

						blockEntries = append(blockEntries, SpellTimelineEntry{
							ID:                  fmt.Sprintf("%s-%s-%d", block.ID, targetSlug, len(blockEntries)+1),
							HeroSlug:            candidate.heroSlug,
							HeroName:            candidate.heroName,
							HeroIconURL:         candidate.heroIconURL,
							HeroIconFallbackURL: candidate.heroIconFallbackURL,
							Changes:             cloneChanges(candidate.changes),
						})
					}
				}
			}

			if len(blockEntries) == 0 {
				continue
			}
			sort.SliceStable(blockEntries, func(i, j int) bool {
				left := strings.ToLower(strings.TrimSpace(blockEntries[i].HeroName))
				right := strings.ToLower(strings.TrimSpace(blockEntries[j].HeroName))
				if left == right {
					return blockEntries[i].ID < blockEntries[j].ID
				}
				if left == "" {
					return false
				}
				if right == "" {
					return true
				}
				return left < right
			})

			timeline = append(timeline, SpellTimelineBlock{
				ID:         block.ID + "-" + targetSlug,
				Kind:       block.Kind,
				Label:      formatTimelineLabel(block.Kind, block.ReleasedAt),
				ReleasedAt: block.ReleasedAt,
				Patch:      TimelinePatchRef{Slug: detail.Slug, Title: detail.Title},
				Source:     block.Source,
				Entries:    blockEntries,
			})
		}
	}

	if meta.Name == "" {
		return SpellChangesResponse{}, ErrSpellNotFound
	}

	sort.SliceStable(timeline, func(i, j int) bool {
		left := parseRFC3339(timeline[i].ReleasedAt)
		right := parseRFC3339(timeline[j].ReleasedAt)
		if left.Equal(right) {
			if timeline[i].Patch.Slug == timeline[j].Patch.Slug {
				return timeline[i].ID < timeline[j].ID
			}
			return timeline[i].Patch.Slug > timeline[j].Patch.Slug
		}
		return left.After(right)
	})

	return SpellChangesResponse{
		Spell: meta,
		Items: timeline,
	}, nil
}

func aggregateSpell(aggregates map[string]*spellAggregate, candidate spellCandidate, releasedAt time.Time) {
	aggregate, ok := aggregates[candidate.id]
	if !ok {
		aggregates[candidate.id] = &spellAggregate{
			slug:            candidate.id,
			name:            candidate.name,
			iconURL:         candidate.iconURL,
			iconFallbackURL: candidate.iconFallbackURL,
			lastChanged:     releasedAt,
		}
		return
	}
	if aggregate.name == "" {
		aggregate.name = candidate.name
	}
	if aggregate.iconURL == "" {
		aggregate.iconURL = candidate.iconURL
	}
	if aggregate.iconFallbackURL == "" {
		aggregate.iconFallbackURL = candidate.iconFallbackURL
	}
	if releasedAt.After(aggregate.lastChanged) {
		aggregate.lastChanged = releasedAt
	}
}

func collectKnownItemNames(details []PatchDetail) map[string]bool {
	knownItems := map[string]bool{}
	for _, detail := range details {
		hydrated := hydratePatchDetail(detail)
		for _, block := range hydrated.Timeline {
			for _, section := range block.Sections {
				if section.Kind != "items" {
					continue
				}
				for _, entry := range section.Entries {
					key := normalizeLookupKey(entry.EntityName)
					if key == "" {
						continue
					}
					knownItems[key] = true
				}
			}
		}
	}
	return knownItems
}

func spellCandidatesFromHeroEntry(entry PatchEntry, knownItems map[string]bool) []spellCandidate {
	candidates := make([]spellCandidate, 0, len(entry.Groups)+1)

	if isHeroTimelineEntry(entry) {
		for _, group := range entry.Groups {
			if !isSpellGroup(group) {
				continue
			}
			candidates = append(candidates, spellCandidate{
				id:              slugifyLookup(group.Title),
				name:            strings.TrimSpace(group.Title),
				iconURL:         strings.TrimSpace(group.IconURL),
				iconFallbackURL: strings.TrimSpace(group.IconFallbackURL),
				heroSlug:        slugifyLookup(entry.EntityName),
				heroName:        strings.TrimSpace(entry.EntityName),
				heroIconURL:     strings.TrimSpace(entry.EntityIconURL),
				heroIconFallbackURL: strings.TrimSpace(entry.EntityIconFallbackURL),
				changes:         cloneChanges(group.Changes),
			})
		}
		return candidates
	}

	if len(entry.Groups) != 0 || len(entry.Changes) == 0 {
		return candidates
	}

	name := strings.TrimSpace(entry.EntityName)
	if name == "" {
		return candidates
	}
	if knownItems[normalizeLookupKey(name)] {
		return candidates
	}

	candidates = append(candidates, spellCandidate{
		id:              slugifyLookup(name),
		name:            name,
		iconURL:         strings.TrimSpace(entry.EntityIconURL),
		iconFallbackURL: strings.TrimSpace(entry.EntityIconFallbackURL),
		changes:         cloneChanges(entry.Changes),
	})
	return candidates
}

func isSpellGroup(group PatchEntryGroup) bool {
	if len(group.Changes) == 0 {
		return false
	}
	title := normalizeLookupKey(group.Title)
	if title == "" {
		return false
	}
	switch title {
	case "talents", "card types":
		return false
	default:
		return true
	}
}
