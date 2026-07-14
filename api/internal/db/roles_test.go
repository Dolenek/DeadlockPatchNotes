package db

import (
	"strings"
	"testing"
)

func TestPostgresQuoting(t *testing.T) {
	if got := quotePostgresIdentifier(`odd"role`); got != `"odd""role"` {
		t.Fatalf("unexpected identifier quote: %s", got)
	}
	if got := quotePostgresLiteral(`pa'ss\word`); got != `'pa''ss\word'` {
		t.Fatalf("unexpected literal quote: %s", got)
	}
}

func TestValidateRolePassword(t *testing.T) {
	if err := validateRolePassword(APIRoleName, "too-short"); err == nil {
		t.Fatal("expected short password to fail")
	}
	if err := validateRolePassword(APIRoleName, "sixteen-characters"); err != nil {
		t.Fatalf("expected valid password: %v", err)
	}
	if err := validateRolePassword(APIRoleName, "sixteen-characters\x00"); err == nil {
		t.Fatal("expected NUL password to fail")
	}
}

func TestRuntimeRolePasswordsMustBeDistinct(t *testing.T) {
	passwords := RuntimeRolePasswords{API: "same-secure-password", Sync: "same-secure-password"}
	if err := ConfigureRuntimeRoles(t.Context(), nil, passwords); err == nil || !strings.Contains(err.Error(), "distinct") {
		t.Fatalf("expected distinct-password validation, got %v", err)
	}
}

func TestRuntimeRolePrivilegesSeparateReadAndWriteAccess(t *testing.T) {
	apiStatements := strings.Join(runtimeRolePrivilegeStatements("deadlock", runtimeRoleSpec{name: APIRoleName}), ";")
	if !strings.Contains(apiStatements, "GRANT SELECT") {
		t.Fatal("expected API role to receive SELECT")
	}
	if strings.Contains(apiStatements, "GRANT INSERT") {
		t.Fatal("API role must not receive write privileges")
	}

	syncStatements := strings.Join(runtimeRolePrivilegeStatements("deadlock", runtimeRoleSpec{name: SyncRoleName, writeAccess: true}), ";")
	if !strings.Contains(syncStatements, "GRANT INSERT, UPDATE, DELETE") || !strings.Contains(syncStatements, "GRANT USAGE, SELECT ON ALL SEQUENCES") {
		t.Fatal("expected sync role to receive required write privileges")
	}
}
