package builder

import (
	"fmt"
	"sort"
	"strings"
)

// InsertBuilder builds INSERT INTO queries.
type InsertBuilder struct {
	table   string
	columns []string
	rows    [][]any
	onDupKV []kv
	err     error
}

type kv struct {
	col string
	val any
}

// InsertInto starts an INSERT INTO builder for the given table.
func InsertInto(table string) *InsertBuilder {
	b := &InsertBuilder{}
	q, err := quoteIdent(table)
	if err != nil {
		b.err = fmt.Errorf("INSERT table: %w", err)
		return b
	}
	b.table = q
	return b
}

func (b *InsertBuilder) setErr(err error) {
	if b.err == nil {
		b.err = err
	}
}

// Values sets a single row from a map. Keys are column names; order is sorted alphabetically
// to ensure stable SQL output.
func (b *InsertBuilder) Values(row map[string]any) *InsertBuilder {
	if len(row) == 0 {
		b.setErr(ErrEmptyValues)
		return b
	}

	keys := make([]string, 0, len(row))
	for k := range row {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	quotedCols := make([]string, len(keys))
	vals := make([]any, len(keys))
	for i, k := range keys {
		q, err := quoteIdent(k)
		if err != nil {
			b.setErr(fmt.Errorf("INSERT column %q: %w", k, err))
			return b
		}
		quotedCols[i] = q
		vals[i] = row[k]
	}

	b.columns = quotedCols
	b.rows = [][]any{vals}
	return b
}

// Columns sets the column list for use with Rows().
func (b *InsertBuilder) Columns(cols ...string) *InsertBuilder {
	if len(cols) == 0 {
		b.setErr(ErrEmptyColumns)
		return b
	}
	quotedCols := make([]string, len(cols))
	for i, c := range cols {
		q, err := quoteIdent(c)
		if err != nil {
			b.setErr(fmt.Errorf("INSERT column %q: %w", c, err))
			return b
		}
		quotedCols[i] = q
	}
	b.columns = quotedCols
	return b
}

// Rows adds one or more value rows for bulk insert.
// Each row must have the same number of values as there are columns set via Columns().
func (b *InsertBuilder) Rows(rows ...[]any) *InsertBuilder {
	if len(rows) == 0 {
		b.setErr(ErrBulkEmptyRows)
		return b
	}
	for i, row := range rows {
		if len(b.columns) > 0 && len(row) != len(b.columns) {
			b.setErr(fmt.Errorf("%w: row %d has %d values, expected %d",
				ErrMismatchedRows, i, len(row), len(b.columns)))
			return b
		}
		b.rows = append(b.rows, row)
	}
	return b
}

// OnDuplicateKeyUpdate sets the ON DUPLICATE KEY UPDATE clause.
// Keys are sorted alphabetically for stable output.
func (b *InsertBuilder) OnDuplicateKeyUpdate(updates map[string]any) *InsertBuilder {
	if len(updates) == 0 {
		return b
	}
	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b.onDupKV = make([]kv, len(keys))
	for i, k := range keys {
		q, err := quoteIdent(k)
		if err != nil {
			b.setErr(fmt.Errorf("ON DUPLICATE KEY UPDATE column %q: %w", k, err))
			return b
		}
		b.onDupKV[i] = kv{col: q, val: updates[k]}
	}
	return b
}

// Build assembles the INSERT query.
func (b *InsertBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	if b.table == "" {
		return "", nil, ErrEmptyTable
	}
	if len(b.columns) == 0 {
		return "", nil, ErrEmptyColumns
	}
	if len(b.rows) == 0 {
		return "", nil, ErrEmptyValues
	}

	colList := "(" + strings.Join(b.columns, ", ") + ")"

	rowParts := make([]string, len(b.rows))
	var args []any
	for i, row := range b.rows {
		rowParts[i] = "(" + placeholders(len(row)) + ")"
		args = append(args, row...)
	}

	sql := "INSERT INTO " + b.table + " " + colList + " VALUES " + strings.Join(rowParts, ", ")

	if len(b.onDupKV) > 0 {
		setParts := make([]string, len(b.onDupKV))
		for i, p := range b.onDupKV {
			setParts[i] = p.col + " = ?"
			args = append(args, p.val)
		}
		sql += "\nON DUPLICATE KEY UPDATE " + strings.Join(setParts, ", ")
	}

	return sql, args, nil
}
