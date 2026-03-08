package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	partnerEventRegex = regexp.MustCompile(`data-partnereventstore="([^"]+)"`)
	ogImageRegex      = regexp.MustCompile(`<meta\s+property="og:image"\s+content="([^"]+)"`)
	steamPatchRegex   = regexp.MustCompile(`(?i)^(\d{2}-\d{2}-\d{4})\s+Patch:\s*$`)
)

type steamEventEnvelope struct {
	GID              string            `json:"gid"`
	EventName        string            `json:"event_name"`
	JSONData         string            `json:"jsondata"`
	AnnouncementBody steamAnnouncement `json:"announcement_body"`
}

type steamAnnouncement struct {
	Headline string `json:"headline"`
	PostTime int64  `json:"posttime"`
	Body     string `json:"body"`
}

type steamJSONData struct {
	LocalizedCapsuleImage []*string `json:"localized_capsule_image"`
}

type SteamEvent struct {
	Title      string
	BodyText   string
	Published  time.Time
	SourceURL  string
	HeroImage  string
	BodyBlocks []SteamBodyBlock
}

type SteamBodyBlock struct {
	Kind       string
	Title      string
	BodyText   string
	ReleasedAt time.Time
}

func FetchSteamEvent(ctx context.Context, client *http.Client, steamURL string, fallbackTime time.Time) (*SteamEvent, error) {
	raw, err := fetchText(ctx, client, steamURL)
	if err != nil {
		return nil, err
	}

	match := partnerEventRegex.FindStringSubmatch(raw)
	if len(match) < 2 {
		return nil, fmt.Errorf("steam metadata missing from %s", steamURL)
	}

	decoded := html.UnescapeString(match[1])
	var events []steamEventEnvelope
	if err := json.Unmarshal([]byte(decoded), &events); err != nil {
		return nil, fmt.Errorf("decode steam event payload: %w", err)
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("empty steam event payload: %s", steamURL)
	}

	event := events[0]
	published := fallbackTime
	if event.AnnouncementBody.PostTime > 0 {
		published = time.Unix(event.AnnouncementBody.PostTime, 0).UTC()
	}

	heroImage := parseSteamImage(raw)
	if heroImage == "" {
		heroImage = parseCapsuleImage(event.JSONData)
	}

	normalizedBody := normalizeSteamBody(event.AnnouncementBody.Body)
	blocks := splitSteamBodyBlocks(normalizedBody, published)

	return &SteamEvent{
		Title:      firstNonEmpty(event.AnnouncementBody.Headline, event.EventName),
		BodyText:   normalizedBody,
		Published:  published,
		SourceURL:  steamURL,
		HeroImage:  heroImage,
		BodyBlocks: blocks,
	}, nil
}

func normalizeSteamBody(value string) string {
	if value == "" {
		return ""
	}

	cleaned := value
	cleaned = strings.ReplaceAll(cleaned, "\\r", "")
	cleaned = strings.ReplaceAll(cleaned, "\\[", "[")
	cleaned = strings.ReplaceAll(cleaned, "\\]", "]")
	cleaned = strings.ReplaceAll(cleaned, "[p][/p]", "\n")
	cleaned = strings.ReplaceAll(cleaned, "[p]", "")
	cleaned = strings.ReplaceAll(cleaned, "[/p]", "\n")
	cleaned = strings.ReplaceAll(cleaned, "[h3]", "\n")
	cleaned = strings.ReplaceAll(cleaned, "[/h3]", "\n")
	cleaned = strings.ReplaceAll(cleaned, "[b]", "")
	cleaned = strings.ReplaceAll(cleaned, "[/b]", "")
	cleaned = strings.ReplaceAll(cleaned, "[u]", "")
	cleaned = strings.ReplaceAll(cleaned, "[/u]", "")
	cleaned = strings.ReplaceAll(cleaned, "\u00a0", " ")

	imageTagRegex := regexp.MustCompile(`\[img\].*?\[/img\]`)
	cleaned = imageTagRegex.ReplaceAllString(cleaned, "\n")

	lines := strings.Split(cleaned, "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = cleanLine(line)
		if line == "" {
			normalized = append(normalized, "")
			continue
		}
		normalized = append(normalized, line)
	}

	return compactLines(normalized)
}

func splitSteamBodyBlocks(body string, defaultTime time.Time) []SteamBodyBlock {
	if body == "" {
		return nil
	}

	lines := strings.Split(body, "\n")
	blocks := make([]SteamBodyBlock, 0, 4)
	currentTitle := "Initial Update"
	currentKind := "initial"
	currentDate := defaultTime
	currentLines := make([]string, 0, len(lines))

	flush := func() {
		text := compactLines(currentLines)
		if text == "" {
			return
		}
		blocks = append(blocks, SteamBodyBlock{
			Kind:       currentKind,
			Title:      currentTitle,
			BodyText:   text,
			ReleasedAt: currentDate,
		})
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		match := steamPatchRegex.FindStringSubmatch(trimmed)
		if len(match) == 2 {
			flush()
			parsed, err := time.Parse("01-02-2006", match[1])
			if err != nil {
				currentDate = defaultTime
			} else {
				currentDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 12, 0, 0, 0, time.UTC)
			}
			currentKind = "hotfix"
			currentTitle = fmt.Sprintf("Hotfix %s", parsedDateLabel(currentDate))
			currentLines = currentLines[:0]
			continue
		}
		currentLines = append(currentLines, line)
	}
	flush()

	if len(blocks) == 0 {
		blocks = append(blocks, SteamBodyBlock{
			Kind:       "initial",
			Title:      "Initial Update",
			BodyText:   body,
			ReleasedAt: defaultTime,
		})
	}

	return blocks
}

func parseSteamImage(raw string) string {
	match := ogImageRegex.FindStringSubmatch(raw)
	if len(match) < 2 {
		return ""
	}
	return html.UnescapeString(strings.TrimSpace(match[1]))
}

func parseCapsuleImage(jsonData string) string {
	if strings.TrimSpace(jsonData) == "" {
		return ""
	}
	var capsule steamJSONData
	if err := json.Unmarshal([]byte(jsonData), &capsule); err != nil {
		return ""
	}
	for _, candidate := range capsule.LocalizedCapsuleImage {
		if candidate == nil || strings.TrimSpace(*candidate) == "" {
			continue
		}
		return "https://clan.fastly.steamstatic.com/images/45164767/" + strings.TrimSpace(*candidate)
	}
	return ""
}

func parsedDateLabel(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.UTC().Format("2006-01-02")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
