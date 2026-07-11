package ingest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSteamMinorUpdatesFiltersAndSortsPatchNews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"appnews":{"newsitems":[
				{"gid":"new","title":"Minor Update - 07-09-2026","url":"https://example.test/new","contents":"[p]- New change[/p]","date":1783625215},
				{"gid":"article","title":"Community spotlight","url":"https://example.test/article","contents":"[p]Ignore[/p]","date":1783500000},
				{"gid":"old","title":"Minor Update - 07-01-2026","url":"https://example.test/old","contents":"[p]- Old change[/p]","date":1782946499},
				{"gid":"new","title":"Minor Update - 07-09-2026","url":"https://example.test/duplicate","contents":"[p]- Duplicate[/p]","date":1783625215}
			]}
		}`))
	}))
	defer server.Close()

	updates, err := FetchSteamMinorUpdates(context.Background(), server.Client(), server.URL)
	if err != nil {
		t.Fatalf("fetch Steam minor updates: %v", err)
	}
	if len(updates) != 2 {
		t.Fatalf("expected 2 unique minor updates, got %d", len(updates))
	}
	if updates[0].GID != "old" || updates[1].GID != "new" {
		t.Fatalf("expected chronological updates, got %q then %q", updates[0].GID, updates[1].GID)
	}
	if updates[0].BodyText != "- Old change" {
		t.Fatalf("expected normalized Steam body, got %q", updates[0].BodyText)
	}
	if !updates[0].PublishedAt.Equal(time.Unix(1782946499, 0).UTC()) {
		t.Fatalf("unexpected publication time %s", updates[0].PublishedAt)
	}
}
