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
	sections = repairMisclassifiedItemAbilities(sections, catalog)
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

func repairMisclassifiedItemAbilities(sections []patches.PatchSection, catalog *AssetCatalog) []patches.PatchSection {
	if catalog == nil {
		return sections
	}

	itemsIndex := -1
	heroesIndex := -1
	for index := range sections {
		switch sections[index].Kind {
		case "items":
			itemsIndex = index
		case "heroes":
			heroesIndex = index
		}
	}
	if itemsIndex == -1 {
		return sections
	}

	if heroesIndex == -1 {
		sections = append(sections, patches.PatchSection{
			ID:    "heroes",
			Title: "Heroes",
			Kind:  "heroes",
		})
		heroesIndex = len(sections) - 1
	}

	heroEntryIndexes := make(map[string]int, len(sections[heroesIndex].Entries))
	for index := range sections[heroesIndex].Entries {
		key := structuredparse.CanonicalHeroKey(sections[heroesIndex].Entries[index].EntityName)
		if key == "" {
			continue
		}
		heroEntryIndexes[key] = index
	}

	filteredItems := make([]patches.PatchEntry, 0, len(sections[itemsIndex].Entries))
	for _, entry := range sections[itemsIndex].Entries {
		if shouldKeepItemEntry(entry, catalog) {
			filteredItems = append(filteredItems, entry)
			continue
		}

		owner, ok := catalog.resolveUniqueAbility(entry.EntityName)
		if !ok {
			filteredItems = append(filteredItems, entry)
			continue
		}

		heroEntry := ensureHeroSectionEntry(&sections[heroesIndex], heroEntryIndexes, owner)
		abilityGroup := ensureHeroAbilityGroup(heroEntry, owner)
		abilityGroup.Changes = append(abilityGroup.Changes, clonePatchChanges(entry.Changes)...)
		for _, group := range entry.Groups {
			abilityGroup.Changes = append(abilityGroup.Changes, clonePatchChanges(group.Changes)...)
		}
	}

	sections[itemsIndex].Entries = filteredItems
	return compactStructuredSections(sections)
}

func shouldKeepItemEntry(entry patches.PatchEntry, catalog *AssetCatalog) bool {
	if _, ok := catalog.resolveNonAbilityItem(entry.EntityName, firstEntryChangeText(entry)); ok {
		return true
	}
	_, ok := catalog.resolveUniqueAbility(entry.EntityName)
	return !ok
}

func firstEntryChangeText(entry patches.PatchEntry) string {
	if len(entry.Changes) > 0 {
		return entry.Changes[0].Text
	}
	return ""
}

func ensureHeroSectionEntry(section *patches.PatchSection, indexes map[string]int, owner abilityOwnerRef) *patches.PatchEntry {
	if existingIndex, ok := indexes[owner.HeroKey]; ok {
		entry := &section.Entries[existingIndex]
		if entry.EntityIconFallbackURL == "" {
			entry.EntityIconFallbackURL = owner.HeroIconFallbackURL
		}
		if entry.EntityName == "" {
			entry.EntityName = owner.HeroName
		}
		return entry
	}

	section.Entries = append(section.Entries, patches.PatchEntry{
		EntityName:            owner.HeroName,
		EntityIconFallbackURL: owner.HeroIconFallbackURL,
	})
	indexes[owner.HeroKey] = len(section.Entries) - 1
	return &section.Entries[len(section.Entries)-1]
}

func ensureHeroAbilityGroup(entry *patches.PatchEntry, owner abilityOwnerRef) *patches.PatchEntryGroup {
	abilityKey := structuredparse.NormalizeLookupKey(owner.AbilityName)
	for index := range entry.Groups {
		if structuredparse.NormalizeLookupKey(entry.Groups[index].Title) != abilityKey {
			continue
		}
		group := &entry.Groups[index]
		if group.IconFallbackURL == "" {
			group.IconFallbackURL = owner.AbilityIconFallbackURL
		}
		if group.Title == "" {
			group.Title = owner.AbilityName
		}
		return group
	}

	entry.Groups = append(entry.Groups, patches.PatchEntryGroup{
		Title:           owner.AbilityName,
		IconFallbackURL: owner.AbilityIconFallbackURL,
	})
	return &entry.Groups[len(entry.Groups)-1]
}

func compactStructuredSections(sections []patches.PatchSection) []patches.PatchSection {
	filtered := make([]patches.PatchSection, 0, len(sections))
	for _, section := range sections {
		if len(section.Entries) == 0 {
			continue
		}
		filtered = append(filtered, section)
	}
	return filtered
}

func clonePatchChanges(changes []patches.PatchChange) []patches.PatchChange {
	if len(changes) == 0 {
		return nil
	}
	cloned := make([]patches.PatchChange, len(changes))
	copy(cloned, changes)
	return cloned
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
