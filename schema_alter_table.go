package builder

import (
	"fmt"
	"strings"
)

// alterOp is a single ALTER TABLE operation.
type alterOp struct {
	sql string
}

// AlterTableBuilder builds ALTER TABLE statements.
// Multiple operations are combined into a single ALTER TABLE statement.
type AlterTableBuilder struct {
	table string
	ops   []alterOp
	err   error
}

// AlterTable starts an ALTER TABLE builder.
func AlterTable(table string) *AlterTableBuilder {
	b := &AlterTableBuilder{}
	q, err := quoteIdent(table)
	if err != nil {
		b.err = fmt.Errorf("ALTER TABLE name: %w", err)
		return b
	}
	b.table = q
	return b
}

func (b *AlterTableBuilder) setErr(err error) {
	if b.err == nil {
		b.err = err
	}
}

func (b *AlterTableBuilder) addOp(sql string) *AlterTableBuilder {
	b.ops = append(b.ops, alterOp{sql: sql})
	return b
}

// AddColumn adds an ADD COLUMN operation.
func (b *AlterTableBuilder) AddColumn(col *ColumnDef) *AlterTableBuilder {
	if col.err != nil {
		b.setErr(col.err)
		return b
	}
	s, err := col.buildSQL()
	if err != nil {
		b.setErr(err)
		return b
	}
	return b.addOp("ADD COLUMN " + s)
}

// DropColumn adds a DROP COLUMN operation.
func (b *AlterTableBuilder) DropColumn(col string) *AlterTableBuilder {
	q, err := quoteIdent(col)
	if err != nil {
		b.setErr(fmt.Errorf("DROP COLUMN name: %w", err))
		return b
	}
	return b.addOp("DROP COLUMN " + q)
}

// ModifyColumn adds a MODIFY COLUMN operation.
func (b *AlterTableBuilder) ModifyColumn(col *ColumnDef) *AlterTableBuilder {
	if col.err != nil {
		b.setErr(col.err)
		return b
	}
	s, err := col.buildSQL()
	if err != nil {
		b.setErr(err)
		return b
	}
	return b.addOp("MODIFY COLUMN " + s)
}

// ChangeColumn adds a CHANGE COLUMN operation (rename + redefine).
func (b *AlterTableBuilder) ChangeColumn(oldName string, col *ColumnDef) *AlterTableBuilder {
	qOld, err := quoteIdent(oldName)
	if err != nil {
		b.setErr(fmt.Errorf("CHANGE COLUMN old name: %w", err))
		return b
	}
	if col.err != nil {
		b.setErr(col.err)
		return b
	}
	s, err := col.buildSQL()
	if err != nil {
		b.setErr(err)
		return b
	}
	return b.addOp("CHANGE COLUMN " + qOld + " " + s)
}

// addIndexOp builds an index definition and records the ADD operation.
func (b *AlterTableBuilder) addIndexOp(kind indexKind, name string, cols []string) *AlterTableBuilder {
	id, err := buildIndexDef(kind, name, cols)
	if err != nil {
		b.setErr(err)
		return b
	}
	s, err := id.buildSQL()
	if err != nil {
		b.setErr(err)
		return b
	}
	return b.addOp("ADD " + s)
}

// AddIndex adds a regular ADD INDEX operation.
func (b *AlterTableBuilder) AddIndex(name string, cols ...string) *AlterTableBuilder {
	return b.addIndexOp(indexKindRegular, name, cols)
}

// AddUniqueIndex adds a ADD UNIQUE INDEX operation.
func (b *AlterTableBuilder) AddUniqueIndex(name string, cols ...string) *AlterTableBuilder {
	return b.addIndexOp(indexKindUnique, name, cols)
}

// AddFullTextIndex adds an ADD FULLTEXT INDEX operation.
func (b *AlterTableBuilder) AddFullTextIndex(name string, cols ...string) *AlterTableBuilder {
	return b.addIndexOp(indexKindFullText, name, cols)
}

// AddSpatialIndex adds an ADD SPATIAL INDEX operation.
func (b *AlterTableBuilder) AddSpatialIndex(name string, cols ...string) *AlterTableBuilder {
	return b.addIndexOp(indexKindSpatial, name, cols)
}

// AddIndexColumns adds an ADD INDEX with full IndexColumnDef control (prefix length, direction).
func (b *AlterTableBuilder) AddIndexColumns(name string, cols ...*IndexColumnDef) *AlterTableBuilder {
	flat := make([]IndexColumnDef, len(cols))
	for i, c := range cols {
		flat[i] = *c
	}
	id, err := buildIndexDefColumns(indexKindRegular, name, flat)
	if err != nil {
		b.setErr(err)
		return b
	}
	s, err := id.buildSQL()
	if err != nil {
		b.setErr(err)
		return b
	}
	return b.addOp("ADD " + s)
}

// DropIndex adds a DROP INDEX operation.
func (b *AlterTableBuilder) DropIndex(name string) *AlterTableBuilder {
	if name == "" {
		b.setErr(ErrEmptyIndex)
		return b
	}
	q, err := quoteIdent(name)
	if err != nil {
		b.setErr(fmt.Errorf("DROP INDEX name: %w", err))
		return b
	}
	return b.addOp("DROP INDEX " + q)
}

// Build assembles the ALTER TABLE statement.
func (b *AlterTableBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	if b.table == "" {
		return "", nil, ErrEmptyTable
	}
	if len(b.ops) == 0 {
		return "", nil, ErrEmptyAlterOps
	}

	opParts := make([]string, len(b.ops))
	for i, op := range b.ops {
		opParts[i] = op.sql
	}

	return "ALTER TABLE " + b.table + "\n" + strings.Join(opParts, ",\n"), nil, nil
}
