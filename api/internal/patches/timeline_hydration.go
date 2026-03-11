package patches

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	structuredSectionHeaderRegex = regexp.MustCompile(`(?i)^\[\s*(general|items|heroes)\s*\]$`)
	structuredPrefixedLineRegex  = regexp.MustCompile(`^([^:]{1,64}):\s*(.*)$`)
	structuredDateHeadingRegex   = regexp.MustCompile(`(?i)^\d{2}-\d{2}-\d{4}\s+patch:\s*$`)
	structuredNonAlphaNumRegex   = regexp.MustCompile(`[^a-z0-9]+`)
	structuredSpaceRegex         = regexp.MustCompile(`\s+`)
)

type parseTemplateCatalog struct {
	itemsByNorm  map[string]PatchEntry
	heroesByNorm map[string]heroTemplate
}

type heroTemplate struct {
	name            string
	iconURL         string
	iconFallbackURL string
	abilities       []abilityTemplate
}

type abilityTemplate struct {
	name            string
	normName        string
	iconURL         string
	iconFallbackURL string
}

type timelineHeroState struct {
	entry       *PatchEntry
	abilities   []abilityTemplate
	groupsByKey map[string]*PatchEntryGroup
	groupOrder  []string
	currentSpecialGroup string
}

func hydratePatchDetail(detail PatchDetail) PatchDetail {
	if len(detail.Timeline) == 0 {
		detail.Timeline = []PatchTimelineBlock{synthesizeInitialTimelineBlock(detail)}
		return detail
	}

	detail.Timeline = hydrateTimelineBlocks(detail.Timeline, detail.Sections)
	return detail
}

func hydrateTimelineBlocks(blocks []PatchTimelineBlock, mergedSections []PatchSection) []PatchTimelineBlock {
	catalog := buildParseTemplateCatalog(mergedSections)
	hydrated := make([]PatchTimelineBlock, 0, len(blocks))
	seenBody := map[string]bool{}

	for _, block := range blocks {
		next := block
		if len(next.Sections) == 0 {
			next.Sections = buildBlockSectionsFromChanges(next, catalog)
		}
		if len(next.Changes) == 0 {
			next.Changes = flattenChangesFromSections(next.Sections, next.ID)
		}
		if strings.TrimSpace(next.Title) == "" {
			next.Title = timelineDefaultTitle(next.Kind, next.ReleasedAt)
		}
		bodySignature := blockBodySignature(next)
		if bodySignature != "" && seenBody[bodySignature] {
			continue
		}
		seenBody[bodySignature] = true
		hydrated = append(hydrated, next)
	}

	sort.SliceStable(hydrated, func(i, j int) bool {
		left := parseRFC3339(hydrated[i].ReleasedAt)
		right := parseRFC3339(hydrated[j].ReleasedAt)
		if left.Equal(right) {
			return hydrated[i].ID < hydrated[j].ID
		}
		if left.IsZero() {
			return true
		}
		if right.IsZero() {
			return false
		}
		return left.Before(right)
	})

	if len(hydrated) > 0 && hydrated[0].Kind != "initial" {
		hydrated[0].Kind = "initial"
		if strings.TrimSpace(hydrated[0].Title) == "" || strings.HasPrefix(strings.ToLower(hydrated[0].Title), "hotfix ") {
			hydrated[0].Title = "Initial Update"
		}
	}

	return hydrated
}

func synthesizeInitialTimelineBlock(detail PatchDetail) PatchTimelineBlock {
	releasedAt := detail.PublishedAt
	if strings.TrimSpace(releasedAt) == "" {
		releasedAt = time.Now().UTC().Format(time.RFC3339)
	}

	return PatchTimelineBlock{
		ID:         fmt.Sprintf("%s-initial", detail.Slug),
		Kind:       "initial",
		Title:      "Initial Update",
		ReleasedAt: releasedAt,
		Source:     detail.Source,
		Sections:   detail.Sections,
		Changes:    flattenChangesFromSections(detail.Sections, detail.Slug+"-initial"),
	}
}

func flattenChangesFromSections(sections []PatchSection, prefix string) []PatchChange {
	changes := make([]PatchChange, 0, 16)

	for _, section := range sections {
		for _, entry := range section.Entries {
			switch section.Kind {
			case "heroes":
				for _, change := range entry.Changes {
					changes = append(changes, PatchChange{
						ID:   fmt.Sprintf("%s-%d", prefix, len(changes)+1),
						Text: fmt.Sprintf("%s: %s", entry.EntityName, strings.TrimSpace(change.Text)),
					})
				}
				for _, group := range entry.Groups {
					for _, change := range group.Changes {
						line := strings.TrimSpace(change.Text)
						if strings.TrimSpace(group.Title) != "" {
							line = strings.TrimSpace(group.Title) + " " + line
						}
						changes = append(changes, PatchChange{
							ID:   fmt.Sprintf("%s-%d", prefix, len(changes)+1),
							Text: fmt.Sprintf("%s: %s", entry.EntityName, line),
						})
					}
				}
			case "items":
				for _, change := range entry.Changes {
					changes = append(changes, PatchChange{
						ID:   fmt.Sprintf("%s-%d", prefix, len(changes)+1),
						Text: fmt.Sprintf("%s: %s", entry.EntityName, strings.TrimSpace(change.Text)),
					})
				}
			default:
				for _, change := range entry.Changes {
					changes = append(changes, PatchChange{
						ID:   fmt.Sprintf("%s-%d", prefix, len(changes)+1),
						Text: strings.TrimSpace(change.Text),
					})
				}
			}
		}
	}

	if len(changes) == 0 {
		changes = append(changes, PatchChange{
			ID:   prefix + "-1",
			Text: "No line-item changes listed.",
		})
	}

	return changes
}

func timelineDefaultTitle(kind, releasedAt string) string {
	if kind == "initial" {
		return "Initial Update"
	}
	date := parseRFC3339(releasedAt)
	if date.IsZero() {
		return "Hotfix"
	}
	return "Hotfix " + date.UTC().Format("2006-01-02")
}

func blockBodySignature(block PatchTimelineBlock) string {
	parts := make([]string, 0, len(block.Changes))
	for _, change := range block.Changes {
		line := normalizeTimelineLine(change.Text)
		if line == "" {
			continue
		}
		parts = append(parts, line)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n")
}

func normalizeTimelineLine(line string) string {
	line = strings.TrimSpace(strings.ToLower(line))
	line = structuredSpaceRegex.ReplaceAllString(line, " ")
	return line
}

func normalizeLookupKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = structuredNonAlphaNumRegex.ReplaceAllString(value, " ")
	value = structuredSpaceRegex.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func slugifyLookup(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = structuredNonAlphaNumRegex.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "entry"
	}
	return value
}

func parseRFC3339(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}
