package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func TestInsert_OneRow_Map(t *testing.T) {
	sql, args, err := sqlbuilder.InsertInto("users").
		Values(map[string]any{
			"name":  "Alex",
			"email": "alex@test.com",
		}).
		Build()

	assertNoErr(t, err)
	// Columns are sorted alphabetically: email, name
	assertEqual(t, "INSERT INTO `users` (`email`, `name`) VALUES (?, ?)", sql)
	assertArgs(t, []any{"alex@test.com", "Alex"}, args)
}

func TestInsert_StableColumnOrder(t *testing.T) {
	// Running multiple times should produce the same column order.
	for range 5 {
		sql, _, err := sqlbuilder.InsertInto("users").
			Values(map[string]any{"z": 1, "a": 2, "m": 3}).
			Build()
		assertNoErr(t, err)
		assertContains(t, sql, "(`a`, `m`, `z`)")
	}
}

func TestInsert_BulkRows(t *testing.T) {
	sql, args, err := sqlbuilder.InsertInto("users").
		Columns("name", "email").
		Rows(
			[]any{"Alex", "alex@test.com"},
			[]any{"Ivan", "ivan@test.com"},
		).
		Build()

	assertNoErr(t, err)
	assertEqual(t, "INSERT INTO `users` (`name`, `email`) VALUES (?, ?), (?, ?)", sql)
	assertArgs(t, []any{"Alex", "alex@test.com", "Ivan", "ivan@test.com"}, args)
}

func TestInsert_OnDuplicateKeyUpdate(t *testing.T) {
	sql, args, err := sqlbuilder.InsertInto("users").
		Values(map[string]any{"id": 1, "name": "Alex"}).
		OnDuplicateKeyUpdate(map[string]any{"name": "Alex Updated"}).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "ON DUPLICATE KEY UPDATE `name` = ?")
	// args: id=1, name="Alex" (sorted: id, name), then "Alex Updated"
	assertArgs(t, []any{1, "Alex", "Alex Updated"}, args)
}

func TestInsert_MismatchedRows_Error(t *testing.T) {
	_, _, err := sqlbuilder.InsertInto("users").
		Columns("name", "email").
		Rows([]any{"Alex"}).
		Build()
	assertErr(t, err)
}

func TestInsert_NoValues_Error(t *testing.T) {
	_, _, err := sqlbuilder.InsertInto("users").Build()
	assertErr(t, err)
}

func TestInsert_NoColumns_Error(t *testing.T) {
	_, _, err := sqlbuilder.InsertInto("users").Rows([]any{"x"}).Build()
	assertErr(t, err)
}

func TestInsert_ValuesNotInSQL(t *testing.T) {
	dangerous := "'; DROP TABLE users; --"
	sql, args, err := sqlbuilder.InsertInto("users").
		Values(map[string]any{"name": dangerous}).
		Build()
	assertNoErr(t, err)
	if containsStr(sql, dangerous) {
		t.Fatal("dangerous value leaked into SQL")
	}
	assertArgs(t, []any{dangerous}, args)
}
