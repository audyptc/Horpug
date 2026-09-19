package postgres

import (
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// The access predicates are built by string templating, so a placeholder left
// unreplaced would only surface as a Postgres syntax error at request time.
func TestRenderAccessLeavesNoPlaceholders(t *testing.T) {
	for name, sql := range map[string]string{
		"visibleUser":   renderAccess(visibleUser, "$2", "$3", "u.role_id"),
		"grantableRole": renderAccess(grantableRole, "$2", "$3", "$1"),
	} {
		if strings.ContainsAny(sql, "{}") {
			t.Errorf("%s still has an unresolved placeholder:\n%s", name, sql)
		}
	}
}

func TestVisibleUserUsesOnlyTheGivenParameters(t *testing.T) {
	sql := renderAccess(visibleUser, "$4", "$5", "u.role_id")

	used := map[string]bool{}
	for _, p := range regexp.MustCompile(`\$\d+`).FindAllString(sql, -1) {
		used[p] = true
	}
	if len(used) != 2 || !used["$4"] || !used["$5"] {
		t.Errorf("expected only $4 and $5, got %v", used)
	}
}

func TestGrantableRoleParameters(t *testing.T) {
	sql := renderAccess(grantableRole, "$2", "$3", "$1")

	for _, want := range []string{"$1", "$2", "$3"} {
		if !strings.Contains(sql, want) {
			t.Errorf("grantableRole should reference %s:\n%s", want, sql)
		}
	}
}

// The condition takes the next two free parameters, so it can sit after the
// filter arguments already collected without renumbering them.
func TestVisibleUserConditionAdvancesArguments(t *testing.T) {
	argIdx := 3
	args := []any{"earlier"}

	cond := visibleUserCondition(uuid.New(), uuid.New(), &argIdx, &args)

	if argIdx != 5 {
		t.Errorf("argIdx = %d, want 5", argIdx)
	}
	if len(args) != 3 {
		t.Errorf("len(args) = %d, want 3", len(args))
	}
	if !strings.Contains(cond, "$3") || !strings.Contains(cond, "$4") || strings.Contains(cond, "$5") {
		t.Errorf("condition should use $3 and $4 only:\n%s", cond)
	}
}
