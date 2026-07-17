package db

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var roleTestPasswords = RuntimeRolePasswords{
	API:  "api-password-is-long-enough",
	Sync: "sync-password-is-long-enough",
}

func TestConfigureRuntimeRolesCommitsLeastPrivilegeRoles(t *testing.T) {
	database, mock := newDBSQLMock(t)
	expectRoleTransactionSetup(mock, `deadlock"prod`)
	expectRuntimeRole(mock, `deadlock"prod`, runtimeRoleSpec{
		name: APIRoleName, password: roleTestPasswords.API, connectionLimit: 20,
	}, false, nil)
	expectRuntimeRole(mock, `deadlock"prod`, runtimeRoleSpec{
		name: SyncRoleName, password: roleTestPasswords.Sync, connectionLimit: 5, writeAccess: true,
	}, true, nil)
	mock.ExpectCommit()

	if err := ConfigureRuntimeRoles(t.Context(), database, roleTestPasswords); err != nil {
		t.Fatalf("configure runtime roles: %v", err)
	}
	assertDBSQLExpectations(t, mock)
}

func TestConfigureRuntimeRolesRollsBackOnPrivilegeFailure(t *testing.T) {
	database, mock := newDBSQLMock(t)
	expectRoleTransactionSetup(mock, "deadlock")
	expectRuntimeRole(mock, "deadlock", runtimeRoleSpec{
		name: APIRoleName, password: roleTestPasswords.API, connectionLimit: 20,
	}, false, errors.New("grant denied"))
	mock.ExpectRollback()

	err := ConfigureRuntimeRoles(t.Context(), database, roleTestPasswords)
	if err == nil || !strings.Contains(err.Error(), "grant privileges") {
		t.Fatalf("expected privilege error, got %v", err)
	}
	assertDBSQLExpectations(t, mock)
}

func TestConfigureRuntimeRolesReportsCommitFailure(t *testing.T) {
	database, mock := newDBSQLMock(t)
	expectRoleTransactionSetup(mock, "deadlock")
	expectRuntimeRole(mock, "deadlock", runtimeRoleSpec{name: APIRoleName, password: roleTestPasswords.API, connectionLimit: 20}, false, nil)
	expectRuntimeRole(mock, "deadlock", runtimeRoleSpec{name: SyncRoleName, password: roleTestPasswords.Sync, connectionLimit: 5, writeAccess: true}, false, nil)
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	err := ConfigureRuntimeRoles(t.Context(), database, roleTestPasswords)
	if err == nil || !strings.Contains(err.Error(), "commit runtime role configuration") {
		t.Fatalf("expected commit error, got %v", err)
	}
	assertDBSQLExpectations(t, mock)
}

func expectRoleTransactionSetup(mock sqlmock.Sqlmock, databaseName string) {
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SET LOCAL standard_conforming_strings = on")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_database()")).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow(databaseName))
	mock.ExpectExec(regexp.QuoteMeta("REVOKE CREATE ON SCHEMA public FROM PUBLIC")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("REVOKE ALL PRIVILEGES ON DATABASE " + quotePostgresIdentifier(databaseName) + " FROM PUBLIC")).WillReturnResult(sqlmock.NewResult(0, 0))
}

func expectRuntimeRole(mock sqlmock.Sqlmock, databaseName string, role runtimeRoleSpec, exists bool, firstPrivilegeError error) {
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs(role.name).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(exists))
	operation := "CREATE ROLE"
	if exists {
		operation = "ALTER ROLE"
	}
	roleSQL := operation + " " + quotePostgresIdentifier(role.name) + " LOGIN PASSWORD " + quotePostgresLiteral(role.password)
	mock.ExpectExec(regexp.QuoteMeta(roleSQL) + `.*`).WillReturnResult(sqlmock.NewResult(0, 0))
	for index, statement := range runtimeRolePrivilegeStatements(databaseName, role) {
		expectation := mock.ExpectExec(regexp.QuoteMeta(statement))
		if index == 0 && firstPrivilegeError != nil {
			expectation.WillReturnError(firstPrivilegeError)
			return
		}
		expectation.WillReturnResult(sqlmock.NewResult(0, 0))
	}
}
