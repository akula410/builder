// Example 14_high_load demonstrates high-load best practices:
//   - One shared *sql.DB per application
//   - Connection pool configuration via connect.Config
//   - Per-request context timeout
//   - Prepared statements for repeated queries
//   - Production-safe logging with SafeLogQuery
//   - db.Stats() for observability
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/14_high_load
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	sqlbuilder "github.com/akula410/builder/v2"
	"github.com/akula410/connect/v2"
)

// app holds the shared database connection for the lifetime of the process.
// Create one instance at startup; do not create a new *sql.DB per request.
type app struct {
	db *sql.DB
}

func main() {
	a, err := newApp()
	if err != nil {
		fmt.Printf("startup: %v\n", err)
		fmt.Println("Set MYSQL_HOST, MYSQL_USER, MYSQL_PASSWORD, MYSQL_DATABASE to run this example.")
		return
	}
	defer a.db.Close()

	// Run a few sample queries to demonstrate the patterns.
	a.listUsers(1, 10)
	a.showPoolStats()
	a.preparedStatementExample()
}

// newApp initialises the application's shared *sql.DB with production pool settings.
// The pool is configured once at startup — not per request, not per goroutine.
func newApp() (*app, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := connect.DefaultConfig()
	cfg.User = getenv("MYSQL_USER", "root")
	cfg.Password = os.Getenv("MYSQL_PASSWORD")
	cfg.Host = getenv("MYSQL_HOST", "127.0.0.1")
	cfg.Port = getenv("MYSQL_PORT", "3306")
	cfg.DBName = getenv("MYSQL_DATABASE", "builder_example")

	// Pool tuning — adjust for your workload and MySQL server settings.
	// Keep MaxIdleConns == MaxOpenConns to avoid connection churn.
	// Keep ConnMaxLifetime below MySQL wait_timeout (default 8 hours).
	cfg.MaxOpenConns = 25
	cfg.MaxIdleConns = 25
	cfg.ConnMaxLifetime = 5 * time.Minute
	cfg.ConnMaxIdleTime = 2 * time.Minute

	db, err := connect.NewMySQLContext(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &app{db: db}, nil
}

// listUsers demonstrates a standard per-request query with:
//   - a fresh builder instance (builders are not safe for concurrent use)
//   - a per-request context timeout
//   - production-safe logging via SafeLogQuery
func (a *app) listUsers(page, pageSize int64) {
	q := sqlbuilder.Select("id", "name", "email").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		OrderBy("id", sqlbuilder.Asc).
		Limit(pageSize).
		Offset((page - 1) * pageSize)

	// Log the SQL template and args count — never log the values in production.
	logEntry, err := sqlbuilder.SafeLogQuery(q)
	if err != nil {
		log.Printf("build error: %v", err)
		return
	}
	log.Printf("query sql=%q args_count=%d", logEntry.SQL, logEntry.ArgsCount)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	exec := sqlbuilder.NewExecutor(a.db)
	rows, err := exec.QueryContext(ctx, q)
	if err != nil {
		log.Printf("query: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== listUsers ===")
	for rows.Next() {
		var id int64
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			log.Printf("scan: %v", err)
			return
		}
		fmt.Printf("  id=%d name=%s email=%s\n", id, name, email)
	}
	if err := rows.Err(); err != nil {
		log.Printf("rows: %v", err)
	}
}

// showPoolStats prints current connection pool statistics.
// Use this in a health-check endpoint or a metrics exporter.
func (a *app) showPoolStats() {
	s := a.db.Stats()
	fmt.Println("\n=== db.Stats() ===")
	fmt.Printf("  MaxOpenConnections: %d\n", s.MaxOpenConnections)
	fmt.Printf("  OpenConnections:    %d\n", s.OpenConnections)
	fmt.Printf("  InUse:              %d\n", s.InUse)
	fmt.Printf("  Idle:               %d\n", s.Idle)
	fmt.Printf("  WaitCount:          %d\n", s.WaitCount)
	fmt.Printf("  WaitDuration:       %s\n", s.WaitDuration)
	fmt.Printf("  MaxIdleClosed:      %d\n", s.MaxIdleClosed)
	fmt.Printf("  MaxLifetimeClosed:  %d\n", s.MaxLifetimeClosed)
}

// preparedStatementExample demonstrates using prepared statements for a query
// that is executed many times with different arguments.
//
// Prepared statements amortise the parse/plan cost across multiple executions.
// Use them when the same parameterised query template runs many times in a tight loop.
// For one-off queries, QueryContext is simpler and equally safe.
func (a *app) preparedStatementExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.Eq("id", 0)) // placeholder — real value supplied below

	exec := sqlbuilder.NewExecutor(a.db)
	stmt, args, err := exec.PrepareContext(ctx, q)
	if err != nil {
		log.Printf("prepare: %v", err)
		return
	}
	defer stmt.Close()

	// args from Build() correspond to the WHERE clause values.
	// Replace with actual runtime values.
	userIDs := []int64{1, 2, 3}
	fmt.Println("\n=== Prepared statement loop ===")
	for _, id := range userIDs {
		args[0] = id // replace the placeholder value

		row := stmt.QueryRowContext(ctx, args...)
		var userID int64
		var name string
		if err := row.Scan(&userID, &name); err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("  id=%d: not found\n", id)
			} else {
				log.Printf("  id=%d scan: %v", id, err)
			}
			continue
		}
		fmt.Printf("  id=%d name=%s\n", userID, name)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
