package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func TestUpdate_Basic(t *testing.T) {
	sql, args, err := sqlbuilder.Update("users").
		Set("name", "Alex").
		Set("status", "active").
		Where(sqlbuilder.Eq("id", 10)).
		Build()

	assertNoErr(t, err)
	assertEqual(t, "UPDATE `users`\nSET `name` = ?, `status` = ?\nWHERE `id` = ?", sql)
	assertArgs(t, []any{"Alex", "active", 10}, args)
}

func TestUpdate_WithLimit(t *testing.T) {
	sql, args, err := sqlbuilder.Update("users").
		Set("status", "archived").
		Where(sqlbuilder.Lt("created_at", "2020-01-01")).
		Limit(100).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "LIMIT 100")
	assertArgs(t, []any{"archived", "2020-01-01"}, args)
}

func TestUpdate_NoSet_Error(t *testing.T) {
	_, _, err := sqlbuilder.Update("users").Where(sqlbuilder.Eq("id", 1)).Build()
	assertErr(t, err)
}

func TestUpdate_NegativeLimit_Error(t *testing.T) {
	_, _, err := sqlbuilder.Update("users").Set("x", 1).Limit(-1).Build()
	assertErr(t, err)
}

func TestUpdate_ValuesInArgs(t *testing.T) {
	dangerous := "'; DROP TABLE users; --"
	sql, args, err := sqlbuilder.Update("users").
		Set("name", dangerous).
		Where(sqlbuilder.Eq("id", 1)).
		Build()
	assertNoErr(t, err)
	if containsStr(sql, dangerous) {
		t.Fatal("dangerous value leaked into SQL")
	}
	assertArgs(t, []any{dangerous, 1}, args)
}
