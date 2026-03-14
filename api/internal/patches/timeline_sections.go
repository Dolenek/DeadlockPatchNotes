package patches

import (
	"fmt"
	"strings"

	"deadlockpatchnotes/api/internal/structuredparse"
)

func buildParseTemplateCatalog(sections []PatchSection) parseTemplateCatalog {
	catalog := parseTemplateCatalog{
		itemsByNorm:  map[string]PatchEntry{},
		heroesByNorm: map[string]heroTemplate{},
	}

	for _, section := range sections {
		switch section.Kind {
		case "items":
			for _, entry := range section.Entries {
				key := structuredparse.NormalizeLookupKey(entry.EntityName)
				if key == "" {
					continue
				}
				catalog.itemsByNorm[key] = entry
			}
		case "heroes":
			for _, entry := range section.Entries {
				key := structuredparse.CanonicalHeroKey(entry.EntityName)
				if key == "" {
					continue
				}
				catalog.heroesByNorm[key] = heroTemplate{
					name:            structuredparse.CanonicalHeroDisplayName(entry.EntityName),
					iconURL:         entry.EntityIconURL,
					iconFallbackURL: entry.EntityIconFallbackURL,
					abilities:       expandTimelineAbilities(key, entry.Groups),
				}
			}
		}
	}

	return catalog
}

func buildBlockSectionsFromChanges(block PatchTimelineBlock, catalog parseTemplateCatalog) []PatchSection {
	lines := make([]string, 0, len(block.Changes))
	for _, change := range block.Changes {
		if strings.TrimSpace(change.Text) == "" {
			continue
		}
		lines = append(lines, change.Text)
	}
	if len(lines) == 0 {
		return emptyTimelineSection(block.ID)
	}

	sections := structuredparse.BuildSections(lines, structuredparse.Resolver{
		ResolveHero: func(name string) (structuredparse.HeroRef, bool) {
			key := structuredparse.CanonicalHeroKey(name)
			template, ok := catalog.heroesByNorm[key]
			if !ok {
				return structuredparse.HeroRef{}, false
			}
			return structuredparse.HeroRef{
				Key:             key,
				Name:            template.name,
				IconURL:         strings.TrimSpace(template.iconURL),
				IconFallbackURL: strings.TrimSpace(template.iconFallbackURL),
				Abilities:       toStructuredTimelineAbilities(template.abilities),
			}, true
		},
		ResolveItem: func(name, _ string) (structuredparse.ItemRef, bool) {
			key := structuredparse.NormalizeLookupKey(name)
			template, ok := catalog.itemsByNorm[key]
			if !ok {
				return structuredparse.ItemRef{}, false
			}
			return structuredparse.ItemRef{
				Key:             key,
				Name:            strings.TrimSpace(template.EntityName),
				IconURL:         strings.TrimSpace(template.EntityIconURL),
				IconFallbackURL: strings.TrimSpace(template.EntityIconFallbackURL),
			}, true
		},
	})
	if len(sections) == 0 {
		return emptyTimelineSection(block.ID)
	}

	output := toTimelinePatchSections(sections)
	assignTimelineIDs(block.ID, output)
	return output
}

func expandTimelineAbilities(heroKey string, groups []PatchEntryGroup) []abilityTemplate {
	abilities := make([]structuredparse.AbilityRef, 0, len(groups))
	for _, group := range groups {
		if strings.TrimSpace(group.Title) == "" {
			continue
		}
		abilities = append(abilities, structuredparse.AbilityRef{
			Name:            group.Title,
			IconURL:         group.IconURL,
			IconFallbackURL: group.IconFallbackURL,
		})
	}

	expanded := structuredparse.ExpandAbilityAliases(heroKey, abilities)
	out := make([]abilityTemplate, 0, len(expanded))
	for _, ability := range expanded {
		out = append(out, abilityTemplate{
			name:            ability.Name,
			normName:        ability.NormName,
			iconURL:         ability.IconURL,
			iconFallbackURL: ability.IconFallbackURL,
		})
	}
	return out
}

func toStructuredTimelineAbilities(input []abilityTemplate) []structuredparse.AbilityRef {
	abilities := make([]structuredparse.AbilityRef, 0, len(input))
	for _, ability := range input {
		abilities = append(abilities, structuredparse.AbilityRef{
			Name:            ability.name,
			NormName:        ability.normName,
			IconURL:         ability.iconURL,
			IconFallbackURL: ability.iconFallbackURL,
		})
	}
	return structuredparse.SortAbilities(abilities)
}

func assignTimelineIDs(blockID string, sections []PatchSection) {
	for sectionIndex := range sections {
		section := &sections[sectionIndex]
		section.ID = blockID + "-" + section.Kind
		for entryIndex := range section.Entries {
			entry := &section.Entries[entryIndex]
			switch section.Kind {
			case "general":
				entry.ID = blockID + "-general-gameplay"
			case "items":
				entry.ID = blockID + "-item-" + slugifyLookup(entry.EntityName)
			case "heroes":
				entry.ID = blockID + "-hero-" + slugifyLookup(entry.EntityName)
			default:
				entry.ID = blockID + "-entry-" + slugifyLookup(entry.EntityName)
			}
			assignTimelineChangeIDs(entry.Changes, entry.ID)

			for groupIndex := range entry.Groups {
				group := &entry.Groups[groupIndex]
				group.ID = entry.ID + "-group-" + slugifyLookup(group.Title)
				assignTimelineChangeIDs(group.Changes, group.ID)
			}
		}
	}
}

func assignTimelineChangeIDs(changes []PatchChange, prefix string) {
	for changeIndex := range changes {
		changes[changeIndex].ID = fmt.Sprintf("%s-%d", prefix, changeIndex+1)
	}
}

func emptyTimelineSection(blockID string) []PatchSection {
	return []PatchSection{
		{
			ID:    blockID + "-general",
			Title: "General",
			Kind:  "general",
			Entries: []PatchEntry{
				{
					ID:         blockID + "-general-gameplay",
					EntityName: "Core Gameplay",
					Changes: []PatchChange{
						{ID: blockID + "-general-1", Text: "No line-item changes listed."},
					},
				},
			},
		},
	}
}

func toTimelinePatchSections(input []structuredparse.Section) []PatchSection {
	sections := make([]PatchSection, 0, len(input))
	for _, section := range input {
		next := PatchSection{
			ID:      section.ID,
			Title:   section.Title,
			Kind:    section.Kind,
			Entries: make([]PatchEntry, 0, len(section.Entries)),
		}
		for _, entry := range section.Entries {
			next.Entries = append(next.Entries, PatchEntry{
				EntityName:            entry.EntityName,
				EntityIconURL:         entry.EntityIconURL,
				EntityIconFallbackURL: entry.EntityIconFallbackURL,
				Changes:               toTimelineChanges(entry.Changes),
				Groups:                toTimelineGroups(entry.Groups),
			})
		}
		sections = append(sections, next)
	}
	return sections
}

func toTimelineGroups(input []structuredparse.Group) []PatchEntryGroup {
	groups := make([]PatchEntryGroup, 0, len(input))
	for _, group := range input {
		groups = append(groups, PatchEntryGroup{
			Title:           group.Title,
			IconURL:         group.IconURL,
			IconFallbackURL: group.IconFallbackURL,
			Changes:         toTimelineChanges(group.Changes),
		})
	}
	return groups
}

func toTimelineChanges(input []structuredparse.Change) []PatchChange {
	changes := make([]PatchChange, 0, len(input))
	for _, change := range input {
		changes = append(changes, PatchChange{Text: change.Text})
	}
	return changes
}
