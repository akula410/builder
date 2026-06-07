// Example 02_insert demonstrates INSERT queries: single row, bulk, and ON DUPLICATE KEY UPDATE.
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/02_insert
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
	// Single row from map (columns sorted alphabetically).
	q1 := sqlbuilder.InsertInto("users").
		Values(map[string]any{
			"name":  "Alex",
			"email": "alex@example.com",
		})
	sql1, args1, err := q1.Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=== INSERT single row (map) ===")
	fmt.Println(sql1)
	fmt.Println("args:", args1)
	// INSERT INTO `users` (`email`, `name`) VALUES (?, ?)

	// Bulk INSERT using Columns + Rows.
	q2 := sqlbuilder.InsertInto("users").
		Columns("name", "email", "status").
		Rows(
			[]any{"Alice", "alice@example.com", "active"},
			[]any{"Bob", "bob@example.com", "inactive"},
			[]any{"Carol", "carol@example.com", "active"},
		)
	sql2, args2, err := q2.Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n=== Bulk INSERT ===")
	fmt.Println(sql2)
	fmt.Println("args:", args2)
	// INSERT INTO `users` (`name`, `email`, `status`) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)

	// INSERT ... ON DUPLICATE KEY UPDATE.
	q3 := sqlbuilder.InsertInto("users").
		Values(map[string]any{
			"email": "alex@example.com",
			"name":  "Alex Updated",
		}).
		OnDuplicateKeyUpdate(map[string]any{
			"name": "Alex Updated",
		})
	sql3, args3, err := q3.Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n=== INSERT ON DUPLICATE KEY UPDATE ===")
	fmt.Println(sql3)
	fmt.Println("args:", args3)
	// INSERT INTO `users` (`email`, `name`) VALUES (?, ?)
	// ON DUPLICATE KEY UPDATE `name` = ?
}

func runnableExamples(exec *sqlbuilder.Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Single row INSERT.
	q := sqlbuilder.InsertInto("users").
		Values(map[string]any{
			"name":   "Example User",
			"email":  "example@builder.test",
			"status": "active",
		})

	result, err := exec.ExecContext(ctx, q)
	if err != nil {
		log.Printf("insert: %v", err)
		return
	}
	lastID, _ := result.LastInsertId()
	affected, _ := result.RowsAffected()
	fmt.Printf("\n=== INSERT result: last_id=%d rows_affected=%d\n", lastID, affected)
}
