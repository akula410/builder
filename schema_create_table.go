package builder

import (
	"fmt"
	"strings"
)

// validEngines is the whitelist of supported MySQL 8 storage engines.
var validEngines = map[string]bool{
	"InnoDB": true, "INNODB": true,
	"MyISAM": true, "MYISAM": true,
	"MEMORY": true, "Memory": true,
	"CSV":       true,
	"ARCHIVE":   true,
	"BLACKHOLE": true,
	"NDB":       true,
	"MERGE":     true,
}

// CreateTableBuilder builds CREATE TABLE statements.
type CreateTableBuilder struct {
	name          string
	ifNotExists   bool
	temporary     bool
	columns       []*ColumnDef
	primaryKey    []string // quoted column names
	indexes       []indexDef
	foreignKeys   []*ForeignKeyDef
	engine        string
	charset       string
	collation     string
	comment       string
	autoIncrement uint64
	hasAI         bool
	err           error
}

// CreateTable starts a CREATE TABLE builder.
func CreateTable(name string) *CreateTableBuilder {
	b := &CreateTableBuilder{}
	q, err := quoteIdent(name)
	if err != nil {
		b.err = fmt.Errorf("CREATE TABLE name: %w", err)
		return b
	}
	b.name = q
	return b
}

func (b *CreateTableBuilder) setErr(err error) {
	if b.err == nil {
		b.err = err
	}
}

// IfNotExists adds IF NOT EXISTS.
func (b *CreateTableBuilder) IfNotExists() *CreateTableBuilder {
	b.ifNotExists = true
	return b
}

// Temporary creates a TEMPORARY table.
func (b *CreateTableBuilder) Temporary() *CreateTableBuilder {
	b.temporary = true
	return b
}

// Column adds a column definition.
func (b *CreateTableBuilder) Column(col *ColumnDef) *CreateTableBuilder {
	if col.err != nil {
		b.setErr(col.err)
		return b
	}
	b.columns = append(b.columns, col)
	return b
}

// Columns adds multiple column definitions.
func (b *CreateTableBuilder) Columns(cols ...*ColumnDef) *CreateTableBuilder {
	for _, c := range cols {
		b.Column(c)
	}
	return b
}

// PrimaryKey sets the PRIMARY KEY columns.
func (b *CreateTableBuilder) PrimaryKey(cols ...string) *CreateTableBuilder {
	quoted := make([]string, len(cols))
	for i, c := range cols {
		q, err := quoteIdent(c)
		if err != nil {
			b.setErr(fmt.Errorf("PRIMARY KEY column %q: %w", c, err))
			return b
		}
		quoted[i] = q
	}
	b.primaryKey = quoted
	return b
}

// Index adds a regular INDEX.
func (b *CreateTableBuilder) Index(name string, cols ...string) *CreateTableBuilder {
	id, err := buildIndexDef(indexKindRegular, name, cols)
	if err != nil {
		b.setErr(err)
		return b
	}
	b.indexes = append(b.indexes, id)
	return b
}

// UniqueIndex adds a UNIQUE INDEX.
func (b *CreateTableBuilder) UniqueIndex(name string, cols ...string) *CreateTableBuilder {
	id, err := buildIndexDef(indexKindUnique, name, cols)
	if err != nil {
		b.setErr(err)
		return b
	}
	b.indexes = append(b.indexes, id)
	return b
}

// FullTextIndex adds a FULLTEXT INDEX.
func (b *CreateTableBuilder) FullTextIndex(name string, cols ...string) *CreateTableBuilder {
	id, err := buildIndexDef(indexKindFullText, name, cols)
	if err != nil {
		b.setErr(err)
		return b
	}
	b.indexes = append(b.indexes, id)
	return b
}

// SpatialIndex adds a SPATIAL INDEX.
func (b *CreateTableBuilder) SpatialIndex(name string, cols ...string) *CreateTableBuilder {
	id, err := buildIndexDef(indexKindSpatial, name, cols)
	if err != nil {
		b.setErr(err)
		return b
	}
	b.indexes = append(b.indexes, id)
	return b
}

// ForeignKey adds a FOREIGN KEY constraint. Call References() on the returned builder,
// then pass it back to AddForeignKey, or chain directly in CreateTable.
func (b *CreateTableBuilder) ForeignKey(fk *ForeignKeyDef) *CreateTableBuilder {
	b.foreignKeys = append(b.foreignKeys, fk)
	return b
}

// Engine sets the storage engine (validated against whitelist).
func (b *CreateTableBuilder) Engine(engine string) *CreateTableBuilder {
	if !validEngines[engine] {
		b.setErr(fmt.Errorf("%w: %q", ErrInvalidEngine, engine))
		return b
	}
	b.engine = engine
	return b
}

// CharacterSet sets the default CHARACTER SET.
func (b *CreateTableBuilder) CharacterSet(cs string) *CreateTableBuilder {
	if !validCharsets[cs] {
		b.setErr(fmt.Errorf("%w: %q", ErrInvalidCharset, cs))
		return b
	}
	b.charset = cs
	return b
}

// Collate sets the COLLATE option.
func (b *CreateTableBuilder) Collate(col string) *CreateTableBuilder {
	if !validCollations[col] {
		b.setErr(fmt.Errorf("%w: %q", ErrInvalidCollation, col))
		return b
	}
	b.collation = col
	return b
}

// Comment sets a table comment.
func (b *CreateTableBuilder) Comment(text string) *CreateTableBuilder {
	b.comment = text
	return b
}

// AutoIncrementValue sets the AUTO_INCREMENT table option (initial value).
func (b *CreateTableBuilder) AutoIncrementValue(n uint64) *CreateTableBuilder {
	b.autoIncrement = n
	b.hasAI = true
	return b
}

// Build assembles CREATE TABLE SQL.
func (b *CreateTableBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	if b.name == "" {
		return "", nil, ErrEmptyTable
	}
	if len(b.columns) == 0 {
		return "", nil, fmt.Errorf("CREATE TABLE %s: %w", b.name, ErrEmptyColumns)
	}

	kw := "CREATE TABLE"
	if b.temporary {
		kw = "CREATE TEMPORARY TABLE"
	}
	if b.ifNotExists {
		kw += " IF NOT EXISTS"
	}

	// Build column + constraint definitions
	var defs []string
	for _, col := range b.columns {
		s, err := col.buildSQL()
		if err != nil {
			return "", nil, err
		}
		defs = append(defs, "  "+s)
	}

	if len(b.primaryKey) > 0 {
		defs = append(defs, "  PRIMARY KEY ("+strings.Join(b.primaryKey, ", ")+")")
	}

	for _, idx := range b.indexes {
		s, err := idx.buildSQL()
		if err != nil {
			return "", nil, err
		}
		defs = append(defs, "  "+s)
	}

	for _, fk := range b.foreignKeys {
		s, err := fk.buildSQL()
		if err != nil {
			return "", nil, err
		}
		defs = append(defs, "  "+s)
	}

	sql := kw + " " + b.name + " (\n" + strings.Join(defs, ",\n") + "\n)"

	// Table options
	if b.engine != "" {
		sql += " ENGINE=" + b.engine
	}
	if b.charset != "" {
		sql += " DEFAULT CHARSET=" + b.charset
	}
	if b.collation != "" {
		sql += " COLLATE=" + b.collation
	}
	if b.comment != "" {
		sql += " COMMENT='" + escapeStringLiteral(b.comment) + "'"
	}
	if b.hasAI {
		sql += fmt.Sprintf(" AUTO_INCREMENT=%d", b.autoIncrement)
	}

	return sql, nil, nil
}
