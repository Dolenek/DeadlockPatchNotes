package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

const DefaultSteamNewsURL = "https://api.steampowered.com/ISteamNews/GetNewsForApp/v2/?appid=1422450&count=100&maxlength=0&format=json"

var minorUpdateTitleRegex = regexp.MustCompile(`(?i)^minor update\s*-\s*\d{2}-\d{2}-\d{4}$`)

type steamNewsResponse struct {
	AppNews struct {
		NewsItems []steamNewsItem `json:"newsitems"`
	} `json:"appnews"`
}

type steamNewsItem struct {
	GID      string `json:"gid"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Contents string `json:"contents"`
	Date     int64  `json:"date"`
}

type SteamMinorUpdate struct {
	GID         string
	Title       string
	SourceURL   string
	BodyText    string
	PublishedAt time.Time
}

func FetchSteamMinorUpdates(ctx context.Context, client *http.Client, sourceURL string) ([]SteamMinorUpdate, error) {
	if strings.TrimSpace(sourceURL) == "" {
		sourceURL = DefaultSteamNewsURL
	}
	raw, err := fetchText(ctx, client, sourceURL)
	if err != nil {
		return nil, fmt.Errorf("fetch Steam news: %w", err)
	}
	return decodeSteamMinorUpdates(raw)
}

func decodeSteamMinorUpdates(raw string) ([]SteamMinorUpdate, error) {
	var response steamNewsResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return nil, fmt.Errorf("decode Steam news: %w", err)
	}
	updates := make([]SteamMinorUpdate, 0, len(response.AppNews.NewsItems))
	seenGIDs := make(map[string]bool, len(response.AppNews.NewsItems))
	for _, item := range response.AppNews.NewsItems {
		if seenGIDs[item.GID] || !minorUpdateTitleRegex.MatchString(strings.TrimSpace(item.Title)) {
			continue
		}
		body := normalizeSteamBody(item.Contents)
		if item.GID == "" || item.Date <= 0 || body == "" {
			continue
		}
		seenGIDs[item.GID] = true
		updates = append(updates, SteamMinorUpdate{
			GID:         item.GID,
			Title:       strings.TrimSpace(item.Title),
			SourceURL:   strings.TrimSpace(item.URL),
			BodyText:    body,
			PublishedAt: time.Unix(item.Date, 0).UTC(),
		})
	}
	sort.SliceStable(updates, func(i, j int) bool {
		return updates[i].PublishedAt.Before(updates[j].PublishedAt)
	})
	if len(updates) == 0 {
		return nil, errors.New("no Steam minor updates discovered")
	}
	return updates, nil
}
