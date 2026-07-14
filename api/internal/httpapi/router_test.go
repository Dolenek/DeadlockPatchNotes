package httpapi

import (
	"context"
	"encoding/json"
	"errors"
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

func TestHealthzReturnsUnavailableWhenDependencyFails(t *testing.T) {
	h := NewRouter(patches.NewStore(), func(context.Context) error {
		return errors.New("database unavailable")
	})
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rr.Code)
	}
	var payload map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["status"] != "unavailable" {
		t.Fatalf("expected unavailable status, got %q", payload["status"])
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
	if !strings.Contains(body, "@scalar/api-reference@1.62.5") {
		t.Fatal("expected pinned Scalar dependency")
	}
	if !strings.Contains(body, "sha384-jVBCKhcCfx34USN27x4iQK1SBNdL/HxKq3KuBAxTS4WPaP5w80K4fjpwB+DezJL5") {
		t.Fatal("expected Scalar subresource integrity hash")
	}
	if !strings.Contains(body, "/api/scalar-init.js") {
		t.Fatal("expected local Scalar initialization script")
	}
}

func TestScalarInitScript(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/scalar-init.js", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "/api/openapi.json") {
		t.Fatalf("unexpected Scalar init response: status=%d body=%q", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/javascript") {
		t.Fatalf("expected JavaScript content type, got %q", rr.Header().Get("Content-Type"))
	}
}

func TestSecurityHeaders(t *testing.T) {
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	for _, header := range []string{
		"Content-Security-Policy",
		"Permissions-Policy",
		"Referrer-Policy",
		"Strict-Transport-Security",
		"X-Content-Type-Options",
		"X-Frame-Options",
	} {
		if rr.Header().Get(header) == "" {
			t.Errorf("expected %s security header", header)
		}
	}
}

func TestHTTPForwardedRequestRedirectsToCanonicalHTTPSHost(t *testing.T) {
	t.Setenv("SITE_URL", "https://www.deadlockpatchnotes.com")
	h := NewRouter(patches.NewStore())
	req := httptest.NewRequest(http.MethodGet, "http://attacker.example/api/v1/patches?page=2", nil)
	req.Header.Set("X-Forwarded-Proto", "http")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusPermanentRedirect {
		t.Fatalf("expected status 308, got %d", rr.Code)
	}
	if location := rr.Header().Get("Location"); location != "https://www.deadlockpatchnotes.com/api/v1/patches?page=2" {
		t.Fatalf("unexpected redirect target: %q", location)
	}
	if rr.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("expected redirect not to be cached")
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
	servers, ok := payload["servers"].([]any)
	if !ok || len(servers) == 0 {
		t.Fatal("expected servers array in openapi response")
	}
	firstServer, ok := servers[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first server object, got %#v", servers[0])
	}
	if firstServer["url"] != "https://www.deadlockpatchnotes.com/api" {
		t.Fatalf("expected production server url on www host, got %#v", firstServer["url"])
	}
	paths, ok := payload["paths"].(map[string]any)
	if !ok {
		t.Fatal("expected paths object in openapi response")
	}
	if _, exists := paths["/v1/patches"]; !exists {
		t.Fatal("expected /v1/patches path in openapi response")
	}
	if _, exists := paths["/v1/days-since-last-update"]; !exists {
		t.Fatal("expected /v1/days-since-last-update path in openapi response")
	}
	if _, exists := paths["/v1/patches/rss.xml"]; !exists {
		t.Fatal("expected /v1/patches/rss.xml path in openapi response")
	}
	if _, exists := paths["/v1/heroes/{heroSlug}/rss.xml"]; !exists {
		t.Fatal("expected /v1/heroes/{heroSlug}/rss.xml path in openapi response")
	}
	if _, exists := paths["/v1/heroes/{heroSlug}/days-without-update/rss.xml"]; !exists {
		t.Fatal("expected /v1/heroes/{heroSlug}/days-without-update/rss.xml path in openapi response")
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

	var payload patches.PatchListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Pagination.Page != 1 || payload.Pagination.PageSize != 1 {
		t.Fatalf("unexpected page data: %+v", payload)
	}
	if payload.Pagination.TotalItems < 1 {
		t.Fatalf("expected total >= 1, got %d", payload.Pagination.TotalItems)
	}
	if len(payload.Patches) != 1 {
		t.Fatalf("expected 1 list item, got %d", len(payload.Patches))
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

	var payload patches.PatchListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Pagination.Page != 1 {
		t.Fatalf("expected fallback page 1, got %d", payload.Pagination.Page)
	}
	if payload.Pagination.PageSize != 12 {
		t.Fatalf("expected fallback limit 12, got %d", payload.Pagination.PageSize)
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

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errorPayload, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured error payload, got %#v", payload["error"])
	}
	if errorPayload["code"] != "invalid_query_param" {
		t.Fatalf("expected invalid_query_param code, got %#v", errorPayload["code"])
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
