package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"deadlockpatchnotes/api/internal/patches"
)

func TestDaysSinceLastUpdateForLatestPatch(t *testing.T) {
	withFixedRSSNow(t, time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC))

	handler := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/days-since-last-update", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("expected text/plain content type, got %q", rr.Header().Get("Content-Type"))
	}
	assertDaysPrefixPayload(t, rr.Body.String())
}

func TestDaysSinceLastUpdateForHero(t *testing.T) {
	withFixedRSSNow(t, time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC))

	handler := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/days-since-last-update?hero=abrams", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	assertDaysPrefixPayload(t, rr.Body.String())
}

func TestDaysSinceLastUpdateMissingHeroReturns404(t *testing.T) {
	handler := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/days-since-last-update?hero=nope", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
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

func withFixedRSSNow(t *testing.T, fixed time.Time) {
	t.Helper()
	originalNow := rssNow
	rssNow = func() time.Time { return fixed }
	t.Cleanup(func() {
		rssNow = originalNow
	})
}

func assertDaysPrefixPayload(t *testing.T, payload string) {
	t.Helper()
	const prefix = "Days since last update: "
	if !strings.HasPrefix(payload, prefix) {
		t.Fatalf("expected payload prefix %q, got %q", prefix, payload)
	}
	daysRaw := strings.TrimSpace(strings.TrimPrefix(payload, prefix))
	days, err := strconv.Atoi(daysRaw)
	if err != nil {
		t.Fatalf("expected numeric day suffix, got %q (%v)", payload, err)
	}
	if days < 0 {
		t.Fatalf("expected non-negative days, got %d", days)
	}
}
