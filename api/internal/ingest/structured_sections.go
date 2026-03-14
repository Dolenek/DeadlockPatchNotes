package ingest

import (
	"fmt"
	"strings"

	"deadlockpatchnotes/api/internal/patches"
	"deadlockpatchnotes/api/internal/structuredparse"
)

func buildStructuredSections(blocks []timelineCandidate, catalog *AssetCatalog) []patches.PatchSection {
	lines := make([]string, 0, len(blocks)*32)
	for _, block := range blocks {
		lines = append(lines, strings.Split(block.BodyText, "\n")...)
	}

	parsedSections := structuredparse.BuildSections(lines, structuredparse.Resolver{
		ResolveHero: func(name string) (structuredparse.HeroRef, bool) {
			hero, ok := catalog.resolveHero(name)
			if !ok {
				return structuredparse.HeroRef{}, false
			}
			displayName := resolveHeroDisplayName(name, hero)
			return structuredparse.HeroRef{
				Key:             structuredparse.CanonicalHeroKey(displayName),
				Name:            displayName,
				IconFallbackURL: hero.Images.IconImageSmall,
				Abilities:       toStructuredAbilities(catalog.heroAbilities(hero.Name)),
			}, true
		},
		ResolveItem: func(name, changeText string) (structuredparse.ItemRef, bool) {
			item, ok := catalog.resolveItem(name, changeText)
			if !ok {
				return structuredparse.ItemRef{}, false
			}
			return structuredparse.ItemRef{
				Key:             structuredparse.NormalizeLookupKey(item.Name),
				Name:            item.Name,
				IconFallbackURL: itemImage(item),
			}, true
		},
	})

	sections := toPatchSections(parsedSections)
	assignStructuredIDs(sections)
	return sections
}

func toPatchSections(input []structuredparse.Section) []patches.PatchSection {
	sections := make([]patches.PatchSection, 0, len(input))
	for _, section := range input {
		next := patches.PatchSection{
			ID:    section.ID,
			Title: section.Title,
			Kind:  section.Kind,
			Entries: make([]patches.PatchEntry, 0, len(section.Entries)),
		}
		for _, entry := range section.Entries {
			converted := patches.PatchEntry{
				EntityName:            entry.EntityName,
				EntityIconURL:         entry.EntityIconURL,
				EntityIconFallbackURL: entry.EntityIconFallbackURL,
				Changes:               toPatchChanges(entry.Changes),
				Groups:                toPatchGroups(entry.Groups),
			}
			next.Entries = append(next.Entries, converted)
		}
		sections = append(sections, next)
	}
	return sections
}

func toPatchGroups(input []structuredparse.Group) []patches.PatchEntryGroup {
	groups := make([]patches.PatchEntryGroup, 0, len(input))
	for _, group := range input {
		groups = append(groups, patches.PatchEntryGroup{
			Title:           group.Title,
			IconURL:         group.IconURL,
			IconFallbackURL: group.IconFallbackURL,
			Changes:         toPatchChanges(group.Changes),
		})
	}
	return groups
}

func toPatchChanges(input []structuredparse.Change) []patches.PatchChange {
	changes := make([]patches.PatchChange, 0, len(input))
	for _, change := range input {
		changes = append(changes, patches.PatchChange{Text: change.Text})
	}
	return changes
}

func toStructuredAbilities(input []abilityRef) []structuredparse.AbilityRef {
	abilities := make([]structuredparse.AbilityRef, 0, len(input))
	for _, ability := range input {
		abilities = append(abilities, structuredparse.AbilityRef{
			Name:            ability.Name,
			NormName:        ability.NormName,
			IconFallbackURL: firstNonEmpty(ability.Image, ability.ImageWebP),
		})
	}
	return structuredparse.SortAbilities(abilities)
}

func assignStructuredIDs(sections []patches.PatchSection) {
	for sectionIndex := range sections {
		section := &sections[sectionIndex]
		for entryIndex := range section.Entries {
			entry := &section.Entries[entryIndex]
			assignStructuredEntryID(section.Kind, entry)
			assignChangeIDs(entry.Changes, entry.ID)

			for groupIndex := range entry.Groups {
				group := &entry.Groups[groupIndex]
				group.ID = fmt.Sprintf("%s-%s", entry.ID, structuredparse.Slugify(group.Title))
				assignChangeIDs(group.Changes, group.ID)
			}
		}
	}
}

func assignStructuredEntryID(kind string, entry *patches.PatchEntry) {
	switch kind {
	case "general":
		entry.ID = "general-gameplay"
	case "items", "heroes":
		entry.ID = structuredparse.Slugify(entry.EntityName)
	default:
		entry.ID = structuredparse.Slugify(entry.EntityName)
	}
}

func assignChangeIDs(changes []patches.PatchChange, prefix string) {
	for changeIndex := range changes {
		changes[changeIndex].ID = fmt.Sprintf("%s-%d", prefix, changeIndex+1)
	}
}
