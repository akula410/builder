// package builder provides a safe, MySQL 8 SQL builder for Go.
//
// All data values are passed through placeholders (?), never concatenated into SQL.
// Identifiers (table names, column names, index names) are validated and quoted with backticks.
// Raw SQL is only available through explicit Raw/Expr/RawCondition/RawQuery APIs.
//
// # Security model
//
// Values → placeholders + []any args.
// Identifiers → validated + backtick-quoted.
// Raw SQL → explicit, clearly marked as dangerous.
package builder

import (
	"context"
	"database/sql"
)

// QueryBuilder is the base interface for all query builders.
// Build returns a parameterized SQL string, a slice of args, and an error.
// The returned SQL must never be executed with concatenated user input.
type QueryBuilder interface {
	Build() (string, []any, error)
}

// DDLBuilder is implemented by DDL builders (CREATE TABLE, ALTER TABLE, etc.).
// DDL builders return empty args slices in most cases, since MySQL does not
// support placeholders for identifiers or column types in DDL statements.
type DDLBuilder interface {
	QueryBuilder
}

// BuiltQuery holds a pre-built SQL statement together with its args.
type BuiltQuery struct {
	SQL  string
	Args []any
}

// DBTX is satisfied by both *sql.DB and *sql.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}
