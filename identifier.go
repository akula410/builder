package builder

import (
	"fmt"
	"regexp"
	"strings"
)

// reIdent matches valid MySQL identifiers: letters, digits, underscore.
// Starts with a letter or underscore. Max 64 characters (MySQL limit).
var reIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,63}$`)

// quoteIdent wraps a single validated identifier in backticks.
func quoteIdent(name string) (string, error) {
	if name == "" {
		return "", ErrInvalidIdent
	}
	if !reIdent.MatchString(name) {
		return "", fmt.Errorf("%w: %q", ErrInvalidIdent, name)
	}
	return "`" + name + "`", nil
}

// mustQuoteIdent panics if name is invalid. For internal use only.
func mustQuoteIdent(name string) string {
	q, err := quoteIdent(name)
	if err != nil {
		panic(err)
	}
	return q
}

// reIntLiteral matches a bare integer literal used in SELECT (e.g. SELECT 1).
var reIntLiteral = regexp.MustCompile(`^\d+$`)

// parseColumnExpr parses a column reference and returns a quoted SQL expression.
//
// Supported formats:
//
//	"*"          → "*"
//	"1"          → "1"   (integer literal, e.g. SELECT 1)
//	"id"         → "`id`"
//	"u.id"       → "`u`.`id`"
//	"users.id"   → "`users`.`id`"
func parseColumnExpr(col string) (string, error) {
	if col == "*" {
		return "*", nil
	}
	if reIntLiteral.MatchString(col) {
		return col, nil
	}
	parts := strings.SplitN(col, ".", 2)
	if len(parts) == 2 {
		q0, err := quoteIdent(parts[0])
		if err != nil {
			return "", fmt.Errorf("column prefix %q: %w", parts[0], err)
		}
		q1, err := quoteIdent(parts[1])
		if err != nil {
			return "", fmt.Errorf("column name %q: %w", parts[1], err)
		}
		return q0 + "." + q1, nil
	}
	return quoteIdent(col)
}

// parseTableExpr parses a table reference with an optional alias.
//
// Supported formats:
//
//	"users"       → "`users`"
//	"users u"     → "`users` AS `u`"
//	"users AS u"  → "`users` AS `u`"
//	"users as u"  → "`users` AS `u`"
func parseTableExpr(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	upper := strings.ToUpper(expr)

	var tablePart, aliasPart string

	if idx := strings.Index(upper, " AS "); idx != -1 {
		tablePart = strings.TrimSpace(expr[:idx])
		aliasPart = strings.TrimSpace(expr[idx+4:])
	} else {
		fields := strings.Fields(expr)
		switch len(fields) {
		case 1:
			tablePart = fields[0]
		case 2:
			tablePart = fields[0]
			aliasPart = fields[1]
		default:
			return "", fmt.Errorf("%w: %q", ErrInvalidIdent, expr)
		}
	}

	quotedTable, err := quoteTableName(tablePart)
	if err != nil {
		return "", err
	}

	if aliasPart != "" {
		qa, err := quoteIdent(aliasPart)
		if err != nil {
			return "", fmt.Errorf("alias %q: %w", aliasPart, err)
		}
		return quotedTable + " AS " + qa, nil
	}
	return quotedTable, nil
}

// quoteTableName quotes a table name, supporting "schema.table" notation.
func quoteTableName(name string) (string, error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 2 {
		q0, err := quoteIdent(parts[0])
		if err != nil {
			return "", err
		}
		q1, err := quoteIdent(parts[1])
		if err != nil {
			return "", err
		}
		return q0 + "." + q1, nil
	}
	return quoteIdent(name)
}
