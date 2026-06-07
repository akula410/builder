// Example 13_debug_sql demonstrates debug and safe-logging helpers.
//
// WARNING: DebugSQL output must NEVER be executed as real SQL.
// It substitutes placeholder values directly into the query for human-readable display.
// Use it only for logging in development/local environments.
// In production, use SafeLogQuery to log only the SQL template and args count.
//
// This example requires no database connection.
package main

import (
	"fmt"
	"log"
	"time"

	sqlbuilder "github.com/akula410/builder/v2"
)

func main() {
	debugBasic()
	debugWithConfig()
	safeLogProduction()
	debugFromBuilder()
}

func debugBasic() {
	q := sqlbuilder.Select("id", "email", "status").
		From("users").
		Where(sqlbuilder.And(
			sqlbuilder.Eq("status", "active"),
			sqlbuilder.Gt("created_at", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
		)).
		OrderBy("id", sqlbuilder.Desc).
		Limit(10)

	sql, args, err := q.Build()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Parameterised SQL (safe to execute) ===")
	fmt.Println(sql)
	fmt.Println("args:", args)

	// DebugSQL: substitutes args into the SQL for human-readable display.
	// WARNING: Do NOT execute this output.
	debug := sqlbuilder.DebugSQL(sql, args)
	fmt.Println("\n=== DebugSQL output (logging only — NEVER execute) ===")
	fmt.Println(debug)
	// SELECT `id`, `email`, `status`
	// FROM `users`
	// WHERE (`status` = 'active' AND `created_at` > '2025-01-01 00:00:00')
	// ORDER BY `id` DESC
	// LIMIT 10
}

func debugWithConfig() {
	q := sqlbuilder.Select("id", "email").
		From("users").
		Where(sqlbuilder.And(
			sqlbuilder.Eq("email", "alex@example.com"),
			sqlbuilder.Eq("password", "s3cr3t"), // will be redacted
			sqlbuilder.Eq("token", "abc123"),    // will be redacted
		))

	sql, args, err := q.Build()
	if err != nil {
		log.Fatal(err)
	}

	cfg := sqlbuilder.DebugConfig{
		MaxStringLength: 50,
		ShowBytes:       false,
		TimeFormat:      "2006-01-02",
		RedactFields:    []string{"password", "token", "api_key"},
		RedactValue:     "[REDACTED]",
	}
	result := sqlbuilder.DebugSQLWithConfig(sql, args, cfg)

	fmt.Println("\n=== DebugSQLWithConfig (redacted fields) ===")
	fmt.Println(result)
	// WHERE (`email` = 'alex@example.com' AND `password` = '[REDACTED]' AND `token` = '[REDACTED]')
}

func safeLogProduction() {
	q := sqlbuilder.InsertInto("users").
		Values(map[string]any{
			"email":    "alex@example.com",
			"name":     "Alex",
			"password": "s3cr3t",
		})

	// SafeLogQuery is the production-safe way to log queries.
	// It returns only the SQL template and the count of args — no values are exposed.
	logEntry, err := sqlbuilder.SafeLogQuery(q)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n=== SafeLogQuery (production-safe, no values) ===")
	fmt.Printf("SQL:        %s\n", logEntry.SQL)
	fmt.Printf("ArgsCount:  %d\n", logEntry.ArgsCount)
	// SQL:       INSERT INTO `users` (`email`, `name`, `password`) VALUES (?, ?, ?)
	// ArgsCount: 3
}

func debugFromBuilder() {
	q := sqlbuilder.Select("id", "name").
		From("orders").
		Where(sqlbuilder.Between("total", 10.0, 500.0))

	// Debug() combines Build() + DebugSQL() in one call.
	// WARNING: Do NOT execute the returned string.
	debug, err := sqlbuilder.Debug(q)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n=== Debug() helper (logging only — NEVER execute) ===")
	fmt.Println(debug)
	// SELECT `id`, `name`
	// FROM `orders`
	// WHERE `total` BETWEEN 10 AND 500
}
