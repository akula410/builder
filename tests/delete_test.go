package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func TestDelete_Basic(t *testing.T) {
	sql, args, err := sqlbuilder.DeleteFrom("users").
		Where(sqlbuilder.Eq("id", 10)).
		Limit(1).
		Build()

	assertNoErr(t, err)
	assertEqual(t, "DELETE FROM `users`\nWHERE `id` = ?\nLIMIT 1", sql)
	assertArgs(t, []any{10}, args)
}

func TestDelete_NoWhere(t *testing.T) {
	sql, args, err := sqlbuilder.DeleteFrom("logs").Build()
	assertNoErr(t, err)
	assertEqual(t, "DELETE FROM `logs`", sql)
	assertArgs(t, nil, args)
}

func TestDelete_WithLimit(t *testing.T) {
	sql, _, err := sqlbuilder.DeleteFrom("users").Limit(50).Build()
	assertNoErr(t, err)
	assertContains(t, sql, "LIMIT 50")
}

func TestDelete_NegativeLimit_Error(t *testing.T) {
	_, _, err := sqlbuilder.DeleteFrom("users").Limit(-1).Build()
	assertErr(t, err)
}

func TestDelete_InvalidTable_Error(t *testing.T) {
	_, _, err := sqlbuilder.DeleteFrom("1bad").Build()
	assertErr(t, err)
}

func TestDelete_ValuesInArgs(t *testing.T) {
	dangerous := "'; DROP TABLE users; --"
	sql, args, err := sqlbuilder.DeleteFrom("users").
		Where(sqlbuilder.Eq("name", dangerous)).
		Build()
	assertNoErr(t, err)
	if containsStr(sql, dangerous) {
		t.Fatal("dangerous value leaked into SQL")
	}
	assertArgs(t, []any{dangerous}, args)
}
