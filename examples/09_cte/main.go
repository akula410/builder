// Example 09_cte demonstrates Common Table Expressions (CTE):
// WITH, multiple CTEs, and WITH RECURSIVE.
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/09_cte
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
	// Single CTE: active users.
	q1 := sqlbuilder.
		With("active_users",
			sqlbuilder.Select("id", "name").
				From("users").
				Where(sqlbuilder.Eq("status", "active")),
		).
		Select("au.id", "au.name").
		From("active_users au")
	printQuery("Single CTE", q1)
	// WITH `active_users` AS (
	//     SELECT `id`, `name` FROM `users` WHERE `status` = ?
	// )
	// SELECT `au`.`id`, `au`.`name`
	// FROM `active_users` AS `au`

	// Multiple CTEs chained.
	activeUsers := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.Eq("status", "active"))

	paidOrders := sqlbuilder.Select("user_id", sqlbuilder.Raw("SUM(total) AS total")).
		From("orders").
		Where(sqlbuilder.Eq("status", "paid")).
		GroupBy("user_id")

	q2 := sqlbuilder.
		With("active_users", activeUsers).
		With("paid_orders", paidOrders).
		Select("au.id", "au.name", "po.total").
		From("active_users au").
		JoinOn("paid_orders po", sqlbuilder.OnEq("po.user_id", "au.id"))
	printQuery("Multiple CTEs", q2)

	// WITH RECURSIVE — walk a category tree.
	// RawQuery is used because UNION ALL is not yet in the typed builder API.
	// CTE args always precede the main query args.
	q3 := sqlbuilder.
		WithRecursive("category_tree",
			sqlbuilder.RawQuery(`
				SELECT id, parent_id, name FROM categories WHERE id = ?
				UNION ALL
				SELECT c.id, c.parent_id, c.name
				FROM categories c
				INNER JOIN category_tree ct ON c.parent_id = ct.id`,
				10,
			),
		).
		Select("id", "parent_id", "name").
		From("category_tree")
	printQuery("WITH RECURSIVE", q3)
}

func runnableExamples(exec *sqlbuilder.Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := sqlbuilder.
		With("active_users",
			sqlbuilder.Select("id", "name").
				From("users").
				Where(sqlbuilder.Eq("status", "active")),
		).
		Select("au.id", "au.name").
		From("active_users au").
		Limit(5)

	rows, err := exec.QueryContext(ctx, q)
	if err != nil {
		log.Printf("query: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Active users via CTE ===")
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("scan: %v", err)
			return
		}
		fmt.Printf("  id=%d name=%s\n", id, name)
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
