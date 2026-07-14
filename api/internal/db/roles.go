package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const (
	APIRoleName  = "deadlock_api"
	SyncRoleName = "deadlock_sync"
)

type RuntimeRolePasswords struct {
	API  string
	Sync string
}

type runtimeRoleSpec struct {
	name            string
	password        string
	connectionLimit int
	writeAccess     bool
}

func ConfigureRuntimeRoles(ctx context.Context, database *sql.DB, passwords RuntimeRolePasswords) error {
	roleSpecs, err := buildRuntimeRoleSpecs(passwords)
	if err != nil {
		return err
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin runtime role configuration: %w", err)
	}
	defer tx.Rollback()

	databaseName, err := prepareRuntimeRoleTransaction(ctx, tx)
	if err != nil {
		return err
	}

	for _, role := range roleSpecs {
		if err := ensureRuntimeRole(ctx, tx, databaseName, role); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit runtime role configuration: %w", err)
	}
	return nil
}

func buildRuntimeRoleSpecs(passwords RuntimeRolePasswords) ([]runtimeRoleSpec, error) {
	roleSpecs := []runtimeRoleSpec{
		{name: APIRoleName, password: passwords.API, connectionLimit: 20},
		{name: SyncRoleName, password: passwords.Sync, connectionLimit: 5, writeAccess: true},
	}
	for _, role := range roleSpecs {
		if err := validateRolePassword(role.name, role.password); err != nil {
			return nil, err
		}
	}
	if passwords.API == passwords.Sync {
		return nil, errors.New("API and sync database passwords must be distinct")
	}
	return roleSpecs, nil
}

func prepareRuntimeRoleTransaction(ctx context.Context, tx *sql.Tx) (string, error) {
	if _, err := tx.ExecContext(ctx, "SET LOCAL standard_conforming_strings = on"); err != nil {
		return "", fmt.Errorf("set safe string parsing: %w", err)
	}
	var databaseName string
	if err := tx.QueryRowContext(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		return "", fmt.Errorf("read current database name: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "REVOKE CREATE ON SCHEMA public FROM PUBLIC"); err != nil {
		return "", fmt.Errorf("lock public schema: %w", err)
	}
	databaseIdentifier := quotePostgresIdentifier(databaseName)
	if _, err := tx.ExecContext(ctx, "REVOKE ALL PRIVILEGES ON DATABASE "+databaseIdentifier+" FROM PUBLIC"); err != nil {
		return "", fmt.Errorf("lock database public privileges: %w", err)
	}
	return databaseName, nil
}

func ensureRuntimeRole(ctx context.Context, tx *sql.Tx, databaseName string, role runtimeRoleSpec) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)", role.name).Scan(&exists); err != nil {
		return fmt.Errorf("check role %s: %w", role.name, err)
	}

	roleIdentifier := quotePostgresIdentifier(role.name)
	passwordLiteral := quotePostgresLiteral(role.password)
	operation := "CREATE ROLE"
	if exists {
		operation = "ALTER ROLE"
	}
	roleSQL := fmt.Sprintf(
		"%s %s LOGIN PASSWORD %s NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOINHERIT CONNECTION LIMIT %d",
		operation,
		roleIdentifier,
		passwordLiteral,
		role.connectionLimit,
	)
	if _, err := tx.ExecContext(ctx, roleSQL); err != nil {
		return fmt.Errorf("configure role %s: %w", role.name, err)
	}

	for _, statement := range runtimeRolePrivilegeStatements(databaseName, role) {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("grant privileges to %s: %w", role.name, err)
		}
	}
	return nil
}

func runtimeRolePrivilegeStatements(databaseName string, role runtimeRoleSpec) []string {
	identifier := quotePostgresIdentifier(role.name)
	databaseIdentifier := quotePostgresIdentifier(databaseName)
	statements := []string{
		fmt.Sprintf("REVOKE ALL PRIVILEGES ON DATABASE %s FROM %s", databaseIdentifier, identifier),
		fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO %s", databaseIdentifier, identifier),
		fmt.Sprintf("REVOKE ALL PRIVILEGES ON SCHEMA public FROM %s", identifier),
		fmt.Sprintf("GRANT USAGE ON SCHEMA public TO %s", identifier),
		fmt.Sprintf("REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM %s", identifier),
		fmt.Sprintf("REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM %s", identifier),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON TABLES FROM %s", identifier),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON SEQUENCES FROM %s", identifier),
		fmt.Sprintf("GRANT SELECT ON ALL TABLES IN SCHEMA public TO %s", identifier),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO %s", identifier),
	}
	if role.writeAccess {
		statements = append(statements,
			fmt.Sprintf("GRANT INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO %s", identifier),
			fmt.Sprintf("GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %s", identifier),
			fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT INSERT, UPDATE, DELETE ON TABLES TO %s", identifier),
			fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO %s", identifier),
		)
	}
	return statements
}

func validateRolePassword(roleName, password string) error {
	if len(password) < 16 {
		return fmt.Errorf("%s password must be at least 16 characters", roleName)
	}
	if strings.ContainsRune(password, '\x00') {
		return errors.New("runtime role password cannot contain a NUL byte")
	}
	return nil
}

func quotePostgresIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func quotePostgresLiteral(value string) string {
	return `'` + strings.ReplaceAll(value, `'`, `''`) + `'`
}
