package ingest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCrawlChangelogThreads_RejectsChallengePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<a href="/.stile/challenge?rung=nojs">Challenge</a>`))
	}))
	defer server.Close()

	_, err := CrawlChangelogThreads(context.Background(), server.Client(), server.URL, 1)
	if err == nil || !strings.Contains(err.Error(), "challenge page") {
		t.Fatalf("expected challenge-page error, got %v", err)
	}
}

func TestCrawlChangelogThreads_RejectsCrossOriginNextPage(t *testing.T) {
	other := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer other.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<a rel="next" href="` + other.URL + `/page/2">Next</a>`))
	}))
	defer server.Close()

	_, err := CrawlChangelogThreads(context.Background(), server.Client(), server.URL, 2)
	if err == nil || !strings.Contains(err.Error(), "cross-origin") {
		t.Fatalf("expected cross-origin error, got %v", err)
	}
}

func TestFetchText_RejectsCrossOriginRedirect(t *testing.T) {
	destinationReached := false
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		destinationReached = true
		_, _ = w.Write([]byte("unexpected"))
	}))
	defer destination.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusFound)
	}))
	defer source.Close()

	_, err := fetchText(context.Background(), source.Client(), source.URL)
	if err == nil || !strings.Contains(err.Error(), "cross-origin redirect") {
		t.Fatalf("expected cross-origin redirect error, got %v", err)
	}
	if destinationReached {
		t.Fatal("cross-origin redirect reached its destination")
	}
}

func TestFetchText_RejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", int(maxSourceResponseBytes)+1)))
	}))
	defer server.Close()

	_, err := fetchText(context.Background(), server.Client(), server.URL)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected response-size error, got %v", err)
	}
}
