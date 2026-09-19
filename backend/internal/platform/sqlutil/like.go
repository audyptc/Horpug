// Package sqlutil holds small helpers for building SQL that the standard
// driver doesn't cover.
package sqlutil

import "strings"

// likeEscaper backslash-escapes the characters that are special to LIKE and
// ILIKE: the wildcards % and _, and the escape character (backslash) itself,
// which must be doubled or it would swallow the character after it. Postgres
// treats backslash as the escape character by default, so no ESCAPE clause is
// needed.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// ContainsPattern turns user input into a bind value for a "contains" match
// with LIKE/ILIKE. The input is matched literally: a "%" or "_" the user types
// is searched for as that character rather than acting as a wildcard, which
// would otherwise make "50%" or "a_b" match far more rows than intended.
func ContainsPattern(input string) string {
	return "%" + likeEscaper.Replace(input) + "%"
}
