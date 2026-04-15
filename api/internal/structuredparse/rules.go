package structuredparse

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

var (
	bracketSectionHeaderRegex = regexp.MustCompile(`^\[\s*(.+?)\s*\]$`)
	datePatchHeadingRegex     = regexp.MustCompile(`(?i)^\d{2}-\d{2}-\d{4}\s+patch:\s*$`)
	prefixedLineRegex         = regexp.MustCompile(`^([^:]{1,64}):\s*(.*)$`)
	nonAlphaNumRegex          = regexp.MustCompile(`[^a-z0-9]+`)
	spaceRegex                = regexp.MustCompile(`\s+`)
)

//go:embed rules.json
var rulesJSON []byte

type parserRules struct {
	HeroAliases        map[string]string              `json:"heroAliases"`
	HeroCanonicalNames map[string]string              `json:"heroCanonicalNames"`
	HeroAbilityAliases map[string]map[string][]string `json:"heroAbilityAliases"`
	CardTypeNames      []string                       `json:"cardTypeNames"`
}

var loadedRules = loadRules()

func loadRules() parserRules {
	var rules parserRules
	if err := json.Unmarshal(rulesJSON, &rules); err != nil {
		panic(err)
	}
	return rules
}

func NormalizeLookupKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = nonAlphaNumRegex.ReplaceAllString(value, " ")
	value = spaceRegex.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func Slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = nonAlphaNumRegex.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "entry"
	}
	return value
}

func ParseSectionHeader(line string) (string, bool) {
	if match := bracketSectionHeaderRegex.FindStringSubmatch(strings.TrimSpace(line)); len(match) == 2 {
		return classifyBracketSectionHeader(match[1]), true
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

func classifyBracketSectionHeader(value string) string {
	tokens := strings.Fields(NormalizeLookupKey(value))
	if len(tokens) == 0 {
		return "general"
	}

	switch tokens[0] {
	case "hero", "heroes":
		return "heroes"
	case "item", "items":
		return "items"
	default:
		return "general"
	}
}

func ParsePrefixedLine(line string) (string, string, bool) {
	match := prefixedLineRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(match) != 3 {
		return "", "", false
	}
	return strings.TrimSpace(match[1]), strings.TrimSpace(match[2]), true
}

func ShouldSkipLine(line string) bool {
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

func CleanLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "-")
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimSpace(line)
	return line
}

func ResolveHeroAlias(name string) string {
	key := NormalizeLookupKey(name)
	if alias, ok := loadedRules.HeroAliases[key]; ok {
		return alias
	}
	return key
}

func CanonicalHeroKey(name string) string {
	key := NormalizeLookupKey(name)
	if alias, ok := loadedRules.HeroAliases[key]; ok {
		key = NormalizeLookupKey(alias)
	}
	if canonical, ok := loadedRules.HeroCanonicalNames[key]; ok {
		return NormalizeLookupKey(canonical)
	}
	return key
}

func CanonicalHeroDisplayName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	key := NormalizeLookupKey(trimmed)
	if alias, ok := loadedRules.HeroAliases[key]; ok {
		key = NormalizeLookupKey(alias)
	}
	if canonical, ok := loadedRules.HeroCanonicalNames[key]; ok {
		return canonical
	}
	return trimmed
}

func AbilityAliases(heroKey, abilityName string) []string {
	heroKey = CanonicalHeroKey(heroKey)
	abilityName = NormalizeLookupKey(abilityName)
	aliases := loadedRules.HeroAbilityAliases[heroKey][abilityName]
	out := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		alias = NormalizeLookupKey(alias)
		if alias == "" {
			continue
		}
		out = append(out, alias)
	}
	return out
}

func IsCardTypeName(value string) bool {
	key := NormalizeLookupKey(value)
	for _, candidate := range loadedRules.CardTypeNames {
		if key == NormalizeLookupKey(candidate) {
			return true
		}
	}
	return false
}

func SortAbilities(abilities []AbilityRef) []AbilityRef {
	sorted := make([]AbilityRef, len(abilities))
	copy(sorted, abilities)
	sort.SliceStable(sorted, func(i, j int) bool {
		return len(sorted[i].NormName) > len(sorted[j].NormName)
	})
	return sorted
}

func ExpandAbilityAliases(heroKey string, abilities []AbilityRef) []AbilityRef {
	expanded := make([]AbilityRef, 0, len(abilities)*2)
	heroKey = CanonicalHeroKey(heroKey)
	for _, ability := range abilities {
		ability.NormName = NormalizeLookupKey(ability.Name)
		expanded = append(expanded, ability)
		for _, alias := range AbilityAliases(heroKey, ability.Name) {
			expanded = append(expanded, AbilityRef{
				Name:            ability.Name,
				NormName:        alias,
				IconURL:         ability.IconURL,
				IconFallbackURL: ability.IconFallbackURL,
			})
		}
	}
	return SortAbilities(expanded)
}

func StripAbilityPrefix(text, abilityName string) string {
	pattern := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(strings.TrimSpace(abilityName)) + `(?:\s+|$)`)
	stripped := strings.TrimSpace(pattern.ReplaceAllString(text, ""))
	if stripped == "" {
		return text
	}
	return stripped
}
