package builder

import (
	"fmt"
	"strings"
)

// ReferentialAction represents ON DELETE / ON UPDATE actions in a FOREIGN KEY.
type ReferentialAction string

const (
	Restrict   ReferentialAction = "RESTRICT"
	Cascade    ReferentialAction = "CASCADE"
	SetNull    ReferentialAction = "SET NULL"
	NoAction   ReferentialAction = "NO ACTION"
	SetDefault ReferentialAction = "SET DEFAULT"
)

// ForeignKeyDef describes a FOREIGN KEY constraint.
type ForeignKeyDef struct {
	name      string   // quoted constraint name
	columns   []string // quoted local columns
	refTable  string   // quoted referenced table
	refCols   []string // quoted referenced columns
	onDelete  ReferentialAction
	onUpdate  ReferentialAction
	hasDelete bool
	hasUpdate bool
	err       error
}

// ForeignKey starts a foreign key definition.
// name is the constraint name; columns are the local columns.
func ForeignKey(name string, columns []string) *ForeignKeyDef {
	fk := &ForeignKeyDef{}
	qn, err := quoteIdent(name)
	if err != nil {
		fk.err = fmt.Errorf("FK constraint name: %w", err)
		return fk
	}
	fk.name = qn
	if len(columns) == 0 {
		fk.err = fmt.Errorf("FK %q: %w", name, ErrEmptyColumns)
		return fk
	}
	fk.columns = make([]string, len(columns))
	for i, c := range columns {
		q, err := quoteIdent(c)
		if err != nil {
			fk.err = fmt.Errorf("FK column %q: %w", c, err)
			return fk
		}
		fk.columns[i] = q
	}
	return fk
}

// References sets the referenced table and columns.
func (fk *ForeignKeyDef) References(table string, columns []string) *ForeignKeyDef {
	if fk.err != nil {
		return fk
	}
	qt, err := quoteIdent(table)
	if err != nil {
		fk.err = fmt.Errorf("FK references table: %w", err)
		return fk
	}
	fk.refTable = qt
	if len(columns) == 0 {
		fk.err = fmt.Errorf("FK references: %w", ErrEmptyColumns)
		return fk
	}
	fk.refCols = make([]string, len(columns))
	for i, c := range columns {
		q, err := quoteIdent(c)
		if err != nil {
			fk.err = fmt.Errorf("FK referenced column %q: %w", c, err)
			return fk
		}
		fk.refCols[i] = q
	}
	return fk
}

// OnDelete sets the ON DELETE referential action.
func (fk *ForeignKeyDef) OnDelete(action ReferentialAction) *ForeignKeyDef {
	fk.onDelete = action
	fk.hasDelete = true
	return fk
}

// OnUpdate sets the ON UPDATE referential action.
func (fk *ForeignKeyDef) OnUpdate(action ReferentialAction) *ForeignKeyDef {
	fk.onUpdate = action
	fk.hasUpdate = true
	return fk
}

// buildSQL returns the CONSTRAINT ... FOREIGN KEY ... SQL fragment.
func (fk *ForeignKeyDef) buildSQL() (string, error) {
	if fk.err != nil {
		return "", fk.err
	}
	if fk.refTable == "" {
		return "", fmt.Errorf("FK %s: References() not called", fk.name)
	}

	localCols := "(" + strings.Join(fk.columns, ", ") + ")"
	refCols := "(" + strings.Join(fk.refCols, ", ") + ")"

	sql := fmt.Sprintf("CONSTRAINT %s FOREIGN KEY %s REFERENCES %s %s",
		fk.name, localCols, fk.refTable, refCols)

	if fk.hasDelete {
		sql += " ON DELETE " + string(fk.onDelete)
	}
	if fk.hasUpdate {
		sql += " ON UPDATE " + string(fk.onUpdate)
	}
	return sql, nil
}
