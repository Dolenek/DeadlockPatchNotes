package ingest

import (
	"fmt"
	"regexp"
	"strings"

	"deadlockpatchnotes/api/internal/patches"
)

var (
	sectionHeaderRegex   = regexp.MustCompile(`(?i)^\[\s*(general|items|heroes)\s*\]$`)
	datePatchHeadingRegex = regexp.MustCompile(`(?i)^\d{2}-\d{2}-\d{4}\s+patch:\s*$`)
	prefixedLineRegex    = regexp.MustCompile(`^([^:]{1,64}):\s*(.*)$`)
	nonAlphaNumRegex     = regexp.MustCompile(`[^a-z0-9]+`)
	spaceRegexLocal      = regexp.MustCompile(`\s+`)
)

var cardTypeNames = map[string]bool{
	"spades":  true,
	"diamond": true,
	"hearts":  true,
	"clubs":   true,
	"joker":   true,
}

type heroEntryState struct {
	entry               *patches.PatchEntry
	groupsByKey         map[string]*patches.PatchEntryGroup
	groupOrder          []string
	abilities           []abilityRef
	currentSpecialGroup string
}

type sectionAccumulator struct {
	catalog *AssetCatalog

	generalEntry *patches.PatchEntry
	itemEntries  map[string]*patches.PatchEntry
	itemOrder    []string
	heroEntries  map[string]*heroEntryState
	heroOrder    []string
}

func buildStructuredSections(blocks []timelineCandidate, catalog *AssetCatalog) []patches.PatchSection {
	acc := &sectionAccumulator{
		catalog: catalog,
		itemEntries: map[string]*patches.PatchEntry{},
		heroEntries: map[string]*heroEntryState{},
	}

	for _, block := range blocks {
		acc.consumeBlock(block)
	}

	return acc.buildSections()
}

func (a *sectionAccumulator) consumeBlock(block timelineCandidate) {
	lines := strings.Split(block.BodyText, "\n")
	mode := "general"
	currentHero := ""
	currentItem := ""

	for _, raw := range lines {
		line := cleanStructuredLine(raw)
		if line == "" || shouldSkipLine(line) {
			continue
		}

		if header, ok := parseSectionHeader(line); ok {
			mode = header
			continue
		}

		if mode == "heroes" && a.consumeHeroHeadingLine(line, &currentHero, &currentItem) {
			continue
		}

		prefix, text, hasPrefix := parsePrefixedLine(line)
		if hasPrefix {
			if hero, ok := a.catalog.resolveHero(prefix); ok {
				heroName := resolveHeroDisplayName(prefix, hero)
				heroImage := hero.Images.IconImageSmall
				state := a.ensureHero(heroName, heroImage)
				currentHero = normalizeLookupKey(heroName)
				currentItem = ""
				state.currentSpecialGroup = ""
				if text != "" {
					applyHeroChange(state, prefix, text)
				}
				continue
			}

			if item, ok := a.catalog.resolveItem(prefix, text); ok || mode == "items" {
				itemName := prefix
				icon := ""
				if ok {
					itemName = item.Name
					icon = itemImage(item)
				}
				entry := a.ensureItem(itemName, icon)
				currentItem = normalizeLookupKey(itemName)
				currentHero = ""
				changeText := text
				if changeText == "" {
					changeText = "Updated."
				}
				appendEntryChange(entry, changeText)
				continue
			}

			if mode == "heroes" && currentHero != "" {
				state := a.heroEntries[currentHero]
				if state != nil {
					applyHeroChange(state, prefix, text)
				}
				continue
			}

			if mode == "items" && currentItem != "" {
				entry := a.itemEntries[currentItem]
				if entry != nil {
					appendEntryChange(entry, fmt.Sprintf("%s: %s", prefix, text))
				}
				continue
			}

			appendEntryChange(a.ensureGeneral(), fmt.Sprintf("%s: %s", prefix, text))
			continue
		}

		switch {
		case mode == "heroes" && currentHero != "":
			if state := a.heroEntries[currentHero]; state != nil {
				applyHeroPlainLine(state, line)
			}
		case mode == "items" && currentItem != "":
			if entry := a.itemEntries[currentItem]; entry != nil {
				appendEntryChange(entry, line)
			}
		default:
			appendEntryChange(a.ensureGeneral(), line)
		}
	}
}

func (a *sectionAccumulator) consumeHeroHeadingLine(line string, currentHero, currentItem *string) bool {
	hero, ok := a.catalog.resolveHero(line)
	if !ok {
		return false
	}

	heroName := resolveHeroDisplayName(line, hero)
	state := a.ensureHero(heroName, hero.Images.IconImageSmall)
	*currentHero = normalizeLookupKey(heroName)
	*currentItem = ""
	state.currentSpecialGroup = ""
	return true
}

func (a *sectionAccumulator) ensureGeneral() *patches.PatchEntry {
	if a.generalEntry != nil {
		return a.generalEntry
	}
	a.generalEntry = &patches.PatchEntry{
		ID:         "general-gameplay",
		EntityName: "Core Gameplay",
	}
	return a.generalEntry
}

func (a *sectionAccumulator) ensureItem(name, iconFallback string) *patches.PatchEntry {
	key := normalizeLookupKey(name)
	if existing, ok := a.itemEntries[key]; ok {
		return existing
	}
	entry := &patches.PatchEntry{
		ID:                    slugifyStructured(name),
		EntityName:            name,
		EntityIconFallbackURL: strings.TrimSpace(iconFallback),
	}
	a.itemEntries[key] = entry
	a.itemOrder = append(a.itemOrder, key)
	return entry
}

func (a *sectionAccumulator) ensureHero(name, iconFallback string) *heroEntryState {
	key := normalizeLookupKey(name)
	if existing, ok := a.heroEntries[key]; ok {
		return existing
	}
	state := &heroEntryState{
		entry: &patches.PatchEntry{
			ID:                    slugifyStructured(name),
			EntityName:            name,
			EntityIconFallbackURL: strings.TrimSpace(iconFallback),
		},
		groupsByKey: map[string]*patches.PatchEntryGroup{},
		abilities:   a.catalog.heroAbilities(name),
	}
	a.heroEntries[key] = state
	a.heroOrder = append(a.heroOrder, key)
	return state
}

func (a *sectionAccumulator) buildSections() []patches.PatchSection {
	sections := make([]patches.PatchSection, 0, 3)

	if a.generalEntry != nil && len(a.generalEntry.Changes) > 0 {
		sections = append(sections, patches.PatchSection{
			ID:      "general",
			Title:   "General",
			Kind:    "general",
			Entries: []patches.PatchEntry{*a.generalEntry},
		})
	}

	if len(a.itemOrder) > 0 {
		itemEntries := make([]patches.PatchEntry, 0, len(a.itemOrder))
		for _, key := range a.itemOrder {
			entry := a.itemEntries[key]
			if entry == nil || len(entry.Changes) == 0 {
				continue
			}
			itemEntries = append(itemEntries, *entry)
		}
		if len(itemEntries) > 0 {
			sections = append(sections, patches.PatchSection{
				ID:      "items",
				Title:   "Items",
				Kind:    "items",
				Entries: itemEntries,
			})
		}
	}

	if len(a.heroOrder) > 0 {
		heroEntries := make([]patches.PatchEntry, 0, len(a.heroOrder))
		for _, key := range a.heroOrder {
			state := a.heroEntries[key]
			if state == nil {
				continue
			}
			state.entry.Groups = make([]patches.PatchEntryGroup, 0, len(state.groupOrder))
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
			heroEntries = append(heroEntries, *state.entry)
		}
		if len(heroEntries) > 0 {
			sections = append(sections, patches.PatchSection{
				ID:      "heroes",
				Title:   "Heroes",
				Kind:    "heroes",
				Entries: heroEntries,
			})
		}
	}

	return sections
}

func applyHeroPlainLine(state *heroEntryState, line string) {
	if normalizeLookupKey(line) == "card types" {
		state.currentSpecialGroup = "card-types"
		ensureHeroGroup(state, "card-types", "Card Types", "")
		return
	}

	applyHeroChange(state, "", line)
}

func applyHeroChange(state *heroEntryState, prefix, text string) {
	prefixKey := normalizeLookupKey(prefix)
	if prefixKey == "card types" {
		state.currentSpecialGroup = "card-types"
		ensureHeroGroup(state, "card-types", "Card Types", "")
		return
	}

	if state.currentSpecialGroup == "card-types" && cardTypeNames[prefixKey] {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		group := ensureHeroGroup(state, "card-types", "Card Types", "")
		appendGroupChange(group, fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), text))
		return
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	if strings.HasPrefix(strings.ToLower(text), "talents ") {
		group := ensureHeroGroup(state, "talents", "Talents", "")
		talentText := strings.TrimSpace(text[len("talents "):])
		appendGroupChange(group, talentText)
		return
	}

	if prefixKey != "" {
		if ability, ok := matchAbility(prefix, state.abilities); ok {
			groupKey := "ability-" + slugifyStructured(ability.Name)
			group := ensureHeroGroup(state, groupKey, ability.Name, firstNonEmpty(ability.Image, ability.ImageWebP))
			appendGroupChange(group, text)
			return
		}
	}

	if ability, ok := matchAbility(text, state.abilities); ok {
		groupKey := "ability-" + slugifyStructured(ability.Name)
		group := ensureHeroGroup(state, groupKey, ability.Name, firstNonEmpty(ability.Image, ability.ImageWebP))
		appendGroupChange(group, stripAbilityPrefix(text, ability.Name))
		return
	}

	if prefix != "" && text != "" && normalizeLookupKey(prefix) != normalizeLookupKey(state.entry.EntityName) {
		appendEntryChange(state.entry, fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), text))
		return
	}
	appendEntryChange(state.entry, text)
}

func ensureHeroGroup(state *heroEntryState, key, title, iconFallback string) *patches.PatchEntryGroup {
	if existing, ok := state.groupsByKey[key]; ok {
		return existing
	}
	group := &patches.PatchEntryGroup{
		ID:              fmt.Sprintf("%s-%s", state.entry.ID, slugifyStructured(title)),
		Title:           title,
		IconFallbackURL: strings.TrimSpace(iconFallback),
	}
	state.groupsByKey[key] = group
	state.groupOrder = append(state.groupOrder, key)
	return group
}

func appendEntryChange(entry *patches.PatchEntry, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	entry.Changes = append(entry.Changes, patches.PatchChange{
		ID:   fmt.Sprintf("%s-%d", entry.ID, len(entry.Changes)+1),
		Text: text,
	})
}

func appendGroupChange(group *patches.PatchEntryGroup, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	group.Changes = append(group.Changes, patches.PatchChange{
		ID:   fmt.Sprintf("%s-%d", group.ID, len(group.Changes)+1),
		Text: text,
	})
}

func parseSectionHeader(line string) (string, bool) {
	if match := sectionHeaderRegex.FindStringSubmatch(strings.TrimSpace(line)); len(match) == 2 {
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

func parsePrefixedLine(line string) (string, string, bool) {
	match := prefixedLineRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(match) != 3 {
		return "", "", false
	}
	return strings.TrimSpace(match[1]), strings.TrimSpace(match[2]), true
}

func shouldSkipLine(line string) bool {
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
	return datePatchHeadingRegex.MatchString(line)
}

func cleanStructuredLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "-")
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimSpace(line)
	return line
}

func matchAbility(text string, abilities []abilityRef) (abilityRef, bool) {
	if len(abilities) == 0 {
		return abilityRef{}, false
	}
	normalized := normalizeLookupKey(text)
	for _, ability := range abilities {
		if normalized == ability.NormName || strings.HasPrefix(normalized, ability.NormName+" ") {
			return ability, true
		}
	}
	return abilityRef{}, false
}

func stripAbilityPrefix(text, abilityName string) string {
	pattern := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(strings.TrimSpace(abilityName)) + `(?:\s+|$)`)
	stripped := strings.TrimSpace(pattern.ReplaceAllString(text, ""))
	if stripped == "" {
		return text
	}
	return stripped
}

func normalizeLookupKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = nonAlphaNumRegex.ReplaceAllString(value, " ")
	value = spaceRegexLocal.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func slugifyStructured(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = nonAlphaNumRegex.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "entry"
	}
	return value
}
