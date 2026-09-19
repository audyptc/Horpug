package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	userdomain "apihorpug/internal/features/user/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// The SQL below is written with {REQ}, {ROLE} and {TARGET} placeholders and
// filled in by renderAccess, because the parameter numbers differ between the
// list queries (where they follow the filter arguments) and the single-row
// checks.
//   {REQ}    the requester's user id
//   {ROLE}   the requester's role id
//   {TARGET} the role being tested (a column or a parameter)

// requesterDormitories lists the dormitories the requester manages, either
// assigned to them directly or granted through their role.
const requesterDormitories = `
	SELECT dormitory_id FROM user_dormitories WHERE user_id = {REQ}
	UNION
	SELECT dormitory_id FROM role_dormitories WHERE role_id = {ROLE}`

// grantableRole holds when the requester may hold or hand out the {TARGET}
// role, i.e. it gives nothing they don't already have: it isn't full dormitory
// access, every menu permission it carries is one the requester's own role has,
// and every dormitory it names is one the requester manages. Without this a
// manager could give themselves, or an accomplice, a more powerful role.
const grantableRole = `(
	NOT COALESCE((SELECT gr.full_dormitory_access FROM roles gr WHERE gr.id = {TARGET}), FALSE)
	AND NOT EXISTS (
		SELECT 1 FROM role_menu_permissions tp
		WHERE tp.role_id = {TARGET}
		AND NOT EXISTS (
			SELECT 1 FROM role_menu_permissions mp
			WHERE mp.role_id = {ROLE} AND mp.menu_id = tp.menu_id AND mp.permission_id = tp.permission_id
		)
	)
	AND NOT EXISTS (
		SELECT 1 FROM role_dormitories td
		WHERE td.role_id = {TARGET} AND td.dormitory_id NOT IN (` + requesterDormitories + `)
	)
)`

// visibleUser restricts a query on users (aliased u) to those a requester
// without full dormitory access may see and act on: themselves, plus users they
// created or who share a dormitory with them — but only when that user's role
// is one the requester could grant (see grantableRole). The second condition
// keeps more privileged accounts out of reach: the seeded admin is registered
// on every dormitory, so sharing a dormitory alone would expose it.
const visibleUser = `(
	u.id = {REQ}
	OR (
		(
			u.created_by = {REQ}
			OR EXISTS (
				SELECT 1 FROM user_dormitories ud
				WHERE ud.user_id = u.id AND ud.dormitory_id IN (` + requesterDormitories + `)
			)
			OR EXISTS (
				SELECT 1 FROM role_dormitories rd
				WHERE rd.role_id = u.role_id AND rd.dormitory_id IN (` + requesterDormitories + `)
			)
		)
		AND ` + grantableRole + `
	)
)`

// renderAccess fills the placeholders in one of the templates above.
func renderAccess(template, req, role, target string) string {
	return strings.NewReplacer("{REQ}", req, "{ROLE}", role, "{TARGET}", target).Replace(template)
}

// visibleUserCondition is the visibleUser predicate for a list query whose
// next free parameters are argIdx and argIdx+1 (requester id, then role id).
func visibleUserCondition(requesterID, roleID uuid.UUID, argIdx *int, args *[]any) string {
	condition := renderAccess(visibleUser, fmt.Sprintf("$%d", *argIdx), fmt.Sprintf("$%d", *argIdx+1), "u.role_id")
	*args = append(*args, requesterID, roleID)
	*argIdx += 2
	return condition
}

// requesterScope reports whether the requester's role is exempt from dormitory
// scoping (sees and manages every user), and their role id.
func (r *Repository) requesterScope(ctx context.Context, requesterID uuid.UUID) (full bool, roleID uuid.UUID, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT r.full_dormitory_access, r.id
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, requesterID).Scan(&full, &roleID)
	if err != nil {
		return false, uuid.Nil, err
	}
	return full, roleID, nil
}

// EnsureAccess confirms the user exists and the requester may see and act on
// them. A missing user and one outside the requester's reach both surface as
// ErrUserNotFound so a scoped-out caller can't tell them apart.
func (r *Repository) EnsureAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.requesterScope(ctx, requesterID)
	if err != nil {
		return err
	}
	if full {
		return r.ensureUserExists(ctx, id)
	}

	var exists int
	err = r.db.QueryRow(ctx,
		`SELECT 1 FROM users u WHERE u.id = $1 AND `+renderAccess(visibleUser, "$2", "$3", "u.role_id"),
		id, requesterID, roleID,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.ErrUserNotFound
		}
		return err
	}
	return nil
}

// EnsureRoleAssignable confirms the role exists and the requester may give it
// to a user. It returns ErrRoleNotFound for a missing role and
// ErrRoleNotAssignable for one that grants more than the requester holds.
func (r *Repository) EnsureRoleAssignable(ctx context.Context, roleID, requesterID uuid.UUID) error {
	if err := r.ensureRoleExists(ctx, roleID); err != nil {
		return err
	}

	full, requesterRoleID, err := r.requesterScope(ctx, requesterID)
	if err != nil {
		return err
	}
	if full {
		return nil
	}

	var assignable bool
	if err := r.db.QueryRow(ctx,
		`SELECT `+renderAccess(grantableRole, "$2", "$3", "$1"),
		roleID, requesterID, requesterRoleID,
	).Scan(&assignable); err != nil {
		return err
	}
	if !assignable {
		return userdomain.ErrRoleNotAssignable
	}
	return nil
}
