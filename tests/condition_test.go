package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func buildCond(t *testing.T, cond sqlbuilder.Condition) (string, []any) {
	t.Helper()
	sql, args, err := sqlbuilder.Select("id").From("t").Where(cond).Build()
	assertNoErr(t, err)
	// Extract just the WHERE clause
	return sql, args
}

func TestCondition_Eq(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.Eq("status", "active"))
	assertContains(t, sql, "`status` = ?")
	assertArgs(t, []any{"active"}, args)
}

func TestCondition_Neq(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.Neq("status", "deleted"))
	assertContains(t, sql, "`status` != ?")
	assertArgs(t, []any{"deleted"}, args)
}

func TestCondition_Gt(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.Gt("age", 18))
	assertContains(t, sql, "`age` > ?")
	assertArgs(t, []any{18}, args)
}

func TestCondition_Gte(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.Gte("age", 18))
	assertContains(t, sql, "`age` >= ?")
	assertArgs(t, []any{18}, args)
}

func TestCondition_Lt(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.Lt("age", 65))
	assertContains(t, sql, "`age` < ?")
	assertArgs(t, []any{65}, args)
}

func TestCondition_Lte(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.Lte("age", 65))
	assertContains(t, sql, "`age` <= ?")
	assertArgs(t, []any{65}, args)
}

func TestCondition_Like(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.Like("name", "%Alex%"))
	assertContains(t, sql, "`name` LIKE ?")
	assertArgs(t, []any{"%Alex%"}, args)
}

func TestCondition_NotLike(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.NotLike("name", "%test%"))
	assertContains(t, sql, "`name` NOT LIKE ?")
	assertArgs(t, []any{"%test%"}, args)
}

func TestCondition_In(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.In("id", []any{1, 2, 3}))
	assertContains(t, sql, "`id` IN (?, ?, ?)")
	assertArgs(t, []any{1, 2, 3}, args)
}

func TestCondition_NotIn(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.NotIn("id", []any{4, 5}))
	assertContains(t, sql, "`id` NOT IN (?, ?)")
	assertArgs(t, []any{4, 5}, args)
}

func TestCondition_In_EmptyError(t *testing.T) {
	_, _, err := sqlbuilder.Select("id").From("t").Where(sqlbuilder.In("id", []any{})).Build()
	assertErr(t, err)
}

func TestCondition_NotIn_EmptyError(t *testing.T) {
	_, _, err := sqlbuilder.Select("id").From("t").Where(sqlbuilder.NotIn("id", []any{})).Build()
	assertErr(t, err)
}

func TestCondition_IsNull(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.IsNull("deleted_at"))
	assertContains(t, sql, "`deleted_at` IS NULL")
	assertArgs(t, nil, args)
}

func TestCondition_IsNotNull(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.IsNotNull("deleted_at"))
	assertContains(t, sql, "`deleted_at` IS NOT NULL")
	assertArgs(t, nil, args)
}

func TestCondition_Between(t *testing.T) {
	sql, args := buildCond(t, sqlbuilder.Between("price", 10, 100))
	assertContains(t, sql, "`price` BETWEEN ? AND ?")
	assertArgs(t, []any{10, 100}, args)
}

func TestCondition_And(t *testing.T) {
	cond := sqlbuilder.And(
		sqlbuilder.Eq("status", "active"),
		sqlbuilder.Gt("age", 18),
	)
	sql, args := buildCond(t, cond)
	assertContains(t, sql, "(`status` = ? AND `age` > ?)")
	assertArgs(t, []any{"active", 18}, args)
}

func TestCondition_Or(t *testing.T) {
	cond := sqlbuilder.Or(
		sqlbuilder.Eq("status", "active"),
		sqlbuilder.IsNull("status"),
	)
	sql, args := buildCond(t, cond)
	assertContains(t, sql, "(`status` = ? OR `status` IS NULL)")
	assertArgs(t, []any{"active"}, args)
}

func TestCondition_AndOr_Nested(t *testing.T) {
	cond := sqlbuilder.And(
		sqlbuilder.Eq("status", "active"),
		sqlbuilder.Or(
			sqlbuilder.Gt("price", 100),
			sqlbuilder.IsNull("price"),
		),
	)
	sql, args := buildCond(t, cond)
	assertContains(t, sql, "(`status` = ? AND (`price` > ? OR `price` IS NULL))")
	assertArgs(t, []any{"active", 100}, args)
}

func TestCondition_RawCondition(t *testing.T) {
	cond := sqlbuilder.RawCondition("o.user_id = u.id")
	sql, _ := buildCond(t, cond)
	assertContains(t, sql, "o.user_id = u.id")
}

func TestCondition_RawCondition_WithArgs(t *testing.T) {
	cond := sqlbuilder.RawCondition("JSON_EXTRACT(`data`, ?) = ?", "$.status", "active")
	sql, args := buildCond(t, cond)
	assertContains(t, sql, "JSON_EXTRACT(`data`, ?) = ?")
	assertArgs(t, []any{"$.status", "active"}, args)
}

func TestCondition_MultipleWhere_ANDed(t *testing.T) {
	sql, args, err := sqlbuilder.Select("id").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		Where(sqlbuilder.Eq("role", "admin")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "`status` = ? AND `role` = ?")
	assertArgs(t, []any{"active", "admin"}, args)
}

// Values always go into args, never into SQL string directly.
func TestCondition_ValuesInArgs_NotInSQL(t *testing.T) {
	dangerousInput := "'; DROP TABLE users; --"
	sql, args, err := sqlbuilder.Select("id").
		From("users").
		Where(sqlbuilder.Eq("name", dangerousInput)).
		Build()
	assertNoErr(t, err)
	// The SQL should contain a placeholder, not the actual value.
	assertContains(t, sql, "?")
	if len(args) != 1 || args[0] != dangerousInput {
		t.Fatalf("expected dangerous input in args, got %v", args)
	}
	if containsStr(sql, dangerousInput) {
		t.Fatal("dangerous input leaked into SQL string")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
