package builder

import (
	"fmt"
	"strings"
)

// DropTableBuilder builds DROP TABLE statements.
type DropTableBuilder struct {
	names    []string
	ifExists bool
	err      error
}

// DropTable starts a DROP TABLE builder. Accepts one or more table names.
//
// Examples:
//
//	DropTable("users").IfExists()
//	DropTable("old_users", "old_orders").IfExists()
func DropTable(names ...string) *DropTableBuilder {
	b := &DropTableBuilder{}
	if len(names) == 0 {
		b.err = ErrEmptyTable
		return b
	}
	quoted := make([]string, len(names))
	for i, n := range names {
		q, err := quoteIdent(n)
		if err != nil {
			b.err = fmt.Errorf("DROP TABLE name %q: %w", n, err)
			return b
		}
		quoted[i] = q
	}
	b.names = quoted
	return b
}

// IfExists adds IF EXISTS.
func (b *DropTableBuilder) IfExists() *DropTableBuilder {
	b.ifExists = true
	return b
}

// Build assembles DROP TABLE SQL.
func (b *DropTableBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	sql := "DROP TABLE"
	if b.ifExists {
		sql += " IF EXISTS"
	}
	return sql + " " + strings.Join(b.names, ", "), nil, nil
}
