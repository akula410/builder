// Example 07_json demonstrates MySQL 8 JSON helpers:
// JSONExtract, JSONUnquote, JSONEq, JSONContains.
//
// These helpers generate parameterised SQL for JSON_EXTRACT, JSON_UNQUOTE,
// and JSON_CONTAINS — values are always passed as placeholders, never concatenated.
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/07_json
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
	// JSONExtract as a SELECT column expression.
	q1 := sqlbuilder.Select("id", sqlbuilder.JSONExtract("data", "$.status")).
		From("orders")
	printQuery("JSONExtract as column", q1)
	// SELECT `id`, JSON_EXTRACT(`data`, ?) FROM `orders`
	// args: ["$.status"]

	// JSONUnquote unwraps the JSON string literal returned by JSON_EXTRACT.
	q2 := sqlbuilder.Select("id", sqlbuilder.JSONUnquote("data", "$.name")).
		From("orders")
	printQuery("JSONUnquote as column", q2)
	// SELECT `id`, JSON_UNQUOTE(JSON_EXTRACT(`data`, ?)) FROM `orders`

	// JSONEq — filter by equality on a JSON field.
	q3 := sqlbuilder.Select("id").
		From("orders").
		Where(sqlbuilder.JSONEq("data", "$.company_id", "ngus"))
	printQuery("JSONEq (WHERE)", q3)
	// WHERE JSON_UNQUOTE(JSON_EXTRACT(`data`, ?)) = ?
	// args: ["$.company_id", "ngus"]

	// JSONContains — check that a JSON array contains a scalar value.
	q4 := sqlbuilder.Select("id").
		From("orders").
		Where(sqlbuilder.JSONContains("data", "$.tags", "urgent"))
	printQuery("JSONContains (WHERE)", q4)
	// WHERE JSON_CONTAINS(JSON_EXTRACT(`data`, ?), JSON_QUOTE(?))
	// args: ["$.tags", "urgent"]

	// Combining JSON conditions with regular conditions.
	q5 := sqlbuilder.Select("id", "email").
		From("users").
		Where(sqlbuilder.And(
			sqlbuilder.Eq("status", "active"),
			sqlbuilder.JSONEq("meta", "$.tier", "premium"),
		))
	printQuery("JSON + regular condition", q5)
}

func runnableExamples(exec *sqlbuilder.Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Example: fetch users where JSON field meta.tier = "premium".
	// Assumes column `meta` of type JSON exists on the users table.
	q := sqlbuilder.Select("id").
		From("users").
		Where(sqlbuilder.JSONEq("meta", "$.tier", "premium")).
		Limit(5)

	rows, err := exec.QueryContext(ctx, q)
	if err != nil {
		log.Printf("query: %v (table may not have a meta column)", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Premium users (JSON filter) ===")
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Printf("scan: %v", err)
			return
		}
		fmt.Printf("  id=%d\n", id)
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
