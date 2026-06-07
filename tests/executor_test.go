package sqlbuilder_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

// fakeDbtx is a minimal fake implementation of DBTX for unit testing.
type fakeDbtx struct {
	lastQuery string
	lastArgs  []any
	execErr   error
	queryErr  error
}

func (f *fakeDbtx) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.lastQuery = query
	f.lastArgs = args
	if f.execErr != nil {
		return nil, f.execErr
	}
	return &fakeResult{}, nil
}

func (f *fakeDbtx) QueryContext(_ context.Context, query string, args ...any) (*sql.Rows, error) {
	f.lastQuery = query
	f.lastArgs = args
	return nil, f.queryErr
}

func (f *fakeDbtx) QueryRowContext(_ context.Context, query string, args ...any) *sql.Row {
	f.lastQuery = query
	f.lastArgs = args
	return nil
}

func (f *fakeDbtx) PrepareContext(_ context.Context, query string) (*sql.Stmt, error) {
	f.lastQuery = query
	return nil, errors.New("fakeDbtx: PrepareContext not supported in unit test")
}

type fakeResult struct{}

func (r *fakeResult) LastInsertId() (int64, error) { return 1, nil }
func (r *fakeResult) RowsAffected() (int64, error) { return 1, nil }

// ---- Tests ----

func TestExecutor_ExecContext_CallsBuild(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	_, err := exec.ExecContext(context.Background(),
		sqlbuilder.Update("users").Set("status", "active").Where(sqlbuilder.Eq("id", 1)),
	)
	assertNoErr(t, err)
	assertContains(t, db.lastQuery, "UPDATE `users`")
	if len(db.lastArgs) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(db.lastArgs), db.lastArgs)
	}
}

func TestExecutor_QueryContext_CallsBuild(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	_, _ = exec.QueryContext(context.Background(),
		sqlbuilder.Select("id").From("users").Where(sqlbuilder.Eq("status", "active")),
	)
	assertContains(t, db.lastQuery, "SELECT `id`")
	assertContains(t, db.lastQuery, "FROM `users`")
}

func TestExecutor_BuildError_NotExecuted(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	// A builder with an error (negative limit) should not reach the DB.
	_, err := exec.ExecContext(context.Background(),
		sqlbuilder.Select("id").From("users").Limit(-1),
	)
	assertErr(t, err)
	if db.lastQuery != "" {
		t.Fatal("query should not have been sent to DB when Build() fails")
	}
}

func TestExecutor_AcceptsDBInterface(t *testing.T) {
	// Verify that *fakeDbtx satisfies the DBTX interface (compile-time check).
	var _ sqlbuilder.DBTX = &fakeDbtx{}
}

func TestExecutor_ToSQL(t *testing.T) {
	q := sqlbuilder.Select("id").From("users").Where(sqlbuilder.Eq("status", "active"))
	query, args, err := sqlbuilder.ToSQL(q)
	assertNoErr(t, err)
	assertContains(t, query, "SELECT `id`")
	assertArgs(t, []any{"active"}, args)
}
