// Example 04_delete demonstrates safe DELETE queries with WHERE and LIMIT.
//
// WARNING: DELETE without a WHERE clause removes ALL rows in the table.
// Always include a WHERE condition and, where possible, a LIMIT.
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/04_delete
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
	// Safe DELETE: always specify WHERE and LIMIT.
	q := sqlbuilder.DeleteFrom("users").
		Where(sqlbuilder.Eq("id", 42)).
		Limit(1)

	sql, args, err := q.Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=== Safe DELETE with WHERE and LIMIT ===")
	fmt.Println(sql)
	fmt.Println("args:", args)
	// DELETE FROM `users`
	// WHERE `id` = ?
	// LIMIT 1

	// DELETE inactive users older than a cutoff — still safe: has WHERE.
	q2 := sqlbuilder.DeleteFrom("users").
		Where(sqlbuilder.And(
			sqlbuilder.Eq("status", "inactive"),
			sqlbuilder.Lt("created_at", "2024-01-01"),
		)).
		Limit(100)
	sql2, args2, _ := q2.Build()
	fmt.Println("\n=== DELETE with compound WHERE ===")
	fmt.Println(sql2)
	fmt.Println("args:", args2)

	// DANGER — no WHERE: would delete ALL rows.
	// Never do this in production without explicit intent.
	dangerousQ := sqlbuilder.DeleteFrom("users")
	dangerSQL, _, _ := dangerousQ.Build()
	fmt.Println("\n=== ⚠ DELETE without WHERE (dangerous!) ===")
	fmt.Println(dangerSQL)
	// DELETE FROM `users`   ← removes every row!
}

func runnableExamples(exec *sqlbuilder.Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Delete only the row we inserted in the insert example.
	q := sqlbuilder.DeleteFrom("users").
		Where(sqlbuilder.Eq("email", "example@builder.test")).
		Limit(1)

	result, err := exec.ExecContext(ctx, q)
	if err != nil {
		log.Printf("delete: %v", err)
		return
	}
	affected, _ := result.RowsAffected()
	fmt.Printf("\n=== DELETE result: rows_affected=%d\n", affected)
}
