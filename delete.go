package builder

import (
	"fmt"
	"strings"
)

// DeleteBuilder builds DELETE FROM queries.
type DeleteBuilder struct {
	table    string
	wheres   []Condition
	limitVal int64
	hasLimit bool
	err      error
}

// DeleteFrom starts a DELETE FROM builder for the given table.
func DeleteFrom(table string) *DeleteBuilder {
	b := &DeleteBuilder{}
	q, err := quoteIdent(table)
	if err != nil {
		b.err = fmt.Errorf("DELETE table: %w", err)
		return b
	}
	b.table = q
	return b
}

func (b *DeleteBuilder) setErr(err error) {
	if b.err == nil {
		b.err = err
	}
}

// Where adds a WHERE condition. Multiple calls are AND-ed together.
func (b *DeleteBuilder) Where(cond Condition) *DeleteBuilder {
	b.wheres = append(b.wheres, cond)
	return b
}

// Limit sets the LIMIT value. Must be >= 0.
func (b *DeleteBuilder) Limit(n int64) *DeleteBuilder {
	if n < 0 {
		b.setErr(ErrNegativeLimit)
		return b
	}
	b.limitVal = n
	b.hasLimit = true
	return b
}

// Build assembles the DELETE query.
func (b *DeleteBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	if b.table == "" {
		return "", nil, ErrEmptyTable
	}

	parts := []string{"DELETE FROM " + b.table}
	var args []any

	if len(b.wheres) > 0 {
		whereSQL, whereArgs, err := buildConditions(b.wheres)
		if err != nil {
			return "", nil, err
		}
		if whereSQL != "" {
			parts = append(parts, "WHERE "+whereSQL)
			args = append(args, whereArgs...)
		}
	}

	if b.hasLimit {
		parts = append(parts, fmt.Sprintf("LIMIT %d", b.limitVal))
	}

	return strings.Join(parts, "\n"), args, nil
}
