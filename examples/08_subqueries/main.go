// Example 08_subqueries demonstrates subquery support:
// WHERE IN subquery, EXISTS / NOT EXISTS, FROM subquery, SubqueryColumn.
//
// Run:
//
//	MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example go run ./examples/08_subqueries
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
	// WHERE IN subquery — users who have at least one high-value order.
	q1 := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.InSubquery("id",
			sqlbuilder.Select("user_id").
				From("orders").
				Where(sqlbuilder.Gt("total", 100)),
		))
	printQuery("WHERE IN subquery", q1)
	// WHERE `id` IN (
	//     SELECT `user_id` FROM `orders` WHERE `total` > ?
	// )
	// args: [100]

	// WHERE NOT IN subquery.
	q2 := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.NotInSubquery("id",
			sqlbuilder.Select("user_id").From("orders"),
		))
	printQuery("WHERE NOT IN subquery", q2)

	// WHERE EXISTS — correlated subquery.
	// NOTE: RawCondition is used here to write the correlated reference o.user_id = u.id.
	// This is a known, hardcoded expression — never pass user input here.
	q3 := sqlbuilder.Select("id", "name").
		From("users u").
		Where(sqlbuilder.Exists(
			sqlbuilder.Select(sqlbuilder.Raw("1")).
				From("orders o").
				Where(sqlbuilder.And(
					sqlbuilder.RawCondition("o.user_id = u.id"),
					sqlbuilder.Gt("o.total", 500),
				)),
		))
	printQuery("WHERE EXISTS (correlated)", q3)

	// WHERE NOT EXISTS.
	q4 := sqlbuilder.Select("id", "name").
		From("users u").
		Where(sqlbuilder.NotExists(
			sqlbuilder.Select(sqlbuilder.Raw("1")).
				From("orders o").
				Where(sqlbuilder.RawCondition("o.user_id = u.id")),
		))
	printQuery("WHERE NOT EXISTS", q4)

	// FROM subquery — aggregate in a derived table.
	q5 := sqlbuilder.Select("t.user_id", "t.total").
		FromSubquery(
			sqlbuilder.Select("user_id", sqlbuilder.Raw("SUM(total) AS total")).
				From("orders").
				GroupBy("user_id"),
			"t",
		).
		Where(sqlbuilder.Gt("t.total", 1000))
	printQuery("FROM subquery", q5)

	// SubqueryColumn — scalar subquery as a SELECT column.
	q6 := sqlbuilder.Select(
		"u.id",
		"u.name",
		sqlbuilder.SubqueryColumn(
			sqlbuilder.Select(sqlbuilder.Raw("COUNT(*)")).
				From("orders o").
				Where(sqlbuilder.RawCondition("o.user_id = u.id")),
			"orders_count",
		),
	).From("users u")
	printQuery("SubqueryColumn", q6)
}

func runnableExamples(exec *sqlbuilder.Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := sqlbuilder.Select("id", "name").
		From("users u").
		Where(sqlbuilder.Exists(
			sqlbuilder.Select(sqlbuilder.Raw("1")).
				From("orders o").
				Where(sqlbuilder.RawCondition("o.user_id = u.id")),
		)).
		OrderBy("u.id", sqlbuilder.Asc).
		Limit(5)

	rows, err := exec.QueryContext(ctx, q)
	if err != nil {
		log.Printf("query: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Users with at least one order ===")
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
}

func printQuery(name string, b sqlbuilder.QueryBuilder) {
	sql, args, err := b.Build()
	if err != nil {
		log.Fatalf("%s: %v", name, err)
	}
	fmt.Printf("=== %s ===\n%s\nargs: %v\n\n", name, sql, args)
}
