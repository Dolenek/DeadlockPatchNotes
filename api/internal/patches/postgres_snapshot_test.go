package patches

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

var patchRowColumns = []string{
	"thread_id", "slug", "title", "published_at", "category", "hero_image_url", "source_type", "source_url", "detail_payload",
}

func TestBuildSnapshotHydratesRelationalFallbackFields(t *testing.T) {
	database, mock := newPatchSQLMock(t)
	latestTime := time.Date(2026, time.July, 2, 12, 0, 0, 0, time.UTC)
	olderTime := latestTime.Add(-24 * time.Hour)
	latestDetail := PatchDetail{Timeline: []PatchTimelineBlock{{ID: "latest", Kind: "initial", ReleasedAt: latestTime.Format(time.RFC3339)}}}
	olderDetail := PatchDetail{Slug: "payload-old", Title: "Payload old", PublishedAt: olderTime.Format(time.RFC3339)}
	rows := sqlmock.NewRows(patchRowColumns).
		AddRow(2, "latest", "Latest row", latestTime, "gameplay", "latest.png", "forum", "https://forum/latest", mustJSON(t, latestDetail)).
		AddRow(1, "older", "Older row", olderTime, "items", "older.png", "steam", "https://steam/older", mustJSON(t, olderDetail))
	mock.ExpectQuery(`(?s)FROM patches\s+ORDER BY published_at DESC, slug DESC`).WillReturnRows(rows)

	snapshot, err := NewPostgresStore(database, time.Minute).buildSnapshot(t.Context())
	if err != nil {
		t.Fatalf("build snapshot: %v", err)
	}
	latest := snapshot.detailBySlug["latest"]
	if latest.ID != "2" || latest.Slug != "latest" || latest.Title != "Latest row" || latest.Category != "gameplay" {
		t.Fatalf("relational fallbacks were not applied: %+v", latest)
	}
	if latest.PublishedAt != latestTime.Format(time.RFC3339) || latest.Source.URL != "https://forum/latest" || latest.HeroImageURL != "latest.png" {
		t.Fatalf("time/source/image fallbacks were not applied: %+v", latest)
	}
	if len(snapshot.patchSummaries) != 2 || snapshot.patchSummaries[0].Slug != "latest" {
		t.Fatalf("unexpected summary order: %+v", snapshot.patchSummaries)
	}
	if len(snapshot.details) != 2 || snapshot.details[0].PublishedAt != olderTime.Format(time.RFC3339) {
		t.Fatalf("aggregate details were not chronological: %+v", snapshot.details)
	}
	assertPatchSQLExpectations(t, mock)
}

func TestBuildSnapshotReportsInvalidDetailJSON(t *testing.T) {
	database, mock := newPatchSQLMock(t)
	rows := sqlmock.NewRows(patchRowColumns).AddRow(
		1, "broken", "Broken", time.Now(), "gameplay", "image.png", "forum", "https://forum/broken", []byte("{"),
	)
	mock.ExpectQuery(`FROM patches`).WillReturnRows(rows)

	_, err := NewPostgresStore(database, time.Minute).buildSnapshot(t.Context())
	if err == nil || !strings.Contains(err.Error(), "decode patch detail broken") {
		t.Fatalf("expected detail decode error, got %v", err)
	}
	assertPatchSQLExpectations(t, mock)
}

func TestBuildSnapshotReportsScanAndIterationErrors(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		rows      *sqlmock.Rows
		errorPart string
	}{
		{name: "scan", rows: sqlmock.NewRows(patchRowColumns).AddRow(
			"not-an-int", "slug", "Title", time.Now(), "category", "image", "forum", "url", []byte(`{}`),
		), errorPart: "scan patch row"},
		{name: "iteration", rows: validPatchRows(t).RowError(0, errors.New("row stream failed")), errorPart: "row stream failed"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			database, mock := newPatchSQLMock(t)
			mock.ExpectQuery(`FROM patches`).WillReturnRows(testCase.rows)
			_, err := NewPostgresStore(database, time.Minute).buildSnapshot(t.Context())
			if err == nil || !strings.Contains(err.Error(), testCase.errorPart) {
				t.Fatalf("expected %q error, got %v", testCase.errorPart, err)
			}
			assertPatchSQLExpectations(t, mock)
		})
	}
}

func newPatchSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database, mock
}

func assertPatchSQLExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	return raw
}

func validPatchRows(t *testing.T) *sqlmock.Rows {
	t.Helper()
	return sqlmock.NewRows(patchRowColumns).AddRow(
		1, "slug", "Title", time.Now(), "category", "image", "forum", "url", mustJSON(t, PatchDetail{Slug: "slug"}),
	)
}
