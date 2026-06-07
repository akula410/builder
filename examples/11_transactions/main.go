// Example 11_transactions demonstrates transactions:
//   - db.BeginTx(ctx, nil) to start a transaction
//   - sqlbuilder.NewExecutor(tx) to wrap the *sql.Tx
//   - exec.ExecContext to execute builders inside the transaction
//   - commit / rollback pattern with defer
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/11_transactions
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

	if err := runTransaction(db); err != nil {
		log.Printf("transaction failed: %v", err)
	} else {
		fmt.Println("Transaction committed successfully.")
	}
}

func buildOnlyExamples() {
	insert := sqlbuilder.InsertInto("orders").
		Values(map[string]any{
			"user_id": 1,
			"total":   99.90,
			"status":  "pending",
		})

	update := sqlbuilder.Update("users").
		Set("status", "has_orders").
		Where(sqlbuilder.Eq("id", 1)).
		Limit(1)

	sql1, args1, _ := insert.Build()
	sql2, args2, _ := update.Build()
	fmt.Println("=== Transaction queries (build-only) ===")
	fmt.Printf("INSERT: %s\n  args: %v\n\n", sql1, args1)
	fmt.Printf("UPDATE: %s\n  args: %v\n\n", sql2, args2)
}

// runTransaction shows the idiomatic Go pattern for transactions with the builder:
//
//  1. Begin a transaction with a timeout context.
//  2. Always defer tx.Rollback() — it is a no-op after a successful Commit.
//  3. Wrap the *sql.Tx in sqlbuilder.NewExecutor to use builders.
//  4. Execute builders with exec.ExecContext.
//  5. Call tx.Commit() at the end.
func runTransaction(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() // no-op after Commit; guards against panics or early returns

	exec := sqlbuilder.NewExecutor(tx)

	// Step 1: insert an order.
	insertOrder := sqlbuilder.InsertInto("orders").
		Values(map[string]any{
			"user_id": 1,
			"total":   99.90,
			"status":  "pending",
		})
	result, err := exec.ExecContext(ctx, insertOrder)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	orderID, _ := result.LastInsertId()
	fmt.Printf("Inserted order id=%d\n", orderID)

	// Step 2: update the user status.
	updateUser := sqlbuilder.Update("users").
		Set("status", "has_orders").
		Where(sqlbuilder.Eq("id", 1)).
		Limit(1)
	res2, err := exec.ExecContext(ctx, updateUser)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	affected, _ := res2.RowsAffected()
	fmt.Printf("Updated %d user row(s)\n", affected)

	// Commit — all steps succeeded.
	return tx.Commit()
}
