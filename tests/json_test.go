package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func TestJSON_Extract(t *testing.T) {
	expr := sqlbuilder.JSONExtract("data", "$.status")
	sql, args, err := expr.Build()
	assertNoErr(t, err)
	assertEqual(t, "JSON_EXTRACT(`data`, ?)", sql)
	assertArgs(t, []any{"$.status"}, args)
}

func TestJSON_Unquote(t *testing.T) {
	expr := sqlbuilder.JSONUnquote("data", "$.name")
	sql, args, err := expr.Build()
	assertNoErr(t, err)
	assertEqual(t, "JSON_UNQUOTE(JSON_EXTRACT(`data`, ?))", sql)
	assertArgs(t, []any{"$.name"}, args)
}

func TestJSON_Eq_InWhere(t *testing.T) {
	sql, args, err := sqlbuilder.Select("id").
		From("orders").
		Where(sqlbuilder.JSONEq("data", "$.company_id", "ngus")).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "JSON_UNQUOTE(JSON_EXTRACT(`data`, ?)) = ?")
	assertArgs(t, []any{"$.company_id", "ngus"}, args)
}

func TestJSON_Contains_InWhere(t *testing.T) {
	sql, args, err := sqlbuilder.Select("id").
		From("orders").
		Where(sqlbuilder.JSONContains("data", "$.tags", "urgent")).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "JSON_CONTAINS(JSON_EXTRACT(`data`, ?), JSON_QUOTE(?))")
	assertArgs(t, []any{"$.tags", "urgent"}, args)
}

func TestJSON_Values_NotInSQL(t *testing.T) {
	dangerous := "'; DROP TABLE orders; --"
	sql, args, err := sqlbuilder.Select("id").
		From("orders").
		Where(sqlbuilder.JSONEq("data", "$.key", dangerous)).
		Build()
	assertNoErr(t, err)
	if containsStr(sql, dangerous) {
		t.Fatal("dangerous value leaked into SQL")
	}
	assertArgs(t, []any{"$.key", dangerous}, args)
}
