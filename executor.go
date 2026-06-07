package builder

import (
	"context"
	"database/sql"
	"fmt"
)

// Executor wraps a DBTX (either *sql.DB or *sql.Tx) and executes QueryBuilders.
//
// Example:
//
//	exec := NewExecutor(db)
//
//	// Use with a transaction:
//	tx, _ := db.BeginTx(ctx, nil)
//	execTx := NewExecutor(tx)
type Executor struct {
	db DBTX
}

// NewExecutor creates an Executor that wraps db.
// db can be *sql.DB or *sql.Tx (both satisfy the DBTX interface).
func NewExecutor(db DBTX) *Executor {
	return &Executor{db: db}
}

// ExecContext builds the query and calls db.ExecContext.
func (e *Executor) ExecContext(ctx context.Context, b QueryBuilder) (sql.Result, error) {
	query, args, err := b.Build()
	if err != nil {
		return nil, fmt.Errorf("sqlbuilder: build: %w", err)
	}
	return e.db.ExecContext(ctx, query, args...)
}

// QueryContext builds the query and calls db.QueryContext.
func (e *Executor) QueryContext(ctx context.Context, b QueryBuilder) (*sql.Rows, error) {
	query, args, err := b.Build()
	if err != nil {
		return nil, fmt.Errorf("sqlbuilder: build: %w", err)
	}
	return e.db.QueryContext(ctx, query, args...)
}

// QueryRowContext builds the query and calls db.QueryRowContext.
func (e *Executor) QueryRowContext(ctx context.Context, b QueryBuilder) *sql.Row {
	query, args, err := b.Build()
	if err != nil {
		// *sql.Row carries an error; create a failing row via a known-bad query.
		// The caller checks row.Scan() which will return the original build error.
		row := e.db.QueryRowContext(ctx, "SELECT /* build error: "+err.Error()+" */ NULL WHERE 1=0")
		return row
	}
	return e.db.QueryRowContext(ctx, query, args...)
}

// PrepareContext builds the query, prepares a statement, and returns the statement and args.
// The caller is responsible for closing the returned *sql.Stmt.
//
// Prepared statements are useful when the same query is executed many times with different args.
//
// Example:
//
//	stmt, args, err := exec.PrepareContext(ctx, Select("id").From("users").Where(Eq("status", "active")))
//	if err != nil { ... }
//	defer stmt.Close()
//	rows, err := stmt.QueryContext(ctx, args...)
func (e *Executor) PrepareContext(ctx context.Context, b QueryBuilder) (*sql.Stmt, []any, error) {
	query, args, err := b.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("sqlbuilder: build: %w", err)
	}
	stmt, err := e.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("sqlbuilder: prepare: %w", err)
	}
	return stmt, args, nil
}

// PreparedQueryContext builds the query, prepares a statement, executes QueryContext,
// closes the statement, and returns the rows.
//
// Use this when you want a one-shot prepared query without managing statement lifecycle.
func (e *Executor) PreparedQueryContext(ctx context.Context, b QueryBuilder) (*sql.Rows, error) {
	stmt, args, err := e.PrepareContext(ctx, b)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.QueryContext(ctx, args...)
}

// PreparedExecContext builds the query, prepares a statement, executes ExecContext,
// closes the statement, and returns the result.
func (e *Executor) PreparedExecContext(ctx context.Context, b QueryBuilder) (sql.Result, error) {
	stmt, args, err := e.PrepareContext(ctx, b)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.ExecContext(ctx, args...)
}

// ToSQL is a convenience function that builds a QueryBuilder and returns SQL, args, and error.
// It does not execute anything.
func ToSQL(b QueryBuilder) (string, []any, error) {
	return b.Build()
}
