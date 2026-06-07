package builder

import (
	"fmt"
	"strings"
)

// cteEntry holds a single Common Table Expression.
type cteEntry struct {
	name      string
	query     QueryBuilder
	recursive bool
}

// CTEBuilder accumulates CTE definitions and creates a SelectBuilder.
type CTEBuilder struct {
	ctes      []cteEntry
	recursive bool
}

// With starts a CTE chain with a named subquery.
//
// Example:
//
//	With("active_users", Select("id", "name").From("users").Where(Eq("status", "active"))).
//	    Select("au.id", "au.name").From("active_users au")
func With(name string, query QueryBuilder) *CTEBuilder {
	return &CTEBuilder{
		ctes: []cteEntry{{name: name, query: query}},
	}
}

// With adds another CTE to the chain.
func (cb *CTEBuilder) With(name string, query QueryBuilder) *CTEBuilder {
	cb.ctes = append(cb.ctes, cteEntry{name: name, query: query})
	return cb
}

// WithRecursive starts a recursive CTE chain.
//
// Example:
//
//	WithRecursive("tree", RawQuery(`SELECT id FROM categories WHERE id = ? UNION ALL ...`, 1)).
//	    Select("*").From("tree")
func WithRecursive(name string, query QueryBuilder) *CTEBuilder {
	return &CTEBuilder{
		ctes:      []cteEntry{{name: name, query: query}},
		recursive: true,
	}
}

// Select creates a SelectBuilder with the accumulated CTEs prepended.
func (cb *CTEBuilder) Select(columns ...any) *SelectBuilder {
	sb := Select(columns...)
	sb.cteList = cb.ctes
	sb.recursive = cb.recursive
	return sb
}

// buildCTEClause generates the WITH [RECURSIVE] ... SQL prefix.
// Args from CTE subqueries are returned in order and must precede main query args.
func buildCTEClause(ctes []cteEntry, recursive bool) (string, []any, error) {
	if len(ctes) == 0 {
		return "", nil, nil
	}

	keyword := "WITH"
	if recursive {
		keyword = "WITH RECURSIVE"
	}

	parts := make([]string, 0, len(ctes))
	var args []any

	for _, c := range ctes {
		qn, err := quoteIdent(c.name)
		if err != nil {
			return "", nil, fmt.Errorf("CTE name %q: %w", c.name, err)
		}
		subSQL, subArgs, err := c.query.Build()
		if err != nil {
			return "", nil, fmt.Errorf("CTE %q: %w", c.name, err)
		}
		parts = append(parts, qn+" AS (\n    "+subSQL+"\n)")
		args = append(args, subArgs...)
	}

	return keyword + " " + strings.Join(parts, ",\n"), args, nil
}
