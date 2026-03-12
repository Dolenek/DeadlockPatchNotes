package patches

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var timelineCardTypeNames = map[string]bool{
	"spades":  true,
	"diamond": true,
	"hearts":  true,
	"clubs":   true,
	"joker":   true,
}

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
					name:            canonicalTimelineHeroName(entry.EntityName),
					iconURL:         entry.EntityIconURL,
					iconFallbackURL: entry.EntityIconFallbackURL,
					abilities:       make([]abilityTemplate, 0, len(entry.Groups)),
				}
				for _, group := range entry.Groups {
					if strings.TrimSpace(group.Title) == "" {
						continue
					}
					ability := abilityTemplate{
						name:            group.Title,
						normName:        normalizeLookupKey(group.Title),
						iconURL:         group.IconURL,
						iconFallbackURL: group.IconFallbackURL,
					}
					template.abilities = append(template.abilities, ability)

					for _, alias := range timelineAbilityAlias[canonicalTimelineHeroKey(key)][ability.normName] {
						template.abilities = append(template.abilities, abilityTemplate{
							name:            ability.name,
							normName:        normalizeLookupKey(alias),
							iconURL:         ability.iconURL,
							iconFallbackURL: ability.iconFallbackURL,
						})
					}
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

		if mode == "heroes" {
			if heroKey, template, ok := resolveTimelineHeroTemplate(catalog, line); ok {
				state := ensureTimelineHeroEntry(block.ID, heroEntries, &heroOrder, heroKey, line, template)
				currentHero = heroKey
				currentItem = ""
				state.currentSpecialGroup = ""
				continue
			}
		}

		prefix, text, hasPrefix := parseStructuredPrefixedLine(line)
		if hasPrefix {
			if heroKey, template, ok := resolveTimelineHeroTemplate(catalog, prefix); ok {
				state := ensureTimelineHeroEntry(block.ID, heroEntries, &heroOrder, heroKey, prefix, template)
				currentHero = heroKey
				currentItem = ""
				state.currentSpecialGroup = ""
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
	entityName = canonicalTimelineHeroName(entityName)

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
	prefixKey := normalizeLookupKey(prefix)
	if prefixKey == "card types" || (prefixKey == "" && normalizeLookupKey(text) == "card types") {
		state.currentSpecialGroup = "card-types"
		ensureHeroTimelineGroup(state, "card-types", "Card Types", "", "")
		return
	}

	if state.currentSpecialGroup == "card-types" && timelineCardTypeNames[prefixKey] {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		group := ensureHeroTimelineGroup(state, "card-types", "Card Types", "", "")
		appendTimelineGroupChange(group, fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), text))
		return
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	if strings.HasPrefix(strings.ToLower(text), "talents ") {
		group := ensureHeroTimelineGroup(state, "talents", "Talents", "", "")
		appendTimelineGroupChange(group, strings.TrimSpace(text[len("talents "):]))
		return
	}

	if prefixKey != "" {
		if ability, ok := matchTimelineAbility(prefix, state.abilities); ok {
			groupKey := "ability-" + slugifyLookup(ability.name)
			group := ensureHeroTimelineGroup(state, groupKey, ability.name, ability.iconURL, ability.iconFallbackURL)
			appendTimelineGroupChange(group, text)
			return
		}
	}

	if ability, ok := matchTimelineAbility(text, state.abilities); ok {
		groupKey := "ability-" + slugifyLookup(ability.name)
		group := ensureHeroTimelineGroup(state, groupKey, ability.name, ability.iconURL, ability.iconFallbackURL)
		appendTimelineGroupChange(group, stripTimelineAbilityPrefix(text, ability.name))
		return
	}

	if prefix != "" && canonicalTimelineHeroKey(prefix) != canonicalTimelineHeroKey(state.entry.EntityName) {
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

func resolveTimelineHeroTemplate(catalog parseTemplateCatalog, raw string) (string, heroTemplate, bool) {
	key := normalizeLookupKey(raw)
	if key == "" {
		return "", heroTemplate{}, false
	}

	if template, ok := catalog.heroesByNorm[key]; ok {
		return canonicalTimelineHeroKey(key), template, true
	}

	if alias, ok := timelineHeroAlias[key]; ok {
		aliasKey := normalizeLookupKey(alias)
		if template, ok := catalog.heroesByNorm[aliasKey]; ok {
			return canonicalTimelineHeroKey(aliasKey), template, true
		}
	}

	if strings.HasPrefix(key, "the ") {
		trimmed := strings.TrimPrefix(key, "the ")
		if template, ok := catalog.heroesByNorm[trimmed]; ok {
			return canonicalTimelineHeroKey(trimmed), template, true
		}
	} else {
		withArticle := "the " + key
		if template, ok := catalog.heroesByNorm[withArticle]; ok {
			return canonicalTimelineHeroKey(withArticle), template, true
		}
	}

	return "", heroTemplate{}, false
}
