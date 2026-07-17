package ingest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"deadlockpatchnotes/api/internal/patches"
	"github.com/DATA-DOG/go-sqlmock"
)

const (
	patchLookupPattern   = `(?s)SELECT\s+id, slug, title, category, intro, excerpt, hero_image_url,.*FROM patches\s+WHERE thread_id = \$1`
	timelineQueryPattern = `(?s)SELECT block_key, kind, title, source_type, source_url, post_id, released_at, body_text.*FROM patch_release_blocks`
)

func newIngestSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database, mock
}

func assertSQLExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func syncWriteFixture() (ForumThread, patchDetailRecord, []timelineCandidate, time.Time, time.Time) {
	published := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
	updated := published.Add(24 * time.Hour)
	thread := ForumThread{ThreadID: 42, Slug: "new-slug", Title: "Patch", URL: "https://forum.test/thread"}
	detail := patchDetailRecord{
		Payload: patches.PatchDetail{
			ID: "42", Slug: thread.Slug, Title: "Patch title", PublishedAt: published.Format(time.RFC3339),
			Category: "gameplay", Intro: "Intro", HeroImageURL: "https://images.test/patch.png",
			Source: patches.PatchSource{Type: "forum", URL: thread.URL},
		},
		Excerpt: "Short excerpt",
	}
	blocks := []timelineCandidate{
		{Key: "initial", Kind: "initial", Title: "Initial Update", SourceType: "forum-post", SourceURL: thread.URL + "/1", PostID: "1", ReleasedAt: published, BodyText: "First"},
		{Key: "hotfix", Kind: "hotfix", Title: "Hotfix", SourceType: "forum-post", SourceURL: thread.URL + "/2", PostID: "2", ReleasedAt: updated, BodyText: "Second"},
	}
	return thread, detail, blocks, published, updated
}

func expectTransactionAndLock(mock sqlmock.Sqlmock, threadID int64) {
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock`).WithArgs(threadID).WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectStoredPatch(mock sqlmock.Sqlmock, patchID int64, thread ForumThread, detail patchDetailRecord, published, updated time.Time, slug string) {
	detailRaw, _ := json.Marshal(detail.Payload)
	rows := sqlmock.NewRows([]string{
		"id", "slug", "title", "category", "intro", "excerpt", "hero_image_url",
		"published_at", "updated_at", "source_type", "source_url", "detail_payload",
	}).AddRow(
		patchID, slug, detail.Payload.Title, detail.Payload.Category, detail.Payload.Intro, detail.Excerpt,
		detail.Payload.HeroImageURL, published, updated, detail.Payload.Source.Type, detail.Payload.Source.URL, detailRaw,
	)
	mock.ExpectQuery(patchLookupPattern).WithArgs(thread.ThreadID).WillReturnRows(rows)
}

func expectStoredTimeline(mock sqlmock.Sqlmock, patchID int64, blocks []timelineCandidate) {
	rows := sqlmock.NewRows([]string{"block_key", "kind", "title", "source_type", "source_url", "post_id", "released_at", "body_text"})
	for _, block := range blocks {
		rows.AddRow(block.Key, block.Kind, block.Title, block.SourceType, block.SourceURL, block.PostID, block.ReleasedAt, block.BodyText)
	}
	mock.ExpectQuery(timelineQueryPattern).WithArgs(patchID).WillReturnRows(rows)
}

func expectReleaseBlockInsert(mock sqlmock.Sqlmock, patchID int64, index int, block timelineCandidate, insertError error) {
	expectation := mock.ExpectExec(`(?s)INSERT INTO patch_release_blocks`).WithArgs(
		patchID, block.Key, block.Kind, block.Title, block.SourceType, block.SourceURL, block.PostID,
		block.ReleasedAt, block.BodyText, hashText(block.BodyText), index+1,
	)
	if insertError != nil {
		expectation.WillReturnError(insertError)
		return
	}
	expectation.WillReturnResult(sqlmock.NewResult(int64(index+1), 1))
}

var errTestDatabase = errors.New("database unavailable")
