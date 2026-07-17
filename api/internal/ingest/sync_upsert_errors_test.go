package ingest

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpsertPatchRollsBackAfterLockFailure(t *testing.T) {
	database, mock := newIngestSQLMock(t)
	thread, detail, blocks, published, updated := syncWriteFixture()
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock`).WithArgs(thread.ThreadID).WillReturnError(errTestDatabase)
	mock.ExpectRollback()

	result, err := upsertPatch(t.Context(), database, thread, detail, blocks, published, updated)
	assertFailedUpsert(t, result, err, "lock patch thread")
	assertSQLExpectations(t, mock)
}

func TestUpsertPatchRollsBackAfterLookupFailure(t *testing.T) {
	database, mock := newIngestSQLMock(t)
	thread, detail, blocks, published, updated := syncWriteFixture()
	expectTransactionAndLock(mock, thread.ThreadID)
	mock.ExpectQuery(patchLookupPattern).WithArgs(thread.ThreadID).WillReturnError(errTestDatabase)
	mock.ExpectRollback()

	result, err := upsertPatch(t.Context(), database, thread, detail, blocks, published, updated)
	assertFailedUpsert(t, result, err, "database unavailable")
	assertSQLExpectations(t, mock)
}

func TestUpsertPatchRollsBackAfterReleaseBlockFailure(t *testing.T) {
	database, mock := newIngestSQLMock(t)
	thread, detail, blocks, published, updated := syncWriteFixture()
	expectNewPatchBeforeReleaseBlocks(mock, thread)
	expectReleaseBlockInsert(mock, 77, 0, blocks[0], errTestDatabase)
	mock.ExpectRollback()

	result, err := upsertPatch(t.Context(), database, thread, detail, blocks, published, updated)
	assertFailedUpsert(t, result, err, "database unavailable")
	assertSQLExpectations(t, mock)
}

func TestUpsertPatchClearsWriteResultAfterCommitFailure(t *testing.T) {
	database, mock := newIngestSQLMock(t)
	thread, detail, blocks, published, updated := syncWriteFixture()
	expectNewPatchBeforeReleaseBlocks(mock, thread)
	for index, block := range blocks {
		expectReleaseBlockInsert(mock, 77, index, block, nil)
	}
	mock.ExpectCommit().WillReturnError(errTestDatabase)

	result, err := upsertPatch(t.Context(), database, thread, detail, blocks, published, updated)
	assertFailedUpsert(t, result, err, "database unavailable")
	assertSQLExpectations(t, mock)
}

func expectNewPatchBeforeReleaseBlocks(mock sqlmock.Sqlmock, thread ForumThread) {
	expectTransactionAndLock(mock, thread.ThreadID)
	mock.ExpectQuery(patchLookupPattern).WithArgs(thread.ThreadID).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`(?s)INSERT INTO patches.*RETURNING id`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(77))
	mock.ExpectExec(`DELETE FROM patch_release_blocks`).WithArgs(int64(77)).WillReturnResult(sqlmock.NewResult(0, 0))
}

func assertFailedUpsert(t *testing.T, result patchWriteResult, err error, messagePart string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), messagePart) {
		t.Fatalf("expected error containing %q, got %v", messagePart, err)
	}
	if result.Inserted || result.Updated {
		t.Fatalf("failed write returned modification counters: %+v", result)
	}
}
