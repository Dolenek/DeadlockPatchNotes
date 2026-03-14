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

func TestDaysSinceLastUpdateHeroBaselineDefaultsToMostRecentPatchOrHero(t *testing.T) {
	now := time.Date(2026, time.March, 14, 12, 0, 0, 0, time.UTC)
	withFixedRSSNow(t, now)
	location := mustLoadBerlinTestLocation(t)

	store := &daysSinceRepoStub{
		patchPublishedAt: "2026-03-13T12:00:00Z",
		heroLastChangedAt: "2026-03-01T12:00:00Z",
	}
	handler := NewRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/days-since-last-update?hero=abrams", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	got := parseDaysPayload(t, rr.Body.String())
	expected := daysSinceLastUpdate(parseTimeRFC3339(store.patchPublishedAt), now, location)
	if got != expected {
		t.Fatalf("expected days=%d from latest patch baseline, got %d", expected, got)
	}
}

func TestDaysSinceLastUpdateHeroOnlyUpdateUsesHeroTimestamp(t *testing.T) {
	now := time.Date(2026, time.March, 14, 12, 0, 0, 0, time.UTC)
	withFixedRSSNow(t, now)
	location := mustLoadBerlinTestLocation(t)

	store := &daysSinceRepoStub{
		patchPublishedAt: "2026-03-13T12:00:00Z",
		heroLastChangedAt: "2026-03-01T12:00:00Z",
	}
	handler := NewRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/days-since-last-update?hero=abrams&onlyUpdate=true", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	got := parseDaysPayload(t, rr.Body.String())
	expected := daysSinceLastUpdate(parseTimeRFC3339(store.heroLastChangedAt), now, location)
	if got != expected {
		t.Fatalf("expected days=%d from hero-only baseline, got %d", expected, got)
	}
}

func TestDaysSinceLastUpdateInvalidOnlyUpdateReturns400(t *testing.T) {
	handler := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/days-since-last-update?onlyUpdate=not-bool", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
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

type daysSinceRepoStub struct {
	patchPublishedAt string
	heroLastChangedAt string
}

func (s *daysSinceRepoStub) List(page, limit int) (patches.PatchListResponse, error) {
	return patches.PatchListResponse{
		Patches: []patches.PatchSummary{
			{
				Slug:        "latest-patch",
				PublishedAt: s.patchPublishedAt,
			},
		},
		Pagination: patches.Pagination{
			Page:       1,
			PageSize:   1,
			TotalItems: 1,
			TotalPages: 1,
		},
	}, nil
}

func (s *daysSinceRepoStub) GetBySlug(string) (patches.PatchDetail, error) {
	return patches.PatchDetail{}, nil
}

func (s *daysSinceRepoStub) ListHeroes() (patches.HeroListResponse, error) {
	return patches.HeroListResponse{
		Items: []patches.HeroSummary{
			{
				Slug:          "abrams",
				Name:          "Abrams",
				LastChangedAt: s.heroLastChangedAt,
			},
		},
	}, nil
}

func (s *daysSinceRepoStub) GetHeroChanges(patches.HeroChangesQuery) (patches.HeroChangesResponse, error) {
	return patches.HeroChangesResponse{}, nil
}

func (s *daysSinceRepoStub) ListItems() (patches.ItemListResponse, error) {
	return patches.ItemListResponse{}, nil
}

func (s *daysSinceRepoStub) GetItemChanges(patches.ItemChangesQuery) (patches.ItemChangesResponse, error) {
	return patches.ItemChangesResponse{}, nil
}

func (s *daysSinceRepoStub) ListSpells() (patches.SpellListResponse, error) {
	return patches.SpellListResponse{}, nil
}

func (s *daysSinceRepoStub) GetSpellChanges(patches.SpellChangesQuery) (patches.SpellChangesResponse, error) {
	return patches.SpellChangesResponse{}, nil
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
	_ = parseDaysPayload(t, payload)
}

func parseDaysPayload(t *testing.T, payload string) int {
	t.Helper()
	const prefix = "Days since last update: "
	daysRaw := strings.TrimSpace(strings.TrimPrefix(payload, prefix))
	days, err := strconv.Atoi(daysRaw)
	if err != nil {
		t.Fatalf("expected numeric day suffix, got %q (%v)", payload, err)
	}
	if days < 0 {
		t.Fatalf("expected non-negative days, got %d", days)
	}
	return days
}
