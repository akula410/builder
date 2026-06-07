package builder

import "fmt"

// validCharsets is the allowed CHARACTER SET whitelist for MySQL 8.
var validCharsets = map[string]bool{
	"utf8mb4": true, "utf8": true, "latin1": true, "ascii": true,
	"binary": true, "ucs2": true, "utf16": true, "utf32": true,
	"cp1251": true, "cp1256": true, "gbk": true, "gb2312": true,
	"armscii8": true, "big5": true, "cp850": true, "cp852": true,
}

// validCollations is the allowed COLLATION whitelist for MySQL 8 (common subset).
var validCollations = map[string]bool{
	"utf8mb4_unicode_ci":     true,
	"utf8mb4_general_ci":     true,
	"utf8mb4_bin":            true,
	"utf8mb4_0900_ai_ci":     true,
	"utf8mb4_0900_as_cs":     true,
	"utf8mb4_unicode_520_ci": true,
	"utf8_general_ci":        true,
	"utf8_unicode_ci":        true,
	"utf8_bin":               true,
	"latin1_swedish_ci":      true,
	"latin1_general_ci":      true,
	"latin1_bin":             true,
	"ascii_general_ci":       true,
	"binary":                 true,
}

// CreateDatabaseBuilder builds CREATE DATABASE statements.
type CreateDatabaseBuilder struct {
	name        string
	ifNotExists bool
	charset     string
	collation   string
	err         error
}

// CreateDatabase starts a CREATE DATABASE builder.
func CreateDatabase(name string) *CreateDatabaseBuilder {
	b := &CreateDatabaseBuilder{}
	q, err := quoteIdent(name)
	if err != nil {
		b.err = fmt.Errorf("CREATE DATABASE name: %w", err)
		return b
	}
	b.name = q
	return b
}

// IfNotExists adds IF NOT EXISTS.
func (b *CreateDatabaseBuilder) IfNotExists() *CreateDatabaseBuilder {
	b.ifNotExists = true
	return b
}

// CharacterSet sets the default CHARACTER SET (validated against whitelist).
func (b *CreateDatabaseBuilder) CharacterSet(cs string) *CreateDatabaseBuilder {
	if !validCharsets[cs] {
		b.err = fmt.Errorf("%w: %q", ErrInvalidCharset, cs)
		return b
	}
	b.charset = cs
	return b
}

// Collate sets the COLLATE (validated against whitelist).
func (b *CreateDatabaseBuilder) Collate(col string) *CreateDatabaseBuilder {
	if !validCollations[col] {
		b.err = fmt.Errorf("%w: %q", ErrInvalidCollation, col)
		return b
	}
	b.collation = col
	return b
}

// Build assembles CREATE DATABASE SQL.
func (b *CreateDatabaseBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	sql := "CREATE DATABASE"
	if b.ifNotExists {
		sql += " IF NOT EXISTS"
	}
	sql += " " + b.name
	if b.charset != "" {
		sql += " CHARACTER SET " + b.charset
	}
	if b.collation != "" {
		sql += " COLLATE " + b.collation
	}
	return sql, nil, nil
}

// DropDatabaseBuilder builds DROP DATABASE statements.
type DropDatabaseBuilder struct {
	name     string
	ifExists bool
	err      error
}

// DropDatabase starts a DROP DATABASE builder.
func DropDatabase(name string) *DropDatabaseBuilder {
	b := &DropDatabaseBuilder{}
	q, err := quoteIdent(name)
	if err != nil {
		b.err = fmt.Errorf("DROP DATABASE name: %w", err)
		return b
	}
	b.name = q
	return b
}

// IfExists adds IF EXISTS.
func (b *DropDatabaseBuilder) IfExists() *DropDatabaseBuilder {
	b.ifExists = true
	return b
}

// Build assembles DROP DATABASE SQL.
func (b *DropDatabaseBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	sql := "DROP DATABASE"
	if b.ifExists {
		sql += " IF EXISTS"
	}
	return sql + " " + b.name, nil, nil
}

// UseDatabaseBuilder builds USE `database` statements.
type UseDatabaseBuilder struct {
	name string
	err  error
}

// UseDatabase starts a USE database builder.
func UseDatabase(name string) *UseDatabaseBuilder {
	b := &UseDatabaseBuilder{}
	q, err := quoteIdent(name)
	if err != nil {
		b.err = fmt.Errorf("USE DATABASE name: %w", err)
		return b
	}
	b.name = q
	return b
}

// Build assembles USE `database` SQL.
func (b *UseDatabaseBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	return "USE " + b.name, nil, nil
}
