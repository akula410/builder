// Example 01_select demonstrates SELECT queries using the builder package.
// It covers simple SELECT, WHERE, ORDER BY, LIMIT/OFFSET, QueryContext, and rows.Scan.
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/01_select
package main

import (
	"context"
	"database/sql"
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
	// Simple SELECT with WHERE, ORDER BY, LIMIT, OFFSET.
	q := sqlbuilder.Select("id", "name", "email").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		OrderBy("id", sqlbuilder.Desc).
		Limit(10).
		Offset(20)

	sql, args, err := q.Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=== Simple SELECT ===")
	fmt.Println(sql)
	fmt.Println("args:", args)
	// SELECT `id`, `name`, `email`
	// FROM `users`
	// WHERE `status` = ?
	// ORDER BY `id` DESC
	// LIMIT 10 OFFSET 20

	// SELECT DISTINCT.
	q2 := sqlbuilder.Select("status").
		Distinct().
		From("users")
	sql2, _, _ := q2.Build()
	fmt.Println("\n=== SELECT DISTINCT ===")
	fmt.Println(sql2)

	// SELECT with GROUP BY and HAVING.
	q3 := sqlbuilder.Select("status", sqlbuilder.Raw("COUNT(*) AS cnt")).
		From("users").
		GroupBy("status").
		Having(sqlbuilder.Gt("cnt", 5))
	sql3, args3, _ := q3.Build()
	fmt.Println("\n=== GROUP BY / HAVING ===")
	fmt.Println(sql3)
	fmt.Println("args:", args3)

	// SELECT with multiple WHERE conditions.
	q4 := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		Where(sqlbuilder.Gte("created_at", "2025-01-01")).
		Where(sqlbuilder.Lte("created_at", "2025-12-31"))
	sql4, args4, _ := q4.Build()
	fmt.Println("\n=== Multiple WHERE (AND) ===")
	fmt.Println(sql4)
	fmt.Println("args:", args4)
}

func runnableExamples(exec *sqlbuilder.Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Query multiple rows.
	q := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		OrderBy("id", sqlbuilder.Asc).
		Limit(5)

	rows, err := exec.QueryContext(ctx, q)
	if err != nil {
		log.Printf("query: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Rows from DB ===")
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

	// Query a single row using QueryRowContextErr.
	single := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.Eq("id", 1)).
		Limit(1)

	row, err := exec.QueryRowContextErr(ctx, single)
	if err != nil {
		log.Printf("build: %v", err)
		return
	}
	var id int64
	var name string
	if err := row.Scan(&id, &name); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("no user with id=1")
			return
		}
		log.Printf("scan: %v", err)
		return
	}
	fmt.Printf("\n=== Single row: id=%d name=%s\n", id, name)
}
