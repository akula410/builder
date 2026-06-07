// Example 12_context_timeout demonstrates per-query context timeouts and cancellation.
// Every real query should have a context with a timeout to prevent runaway queries
// from holding connections indefinitely.
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/12_context_timeout
package main

import (
	"context"
	"database/sql"
	"errors"
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
	runnableExamples(db, exec)
}

func buildOnlyExamples() {
	// Demonstrate what the queries look like.
	q := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		Limit(100)

	sql, args, _ := q.Build()
	fmt.Println("=== Query that will run with a timeout ===")
	fmt.Println(sql)
	fmt.Println("args:", args)
}

func runnableExamples(db *sql.DB, exec *sqlbuilder.Executor) {
	// SELECT with a 2-second timeout.
	{
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		q := sqlbuilder.Select("id", "name").
			From("users").
			Where(sqlbuilder.Eq("status", "active")).
			Limit(10)

		rows, err := exec.QueryContext(ctx, q)
		if err != nil {
			log.Printf("query: %v", err)
		} else {
			defer rows.Close()
			count := 0
			for rows.Next() {
				var id int64
				var name string
				if err := rows.Scan(&id, &name); err != nil {
					log.Printf("scan: %v", err)
					break
				}
				count++
			}
			fmt.Printf("\nSELECT with 2s timeout: %d row(s)\n", count)
			if err := rows.Err(); err != nil {
				log.Printf("rows: %v", err)
			}
		}
	}

	// Single row with QueryRowContextErr and a timeout.
	{
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		q := sqlbuilder.Select("id", "name").
			From("users").
			Where(sqlbuilder.Eq("id", 1))

		row, err := exec.QueryRowContextErr(ctx, q)
		if err != nil {
			log.Printf("build error: %v", err)
			return
		}
		var id int64
		var name string
		if err := row.Scan(&id, &name); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("QueryRowContextErr: no user with id=1")
			} else {
				log.Printf("scan: %v", err)
			}
			return
		}
		fmt.Printf("QueryRowContextErr: id=%d name=%s\n", id, name)
	}

	// ExecContext (INSERT) with cancel.
	{
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		q := sqlbuilder.Update("users").
			Set("status", "active").
			Where(sqlbuilder.Eq("id", 999)) // non-existent row — safe

		result, err := exec.ExecContext(ctx, q)
		if err != nil {
			log.Printf("update: %v", err)
			return
		}
		affected, _ := result.RowsAffected()
		fmt.Printf("Update with cancel context: rows_affected=%d\n", affected)

		// Explicitly cancel after use (also called by defer, but shown for clarity).
		cancel()
	}

	// Demonstrate that a query correctly times out.
	{
		// Very short timeout — most queries will exceed it.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		q := sqlbuilder.Select("id").From("users").Limit(1)
		_, err := exec.QueryContext(ctx, q)
		if err != nil {
			fmt.Printf("Expected timeout error: %v\n", err)
		}
	}
}
