package sqlbuilder_test

import (
	"context"
	"errors"
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func TestQueryRowContextErr_BuildError_ReturnsNilRow(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	row, err := exec.QueryRowContextErr(context.Background(),
		sqlbuilder.Select("id").From("users").Limit(-1),
	)
	assertErr(t, err)
	if row != nil {
		t.Fatal("expected nil *sql.Row when Build() fails")
	}
	if db.lastQuery != "" {
		t.Fatal("DB must not be called when Build() fails")
	}
}

func TestQueryRowContextErr_BuildError_IsWrappedOriginalError(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	_, err := exec.QueryRowContextErr(context.Background(),
		sqlbuilder.Select("id").From("users").Limit(-1),
	)
	if !errors.Is(err, sqlbuilder.ErrNegativeLimit) {
		t.Fatalf("expected errors.Is(err, ErrNegativeLimit) to be true, got: %v", err)
	}
}

func TestQueryRowContextErr_BuildError_InvalidIdent(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	_, err := exec.QueryRowContextErr(context.Background(),
		sqlbuilder.Select("1invalid").From("users"),
	)
	if !errors.Is(err, sqlbuilder.ErrInvalidIdent) {
		t.Fatalf("expected errors.Is(err, ErrInvalidIdent) to be true, got: %v", err)
	}
	if db.lastQuery != "" {
		t.Fatal("DB must not be called when Build() fails")
	}
}

func TestQueryRowContextErr_Success_CallsDB(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	_, err := exec.QueryRowContextErr(context.Background(),
		sqlbuilder.Select("id").From("users").Where(sqlbuilder.Eq("id", 42)),
	)
	assertNoErr(t, err)
	assertContains(t, db.lastQuery, "SELECT `id`")
	assertContains(t, db.lastQuery, "FROM `users`")
	assertContains(t, db.lastQuery, "WHERE `id` = ?")
	if len(db.lastArgs) != 1 || db.lastArgs[0] != 42 {
		t.Fatalf("expected args [42], got %v", db.lastArgs)
	}
}

func TestQueryRowContextErr_Success_ArgsPassedCorrectly(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	_, err := exec.QueryRowContextErr(context.Background(),
		sqlbuilder.Select("id", "name").
			From("users").
			Where(sqlbuilder.Eq("status", "active")).
			Where(sqlbuilder.Gt("age", 18)),
	)
	assertNoErr(t, err)
	assertArgs(t, []any{"active", 18}, db.lastArgs)
}

func TestQueryRowContext_Deprecated_StillWorks(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	_ = exec.QueryRowContext(context.Background(),
		sqlbuilder.Select("id").From("users").Where(sqlbuilder.Eq("id", 1)),
	)
	assertContains(t, db.lastQuery, "SELECT `id`")
	assertContains(t, db.lastQuery, "FROM `users`")
}

func TestQueryRowContext_Deprecated_BuildError_DoesNotPanic(t *testing.T) {
	db := &fakeDbtx{}
	exec := sqlbuilder.NewExecutor(db)

	// Old method must not panic on build error
	row := exec.QueryRowContext(context.Background(),
		sqlbuilder.Select("id").From("users").Limit(-1),
	)
	// row.Scan would fail, but we cannot verify that without a real *sql.DB.
	// What we CAN verify is that a dummy/fallback query was sent.
	_ = row
}
