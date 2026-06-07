// Package db provides a shared MySQL connection helper for builder examples.
// It reads connection settings from environment variables and uses
// github.com/akula410/connect/v2 to open a *sql.DB.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/akula410/connect/v2"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// HasEnvOverride returns true when at least one MySQL env var is explicitly set.
// Examples use this to decide whether to attempt a real connection.
func HasEnvOverride() bool {
	return os.Getenv("MYSQL_HOST") != "" ||
		os.Getenv("MYSQL_USER") != "" ||
		os.Getenv("MYSQL_PASSWORD") != "" ||
		os.Getenv("MYSQL_DATABASE") != ""
}

// SkipIfNoDatabaseConfig prints a usage hint and returns false when no MySQL env
// vars are set. Call this at the start of any example that requires a live DB.
func SkipIfNoDatabaseConfig() bool {
	if !HasEnvOverride() {
		fmt.Println("Skipping: no MySQL environment variables are set.")
		fmt.Println("To run this example, set:")
		fmt.Println("  MYSQL_HOST     (default: 127.0.0.1)")
		fmt.Println("  MYSQL_PORT     (default: 3306)")
		fmt.Println("  MYSQL_USER     (default: root)")
		fmt.Println("  MYSQL_PASSWORD")
		fmt.Println("  MYSQL_DATABASE (default: builder_example)")
		return false
	}
	return true
}

// Open returns a *sql.DB configured from environment variables.
//
// Defaults:
//
//	MYSQL_HOST     = 127.0.0.1
//	MYSQL_PORT     = 3306
//	MYSQL_USER     = root
//	MYSQL_DATABASE = builder_example
//
// The password is read from MYSQL_PASSWORD and is never logged.
func Open() (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := connect.DefaultConfig()
	cfg.User = getenv("MYSQL_USER", "root")
	cfg.Password = os.Getenv("MYSQL_PASSWORD")
	cfg.Host = getenv("MYSQL_HOST", "127.0.0.1")
	cfg.Port = getenv("MYSQL_PORT", "3306")
	cfg.DBName = getenv("MYSQL_DATABASE", "builder_example")

	// Pool settings — tuned for examples; adjust for production workloads.
	cfg.MaxOpenConns = 10
	cfg.MaxIdleConns = 10
	cfg.ConnMaxLifetime = 5 * time.Minute
	cfg.ConnMaxIdleTime = 2 * time.Minute

	return connect.NewMySQLContext(ctx, cfg)
}

// MustOpen opens a MySQL connection. On failure it prints an error message and
// returns nil (the caller must check for nil before using the result).
func MustOpen() *sql.DB {
	db, err := Open()
	if err != nil {
		fmt.Printf("connect: %v\n", err)
		fmt.Println("Check your MySQL env vars and ensure the server is running.")
		return nil
	}
	return db
}
