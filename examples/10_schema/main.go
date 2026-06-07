// Example 10_schema demonstrates the Schema / DDL builder:
// CREATE DATABASE, CREATE TABLE, indexes, foreign keys,
// ALTER TABLE, DROP TABLE, TRUNCATE, RENAME, and Migration.
//
// All DDL is built and printed by default (no DB required).
// To execute the migration against a live database set EXECUTE_DDL=true.
//
// WARNING: DROP TABLE, DROP DATABASE, and TRUNCATE are destructive and irreversible.
// MySQL DDL causes an implicit COMMIT — DDL cannot be rolled back inside a transaction.
// Always back up your database before running migrations.
//
// Run (build only, no DB):
//
//	go run ./examples/10_schema
//
// Run (execute against live DB):
//
//	EXECUTE_DDL=true MYSQL_USER=root MYSQL_PASSWORD=secret MYSQL_DATABASE=builder_example \
//	    go run ./examples/10_schema
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	sqlbuilder "github.com/akula410/builder/v2"
	dbhelper "github.com/akula410/builder/v2/examples/internal/db"
)

func main() {
	fmt.Println("=== Schema Builder — build-only output ===")
	fmt.Println()
	createDB()
	createTables()
	alterTableExamples()
	dropRenameExamples()
	migrationBuildOnly()

	if os.Getenv("EXECUTE_DDL") != "true" {
		fmt.Println("\n--- DDL not executed. Set EXECUTE_DDL=true to run against a live DB. ---")
		return
	}

	if !dbhelper.SkipIfNoDatabaseConfig() {
		return
	}
	db := dbhelper.MustOpen()
	if db == nil {
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exec := sqlbuilder.NewExecutor(db)

	m := buildMigration()
	fmt.Println("\n=== Running migration Up ===")
	if err := m.RunUp(ctx, exec); err != nil {
		log.Printf("RunUp: %v", err)
		return
	}
	fmt.Println("Migration Up executed successfully.")
}

func createDB() {
	printDDL("CREATE DATABASE", sqlbuilder.CreateDatabase("builder_example").
		IfNotExists().
		CharacterSet("utf8mb4").
		Collate("utf8mb4_unicode_ci"))

	printDDL("USE DATABASE", sqlbuilder.UseDatabase("builder_example"))
}

func createTables() {
	fk := sqlbuilder.ForeignKey("fk_orders_user", []string{"user_id"}).
		References("users", []string{"id"}).
		OnDelete(sqlbuilder.Cascade).
		OnUpdate(sqlbuilder.Restrict)

	printDDL("CREATE TABLE users", sqlbuilder.CreateTable("users").
		IfNotExists().
		Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
		Column(sqlbuilder.Column("email", "VARCHAR(255)").NotNull()).
		Column(sqlbuilder.Column("name", "VARCHAR(100)").NotNull()).
		Column(sqlbuilder.Column("status", "VARCHAR(20)").NotNull().Default("active")).
		Column(sqlbuilder.Column("meta", "JSON").Nullable()).
		Column(sqlbuilder.Column("created_at", "DATETIME").NotNull().DefaultRaw("CURRENT_TIMESTAMP")).
		PrimaryKey("id").
		UniqueIndex("uniq_users_email", "email").
		Index("idx_users_status", "status").
		Engine("InnoDB").
		CharacterSet("utf8mb4").
		Collate("utf8mb4_unicode_ci"))

	printDDL("CREATE TABLE orders", sqlbuilder.CreateTable("orders").
		IfNotExists().
		Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
		Column(sqlbuilder.Column("user_id", "BIGINT UNSIGNED").NotNull()).
		Column(sqlbuilder.Column("total", "DECIMAL(10,2)").NotNull()).
		Column(sqlbuilder.Column("status", "VARCHAR(20)").NotNull().Default("pending")).
		Column(sqlbuilder.Column("created_at", "DATETIME").NotNull().DefaultRaw("CURRENT_TIMESTAMP")).
		PrimaryKey("id").
		Index("idx_orders_user_id", "user_id").
		ForeignKey(fk).
		Engine("InnoDB").
		CharacterSet("utf8mb4").
		Collate("utf8mb4_unicode_ci"))
}

func alterTableExamples() {
	printDDL("ADD COLUMN", sqlbuilder.AlterTable("users").
		AddColumn(sqlbuilder.Column("phone", "VARCHAR(32)").Nullable()))

	printDDL("MODIFY COLUMN", sqlbuilder.AlterTable("users").
		ModifyColumn(sqlbuilder.Column("phone", "VARCHAR(64)").Nullable()))

	printDDL("DROP COLUMN", sqlbuilder.AlterTable("users").DropColumn("phone"))

	printDDL("ADD INDEX", sqlbuilder.AlterTable("users").AddIndex("idx_users_name", "name"))

	printDDL("ADD UNIQUE INDEX", sqlbuilder.AlterTable("users").AddUniqueIndex("uniq_users_name", "name"))

	printDDL("DROP INDEX", sqlbuilder.AlterTable("users").DropIndex("idx_users_name"))

	printDDL("ADD FULLTEXT INDEX", sqlbuilder.AlterTable("users").
		AddFullTextIndex("ft_users_name", "name"))

	printDDL("MULTIPLE ALTER OPS", sqlbuilder.AlterTable("users").
		DropColumn("phone").
		AddColumn(sqlbuilder.Column("avatar_url", "VARCHAR(512)").Nullable()))
}

func dropRenameExamples() {
	printDDL("TRUNCATE TABLE", sqlbuilder.TruncateTable("logs"))

	printDDL("RENAME TABLE", sqlbuilder.RenameTable("users_old", "users_archive"))

	printDDL("RENAME TABLES (multi)", sqlbuilder.RenameTables(
		sqlbuilder.TableRename("users_old", "users_archive"),
		sqlbuilder.TableRename("orders_old", "orders_archive"),
	))

	printDDL("DROP TABLE IF EXISTS", sqlbuilder.DropTable("tmp_export").IfExists())
}

func migrationBuildOnly() {
	m := buildMigration()
	fmt.Println("=== Migration.BuildUp() ===")
	steps, err := m.BuildUp()
	if err != nil {
		log.Fatal(err)
	}
	for i, s := range steps {
		fmt.Printf("Up step %d:\n%s\n\n", i+1, s.SQL)
	}

	stepsDown, _ := m.BuildDown()
	fmt.Println("=== Migration.BuildDown() ===")
	for i, s := range stepsDown {
		fmt.Printf("Down step %d:\n%s\n\n", i+1, s.SQL)
	}
}

func buildMigration() *sqlbuilder.Migration {
	return sqlbuilder.NewMigration("20260607_create_users").
		Up(
			sqlbuilder.CreateTable("users").
				IfNotExists().
				Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
				Column(sqlbuilder.Column("email", "VARCHAR(255)").NotNull()).
				Column(sqlbuilder.Column("name", "VARCHAR(100)").NotNull()).
				Column(sqlbuilder.Column("status", "VARCHAR(20)").NotNull().Default("active")).
				Column(sqlbuilder.Column("created_at", "DATETIME").NotNull().DefaultRaw("CURRENT_TIMESTAMP")).
				PrimaryKey("id").
				UniqueIndex("uniq_users_email", "email").
				Engine("InnoDB").
				CharacterSet("utf8mb4").
				Collate("utf8mb4_unicode_ci"),
			sqlbuilder.CreateTable("orders").
				IfNotExists().
				Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
				Column(sqlbuilder.Column("user_id", "BIGINT UNSIGNED").NotNull()).
				Column(sqlbuilder.Column("total", "DECIMAL(10,2)").NotNull()).
				Column(sqlbuilder.Column("status", "VARCHAR(20)").NotNull().Default("pending")).
				Column(sqlbuilder.Column("created_at", "DATETIME").NotNull().DefaultRaw("CURRENT_TIMESTAMP")).
				PrimaryKey("id").
				Index("idx_orders_user_id", "user_id").
				Engine("InnoDB").
				CharacterSet("utf8mb4").
				Collate("utf8mb4_unicode_ci"),
		).
		Down(
			sqlbuilder.DropTable("orders").IfExists(),
			sqlbuilder.DropTable("users").IfExists(),
		)
}

func printDDL(name string, b sqlbuilder.QueryBuilder) {
	sql, _, err := b.Build()
	if err != nil {
		log.Fatalf("%s: %v", name, err)
	}
	fmt.Printf("=== %s ===\n%s\n\n", name, sql)
}
