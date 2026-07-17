package ingest

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpsertPatchInsertsPatchAndOrderedReleaseBlocks(t *testing.T) {
	database, mock := newIngestSQLMock(t)
	thread, detail, blocks, published, updated := syncWriteFixture()

	expectTransactionAndLock(mock, thread.ThreadID)
	mock.ExpectQuery(patchLookupPattern).WithArgs(thread.ThreadID).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`(?s)INSERT INTO patches.*RETURNING id`).WithArgs(
		thread.ThreadID, thread.Slug, detail.Payload.Title, detail.Payload.Category, detail.Payload.Intro,
		detail.Excerpt, detail.Payload.HeroImageURL, published, updated, detail.Payload.Source.Type,
		detail.Payload.Source.URL, sqlmock.AnyArg(),
	).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(77))
	mock.ExpectExec(`DELETE FROM patch_release_blocks`).WithArgs(int64(77)).WillReturnResult(sqlmock.NewResult(0, 0))
	for index, block := range blocks {
		expectReleaseBlockInsert(mock, 77, index, block, nil)
	}
	mock.ExpectCommit()

	result, err := upsertPatch(t.Context(), database, thread, detail, blocks, published, updated)
	if err != nil {
		t.Fatalf("upsert patch: %v", err)
	}
	if !result.Inserted || result.Updated {
		t.Fatalf("unexpected write result: %+v", result)
	}
	assertSQLExpectations(t, mock)
}

func TestUpsertPatchLeavesUnchangedReleaseBlocksIntact(t *testing.T) {
	database, mock := newIngestSQLMock(t)
	thread, detail, blocks, published, updated := syncWriteFixture()

	expectTransactionAndLock(mock, thread.ThreadID)
	expectStoredPatch(mock, 77, thread, detail, published, updated, thread.Slug)
	expectStoredTimeline(mock, 77, blocks)
	mock.ExpectExec(`UPDATE patches SET last_synced_at`).WithArgs(int64(77)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := upsertPatch(t.Context(), database, thread, detail, blocks, published, updated)
	if err != nil {
		t.Fatalf("upsert unchanged patch: %v", err)
	}
	if result.Inserted || result.Updated {
		t.Fatalf("unchanged patch was reported as modified: %+v", result)
	}
	assertSQLExpectations(t, mock)
}

func TestUpsertPatchFindsStableThreadIDWhenSlugChanges(t *testing.T) {
	database, mock := newIngestSQLMock(t)
	thread, detail, blocks, published, updated := syncWriteFixture()

	expectTransactionAndLock(mock, thread.ThreadID)
	expectStoredPatch(mock, 77, thread, detail, published, updated, "old-slug")
	expectStoredTimeline(mock, 77, blocks)
	mock.ExpectExec(`(?s)UPDATE patches\s+SET\s+slug = \$2`).WithArgs(
		int64(77), thread.Slug, detail.Payload.Title, detail.Payload.Category, detail.Payload.Intro,
		detail.Excerpt, detail.Payload.HeroImageURL, published, updated, detail.Payload.Source.Type,
		detail.Payload.Source.URL, sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM patch_release_blocks`).WithArgs(int64(77)).WillReturnResult(sqlmock.NewResult(0, 2))
	for index, block := range blocks {
		expectReleaseBlockInsert(mock, 77, index, block, nil)
	}
	mock.ExpectCommit()

	result, err := upsertPatch(t.Context(), database, thread, detail, blocks, published, updated)
	if err != nil {
		t.Fatalf("upsert renamed patch: %v", err)
	}
	if result.Inserted || !result.Updated {
		t.Fatalf("slug change was not reported as update: %+v", result)
	}
	assertSQLExpectations(t, mock)
}
