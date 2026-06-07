// Example 06_join demonstrates JOIN queries using the safe typed API (JoinOn, LeftJoinOn)
// and the raw API (JoinRaw) with appropriate warnings.
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/06_join
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	sqlbuilder "github.com/akula410/builder/v2"
	dbhelper "github.com/akula410/builder/v2/examples/internal/db"
)

func main() {
	buildOnlyExamples()

	if !dbhelper.SkipIfNoDatabaseConfig() {
		return
	}
	db := dbhelper.MustOpen()
	if db == nil {
		return
	}
	defer db.Close()

	exec := sqlbuilder.NewExecutor(db)
	runnableExamples(exec)
}

func buildOnlyExamples() {
	// === Safe JOIN API (recommended) ===
	// JoinOn validates and backtick-quotes both column references; no raw SQL accepted.

	// INNER JOIN with OnEq.
	q1 := sqlbuilder.Select("u.id", "u.name", "o.total").
		From("users u").
		JoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id"))
	printQuery("INNER JOIN (safe)", q1)
	// INNER JOIN `orders` AS `o` ON `o`.`user_id` = `u`.`id`

	// LEFT JOIN — user rows even when no orders exist.
	q2 := sqlbuilder.Select("u.id", "u.name", sqlbuilder.Raw("COUNT(o.id) AS order_count")).
		From("users u").
		LeftJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
		GroupBy("u.id", "u.name")
	printQuery("LEFT JOIN (safe)", q2)

	// Compound ON condition using OnAnd.
	q3 := sqlbuilder.Select("u.id", "o.id", "o.status").
		From("users u").
		LeftJoinOn("orders o", sqlbuilder.OnAnd(
			sqlbuilder.OnEq("o.user_id", "u.id"),
			sqlbuilder.OnEq("o.status", "u.status"),
		))
	printQuery("LEFT JOIN with compound ON (OnAnd)", q3)
	// ON (`o`.`user_id` = `u`.`id` AND `o`.`status` = `u`.`status`)

	// RIGHT JOIN.
	q4 := sqlbuilder.Select("u.id", "o.id").
		From("users u").
		RightJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id"))
	printQuery("RIGHT JOIN (safe)", q4)

	// === Raw JOIN API (use only for trusted, hardcoded ON expressions) ===
	// WARNING: The ON string is embedded into SQL as-is — no validation, no escaping.
	// Only pass hardcoded column comparison expressions (e.g. "o.user_id = u.id").
	// NEVER pass user-controlled input as the ON argument — SQL injection risk.
	// To filter by a value use a Where condition instead.

	q5 := sqlbuilder.Select("u.id", "o.total").
		From("users u").
		JoinRaw("orders o", "o.user_id = u.id") // safe: hardcoded expression
	printQuery("INNER JOIN (raw — hardcoded ON only)", q5)
}

func runnableExamples(exec *sqlbuilder.Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := sqlbuilder.Select("u.id", "u.name", sqlbuilder.Raw("COUNT(o.id) AS orders")).
		From("users u").
		LeftJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
		GroupBy("u.id", "u.name").
		OrderBy("u.id", sqlbuilder.Asc).
		Limit(5)

	rows, err := exec.QueryContext(ctx, q)
	if err != nil {
		log.Printf("query: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Users with order count ===")
	for rows.Next() {
		var id int64
		var name string
		var orders int64
		if err := rows.Scan(&id, &name, &orders); err != nil {
			log.Printf("scan: %v", err)
			return
		}
		fmt.Printf("  id=%d name=%s orders=%d\n", id, name, orders)
	}
	if err := rows.Err(); err != nil {
		log.Printf("rows: %v", err)
	}
}

func printQuery(name string, b sqlbuilder.QueryBuilder) {
	sql, args, err := b.Build()
	if err != nil {
		log.Fatalf("%s: %v", name, err)
	}
	fmt.Printf("=== %s ===\n%s\nargs: %v\n\n", name, sql, args)
}
