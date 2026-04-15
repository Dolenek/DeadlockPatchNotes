package structuredparse

import (
	"fmt"
	"strings"
)

type AbilityRef struct {
	Name            string
	NormName        string
	IconURL         string
	IconFallbackURL string
}

type HeroRef struct {
	Key             string
	Name            string
	IconURL         string
	IconFallbackURL string
	Abilities       []AbilityRef
}

type ItemRef struct {
	Key             string
	Name            string
	IconURL         string
	IconFallbackURL string
}

type Resolver struct {
	ResolveHero func(name string) (HeroRef, bool)
	ResolveItem func(name, changeText string) (ItemRef, bool)
}

type Change struct {
	Text string
}

type Group struct {
	Title           string
	IconURL         string
	IconFallbackURL string
	Changes         []Change
}

type Entry struct {
	EntityName            string
	EntityIconURL         string
	EntityIconFallbackURL string
	Changes               []Change
	Groups                []Group
}

type Section struct {
	ID      string
	Title   string
	Kind    string
	Entries []Entry
}

type heroEntryState struct {
	entry               *Entry
	groupsByKey         map[string]*Group
	groupOrder          []string
	abilities           []AbilityRef
	currentSpecialGroup string
}

type sectionAccumulator struct {
	resolver Resolver

	generalEntry *Entry
	itemEntries  map[string]*Entry
	itemOrder    []string
	heroEntries  map[string]*heroEntryState
	heroOrder    []string
}

func BuildSections(lines []string, resolver Resolver) []Section {
	acc := &sectionAccumulator{
		resolver:    resolver,
		itemEntries: map[string]*Entry{},
		heroEntries: map[string]*heroEntryState{},
	}

	mode := "general"
	currentHero := ""
	currentItem := ""

	for _, raw := range lines {
		line := CleanLine(raw)
		if line == "" || ShouldSkipLine(line) {
			continue
		}

		if header, ok := ParseSectionHeader(line); ok {
			mode = header
			currentHero = ""
			currentItem = ""
			continue
		}

		if mode == "heroes" && acc.consumeHeroHeadingLine(line, &currentHero, &currentItem) {
			continue
		}
		if mode == "items" && acc.consumeItemHeadingLine(line, &currentHero, &currentItem) {
			continue
		}

		prefix, text, hasPrefix := ParsePrefixedLine(line)
		if hasPrefix {
			if acc.consumePrefixedLine(mode, prefix, text, &currentHero, &currentItem) {
				continue
			}
		}

		switch {
		case mode == "heroes" && currentHero != "":
			if state := acc.heroEntries[currentHero]; state != nil {
				applyHeroPlainLine(state, line)
			}
		case mode == "items" && currentItem != "":
			if entry := acc.itemEntries[currentItem]; entry != nil {
				appendEntryChange(entry, line)
			}
		default:
			appendEntryChange(acc.ensureGeneral(), line)
		}
	}

	return acc.buildSections()
}

func (a *sectionAccumulator) consumeHeroHeadingLine(line string, currentHero, currentItem *string) bool {
	if a.resolver.ResolveHero == nil {
		return false
	}
	hero, ok := a.resolver.ResolveHero(line)
	if !ok {
		return false
	}

	state := a.ensureHero(hero)
	*currentHero = hero.Key
	*currentItem = ""
	state.currentSpecialGroup = ""
	return true
}

func (a *sectionAccumulator) consumeItemHeadingLine(line string, currentHero, currentItem *string) bool {
	if a.resolver.ResolveItem == nil {
		return false
	}

	item, ok := a.resolver.ResolveItem(line, "")
	if !ok {
		return false
	}

	a.ensureItem(item)
	*currentItem = entryKey(item.Key, item.Name)
	*currentHero = ""
	return true
}

func (a *sectionAccumulator) consumePrefixedLine(mode, prefix, text string, currentHero, currentItem *string) bool {
	if a.resolver.ResolveHero != nil {
		if hero, ok := a.resolver.ResolveHero(prefix); ok {
			state := a.ensureHero(hero)
			*currentHero = hero.Key
			*currentItem = ""
			state.currentSpecialGroup = ""
			if strings.TrimSpace(text) != "" {
				applyHeroChange(state, prefix, text)
			}
			return true
		}
	}

	if mode == "heroes" && *currentHero != "" {
		if state := a.heroEntries[*currentHero]; state != nil {
			applyHeroChange(state, prefix, text)
			return true
		}
	}

	item, itemResolved := ItemRef{}, false
	if a.resolver.ResolveItem != nil {
		item, itemResolved = a.resolver.ResolveItem(prefix, text)
	}
	if itemResolved || mode == "items" {
		if !itemResolved {
			item = ItemRef{Key: NormalizeLookupKey(prefix), Name: strings.TrimSpace(prefix)}
		}
		entry := a.ensureItem(item)
		*currentItem = entryKey(item.Key, item.Name)
		*currentHero = ""
		changeText := strings.TrimSpace(text)
		if changeText == "" {
			changeText = "Updated."
		}
		appendEntryChange(entry, changeText)
		return true
	}

	if mode == "items" && *currentItem != "" {
		if entry := a.itemEntries[*currentItem]; entry != nil {
			appendEntryChange(entry, fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), strings.TrimSpace(text)))
			return true
		}
	}

	appendEntryChange(a.ensureGeneral(), fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), strings.TrimSpace(text)))
	return true
}

func (a *sectionAccumulator) ensureGeneral() *Entry {
	if a.generalEntry != nil {
		return a.generalEntry
	}
	a.generalEntry = &Entry{EntityName: "Core Gameplay"}
	return a.generalEntry
}

func (a *sectionAccumulator) ensureItem(item ItemRef) *Entry {
	key := entryKey(item.Key, item.Name)
	if existing, ok := a.itemEntries[key]; ok {
		return existing
	}
	entry := &Entry{
		EntityName:            strings.TrimSpace(item.Name),
		EntityIconURL:         strings.TrimSpace(item.IconURL),
		EntityIconFallbackURL: strings.TrimSpace(item.IconFallbackURL),
	}
	a.itemEntries[key] = entry
	a.itemOrder = append(a.itemOrder, key)
	return entry
}

func entryKey(key, fallbackName string) string {
	if key != "" {
		return key
	}
	return NormalizeLookupKey(fallbackName)
}

func (a *sectionAccumulator) ensureHero(hero HeroRef) *heroEntryState {
	key := hero.Key
	if key == "" {
		key = CanonicalHeroKey(hero.Name)
	}
	if existing, ok := a.heroEntries[key]; ok {
		return existing
	}
	state := &heroEntryState{
		entry: &Entry{
			EntityName:            strings.TrimSpace(hero.Name),
			EntityIconURL:         strings.TrimSpace(hero.IconURL),
			EntityIconFallbackURL: strings.TrimSpace(hero.IconFallbackURL),
		},
		groupsByKey: map[string]*Group{},
		abilities:   hero.Abilities,
	}
	a.heroEntries[key] = state
	a.heroOrder = append(a.heroOrder, key)
	return state
}

func (a *sectionAccumulator) buildSections() []Section {
	sections := make([]Section, 0, 3)

	if a.generalEntry != nil && len(a.generalEntry.Changes) > 0 {
		sections = append(sections, Section{
			ID:      "general",
			Title:   "General",
			Kind:    "general",
			Entries: []Entry{*a.generalEntry},
		})
	}

	if len(a.itemOrder) > 0 {
		itemEntries := make([]Entry, 0, len(a.itemOrder))
		for _, key := range a.itemOrder {
			entry := a.itemEntries[key]
			if entry == nil || len(entry.Changes) == 0 {
				continue
			}
			itemEntries = append(itemEntries, *entry)
		}
		if len(itemEntries) > 0 {
			sections = append(sections, Section{
				ID:      "items",
				Title:   "Items",
				Kind:    "items",
				Entries: itemEntries,
			})
		}
	}

	if len(a.heroOrder) > 0 {
		heroEntries := make([]Entry, 0, len(a.heroOrder))
		for _, key := range a.heroOrder {
			state := a.heroEntries[key]
			if state == nil {
				continue
			}
			state.entry.Groups = make([]Group, 0, len(state.groupOrder))
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
			sections = append(sections, Section{
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
	if CanonicalHeroKey(line) == "card types" {
		state.currentSpecialGroup = "card-types"
		ensureHeroGroup(state, "card-types", "Card Types", "", "")
		return
	}

	applyHeroChange(state, "", line)
}

func applyHeroChange(state *heroEntryState, prefix, text string) {
	prefixKey := NormalizeLookupKey(prefix)
	if prefixKey == "card types" {
		state.currentSpecialGroup = "card-types"
		ensureHeroGroup(state, "card-types", "Card Types", "", "")
		return
	}

	if state.currentSpecialGroup == "card-types" && IsCardTypeName(prefixKey) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		group := ensureHeroGroup(state, "card-types", "Card Types", "", "")
		appendGroupChange(group, fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), text))
		return
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	if strings.HasPrefix(strings.ToLower(text), "talents ") {
		group := ensureHeroGroup(state, "talents", "Talents", "", "")
		appendGroupChange(group, strings.TrimSpace(text[len("talents "):]))
		return
	}

	if prefixKey != "" {
		if ability, ok := matchAbility(prefix, state.abilities); ok {
			groupKey := "ability-" + Slugify(ability.Name)
			group := ensureHeroGroup(state, groupKey, ability.Name, ability.IconURL, ability.IconFallbackURL)
			appendGroupChange(group, text)
			return
		}
	}

	if ability, ok := matchAbility(text, state.abilities); ok {
		groupKey := "ability-" + Slugify(ability.Name)
		group := ensureHeroGroup(state, groupKey, ability.Name, ability.IconURL, ability.IconFallbackURL)
		appendGroupChange(group, StripAbilityPrefix(text, ability.Name))
		return
	}

	if prefix != "" && CanonicalHeroKey(prefix) != CanonicalHeroKey(state.entry.EntityName) {
		appendEntryChange(state.entry, fmt.Sprintf("%s: %s", strings.TrimSpace(prefix), text))
		return
	}
	appendEntryChange(state.entry, text)
}

func ensureHeroGroup(state *heroEntryState, key, title, iconURL, iconFallbackURL string) *Group {
	if existing, ok := state.groupsByKey[key]; ok {
		return existing
	}
	group := &Group{
		Title:           title,
		IconURL:         strings.TrimSpace(iconURL),
		IconFallbackURL: strings.TrimSpace(iconFallbackURL),
	}
	state.groupsByKey[key] = group
	state.groupOrder = append(state.groupOrder, key)
	return group
}

func appendEntryChange(entry *Entry, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	entry.Changes = append(entry.Changes, Change{Text: text})
}

func appendGroupChange(group *Group, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	group.Changes = append(group.Changes, Change{Text: text})
}

func matchAbility(text string, abilities []AbilityRef) (AbilityRef, bool) {
	normalized := NormalizeLookupKey(text)
	for _, ability := range abilities {
		if normalized == ability.NormName || strings.HasPrefix(normalized, ability.NormName+" ") {
			return ability, true
		}
	}
	return AbilityRef{}, false
}
