package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"deadlockpatchnotes/api/internal/patches"
)

func TestHealthz(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", payload["status"])
	}
}

func TestScalarDocs(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/scalar", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected html content-type, got %q", rr.Header().Get("Content-Type"))
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Scalar.createApiReference") {
		t.Fatal("expected scalar initialization in html")
	}
	if !strings.Contains(body, "/api/openapi.json") {
		t.Fatal("expected scalar page to reference /api/openapi.json")
	}
}

func TestOpenAPISpec(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected json content-type, got %q", rr.Header().Get("Content-Type"))
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode openapi response: %v", err)
	}
	if payload["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi 3.1.0, got %v", payload["openapi"])
	}
	paths, ok := payload["paths"].(map[string]any)
	if !ok {
		t.Fatal("expected paths object in openapi response")
	}
	if _, exists := paths["/v1/patches"]; !exists {
		t.Fatal("expected /v1/patches path in openapi response")
	}
}

func TestListPatchesPagination(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patches?page=1&limit=1", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload patches.ListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Page != 1 || payload.Limit != 1 {
		t.Fatalf("unexpected page data: %+v", payload)
	}
	if payload.Total < 1 {
		t.Fatalf("expected total >= 1, got %d", payload.Total)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 list item, got %d", len(payload.Items))
	}
}

func TestListPatchesInvalidQueryFallsBack(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patches?page=oops&limit=bad", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload patches.ListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Page != 1 {
		t.Fatalf("expected fallback page 1, got %d", payload.Page)
	}
	if payload.Limit != 12 {
		t.Fatalf("expected fallback limit 12, got %d", payload.Limit)
	}
}

func TestGetPatchBySlug(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patches/2026-03-06-update", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload patches.PatchDetail
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Slug != "2026-03-06-update" {
		t.Fatalf("unexpected slug: %s", payload.Slug)
	}
	if len(payload.Sections) == 0 {
		t.Fatal("expected at least one section")
	}
}

func TestGetPatchMissingReturns404(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patches/nope", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestListHeroes(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/heroes", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload patches.HeroListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) == 0 {
		t.Fatal("expected at least one hero")
	}
}

func TestGetHeroChanges(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/heroes/abrams/changes?skill=Shoulder%20Charge", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload patches.HeroChangesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Hero.Slug != "abrams" {
		t.Fatalf("unexpected hero slug: %s", payload.Hero.Slug)
	}
	if len(payload.Items) == 0 {
		t.Fatal("expected hero timeline items")
	}
}

func TestGetHeroChangesInvalidDateReturns400(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/heroes/abrams/changes?from=not-a-date", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestGetHeroChangesMissingReturns404(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/heroes/nope/changes", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestListItems(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload patches.ItemListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) == 0 {
		t.Fatal("expected at least one item")
	}
}

func TestGetItemChanges(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/active-reload/changes", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload patches.ItemChangesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Item.Slug != "active-reload" {
		t.Fatalf("unexpected item slug: %s", payload.Item.Slug)
	}
	if len(payload.Items) == 0 {
		t.Fatal("expected item timeline items")
	}
}

func TestGetItemChangesInvalidDateReturns400(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/active-reload/changes?from=not-a-date", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestGetItemChangesMissingReturns404(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/nope/changes", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestListSpells(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/spells", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload patches.SpellListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) == 0 {
		t.Fatal("expected at least one spell")
	}
}

func TestGetSpellChanges(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/spells/shoulder-charge/changes", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload patches.SpellChangesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Spell.Slug != "shoulder-charge" {
		t.Fatalf("unexpected spell slug: %s", payload.Spell.Slug)
	}
	if len(payload.Items) == 0 {
		t.Fatal("expected spell timeline items")
	}
}

func TestGetSpellChangesInvalidDateReturns400(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/spells/shoulder-charge/changes?from=not-a-date", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestGetSpellChangesMissingReturns404(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/spells/nope/changes", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}
