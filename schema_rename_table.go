package builder

import (
	"fmt"
	"strings"
)

// TableRenamePair holds a from→to pair for RENAME TABLE.
type TableRenamePair struct {
	from string // quoted
	to   string // quoted
	err  error
}

// TableRename creates a from→to rename pair.
// The returned pair carries a validation error if either name is invalid;
// RenameTables.Build() will surface it.
func TableRename(from, to string) TableRenamePair {
	qf, err := quoteIdent(from)
	if err != nil {
		return TableRenamePair{err: fmt.Errorf("RENAME TABLE from %q: %w", from, err)}
	}
	qt, err := quoteIdent(to)
	if err != nil {
		return TableRenamePair{err: fmt.Errorf("RENAME TABLE to %q: %w", to, err)}
	}
	return TableRenamePair{from: qf, to: qt}
}

// RenameTableBuilder builds RENAME TABLE statements.
type RenameTableBuilder struct {
	pairs []TableRenamePair
	err   error
}

// RenameTable starts a RENAME TABLE builder for a single pair.
func RenameTable(from, to string) *RenameTableBuilder {
	b := &RenameTableBuilder{}
	qf, err := quoteIdent(from)
	if err != nil {
		b.err = fmt.Errorf("RENAME TABLE from %q: %w", from, err)
		return b
	}
	qt, err := quoteIdent(to)
	if err != nil {
		b.err = fmt.Errorf("RENAME TABLE to %q: %w", to, err)
		return b
	}
	b.pairs = []TableRenamePair{{from: qf, to: qt}}
	return b
}

// RenameTables creates a RENAME TABLE builder for multiple pairs at once.
func RenameTables(pairs ...TableRenamePair) *RenameTableBuilder {
	b := &RenameTableBuilder{}
	if len(pairs) == 0 {
		b.err = ErrEmptyRename
		return b
	}
	b.pairs = pairs
	return b
}

// Build assembles RENAME TABLE SQL.
func (b *RenameTableBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	if len(b.pairs) == 0 {
		return "", nil, ErrEmptyRename
	}
	parts := make([]string, len(b.pairs))
	for i, p := range b.pairs {
		if p.err != nil {
			return "", nil, p.err
		}
		parts[i] = p.from + " TO " + p.to
	}
	return "RENAME TABLE " + strings.Join(parts, ", "), nil, nil
}
