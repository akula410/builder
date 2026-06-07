// Example 05_conditions demonstrates all condition helpers:
// Eq, Neq, Gt, Gte, Lt, Lte, Like, NotLike, In, NotIn,
// IsNull, IsNotNull, Between, And, Or, RawCondition.
//
// All examples are build-only — no database connection is required.
package main

import (
	"fmt"
	"log"

	sqlbuilder "github.com/akula410/builder/v2"
)

func main() {
	// --- Comparison operators ---
	printQuery("Eq", sqlbuilder.Select("id").From("users").
		Where(sqlbuilder.Eq("status", "active")))

	printQuery("Neq", sqlbuilder.Select("id").From("users").
		Where(sqlbuilder.Neq("status", "banned")))

	printQuery("Gt", sqlbuilder.Select("id").From("orders").
		Where(sqlbuilder.Gt("total", 100)))

	printQuery("Gte", sqlbuilder.Select("id").From("orders").
		Where(sqlbuilder.Gte("total", 100)))

	printQuery("Lt", sqlbuilder.Select("id").From("orders").
		Where(sqlbuilder.Lt("total", 10)))

	printQuery("Lte", sqlbuilder.Select("id").From("orders").
		Where(sqlbuilder.Lte("total", 10)))

	// --- Pattern matching ---
	printQuery("Like", sqlbuilder.Select("id", "name").From("users").
		Where(sqlbuilder.Like("name", "Alex%")))

	printQuery("NotLike", sqlbuilder.Select("id", "name").From("users").
		Where(sqlbuilder.NotLike("email", "%@spam.%")))

	// --- IN / NOT IN ---
	printQuery("In", sqlbuilder.Select("id").From("users").
		Where(sqlbuilder.In("id", []any{1, 2, 3, 4})))

	printQuery("NotIn", sqlbuilder.Select("id").From("users").
		Where(sqlbuilder.NotIn("status", []any{"banned", "deleted"})))

	// --- NULL checks ---
	printQuery("IsNull", sqlbuilder.Select("id").From("users").
		Where(sqlbuilder.IsNull("deleted_at")))

	printQuery("IsNotNull", sqlbuilder.Select("id").From("users").
		Where(sqlbuilder.IsNotNull("email")))

	// --- BETWEEN ---
	printQuery("Between", sqlbuilder.Select("id").From("orders").
		Where(sqlbuilder.Between("total", 10, 500)))

	// --- AND / OR compound conditions ---
	printQuery("And+Or compound", sqlbuilder.Select("id", "name").From("users").
		Where(sqlbuilder.And(
			sqlbuilder.Eq("status", "active"),
			sqlbuilder.Or(
				sqlbuilder.Gt("price", 100),
				sqlbuilder.IsNull("price"),
			),
		)))
	// WHERE (`status` = ? AND (`price` > ? OR `price` IS NULL))

	// --- Multiple Where calls (implicit AND at top level) ---
	printQuery("Multiple Where", sqlbuilder.Select("id").From("users").
		Where(sqlbuilder.Eq("status", "active")).
		Where(sqlbuilder.Gt("created_at", "2025-01-01")).
		Where(sqlbuilder.IsNotNull("email")))

	// --- RawCondition: use only for hardcoded expressions ---
	// WARNING: Never pass user input directly in the sql string.
	// Values must use ? placeholders and be passed through args.
	printQuery("RawCondition (hardcoded)", sqlbuilder.Select("id").From("orders o").
		Where(sqlbuilder.RawCondition("o.user_id = u.id")))

	// Safe RawCondition with parameterised value.
	printQuery("RawCondition (parameterised)", sqlbuilder.Select("id").From("users").
		Where(sqlbuilder.RawCondition("JSON_EXTRACT(`data`, ?) = ?", "$.status", "active")))
}

func printQuery(name string, b sqlbuilder.QueryBuilder) {
	sql, args, err := b.Build()
	if err != nil {
		log.Fatalf("%s: %v", name, err)
	}
	fmt.Printf("=== %s ===\n%s\nargs: %v\n\n", name, sql, args)
}
