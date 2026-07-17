package ingest

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRunPatchSyncRecordsSuccessfulRun(t *testing.T) {
	database, mock := newIngestSQLMock(t)
	expectSyncRunStart(mock, 9)
	expectSyncRunFinish(mock, 9, "success", "sync complete", SyncStats{DiscoveredThreads: 2, ProcessedThreads: 2, InsertedPatches: 1, UpdatedPatches: 1}, nil)
	dependencies := syncTestDependencies()
	dependencies.crawlChangelogThreads = func(context.Context, *http.Client, string, int) ([]ForumThreadRef, error) {
		return []ForumThreadRef{{URL: "one"}, {URL: "two"}}, nil
	}
	dependencies.syncDiscoveredThreads = func(_ context.Context, _ *sql.DB, _ *http.Client, _ *AssetCatalog, _ []ForumThreadRef, stats SyncStats) (SyncStats, []string) {
		stats.ProcessedThreads, stats.InsertedPatches, stats.UpdatedPatches = 2, 1, 1
		return stats, nil
	}

	stats, err := runPatchSync(t.Context(), database, http.DefaultClient, SyncConfig{}, dependencies)
	if err != nil || stats.ProcessedThreads != 2 || stats.FailedThreads != 0 {
		t.Fatalf("unexpected successful sync: stats=%+v err=%v", stats, err)
	}
	assertSQLExpectations(t, mock)
}

func TestRunPatchSyncRecordsPartialAndFailedRuns(t *testing.T) {
	testCases := []struct {
		name      string
		processed int
		failed    int
		status    string
	}{
		{name: "partial", processed: 1, failed: 1, status: "partial"},
		{name: "failed", processed: 0, failed: 2, status: "failed"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testSyncFailureStatus(t, testCase.processed, testCase.failed, testCase.status)
		})
	}
}

func testSyncFailureStatus(t *testing.T, processed, failed int, status string) {
	database, mock := newIngestSQLMock(t)
	stats := SyncStats{DiscoveredThreads: 2, ProcessedThreads: processed, FailedThreads: failed, InsertedPatches: processed}
	expectSyncRunStart(mock, 10)
	expectSyncRunFinish(mock, 10, status, "", stats, sqlmock.AnyArg())
	dependencies := syncTestDependencies()
	dependencies.crawlChangelogThreads = func(context.Context, *http.Client, string, int) ([]ForumThreadRef, error) {
		return []ForumThreadRef{{URL: "one"}, {URL: "two"}}, nil
	}
	dependencies.syncDiscoveredThreads = func(context.Context, *sql.DB, *http.Client, *AssetCatalog, []ForumThreadRef, SyncStats) (SyncStats, []string) {
		return stats, []string{"thread failed"}
	}

	actual, err := runPatchSync(t.Context(), database, http.DefaultClient, SyncConfig{}, dependencies)
	if err == nil || !strings.Contains(err.Error(), "thread failed") || actual != stats {
		t.Fatalf("unexpected failed sync: stats=%+v err=%v", actual, err)
	}
	assertSQLExpectations(t, mock)
}

func TestRunPatchSyncUsesSteamFallbackForUnavailableDiscovery(t *testing.T) {
	for _, crawlResult := range []struct {
		name string
		err  error
	}{
		{name: "forum error", err: errors.New("forum unavailable")},
		{name: "empty discovery"},
	} {
		t.Run(crawlResult.name, func(t *testing.T) {
			testSteamFallbackSuccess(t, crawlResult.err)
		})
	}
}

func testSteamFallbackSuccess(t *testing.T, crawlError error) {
	database, mock := newIngestSQLMock(t)
	expected := SyncStats{DiscoveredThreads: 3, ProcessedThreads: 1, UpdatedPatches: 1}
	expectSyncRunStart(mock, 11)
	expectSyncRunFinish(mock, 11, "success", "Steam fallback complete: discovered=3 added_blocks=2", expected, nil)
	dependencies := syncTestDependencies()
	dependencies.crawlChangelogThreads = func(context.Context, *http.Client, string, int) ([]ForumThreadRef, error) {
		return nil, crawlError
	}
	dependencies.syncLatestPatchFromSteamNews = func(context.Context, *sql.DB, *http.Client, *AssetCatalog, string) (steamFallbackResult, error) {
		return steamFallbackResult{DiscoveredNews: 3, AddedBlocks: 2}, nil
	}

	stats, err := runPatchSync(t.Context(), database, http.DefaultClient, SyncConfig{SteamNewsURL: "steam"}, dependencies)
	if err != nil || stats != expected {
		t.Fatalf("unexpected fallback result: stats=%+v err=%v", stats, err)
	}
	assertSQLExpectations(t, mock)
}

func TestRunPatchSyncReportsCatalogAndFallbackFailures(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		refs         []ForumThreadRef
		catalogError error
		steamError   error
		errorPart    string
		expected     SyncStats
		fallback     steamFallbackResult
	}{
		{name: "catalog", refs: []ForumThreadRef{{URL: "one"}}, catalogError: errors.New("bad assets"), errorPart: "load asset catalog", expected: SyncStats{DiscoveredThreads: 1}},
		{name: "fallback catalog", catalogError: errors.New("bad assets"), errorPart: "Steam fallback"},
		{name: "fallback fetch", steamError: errors.New("bad steam"), errorPart: "Steam fallback", fallback: steamFallbackResult{DiscoveredNews: 3}, expected: SyncStats{DiscoveredThreads: 3}},
	} {
		t.Run(testCase.name, func(t *testing.T) { testSyncDependencyFailure(t, testCase) })
	}
}

func testSyncDependencyFailure(t *testing.T, testCase struct {
	name                     string
	refs                     []ForumThreadRef
	catalogError, steamError error
	errorPart                string
	expected                 SyncStats
	fallback                 steamFallbackResult
}) {
	database, mock := newIngestSQLMock(t)
	expectSyncRunStart(mock, 12)
	expectSyncRunFinish(mock, 12, "failed", "", testCase.expected, sqlmock.AnyArg())
	dependencies := syncTestDependencies()
	dependencies.crawlChangelogThreads = func(context.Context, *http.Client, string, int) ([]ForumThreadRef, error) { return testCase.refs, nil }
	dependencies.loadAssetCatalog = func(context.Context, *http.Client) (*AssetCatalog, error) {
		return &AssetCatalog{}, testCase.catalogError
	}
	dependencies.syncLatestPatchFromSteamNews = func(context.Context, *sql.DB, *http.Client, *AssetCatalog, string) (steamFallbackResult, error) {
		return testCase.fallback, testCase.steamError
	}

	_, err := runPatchSync(t.Context(), database, http.DefaultClient, SyncConfig{}, dependencies)
	if err == nil || !strings.Contains(err.Error(), testCase.errorPart) {
		t.Fatalf("expected %q error, got %v", testCase.errorPart, err)
	}
	assertSQLExpectations(t, mock)
}

func TestFinalizeSyncRunJoinsOriginalAndPersistenceErrors(t *testing.T) {
	database, mock := newIngestSQLMock(t)
	original := errors.New("sync failed")
	finishFailure := errors.New("finish failed")
	mock.ExpectExec(`UPDATE sync_runs`).WillReturnError(finishFailure)

	_, err := finalizeSyncRun(t.Context(), database, 1, "failed", "message", SyncStats{}, original)
	if !errors.Is(err, original) || !errors.Is(err, finishFailure) {
		t.Fatalf("expected joined errors, got %v", err)
	}
	assertSQLExpectations(t, mock)
}

func syncTestDependencies() syncDependencies {
	return syncDependencies{
		crawlChangelogThreads: func(context.Context, *http.Client, string, int) ([]ForumThreadRef, error) { return nil, nil },
		loadAssetCatalog:      func(context.Context, *http.Client) (*AssetCatalog, error) { return &AssetCatalog{}, nil },
		syncDiscoveredThreads: func(_ context.Context, _ *sql.DB, _ *http.Client, _ *AssetCatalog, _ []ForumThreadRef, stats SyncStats) (SyncStats, []string) {
			return stats, nil
		},
		syncLatestPatchFromSteamNews: func(context.Context, *sql.DB, *http.Client, *AssetCatalog, string) (steamFallbackResult, error) {
			return steamFallbackResult{}, nil
		},
	}
}

func expectSyncRunStart(mock sqlmock.Sqlmock, runID int64) {
	mock.ExpectQuery(`INSERT INTO sync_runs`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(runID))
}

func expectSyncRunFinish(mock sqlmock.Sqlmock, runID int64, status, message string, stats SyncStats, messageMatcher interface{}) {
	messageArgument := interface{}(message)
	if messageMatcher != nil {
		messageArgument = messageMatcher
	}
	mock.ExpectExec(`UPDATE sync_runs`).WithArgs(
		runID, status, stats.DiscoveredThreads, stats.ProcessedThreads, stats.FailedThreads,
		stats.InsertedPatches, stats.UpdatedPatches, messageArgument,
	).WillReturnResult(sqlmock.NewResult(0, 1))
}
