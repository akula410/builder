package builder

import (
	"fmt"
	"strings"
)

// ColumnDef describes a single column definition used in CREATE TABLE or ALTER TABLE.
type ColumnDef struct {
	name          string // quoted
	typSQL        string // validated type SQL
	nullable      *bool  // nil = not specified, true = NULL, false = NOT NULL
	defaultVal    *string
	defaultIsRaw  bool
	autoIncrement bool
	comment       string
	hasComment    bool
	afterCol      string // AFTER `col`
	first         bool   // FIRST
	err           error
}

// Column creates a new column definition.
//
// typ can be:
//   - a plain string validated against the MySQL 8 type whitelist ("VARCHAR(255)", "BIGINT UNSIGNED")
//   - a ColumnTypeDef returned by RawType() to bypass validation
func Column(name string, typ interface{}) *ColumnDef {
	cd := &ColumnDef{}

	qn, err := quoteIdent(name)
	if err != nil {
		cd.err = fmt.Errorf("column name: %w", err)
		return cd
	}
	cd.name = qn

	switch v := typ.(type) {
	case string:
		if err := validateColumnType(v); err != nil {
			cd.err = err
			return cd
		}
		cd.typSQL = normalizeColumnType(v)
	case ColumnTypeDef:
		if v.sql == "" {
			cd.err = ErrEmptyColumnType
			return cd
		}
		cd.typSQL = v.sql
	default:
		cd.err = fmt.Errorf("sqlbuilder: unsupported column type argument %T", typ)
	}
	return cd
}

// NotNull marks the column as NOT NULL.
func (c *ColumnDef) NotNull() *ColumnDef {
	f := false
	c.nullable = &f
	return c
}

// Nullable marks the column as NULL.
func (c *ColumnDef) Nullable() *ColumnDef {
	t := true
	c.nullable = &t
	return c
}

// Default sets a quoted string default value.
//
// Example: Default("active") → DEFAULT 'active'
func (c *ColumnDef) Default(val string) *ColumnDef {
	c.defaultVal = &val
	c.defaultIsRaw = false
	return c
}

// DefaultRaw sets a raw SQL default expression (not quoted).
//
// Example: DefaultRaw("CURRENT_TIMESTAMP") → DEFAULT CURRENT_TIMESTAMP
//
// WARNING: Do not pass user input to DefaultRaw.
func (c *ColumnDef) DefaultRaw(expr string) *ColumnDef {
	c.defaultVal = &expr
	c.defaultIsRaw = true
	return c
}

// AutoIncrement marks the column as AUTO_INCREMENT.
func (c *ColumnDef) AutoIncrement() *ColumnDef {
	c.autoIncrement = true
	return c
}

// Comment adds a column comment.
func (c *ColumnDef) Comment(text string) *ColumnDef {
	c.comment = text
	c.hasComment = true
	return c
}

// After positions the column AFTER the named column (for ALTER TABLE ADD/MODIFY COLUMN).
func (c *ColumnDef) After(col string) *ColumnDef {
	q, err := quoteIdent(col)
	if err != nil {
		if c.err == nil {
			c.err = fmt.Errorf("AFTER column: %w", err)
		}
		return c
	}
	c.afterCol = q
	return c
}

// First positions the column first (for ALTER TABLE ADD/MODIFY COLUMN).
func (c *ColumnDef) First() *ColumnDef {
	c.first = true
	return c
}

// BuildSQL returns the SQL fragment for the column definition (without a leading DDL keyword).
// Used internally by CreateTable and AlterTable, and available for introspection/testing.
func (c *ColumnDef) BuildSQL() (string, error) {
	return c.buildSQL()
}

// buildSQL is the unexported implementation.
func (c *ColumnDef) buildSQL() (string, error) {
	if c.err != nil {
		return "", c.err
	}

	parts := []string{c.name, c.typSQL}

	if c.nullable != nil {
		if *c.nullable {
			parts = append(parts, "NULL")
		} else {
			parts = append(parts, "NOT NULL")
		}
	}

	if c.autoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}

	if c.defaultVal != nil {
		if c.defaultIsRaw {
			parts = append(parts, "DEFAULT "+*c.defaultVal)
		} else {
			parts = append(parts, "DEFAULT '"+escapeStringLiteral(*c.defaultVal)+"'")
		}
	}

	if c.hasComment {
		parts = append(parts, "COMMENT '"+escapeStringLiteral(c.comment)+"'")
	}

	if c.first {
		parts = append(parts, "FIRST")
	} else if c.afterCol != "" {
		parts = append(parts, "AFTER "+c.afterCol)
	}

	return strings.Join(parts, " "), nil
}

// escapeStringLiteral escapes single quotes for use inside SQL string literals.
func escapeStringLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
