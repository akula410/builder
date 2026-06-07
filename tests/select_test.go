package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func TestSelect_Simple(t *testing.T) {
	sql, args, err := sqlbuilder.Select("id", "name", "email").
		From("users").
		Build()

	assertNoErr(t, err)
	assertEqual(t, "SELECT `id`, `name`, `email`\nFROM `users`", sql)
	assertArgs(t, nil, args)
}

func TestSelect_Star(t *testing.T) {
	sql, args, err := sqlbuilder.Select("*").From("users").Build()
	assertNoErr(t, err)
	assertEqual(t, "SELECT *\nFROM `users`", sql)
	assertArgs(t, nil, args)
}

func TestSelect_NoColumns_DefaultsStar(t *testing.T) {
	sql, args, err := sqlbuilder.Select().From("users").Build()
	assertNoErr(t, err)
	assertEqual(t, "SELECT *\nFROM `users`", sql)
	assertArgs(t, nil, args)
}

func TestSelect_Distinct(t *testing.T) {
	sql, _, err := sqlbuilder.Select("status").Distinct().From("users").Build()
	assertNoErr(t, err)
	assertContains(t, sql, "SELECT DISTINCT")
}

func TestSelect_Where(t *testing.T) {
	sql, args, err := sqlbuilder.Select("id", "name", "email").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		Where(sqlbuilder.Gt("created_at", "2025-01-01")).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "WHERE `status` = ? AND `created_at` > ?")
	assertArgs(t, []any{"active", "2025-01-01"}, args)
}

func TestSelect_OrderBy(t *testing.T) {
	sql, _, err := sqlbuilder.Select("id", "name", "email").
		From("users").
		OrderBy("id", sqlbuilder.Desc).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "ORDER BY `id` DESC")
}

func TestSelect_LimitOffset(t *testing.T) {
	sql, args, err := sqlbuilder.Select("id", "name", "email").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		Where(sqlbuilder.Gt("created_at", "2025-01-01")).
		OrderBy("id", sqlbuilder.Desc).
		Limit(10).
		Offset(20).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "LIMIT 10")
	assertContains(t, sql, "OFFSET 20")
	assertArgs(t, []any{"active", "2025-01-01"}, args)
}

func TestSelect_GroupBy(t *testing.T) {
	sql, args, err := sqlbuilder.Select("status", sqlbuilder.Raw("COUNT(*) AS cnt")).
		From("users").
		GroupBy("status").
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "GROUP BY `status`")
	assertArgs(t, nil, args)
}

func TestSelect_Having(t *testing.T) {
	sql, args, err := sqlbuilder.Select("status", sqlbuilder.Raw("COUNT(*) AS cnt")).
		From("users").
		GroupBy("status").
		Having(sqlbuilder.Gt("cnt", 5)).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "HAVING `cnt` > ?")
	assertArgs(t, []any{5}, args)
}

func TestSelect_LeftJoin(t *testing.T) {
	sql, args, err := sqlbuilder.Select("u.id", "u.name").
		From("users u").
		LeftJoin("orders o", "o.user_id = u.id").
		Where(sqlbuilder.Eq("u.status", "active")).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "FROM `users` AS `u`")
	assertContains(t, sql, "LEFT JOIN `orders` AS `o` ON o.user_id = u.id")
	assertContains(t, sql, "WHERE `u`.`status` = ?")
	assertArgs(t, []any{"active"}, args)
}

func TestSelect_QualifiedColumns(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id", "u.name").From("users u").Build()
	assertNoErr(t, err)
	assertContains(t, sql, "SELECT `u`.`id`, `u`.`name`")
}

func TestSelect_NegativeLimit_Error(t *testing.T) {
	_, _, err := sqlbuilder.Select("id").From("users").Limit(-1).Build()
	assertErr(t, err)
}

func TestSelect_NegativeOffset_Error(t *testing.T) {
	_, _, err := sqlbuilder.Select("id").From("users").Offset(-1).Build()
	assertErr(t, err)
}

func TestSelect_MissingFrom_Error(t *testing.T) {
	_, _, err := sqlbuilder.Select("id").Build()
	assertErr(t, err)
}

func TestSelect_RawColumn(t *testing.T) {
	sql, _, err := sqlbuilder.Select("user_id", sqlbuilder.Raw("SUM(total) AS total")).
		From("orders").
		GroupBy("user_id").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "SUM(total) AS total")
}

func TestSelect_InvalidColumnName_Error(t *testing.T) {
	_, _, err := sqlbuilder.Select("1invalid").From("users").Build()
	assertErr(t, err)
}
