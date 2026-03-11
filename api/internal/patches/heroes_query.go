package patches

import (
	"sort"
	"strings"
	"time"
)

type heroAggregate struct {
	slug            string
	name            string
	iconURL         string
	iconFallbackURL string
	lastChanged     time.Time
}

func buildHeroList(details []PatchDetail) HeroListResponse {
	aggregates := map[string]*heroAggregate{}

	for _, detail := range details {
		hydrated := hydratePatchDetail(detail)
		for _, block := range hydrated.Timeline {
			releasedAt := parseRFC3339(block.ReleasedAt)
			for _, section := range block.Sections {
				if section.Kind != "heroes" {
					continue
				}
				for _, entry := range section.Entries {
					if !isHeroTimelineEntry(entry) {
						continue
					}
					if len(entry.Changes) == 0 && len(entry.Groups) == 0 {
						continue
					}
					name := strings.TrimSpace(entry.EntityName)
					if name == "" {
						continue
					}
					slug := slugifyLookup(name)
					aggregate, ok := aggregates[slug]
					if !ok {
						aggregate = &heroAggregate{
							slug:            slug,
							name:            name,
							iconURL:         strings.TrimSpace(entry.EntityIconURL),
							iconFallbackURL: strings.TrimSpace(entry.EntityIconFallbackURL),
							lastChanged:     releasedAt,
						}
						aggregates[slug] = aggregate
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

	items := make([]HeroSummary, 0, len(aggregates))
	for _, aggregate := range aggregates {
		lastChanged := ""
		if !aggregate.lastChanged.IsZero() {
			lastChanged = aggregate.lastChanged.UTC().Format(time.RFC3339)
		}
		items = append(items, HeroSummary{
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

	return HeroListResponse{Items: items}
}

func buildHeroChanges(details []PatchDetail, query HeroChangesQuery) (HeroChangesResponse, error) {
	targetSlug := strings.TrimSpace(strings.ToLower(query.HeroSlug))
	if targetSlug == "" {
		return HeroChangesResponse{}, ErrHeroNotFound
	}

	skillFilter := strings.TrimSpace(query.Skill)
	matchGeneralOnly := strings.EqualFold(skillFilter, "general")
	filterSkillNorm := normalizeLookupKey(skillFilter)

	timeline := make([]HeroTimelineBlock, 0, 64)
	heroMeta := HeroSummary{Slug: targetSlug}

	for _, detail := range details {
		hydrated := hydratePatchDetail(detail)
		for _, block := range hydrated.Timeline {
			releasedAt := parseRFC3339(block.ReleasedAt)
			if !withinDateRange(releasedAt, query.From, query.To) {
				continue
			}
			for _, section := range block.Sections {
				if section.Kind != "heroes" {
					continue
				}
				for _, entry := range section.Entries {
					if !isHeroTimelineEntry(entry) {
						continue
					}
					entrySlug := slugifyLookup(entry.EntityName)
					if entrySlug != targetSlug {
						continue
					}
					if len(entry.Changes) == 0 && len(entry.Groups) == 0 {
						continue
					}

					if heroMeta.Name == "" {
						heroMeta.Name = entry.EntityName
					}
					if heroMeta.IconURL == "" {
						heroMeta.IconURL = strings.TrimSpace(entry.EntityIconURL)
					}
					if heroMeta.IconFallbackURL == "" {
						heroMeta.IconFallbackURL = strings.TrimSpace(entry.EntityIconFallbackURL)
					}
					if releasedAt.After(parseRFC3339(heroMeta.LastChangedAt)) {
						heroMeta.LastChangedAt = releasedAt.UTC().Format(time.RFC3339)
					}

					generalChanges := cloneChanges(entry.Changes)
					skills := make([]HeroTimelineSkill, 0, len(entry.Groups))
					for _, group := range entry.Groups {
						if len(group.Changes) == 0 {
							continue
						}
						skills = append(skills, HeroTimelineSkill{
							ID:              group.ID,
							Title:           group.Title,
							IconURL:         strings.TrimSpace(group.IconURL),
							IconFallbackURL: strings.TrimSpace(group.IconFallbackURL),
							Changes:         cloneChanges(group.Changes),
						})
					}

					if filterSkillNorm != "" {
						filtered := make([]HeroTimelineSkill, 0, len(skills))
						for _, skill := range skills {
							if normalizeLookupKey(skill.Title) == filterSkillNorm {
								filtered = append(filtered, skill)
							}
						}
						skills = filtered
						if matchGeneralOnly {
							skills = nil
						}
						if !matchGeneralOnly {
							generalChanges = nil
						}
						if matchGeneralOnly && len(generalChanges) == 0 {
							continue
						}
						if !matchGeneralOnly && len(skills) == 0 {
							continue
						}
					}

					timeline = append(timeline, HeroTimelineBlock{
						ID:             block.ID + "-" + targetSlug,
						Kind:           block.Kind,
						Label:          formatTimelineLabel(block.Kind, block.ReleasedAt),
						ReleasedAt:     block.ReleasedAt,
						Patch:          TimelinePatchRef{Slug: detail.Slug, Title: detail.Title},
						Source:         block.Source,
						GeneralChanges: generalChanges,
						Skills:         skills,
					})
				}
			}
		}
	}

	if heroMeta.Name == "" {
		return HeroChangesResponse{}, ErrHeroNotFound
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

	return HeroChangesResponse{
		Hero:  heroMeta,
		Items: timeline,
	}, nil
}

func withinDateRange(releasedAt time.Time, from, to *time.Time) bool {
	if from != nil && !releasedAt.IsZero() && releasedAt.Before(from.UTC()) {
		return false
	}
	if to != nil && !releasedAt.IsZero() && releasedAt.After(to.UTC()) {
		return false
	}
	return true
}

func isHeroTimelineEntry(entry PatchEntry) bool {
	if len(entry.Groups) > 0 {
		return true
	}
	if strings.TrimSpace(entry.EntityIconURL) != "" {
		return true
	}
	return strings.TrimSpace(entry.EntityIconFallbackURL) != ""
}

func formatTimelineLabel(kind, releasedAt string) string {
	parsed := parseRFC3339(releasedAt)
	date := "Unknown Date"
	if !parsed.IsZero() {
		date = parsed.UTC().Format("01-02-2006")
	}
	if kind == "initial" {
		return "Update " + date
	}
	return "Patch " + date
}

func cloneChanges(changes []PatchChange) []PatchChange {
	if len(changes) == 0 {
		return nil
	}
	cloned := make([]PatchChange, len(changes))
	copy(cloned, changes)
	return cloned
}
