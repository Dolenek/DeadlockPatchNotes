package patches

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
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
				key := normalizeLookupKey(entry.EntityName)
				if key == "" {
					continue
				}
				catalog.itemsByNorm[key] = entry
			}
		case "heroes":
			for _, entry := range section.Entries {
				key := normalizeLookupKey(entry.EntityName)
				if key == "" {
					continue
				}
				template := heroTemplate{
					name:            entry.EntityName,
					iconURL:         entry.EntityIconURL,
					iconFallbackURL: entry.EntityIconFallbackURL,
					abilities:       make([]abilityTemplate, 0, len(entry.Groups)),
				}
				for _, group := range entry.Groups {
					if strings.TrimSpace(group.Title) == "" {
						continue
					}
					template.abilities = append(template.abilities, abilityTemplate{
						name:            group.Title,
						normName:        normalizeLookupKey(group.Title),
						iconURL:         group.IconURL,
						iconFallbackURL: group.IconFallbackURL,
					})
				}
				sort.SliceStable(template.abilities, func(i, j int) bool {
					return len(template.abilities[i].normName) > len(template.abilities[j].normName)
				})
				catalog.heroesByNorm[key] = template
			}
		}
	}

	return catalog
}

func buildBlockSectionsFromChanges(block PatchTimelineBlock, catalog parseTemplateCatalog) []PatchSection {
	lines := make([]string, 0, len(block.Changes))
	for _, change := range block.Changes {
		text := strings.TrimSpace(change.Text)
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}

	if len(lines) == 0 {
		return emptyTimelineSection(block.ID)
	}

	generalEntry := PatchEntry{ID: block.ID + "-general-gameplay", EntityName: "Core Gameplay"}
	itemEntries := map[string]*PatchEntry{}
	itemOrder := make([]string, 0, 16)
	heroEntries := map[string]*timelineHeroState{}
	heroOrder := make([]string, 0, 32)

	mode := "general"
	currentHero := ""
	currentItem := ""

	for _, raw := range lines {
		line := cleanTimelineLine(raw)
		if shouldSkipTimelineLine(line) {
			continue
		}
		if header, ok := parseStructuredSectionHeader(line); ok {
			mode = header
			continue
		}

		prefix, text, hasPrefix := parseStructuredPrefixedLine(line)
		if hasPrefix {
			heroKey := normalizeLookupKey(prefix)
			if template, ok := catalog.heroesByNorm[heroKey]; ok || mode == "heroes" {
				state := ensureTimelineHeroEntry(block.ID, heroEntries, &heroOrder, heroKey, prefix, template)
				currentHero = heroKey
				currentItem = ""
				if strings.TrimSpace(text) != "" {
					appendHeroTimelineLine(state, prefix, text)
				}
				continue
			}

			itemKey := normalizeLookupKey(prefix)
			if template, ok := catalog.itemsByNorm[itemKey]; ok || mode == "items" {
				entry := ensureTimelineItemEntry(block.ID, itemEntries, &itemOrder, itemKey, prefix, template)
				currentItem = itemKey
				currentHero = ""
				changeText := strings.TrimSpace(text)
				if changeText == "" {
					changeText = "Updated."
				}
				appendTimelineEntryChange(entry, changeText)
				continue
			}

			if mode == "heroes" && currentHero != "" {
				if state := heroEntries[currentHero]; state != nil {
					appendHeroTimelineLine(state, prefix, text)
					continue
				}
			}
			if mode == "items" && currentItem != "" {
				if entry := itemEntries[currentItem]; entry != nil {
					appendTimelineEntryChange(entry, fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), strings.TrimSpace(text)))
					continue
				}
			}

			appendTimelineEntryChange(&generalEntry, fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), strings.TrimSpace(text)))
			continue
		}

		switch {
		case mode == "heroes" && currentHero != "":
			if state := heroEntries[currentHero]; state != nil {
				appendHeroTimelineLine(state, "", line)
			}
		case mode == "items" && currentItem != "":
			if entry := itemEntries[currentItem]; entry != nil {
				appendTimelineEntryChange(entry, line)
			}
		default:
			appendTimelineEntryChange(&generalEntry, line)
		}
	}

	sections := make([]PatchSection, 0, 3)
	if len(generalEntry.Changes) > 0 {
		sections = append(sections, PatchSection{
			ID:      block.ID + "-general",
			Title:   "General",
			Kind:    "general",
			Entries: []PatchEntry{generalEntry},
		})
	}

	items := collectTimelineItems(block.ID, itemEntries, itemOrder)
	if len(items) > 0 {
		sections = append(sections, PatchSection{
			ID:      block.ID + "-items",
			Title:   "Items",
			Kind:    "items",
			Entries: items,
		})
	}

	heroes := collectTimelineHeroes(block.ID, heroEntries, heroOrder)
	if len(heroes) > 0 {
		sections = append(sections, PatchSection{
			ID:      block.ID + "-heroes",
			Title:   "Heroes",
			Kind:    "heroes",
			Entries: heroes,
		})
	}

	if len(sections) == 0 {
		return emptyTimelineSection(block.ID)
	}

	return sections
}

func collectTimelineItems(blockID string, entries map[string]*PatchEntry, order []string) []PatchEntry {
	items := make([]PatchEntry, 0, len(order))
	for _, key := range order {
		entry := entries[key]
		if entry == nil || len(entry.Changes) == 0 {
			continue
		}
		items = append(items, *entry)
	}
	return items
}

func collectTimelineHeroes(blockID string, entries map[string]*timelineHeroState, order []string) []PatchEntry {
	heroes := make([]PatchEntry, 0, len(order))
	for _, key := range order {
		state := entries[key]
		if state == nil {
			continue
		}
		state.entry.Groups = make([]PatchEntryGroup, 0, len(state.groupOrder))
		for _, groupKey := range state.groupOrder {
			group := state.groupsByKey[groupKey]
			if group == nil || len(group.Changes) == 0 {
				continue
			}
			state.entry.Groups = append(state.entry.Groups, *group)
		}
		if len(state.entry.Changes) == 0 && len(state.entry.Groups) == 0 {
			continue
		}
		heroes = append(heroes, *state.entry)
	}
	return heroes
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

func ensureTimelineItemEntry(blockID string, entries map[string]*PatchEntry, order *[]string, key, fallback string, template PatchEntry) *PatchEntry {
	if existing, ok := entries[key]; ok {
		return existing
	}

	entityName := strings.TrimSpace(template.EntityName)
	iconURL := strings.TrimSpace(template.EntityIconURL)
	iconFallbackURL := strings.TrimSpace(template.EntityIconFallbackURL)
	if entityName == "" {
		entityName = strings.TrimSpace(fallback)
	}

	entry := &PatchEntry{
		ID:                    blockID + "-item-" + slugifyLookup(entityName),
		EntityName:            entityName,
		EntityIconURL:         iconURL,
		EntityIconFallbackURL: iconFallbackURL,
	}
	entries[key] = entry
	*order = append(*order, key)
	return entry
}

func ensureTimelineHeroEntry(blockID string, entries map[string]*timelineHeroState, order *[]string, key, fallback string, template heroTemplate) *timelineHeroState {
	if existing, ok := entries[key]; ok {
		return existing
	}

	entityName := strings.TrimSpace(template.name)
	if entityName == "" {
		entityName = strings.TrimSpace(fallback)
	}

	state := &timelineHeroState{
		entry: &PatchEntry{
			ID:                    blockID + "-hero-" + slugifyLookup(entityName),
			EntityName:            entityName,
			EntityIconURL:         strings.TrimSpace(template.iconURL),
			EntityIconFallbackURL: strings.TrimSpace(template.iconFallbackURL),
		},
		abilities:   template.abilities,
		groupsByKey: map[string]*PatchEntryGroup{},
	}
	entries[key] = state
	*order = append(*order, key)
	return state
}

func appendHeroTimelineLine(state *timelineHeroState, prefix, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	if strings.HasPrefix(strings.ToLower(text), "talents ") {
		group := ensureHeroTimelineGroup(state, "talents", "Talents", "", "")
		appendTimelineGroupChange(group, strings.TrimSpace(text[len("talents "):]))
		return
	}

	if ability, ok := matchTimelineAbility(text, state.abilities); ok {
		groupKey := "ability-" + slugifyLookup(ability.name)
		group := ensureHeroTimelineGroup(state, groupKey, ability.name, ability.iconURL, ability.iconFallbackURL)
		appendTimelineGroupChange(group, stripTimelineAbilityPrefix(text, ability.name))
		return
	}

	if prefix != "" && normalizeLookupKey(prefix) != normalizeLookupKey(state.entry.EntityName) {
		appendTimelineEntryChange(state.entry, fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), text))
		return
	}
	appendTimelineEntryChange(state.entry, text)
}

func ensureHeroTimelineGroup(state *timelineHeroState, key, title, iconURL, iconFallbackURL string) *PatchEntryGroup {
	if existing, ok := state.groupsByKey[key]; ok {
		return existing
	}

	group := &PatchEntryGroup{
		ID:              state.entry.ID + "-group-" + slugifyLookup(title),
		Title:           title,
		IconURL:         strings.TrimSpace(iconURL),
		IconFallbackURL: strings.TrimSpace(iconFallbackURL),
	}
	state.groupsByKey[key] = group
	state.groupOrder = append(state.groupOrder, key)
	return group
}

func matchTimelineAbility(text string, abilities []abilityTemplate) (abilityTemplate, bool) {
	normalized := normalizeLookupKey(text)
	for _, ability := range abilities {
		if normalized == ability.normName || strings.HasPrefix(normalized, ability.normName+" ") {
			return ability, true
		}
	}
	return abilityTemplate{}, false
}

func stripTimelineAbilityPrefix(text, ability string) string {
	pattern := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(strings.TrimSpace(ability)) + `(?:\s+|$)`)
	stripped := strings.TrimSpace(pattern.ReplaceAllString(text, ""))
	if stripped == "" {
		return text
	}
	return stripped
}

func appendTimelineEntryChange(entry *PatchEntry, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	entry.Changes = append(entry.Changes, PatchChange{
		ID:   fmt.Sprintf("%s-%d", entry.ID, len(entry.Changes)+1),
		Text: text,
	})
}

func appendTimelineGroupChange(group *PatchEntryGroup, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	group.Changes = append(group.Changes, PatchChange{
		ID:   fmt.Sprintf("%s-%d", group.ID, len(group.Changes)+1),
		Text: text,
	})
}

func parseStructuredSectionHeader(line string) (string, bool) {
	match := structuredSectionHeaderRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(match) == 2 {
		return strings.ToLower(match[1]), true
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "general":
		return "general", true
	case "items":
		return "items", true
	case "heroes":
		return "heroes", true
	}
	return "", false
}

func parseStructuredPrefixedLine(line string) (string, string, bool) {
	match := structuredPrefixedLineRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(match) != 3 {
		return "", "", false
	}
	return strings.TrimSpace(match[1]), strings.TrimSpace(match[2]), true
}

func shouldSkipTimelineLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" {
		return true
	}
	if lower == "read more" {
		return true
	}
	if strings.HasPrefix(lower, "deadlock - ") && strings.Contains(lower, "steam news") {
		return true
	}
	return structuredDateHeadingRegex.MatchString(line)
}

func cleanTimelineLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "-")
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimSpace(line)
	return line
}
