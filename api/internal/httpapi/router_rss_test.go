package httpapi

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"deadlockpatchnotes/api/internal/patches"
)

type rssTestDocument struct {
	Channel rssTestChannel `xml:"channel"`
}

type rssTestChannel struct {
	Title string        `xml:"title"`
	Link  string        `xml:"link"`
	Items []rssTestItem `xml:"item"`
}

type rssTestItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func TestPatchFeedRSS(t *testing.T) {
	t.Setenv("SITE_URL", "https://www.deadlockpatchnotes.com")

	handler := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patches/rss.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/rss+xml") {
		t.Fatalf("expected rss content type, got %q", rr.Header().Get("Content-Type"))
	}

	feed := decodeRSSTestDocument(t, rr.Body.Bytes())
	if feed.Channel.Title == "" {
		t.Fatal("expected channel title")
	}
	if len(feed.Channel.Items) == 0 {
		t.Fatal("expected at least one feed item")
	}
	if !strings.Contains(feed.Channel.Items[0].Link, "/patches/") {
		t.Fatalf("expected patch detail link, got %q", feed.Channel.Items[0].Link)
	}
	if _, err := time.Parse(time.RFC1123Z, feed.Channel.Items[0].PubDate); err != nil {
		t.Fatalf("expected RFC1123Z pubDate, got %q (%v)", feed.Channel.Items[0].PubDate, err)
	}
}

func TestHeroFeedRSS(t *testing.T) {
	t.Setenv("SITE_URL", "https://www.deadlockpatchnotes.com")

	handler := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/heroes/abrams/rss.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/rss+xml") {
		t.Fatalf("expected rss content type, got %q", rr.Header().Get("Content-Type"))
	}

	feed := decodeRSSTestDocument(t, rr.Body.Bytes())
	if len(feed.Channel.Items) == 0 {
		t.Fatal("expected at least one hero feed item")
	}

	seenGUIDs := map[string]bool{}
	for _, item := range feed.Channel.Items {
		if !strings.HasSuffix(item.Link, "/heroes/abrams") {
			t.Fatalf("expected hero page link, got %q", item.Link)
		}
		if strings.TrimSpace(item.Description) == "" {
			t.Fatalf("expected non-empty hero feed description for item %q", item.Title)
		}
		if seenGUIDs[item.GUID] {
			t.Fatalf("expected unique GUID per patch item, duplicate %q", item.GUID)
		}
		seenGUIDs[item.GUID] = true
	}
}

func TestHeroFeedRSSMissingReturns404(t *testing.T) {
	handler := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/heroes/nope/rss.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected json content type, got %q", rr.Header().Get("Content-Type"))
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errorPayload, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured error payload, got %#v", payload["error"])
	}
	if errorPayload["code"] != "resource_not_found" {
		t.Fatalf("expected resource_not_found code, got %#v", errorPayload["code"])
	}
}

func TestHeroDaysWithoutUpdateRSS(t *testing.T) {
	t.Setenv("SITE_URL", "https://www.deadlockpatchnotes.com")

	originalNow := rssNow
	rssNow = func() time.Time {
		return time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	}
	t.Cleanup(func() {
		rssNow = originalNow
	})

	handler := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/heroes/abrams/days-without-update/rss.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/rss+xml") {
		t.Fatalf("expected rss content type, got %q", rr.Header().Get("Content-Type"))
	}

	feed := decodeRSSTestDocument(t, rr.Body.Bytes())
	if len(feed.Channel.Items) != 1 {
		t.Fatalf("expected exactly one streak item, got %d", len(feed.Channel.Items))
	}
	item := feed.Channel.Items[0]
	if !strings.Contains(item.Title, "Days since last update:") {
		t.Fatalf("unexpected streak item title: %q", item.Title)
	}
	if !strings.HasSuffix(item.Link, "/heroes/abrams") {
		t.Fatalf("expected hero page link, got %q", item.Link)
	}
	if _, err := time.Parse(time.RFC1123Z, item.PubDate); err != nil {
		t.Fatalf("expected RFC1123Z pubDate, got %q (%v)", item.PubDate, err)
	}
}

func TestHeroDaysWithoutUpdateRSSMissingReturns404(t *testing.T) {
	handler := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/heroes/nope/days-without-update/rss.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func decodeRSSTestDocument(t *testing.T, raw []byte) rssTestDocument {
	t.Helper()
	var doc rssTestDocument
	if err := xml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode rss payload: %v", err)
	}
	return doc
}
