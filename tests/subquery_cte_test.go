package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

// ---- Subquery tests ----

func TestSubquery_InSubquery(t *testing.T) {
	inner := sqlbuilder.Select("user_id").
		From("orders").
		Where(sqlbuilder.Gt("total", 100))

	sql, args, err := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.InSubquery("id", inner)).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "`id` IN (")
	assertContains(t, sql, "SELECT `user_id`")
	assertContains(t, sql, "FROM `orders`")
	assertArgs(t, []any{100}, args)
}

func TestSubquery_Exists(t *testing.T) {
	inner := sqlbuilder.Select("1").
		From("orders o").
		Where(sqlbuilder.RawCondition("o.user_id = u.id"))

	sql, args, err := sqlbuilder.Select("id", "name").
		From("users u").
		Where(sqlbuilder.Exists(inner)).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "EXISTS (")
	assertContains(t, sql, "SELECT 1")
	assertArgs(t, nil, args)
}

func TestSubquery_NotExists(t *testing.T) {
	inner := sqlbuilder.Select("1").
		From("orders o").
		Where(sqlbuilder.RawCondition("o.user_id = u.id"))

	sql, _, err := sqlbuilder.Select("id").
		From("users u").
		Where(sqlbuilder.NotExists(inner)).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "NOT EXISTS (")
}

func TestSubquery_FromSubquery(t *testing.T) {
	inner := sqlbuilder.Select("user_id", sqlbuilder.Raw("SUM(total) AS total")).
		From("orders").
		GroupBy("user_id")

	sql, _, err := sqlbuilder.Select("t.user_id", "t.total").
		FromSubquery(inner, "t").
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "FROM (")
	assertContains(t, sql, ") AS `t`")
	assertContains(t, sql, "SELECT `user_id`, SUM(total) AS total")
}

func TestSubquery_SubqueryColumn(t *testing.T) {
	inner := sqlbuilder.Select(sqlbuilder.Raw("COUNT(*)")).
		From("orders o").
		Where(sqlbuilder.RawCondition("o.user_id = u.id"))

	sql, _, err := sqlbuilder.Select(
		"u.id",
		"u.name",
		sqlbuilder.SubqueryColumn(inner, "orders_count"),
	).From("users u").Build()

	assertNoErr(t, err)
	assertContains(t, sql, "AS `orders_count`")
	assertContains(t, sql, "SELECT COUNT(*)")
}

func TestSubquery_ArgsOrder(t *testing.T) {
	inner := sqlbuilder.Select("user_id").
		From("orders").
		Where(sqlbuilder.Gt("total", 500))

	sql, args, err := sqlbuilder.Select("id").
		From("users").
		Where(sqlbuilder.InSubquery("id", inner)).
		Where(sqlbuilder.Eq("status", "active")).
		Build()

	assertNoErr(t, err)
	_ = sql
	// Inner query arg (500) comes before outer WHERE arg ("active")
	assertArgs(t, []any{500, "active"}, args)
}

// ---- CTE tests ----

func TestCTE_Simple(t *testing.T) {
	active := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.Eq("status", "active"))

	sql, args, err := sqlbuilder.With("active_users", active).
		Select("au.id", "au.name").
		From("active_users au").
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "WITH `active_users` AS (")
	assertContains(t, sql, "SELECT `au`.`id`, `au`.`name`")
	assertContains(t, sql, "FROM `active_users` AS `au`")
	assertArgs(t, []any{"active"}, args)
}

func TestCTE_Multiple(t *testing.T) {
	q1 := sqlbuilder.Select("id").From("users").Where(sqlbuilder.Eq("status", "active"))
	q2 := sqlbuilder.Select("id").From("orders").Where(sqlbuilder.Gt("total", 100))

	sql, args, err := sqlbuilder.With("active_users", q1).
		With("big_orders", q2).
		Select("*").
		From("active_users").
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "`active_users` AS (")
	assertContains(t, sql, "`big_orders` AS (")
	// CTE args come first: "active" then 100
	assertArgs(t, []any{"active", 100}, args)
}

func TestCTE_Recursive(t *testing.T) {
	recQuery := sqlbuilder.RawQuery(`SELECT id, parent_id, name
FROM categories
WHERE id = ?

UNION ALL

SELECT c.id, c.parent_id, c.name
FROM categories c
INNER JOIN category_tree ct ON c.parent_id = ct.id`, 10)

	sql, args, err := sqlbuilder.WithRecursive("category_tree", recQuery).
		Select("*").
		From("category_tree").
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "WITH RECURSIVE `category_tree` AS (")
	assertContains(t, sql, "FROM `category_tree`")
	assertArgs(t, []any{10}, args)
}

func TestCTE_ArgsBeforeMainQuery(t *testing.T) {
	cteQ := sqlbuilder.Select("id").From("users").Where(sqlbuilder.Eq("status", "vip"))
	sql, args, err := sqlbuilder.With("vip_users", cteQ).
		Select("id").
		From("vip_users").
		Where(sqlbuilder.Eq("country", "US")).
		Build()

	assertNoErr(t, err)
	_ = sql
	// CTE arg ("vip") must come before main WHERE arg ("US")
	assertArgs(t, []any{"vip", "US"}, args)
}
