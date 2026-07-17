package db

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestApplyMigrationsExecutesEmbeddedFilesInOrder(t *testing.T) {
	database, mock := newDBSQLMock(t)
	first := readTestMigration(t, "001_patchnotes.sql")
	second := readTestMigration(t, "002_sync_failure_tracking.sql")
	mock.ExpectExec(regexp.QuoteMeta(first)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(second)).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := ApplyMigrations(t.Context(), database); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	assertDBSQLExpectations(t, mock)
}

func TestApplyMigrationsStopsAtFirstExecutionFailure(t *testing.T) {
	database, mock := newDBSQLMock(t)
	first := readTestMigration(t, "001_patchnotes.sql")
	mock.ExpectExec(regexp.QuoteMeta(first)).WillReturnError(errors.New("migration failed"))

	err := ApplyMigrations(t.Context(), database)
	if err == nil || !strings.Contains(err.Error(), "execute migration 001_patchnotes.sql") {
		t.Fatalf("expected named migration error, got %v", err)
	}
	assertDBSQLExpectations(t, mock)
}

func readTestMigration(t *testing.T, name string) string {
	t.Helper()
	raw, err := migrationFS.ReadFile("migrations/" + name)
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return string(raw)
}
