package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"deadlockpatchnotes/api/internal/patches"
)

var (
	berlinLocationOnce sync.Once
	berlinLocation     *time.Location
	berlinLocationErr  error
)

type heroPatchGroup struct {
	PatchSlug   string
	PatchTitle  string
	PublishedAt time.Time
	Lines       []string
}

func listAllPatchSummaries(repository patches.Repository) ([]patches.PatchSummary, error) {
	firstPage, err := repository.List(1, 1)
	if err != nil {
		return nil, err
	}
	if firstPage.Pagination.TotalItems <= 0 {
		return nil, nil
	}

	fullPage, err := repository.List(1, firstPage.Pagination.TotalItems)
	if err != nil {
		return nil, err
	}
	return fullPage.Patches, nil
}

func preparePatchSummariesForFeed(summaries []patches.PatchSummary, maxItems int) []patches.PatchSummary {
	sort.SliceStable(summaries, func(i, j int) bool {
		left := parseTimeRFC3339(summaries[i].PublishedAt)
		right := parseTimeRFC3339(summaries[j].PublishedAt)
		if left.Equal(right) {
			return summaries[i].Slug > summaries[j].Slug
		}
		return left.After(right)
	})

	if maxItems > 0 && len(summaries) > maxItems {
		return summaries[:maxItems]
	}
	return summaries
}

func buildPatchRSSItems(store patches.Repository, summaries []patches.PatchSummary, siteBaseURL string, now time.Time) ([]rssItem, error) {
	items := make([]rssItem, 0, len(summaries))
	for _, patchSummary := range summaries {
		detail, err := store.GetBySlug(patchSummary.Slug)
		if err != nil {
			return nil, err
		}

		link := buildAbsoluteURL(siteBaseURL, "/patches/"+patchSummary.Slug)
		description := strings.TrimSpace(detail.Intro)
		if description == "" {
			description = patchSummary.Title
		}

		items = append(items, rssItem{
			Title:       patchSummary.Title,
			Link:        link,
			GUID:        rssGUID{IsPermaLink: true, Value: link},
			Description: description,
			PubDate:     formatRSSDate(patchSummary.PublishedAt, now),
		})
	}
	return items, nil
}

func buildHeroRSSItems(hero patches.HeroSummary, grouped []heroPatchGroup, siteBaseURL string) ([]rssItem, string, string) {
	heroName := strings.TrimSpace(hero.Name)
	if heroName == "" {
		heroName = hero.Slug
	}

	heroLink := buildAbsoluteURL(siteBaseURL, "/heroes/"+hero.Slug)
	items := make([]rssItem, 0, len(grouped))
	for _, group := range grouped {
		patchTitle := strings.TrimSpace(group.PatchTitle)
		if patchTitle == "" {
			patchTitle = group.PatchSlug
		}
		description := strings.Join(group.Lines, "\n")
		if strings.TrimSpace(description) == "" {
			description = "No line-item changes listed."
		}

		items = append(items, rssItem{
			Title:       fmt.Sprintf("%s - %s", heroName, patchTitle),
			Link:        heroLink,
			GUID:        rssGUID{IsPermaLink: false, Value: fmt.Sprintf("hero:%s:patch:%s", hero.Slug, group.PatchSlug)},
			Description: description,
			PubDate:     group.PublishedAt.UTC().Format(time.RFC1123Z),
		})
	}

	return items, heroLink, heroName
}

func buildHeroDaysWithoutUpdateRSSItem(hero patches.HeroSummary, daysWithoutUpdate int, lastChanged time.Time, location *time.Location, siteBaseURL string, now time.Time) (rssItem, string) {
	heroLink := buildAbsoluteURL(siteBaseURL, "/heroes/"+hero.Slug)
	return rssItem{
		Title:       fmt.Sprintf("Days since last update: %d", daysWithoutUpdate),
		Link:        heroLink,
		GUID:        rssGUID{IsPermaLink: false, Value: "hero-days-without-update:" + hero.Slug},
		Description: buildDaysWithoutUpdateDescription(hero.Name, lastChanged, location),
		PubDate:     now.Format(time.RFC1123Z),
	}, heroLink
}

func newHeroRSSDocument(heroName, heroLink, selfURL string, items []rssItem, now time.Time) rssDocument {
	return rssDocument{
		Version: "2.0",
		AtomNS:  "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:         "Deadlock Champion Updates - " + heroName,
			Link:          heroLink,
			Description:   "Patch-level update feed for " + heroName + ".",
			Language:      "en-us",
			LastBuildDate: now.Format(time.RFC1123Z),
			AtomLink: rssAtomLink{
				Href: selfURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}
}

func newHeroDaysWithoutUpdateRSSDocument(heroName, heroLink, selfURL string, item rssItem, now time.Time) rssDocument {
	return rssDocument{
		Version: "2.0",
		AtomNS:  "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:         "Days Without Update - " + heroName,
			Link:          heroLink,
			Description:   "Live streak counter feed.",
			Language:      "en-us",
			LastBuildDate: now.Format(time.RFC1123Z),
			AtomLink: rssAtomLink{
				Href: selfURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: []rssItem{item},
		},
	}
}

func formatRSSDate(raw string, fallback time.Time) string {
	parsed := parseTimeRFC3339(raw)
	if parsed.IsZero() {
		return fallback.UTC().Format(time.RFC1123Z)
	}
	return parsed.UTC().Format(time.RFC1123Z)
}

func parseTimeRFC3339(raw string) time.Time {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

func resolveFeedSiteBaseURL(r *http.Request) string {
	envSiteURL := normalizeSiteURL(os.Getenv("SITE_URL"))
	if envSiteURL != "" {
		return envSiteURL
	}
	return requestOrigin(r)
}

func normalizeSiteURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	if parsed.Host == "" {
		return ""
	}
	parsed.Path = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimSuffix(parsed.String(), "/")
}

func requestOrigin(r *http.Request) string {
	scheme := firstForwardedHeaderValue(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := firstForwardedHeaderValue(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}
	if host == "" {
		host = "localhost"
	}
	return scheme + "://" + host
}

func resolveRequestURL(r *http.Request) string {
	return requestOrigin(r) + r.URL.RequestURI()
}

func firstForwardedHeaderValue(raw string) string {
	if raw == "" {
		return ""
	}
	if comma := strings.Index(raw, ","); comma >= 0 {
		raw = raw[:comma]
	}
	return strings.TrimSpace(raw)
}

func buildAbsoluteURL(baseURL, path string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "https://www.deadlockpatchnotes.com"
	}
	normalizedPath := strings.TrimSpace(path)
	if !strings.HasPrefix(normalizedPath, "/") {
		normalizedPath = "/" + normalizedPath
	}
	return base + normalizedPath
}

func groupHeroTimelineByPatch(blocks []patches.HeroTimelineBlock) []heroPatchGroup {
	byPatch := make(map[string]*heroPatchGroup, len(blocks))
	order := make([]string, 0, len(blocks))

	for _, block := range blocks {
		patchSlug := strings.TrimSpace(block.Patch.Slug)
		if patchSlug == "" {
			continue
		}
		if _, seen := byPatch[patchSlug]; !seen {
			order = append(order, patchSlug)
		}
		upsertHeroPatchGroup(byPatch, patchSlug, block)
	}

	grouped := make([]heroPatchGroup, 0, len(order))
	for _, patchSlug := range order {
		group := byPatch[patchSlug]
		if group.PublishedAt.IsZero() {
			group.PublishedAt = rssNow().UTC()
		}
		grouped = append(grouped, *group)
	}

	sort.SliceStable(grouped, func(i, j int) bool {
		if grouped[i].PublishedAt.Equal(grouped[j].PublishedAt) {
			return grouped[i].PatchSlug > grouped[j].PatchSlug
		}
		return grouped[i].PublishedAt.After(grouped[j].PublishedAt)
	})

	return grouped
}

func upsertHeroPatchGroup(byPatch map[string]*heroPatchGroup, patchSlug string, block patches.HeroTimelineBlock) {
	group, ok := byPatch[patchSlug]
	if !ok {
		group = &heroPatchGroup{
			PatchSlug:   patchSlug,
			PatchTitle:  strings.TrimSpace(block.Patch.Title),
			PublishedAt: parseTimeRFC3339(block.ReleasedAt),
			Lines:       make([]string, 0, 8),
		}
		byPatch[patchSlug] = group
	}

	releasedAt := parseTimeRFC3339(block.ReleasedAt)
	if group.PublishedAt.IsZero() || (!releasedAt.IsZero() && releasedAt.Before(group.PublishedAt)) {
		group.PublishedAt = releasedAt
	}
	if group.PatchTitle == "" {
		group.PatchTitle = strings.TrimSpace(block.Patch.Title)
	}
	group.Lines = append(group.Lines, heroBlockLines(block)...)
}

func heroBlockLines(block patches.HeroTimelineBlock) []string {
	lines := make([]string, 0, len(block.GeneralChanges)+8)
	label := strings.TrimSpace(block.Label)
	if label == "" {
		label = strings.TrimSpace(block.ReleasedAt)
	}

	for _, change := range block.GeneralChanges {
		text := strings.TrimSpace(change.Text)
		if text == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("[%s] General: %s", label, text))
	}

	for _, skill := range block.Skills {
		skillTitle := strings.TrimSpace(skill.Title)
		if skillTitle == "" {
			skillTitle = "Skill"
		}
		for _, change := range skill.Changes {
			text := strings.TrimSpace(change.Text)
			if text == "" {
				continue
			}
			lines = append(lines, fmt.Sprintf("[%s] %s: %s", label, skillTitle, text))
		}
	}

	return lines
}

func findHeroBySlug(items []patches.HeroSummary, heroSlug string) (patches.HeroSummary, bool) {
	target := strings.TrimSpace(strings.ToLower(heroSlug))
	for _, item := range items {
		if strings.TrimSpace(strings.ToLower(item.Slug)) == target {
			return item, true
		}
	}
	return patches.HeroSummary{}, false
}

func resolveBerlinLocation() (*time.Location, error) {
	berlinLocationOnce.Do(func() {
		berlinLocation, berlinLocationErr = time.LoadLocation("Europe/Berlin")
	})
	return berlinLocation, berlinLocationErr
}

func daysSinceLastUpdate(lastChanged, now time.Time, location *time.Location) int {
	if lastChanged.IsZero() || location == nil {
		return 0
	}
	lastLocal := lastChanged.In(location)
	nowLocal := now.In(location)
	if !nowLocal.After(lastLocal) {
		return 0
	}

	checkpoint := nextNoonCheckpoint(lastLocal, location)
	days := 0
	for !checkpoint.After(nowLocal) {
		days++
		checkpoint = checkpoint.AddDate(0, 0, 1)
	}
	return days
}

func nextNoonCheckpoint(from time.Time, location *time.Location) time.Time {
	candidate := time.Date(from.Year(), from.Month(), from.Day(), 12, 0, 0, 0, location)
	if !candidate.After(from) {
		candidate = candidate.AddDate(0, 0, 1)
	}
	return candidate
}

func buildDaysWithoutUpdateDescription(heroName string, lastChanged time.Time, location *time.Location) string {
	if lastChanged.IsZero() {
		return heroName + " has no recorded update timestamp."
	}
	localLastChanged := lastChanged.In(location)
	return fmt.Sprintf("Last recorded update for %s: %s", heroName, localLastChanged.Format(time.RFC1123))
}
