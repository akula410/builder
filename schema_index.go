package builder

import (
	"fmt"
	"strings"
)

// indexKind represents the type of index.
type indexKind string

const (
	indexKindRegular  indexKind = "INDEX"
	indexKindUnique   indexKind = "UNIQUE INDEX"
	indexKindFullText indexKind = "FULLTEXT INDEX"
	indexKindSpatial  indexKind = "SPATIAL INDEX"
)

// IndexColumnDef describes a single column in an index, with optional prefix length and direction.
type IndexColumnDef struct {
	name   string // quoted
	length int    // 0 = no prefix length
	desc   bool
	err    error
}

// IndexColumn creates an index column definition.
func IndexColumn(name string) *IndexColumnDef {
	q, err := quoteIdent(name)
	if err != nil {
		// Do not embed the caller-supplied name in the SQL fragment —
		// a name containing '*/' could close the comment and inject SQL.
		return &IndexColumnDef{err: fmt.Errorf("index column: %w", err)}
	}
	return &IndexColumnDef{name: q}
}

// Length sets a prefix length for the index column (e.g., for TEXT/BLOB columns).
func (ic *IndexColumnDef) Length(n int) *IndexColumnDef {
	ic.length = n
	return ic
}

// Desc marks the index column as descending (MySQL 8 supports per-column direction).
func (ic *IndexColumnDef) Desc() *IndexColumnDef {
	ic.desc = true
	return ic
}

// buildSQL returns the SQL fragment for this index column.
func (ic *IndexColumnDef) buildSQL() (string, error) {
	if ic.err != nil {
		return "", ic.err
	}
	s := ic.name
	if ic.length > 0 {
		s += fmt.Sprintf("(%d)", ic.length)
	}
	if ic.desc {
		s += " DESC"
	}
	return s, nil
}

// indexDef holds a complete index definition (used in CREATE TABLE and ALTER TABLE).
type indexDef struct {
	kind    indexKind
	name    string // quoted index name
	columns []IndexColumnDef
}

// buildSQL returns the SQL fragment for the index (used inside CREATE TABLE or ALTER TABLE ADD).
func (id indexDef) buildSQL() (string, error) {
	cols := make([]string, len(id.columns))
	for i, c := range id.columns {
		s, err := c.buildSQL()
		if err != nil {
			return "", err
		}
		cols[i] = s
	}
	return string(id.kind) + " " + id.name + " (" + strings.Join(cols, ", ") + ")", nil
}

// buildIndexDef validates and constructs an indexDef from simple column name strings.
func buildIndexDef(kind indexKind, name string, cols []string) (indexDef, error) {
	if name == "" {
		return indexDef{}, ErrEmptyIndex
	}
	if len(cols) == 0 {
		return indexDef{}, ErrEmptyIndexColumns
	}
	qn, err := quoteIdent(name)
	if err != nil {
		return indexDef{}, fmt.Errorf("index name: %w", err)
	}
	idxCols := make([]IndexColumnDef, len(cols))
	for i, c := range cols {
		q, err := quoteIdent(c)
		if err != nil {
			return indexDef{}, fmt.Errorf("index column %q: %w", c, err)
		}
		idxCols[i] = IndexColumnDef{name: q}
	}
	return indexDef{kind: kind, name: qn, columns: idxCols}, nil
}

// buildIndexDefColumns constructs an indexDef from IndexColumnDef values.
func buildIndexDefColumns(kind indexKind, name string, cols []IndexColumnDef) (indexDef, error) {
	if name == "" {
		return indexDef{}, ErrEmptyIndex
	}
	if len(cols) == 0 {
		return indexDef{}, ErrEmptyIndexColumns
	}
	qn, err := quoteIdent(name)
	if err != nil {
		return indexDef{}, fmt.Errorf("index name: %w", err)
	}
	return indexDef{kind: kind, name: qn, columns: cols}, nil
}
