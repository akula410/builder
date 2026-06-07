package builder

import "fmt"

// TruncateTableBuilder builds TRUNCATE TABLE statements.
type TruncateTableBuilder struct {
	name string
	err  error
}

// TruncateTable starts a TRUNCATE TABLE builder.
func TruncateTable(name string) *TruncateTableBuilder {
	b := &TruncateTableBuilder{}
	q, err := quoteIdent(name)
	if err != nil {
		b.err = fmt.Errorf("TRUNCATE TABLE name: %w", err)
		return b
	}
	b.name = q
	return b
}

// Build assembles TRUNCATE TABLE SQL.
func (b *TruncateTableBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	return "TRUNCATE TABLE " + b.name, nil, nil
}
