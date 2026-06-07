package builder

import (
	"fmt"
	"strings"
)

// UpdateBuilder builds UPDATE queries.
type UpdateBuilder struct {
	table    string
	sets     []kv
	wheres   []Condition
	limitVal int64
	hasLimit bool
	err      error
}

// Update starts an UPDATE builder for the given table.
func Update(table string) *UpdateBuilder {
	b := &UpdateBuilder{}
	q, err := quoteIdent(table)
	if err != nil {
		b.err = fmt.Errorf("UPDATE table: %w", err)
		return b
	}
	b.table = q
	return b
}

func (b *UpdateBuilder) setErr(err error) {
	if b.err == nil {
		b.err = err
	}
}

// Set adds a SET assignment.
func (b *UpdateBuilder) Set(col string, val any) *UpdateBuilder {
	q, err := quoteIdent(col)
	if err != nil {
		b.setErr(fmt.Errorf("UPDATE SET column %q: %w", col, err))
		return b
	}
	b.sets = append(b.sets, kv{col: q, val: val})
	return b
}

// Where adds a WHERE condition. Multiple calls are AND-ed together.
func (b *UpdateBuilder) Where(cond Condition) *UpdateBuilder {
	b.wheres = append(b.wheres, cond)
	return b
}

// Limit sets the LIMIT value. Must be >= 0.
func (b *UpdateBuilder) Limit(n int64) *UpdateBuilder {
	if n < 0 {
		b.setErr(ErrNegativeLimit)
		return b
	}
	b.limitVal = n
	b.hasLimit = true
	return b
}

// Build assembles the UPDATE query.
func (b *UpdateBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	if b.table == "" {
		return "", nil, ErrEmptyTable
	}
	if len(b.sets) == 0 {
		return "", nil, ErrEmptySet
	}

	setParts := make([]string, len(b.sets))
	var args []any
	for i, s := range b.sets {
		setParts[i] = s.col + " = ?"
		args = append(args, s.val)
	}

	sql := "UPDATE " + b.table + "\nSET " + strings.Join(setParts, ", ")

	if len(b.wheres) > 0 {
		whereSQL, whereArgs, err := buildConditions(b.wheres)
		if err != nil {
			return "", nil, err
		}
		if whereSQL != "" {
			sql += "\nWHERE " + whereSQL
			args = append(args, whereArgs...)
		}
	}

	if b.hasLimit {
		sql += fmt.Sprintf("\nLIMIT %d", b.limitVal)
	}

	return sql, args, nil
}
