// Example 03_update demonstrates UPDATE queries: Set, Where, Limit, ExecContext, RowsAffected.
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/03_update
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
	// UPDATE with multiple SET, WHERE, LIMIT.
	q := sqlbuilder.Update("users").
		Set("name", "Alex Renamed").
		Set("status", "inactive").
		Where(sqlbuilder.Eq("id", 10)).
		Limit(1)

	sql, args, err := q.Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=== UPDATE with WHERE and LIMIT ===")
	fmt.Println(sql)
	fmt.Println("args:", args)
	// UPDATE `users`
	// SET `name` = ?, `status` = ?
	// WHERE `id` = ?
	// LIMIT 1

	// UPDATE with complex WHERE.
	q2 := sqlbuilder.Update("users").
		Set("status", "inactive").
		Where(sqlbuilder.And(
			sqlbuilder.Eq("status", "active"),
			sqlbuilder.Lt("created_at", "2024-01-01"),
		))
	sql2, args2, _ := q2.Build()
	fmt.Println("\n=== UPDATE with compound WHERE ===")
	fmt.Println(sql2)
	fmt.Println("args:", args2)
}

func runnableExamples(exec *sqlbuilder.Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := sqlbuilder.Update("users").
		Set("status", "inactive").
		Where(sqlbuilder.Eq("email", "example@builder.test")).
		Limit(1)

	result, err := exec.ExecContext(ctx, q)
	if err != nil {
		log.Printf("update: %v", err)
		return
	}
	affected, _ := result.RowsAffected()
	fmt.Printf("\n=== UPDATE result: rows_affected=%d\n", affected)
}
