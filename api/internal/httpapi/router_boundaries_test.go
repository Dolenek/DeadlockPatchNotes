package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"deadlockpatchnotes/api/internal/patches"
)

func TestListPatchesCapsLimitAndFallsBackForInvalidPages(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		target        string
		expectedPage  int
		expectedLimit int
	}{
		{name: "cap limit", target: "/api/v1/patches?page=2&limit=999", expectedPage: 2, expectedLimit: 50},
		{name: "not numeric", target: "/api/v1/patches?page=nope", expectedPage: 1, expectedLimit: 12},
		{name: "integer overflow", target: "/api/v1/patches?page=999999999999999999999999", expectedPage: 1, expectedLimit: 12},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &recordingListRepository{}
			response := serveTestRequest(NewRouter(repository), testCase.target)
			if response.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", response.Code)
			}
			if repository.page != testCase.expectedPage || repository.limit != testCase.expectedLimit {
				t.Fatalf("unexpected pagination: page=%d limit=%d", repository.page, repository.limit)
			}
		})
	}
}

func TestHeroDateQueriesUseInclusiveUTCDayBoundaries(t *testing.T) {
	repository := &recordingHeroRepository{}
	query := url.Values{
		"from":  {"2026-07-01"},
		"to":    {"2026-07-02"},
		"skill": {"  Power Slash  "},
	}
	response := serveTestRequest(NewRouter(repository), "/api/v1/heroes/doorman/changes?"+query.Encode())
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	expectedFrom := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2026, time.July, 3, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if repository.query.From == nil || !repository.query.From.Equal(expectedFrom) {
		t.Fatalf("unexpected from boundary: %v", repository.query.From)
	}
	if repository.query.To == nil || !repository.query.To.Equal(expectedTo) || repository.query.Skill != "Power Slash" {
		t.Fatalf("unexpected to/skill query: %+v", repository.query)
	}
}

func TestParseTimeQueryNormalizesRFC3339OffsetToUTC(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/?from="+url.QueryEscape("2026-07-01T01:30:00+02:00"), nil)
	parsed, err := parseTimeQuery(request, "from", true)
	if err != nil {
		t.Fatalf("parse RFC3339 query: %v", err)
	}
	expected := time.Date(2026, time.June, 30, 23, 30, 0, 0, time.UTC)
	if parsed == nil || !parsed.Equal(expected) || parsed.Location() != time.UTC {
		t.Fatalf("unexpected normalized timestamp: %v", parsed)
	}
}

func TestRepositoryFailureReturnsSanitizedRequestScopedError(t *testing.T) {
	repository := &recordingListRepository{listError: errors.New("secret database DSN")}
	response := serveTestRequest(NewRouter(repository), "/api/v1/patches")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.Code)
	}
	body := response.Body.String()
	if strings.Contains(body, "secret database DSN") || !strings.Contains(body, `"code":"internal_error"`) {
		t.Fatalf("unexpected error body: %s", body)
	}
	if !strings.Contains(body, `"message":"failed to list patches"`) || !strings.Contains(body, `"requestId":`) {
		t.Fatalf("missing sanitized message or request ID: %s", body)
	}
}

type recordingListRepository struct {
	patches.Repository
	page, limit int
	listError   error
}

func (repository *recordingListRepository) List(_ context.Context, page, limit int) (patches.PatchListResponse, error) {
	repository.page, repository.limit = page, limit
	return patches.PatchListResponse{}, repository.listError
}

type recordingHeroRepository struct {
	patches.Repository
	query patches.HeroChangesQuery
}

func (repository *recordingHeroRepository) GetHeroChanges(_ context.Context, query patches.HeroChangesQuery) (patches.HeroChangesResponse, error) {
	repository.query = query
	return patches.HeroChangesResponse{}, nil
}

func serveTestRequest(handler http.Handler, target string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
