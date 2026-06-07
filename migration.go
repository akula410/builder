package builder

import (
	"context"
	"fmt"
)

// Migration describes a named database migration with Up and Down steps.
//
// # MySQL DDL and transactions
//
// Most DDL statements (CREATE TABLE, ALTER TABLE, DROP TABLE, etc.) cause an
// implicit COMMIT in MySQL. They cannot be safely rolled back inside a
// transaction. RunUp / RunDown execute each step in sequence; if a step fails
// the remaining steps are skipped but already-executed DDL cannot be undone.
// Always back up your database before running destructive migrations.
type Migration struct {
	name  string
	ups   []QueryBuilder
	downs []QueryBuilder
}

// NewMigration creates a new Migration with the given name/ID.
// The name is typically a timestamp prefix: "20260607_120000_create_users".
func NewMigration(name string) *Migration {
	return &Migration{name: name}
}

// Up registers one or more builders to execute on migration apply.
func (m *Migration) Up(builders ...QueryBuilder) *Migration {
	m.ups = append(m.ups, builders...)
	return m
}

// Down registers one or more builders to execute on migration rollback.
func (m *Migration) Down(builders ...QueryBuilder) *Migration {
	m.downs = append(m.downs, builders...)
	return m
}

// Name returns the migration identifier.
func (m *Migration) Name() string { return m.name }

// BuildUp builds all Up steps and returns a slice of BuiltQuery.
func (m *Migration) BuildUp() ([]BuiltQuery, error) {
	return buildSteps(m.ups, "Up")
}

// BuildDown builds all Down steps and returns a slice of BuiltQuery.
func (m *Migration) BuildDown() ([]BuiltQuery, error) {
	return buildSteps(m.downs, "Down")
}

func buildSteps(builders []QueryBuilder, direction string) ([]BuiltQuery, error) {
	out := make([]BuiltQuery, 0, len(builders))
	for i, b := range builders {
		sql, args, err := b.Build()
		if err != nil {
			return nil, fmt.Errorf("migration %s step %d: %w", direction, i+1, err)
		}
		out = append(out, BuiltQuery{SQL: sql, Args: args})
	}
	return out, nil
}

// RunUp executes all Up steps using the provided executor.
// Steps are executed in order. On error, execution stops and the error is returned
// with a step index. Already-executed DDL steps cannot be rolled back in MySQL.
func (m *Migration) RunUp(ctx context.Context, exec *Executor) error {
	return m.run(ctx, exec, m.ups, "Up")
}

// RunDown executes all Down steps using the provided executor.
func (m *Migration) RunDown(ctx context.Context, exec *Executor) error {
	return m.run(ctx, exec, m.downs, "Down")
}

func (m *Migration) run(ctx context.Context, exec *Executor, builders []QueryBuilder, dir string) error {
	for i, b := range builders {
		if _, err := exec.ExecContext(ctx, b); err != nil {
			return fmt.Errorf("migration %q %s step %d: %w", m.name, dir, i+1, err)
		}
	}
	return nil
}
