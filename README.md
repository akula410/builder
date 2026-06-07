# builder

Safe MySQL 8 SQL builder for Go 1.26.4.

Generates parameterised SQL using `?` placeholders and `[]any` args. Identifiers (table names, column names, index names) are validated and quoted with backticks. Raw SQL requires an explicit `Raw()` / `RawCondition()` call.

**This is not an ORM.** There is no struct-to-table mapping. The library helps you write SQL safely — nothing more.

---

## What the library does NOT do

- No struct ↔ table mapping
- No automatic schema introspection
- No connection pooling (use `database/sql`)
- No query result scanning helpers (use `rows.Scan` directly)
- No migration history table or applied-migration tracking (implement in your application)
- `DebugSQL` output is **never** executed — it is for logging only

---

## Installation

```bash
go get github.com/akula410/builder/v2
```

Import the package:

```go
import sqlbuilder "github.com/akula410/builder/v2"
```

---

## Database connection

`mysqli` and `PDO` are PHP tools. For Go use the standard `database/sql` package together with a MySQL driver.

```bash
go get github.com/go-sql-driver/mysql
```

```go
import (
    "context"
    "database/sql"
    "time"

    _ "github.com/go-sql-driver/mysql"
    sqlbuilder "github.com/akula410/builder/v2"
)

func main() {
    dsn := "user:password@tcp(127.0.0.1:3306)/dbname?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"

    db, err := sql.Open("mysql", dsn)
    if err != nil { panic(err) }
    defer db.Close()

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)

    ctx := context.Background()
    if err := db.PingContext(ctx); err != nil { panic(err) }

    exec := sqlbuilder.NewExecutor(db)

    rows, err := exec.QueryContext(ctx,
        sqlbuilder.Select("id", "name").
            From("users").
            Where(sqlbuilder.Eq("status", "active")),
    )
    if err != nil { panic(err) }
    defer rows.Close()
}
```

---

## Quick start

```go
// Build only — does not execute
sql, args, err := sqlbuilder.Select("id", "name", "email").
    From("users").
    Where(sqlbuilder.Eq("status", "active")).
    OrderBy("id", sqlbuilder.Desc).
    Limit(10).
    Build()

// Execute via Executor
exec := sqlbuilder.NewExecutor(db)
rows, err := exec.QueryContext(ctx, query)

// Debug representation (NEVER execute this)
debug, err := sqlbuilder.Debug(query)
```

---

## Thread-safety

Builder instances are **mutable** and are **not safe for concurrent use**.

Do not share the same builder instance between goroutines unless you protect it with
external synchronization. Reusing a mutable builder from multiple goroutines is a data race.

**Correct approach — create a new builder per query / request / goroutine:**

```go
func handleRequest(userID int) error {
    q := sqlbuilder.Select("id", "name").
        From("users").
        Where(sqlbuilder.Eq("id", userID))
    // q is local to this goroutine — safe
    rows, err := exec.QueryContext(ctx, q)
    // ...
}
```

**Do NOT do this — shared mutable builder:**

```go
// WRONG: global shared builder — data race under concurrent use
var base = sqlbuilder.Select("id").From("users")

func handleRequest(status string) {
    base.Where(sqlbuilder.Eq("status", status)) // DATA RACE
}
```

**After Build():** the returned `sql string` and `args []any` are plain Go values. They can be
passed between goroutines freely as long as `args` contains only immutable values (strings,
integers, etc.) and you do not mutate the slice itself.

---

## SELECT

```go
sql, args, err := sqlbuilder.Select("id", "name", "email").
    From("users").
    Where(sqlbuilder.Eq("status", "active")).
    Where(sqlbuilder.Gt("created_at", "2025-01-01")).
    OrderBy("id", sqlbuilder.Desc).
    Limit(10).
    Offset(20).
    Build()
// SELECT `id`, `name`, `email`
// FROM `users`
// WHERE `status` = ? AND `created_at` > ?
// ORDER BY `id` DESC
// LIMIT 10 OFFSET 20
// args: ["active", "2025-01-01"]
```

### DISTINCT

```go
sqlbuilder.Select("status").Distinct().From("users")
```

### Aliases and qualified columns

```go
sqlbuilder.Select("u.id", "u.name").
    From("users u").
    LeftJoin("orders o", "o.user_id = u.id").
    Where(sqlbuilder.Eq("u.status", "active"))
// SELECT `u`.`id`, `u`.`name`
// FROM `users` AS `u`
// LEFT JOIN `orders` AS `o` ON o.user_id = u.id
// WHERE `u`.`status` = ?
```

### Raw expression as column

```go
sqlbuilder.Select("user_id", sqlbuilder.Raw("SUM(total) AS total")).
    From("orders").
    GroupBy("user_id")
```

### GROUP BY / HAVING

```go
sqlbuilder.Select("status", sqlbuilder.Raw("COUNT(*) AS cnt")).
    From("users").
    GroupBy("status").
    Having(sqlbuilder.Gt("cnt", 5))
```

---

## INSERT

```go
// Single row from map (columns sorted alphabetically for stable output)
sql, args, err := sqlbuilder.InsertInto("users").
    Values(map[string]any{
        "name":  "Alex",
        "email": "alex@test.com",
    }).
    Build()
// INSERT INTO `users` (`email`, `name`) VALUES (?, ?)

// Bulk insert
sqlbuilder.InsertInto("users").
    Columns("name", "email").
    Rows(
        []any{"Alex", "alex@test.com"},
        []any{"Ivan", "ivan@test.com"},
    )
// INSERT INTO `users` (`name`, `email`) VALUES (?, ?), (?, ?)

// ON DUPLICATE KEY UPDATE
sqlbuilder.InsertInto("users").
    Values(map[string]any{"id": 1, "name": "Alex"}).
    OnDuplicateKeyUpdate(map[string]any{"name": "Alex Updated"})
// INSERT INTO `users` (`id`, `name`) VALUES (?, ?)
// ON DUPLICATE KEY UPDATE `name` = ?
```

---

## UPDATE

```go
sqlbuilder.Update("users").
    Set("name", "Alex").
    Set("status", "active").
    Where(sqlbuilder.Eq("id", 10)).
    Limit(1)
// UPDATE `users`
// SET `name` = ?, `status` = ?
// WHERE `id` = ?
// LIMIT 1
```

---

## DELETE

```go
sqlbuilder.DeleteFrom("users").
    Where(sqlbuilder.Eq("id", 10)).
    Limit(1)
// DELETE FROM `users`
// WHERE `id` = ?
// LIMIT 1
```

---

## Conditions API

| Function | SQL |
|---|---|
| `Eq("f", v)` | `` `f` = ? `` |
| `Neq("f", v)` | `` `f` != ? `` |
| `Gt("f", v)` | `` `f` > ? `` |
| `Gte("f", v)` | `` `f` >= ? `` |
| `Lt("f", v)` | `` `f` < ? `` |
| `Lte("f", v)` | `` `f` <= ? `` |
| `Like("f", v)` | `` `f` LIKE ? `` |
| `NotLike("f", v)` | `` `f` NOT LIKE ? `` |
| `In("f", []any{…})` | `` `f` IN (?, ?, …) `` |
| `NotIn("f", []any{…})` | `` `f` NOT IN (?, ?, …) `` |
| `IsNull("f")` | `` `f` IS NULL `` |
| `IsNotNull("f")` | `` `f` IS NOT NULL `` |
| `Between("f", a, b)` | `` `f` BETWEEN ? AND ? `` |
| `And(c1, c2, …)` | `(c1 AND c2 …)` |
| `Or(c1, c2, …)` | `(c1 OR c2 …)` |
| `RawCondition(sql, args…)` | raw SQL ⚠️ |

```go
sqlbuilder.And(
    sqlbuilder.Eq("status", "active"),
    sqlbuilder.Or(
        sqlbuilder.Gt("price", 100),
        sqlbuilder.IsNull("price"),
    ),
)
// (`status` = ? AND (`price` > ? OR `price` IS NULL))
```

---

## JOIN

### Safe JOIN API (recommended)

Use `JoinOn` / `LeftJoinOn` / `RightJoinOn` / `InnerJoinOn` with typed ON condition helpers.
Column references are validated against the identifier whitelist and backtick-quoted.
No raw SQL is accepted; no value placeholders are generated.

```go
// INNER JOIN with a safe ON condition
sqlbuilder.Select("u.id", "u.name", "o.total").
    From("users u").
    JoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id"))
// INNER JOIN `orders` AS `o` ON `o`.`user_id` = `u`.`id`

// LEFT JOIN
sqlbuilder.Select("u.id", "u.name").
    From("users u").
    LeftJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id"))

// Compound ON condition
sqlbuilder.Select("u.id").
    From("users u").
    LeftJoinOn("orders o", sqlbuilder.OnAnd(
        sqlbuilder.OnEq("o.user_id", "u.id"),
        sqlbuilder.OnEq("o.status", "u.status"),
    ))
// ON (`o`.`user_id` = `u`.`id` AND `o`.`status` = `u`.`status`)
```

**ON condition helpers:**

| Function | SQL |
|---|---|
| `OnEq(left, right)` | `left = right` |
| `OnNe(left, right)` | `left != right` |
| `OnGt(left, right)` | `left > right` |
| `OnGte(left, right)` | `left >= right` |
| `OnLt(left, right)` | `left < right` |
| `OnLte(left, right)` | `left <= right` |
| `OnAnd(conds…)` | `(c1 AND c2 …)` |
| `OnOr(conds…)` | `(c1 OR c2 …)` |

Supported column formats: `column`, `table.column` (both parts validated and quoted).
Three-part dotted identifiers (`db.table.column`) are not supported and return an error.

### Raw JOIN API (use only for trusted, hardcoded ON expressions)

```go
.JoinRaw("orders o", "o.user_id = u.id")         // INNER JOIN — raw ON
.LeftJoinRaw("orders o", "o.user_id = u.id")     // LEFT JOIN  — raw ON
.RightJoinRaw("orders o", "o.user_id = u.id")    // RIGHT JOIN — raw ON
.InnerJoinRaw("orders o", "o.user_id = u.id")    // INNER JOIN — raw ON (alias)
```

> **WARNING — raw JOIN ON clause is embedded into SQL as-is, without validation or escaping.**
> Only pass hardcoded, trusted column comparison expressions (e.g. `"o.user_id = u.id"`).
> Never pass user-controlled input as the `on` argument — it is a SQL injection risk.
> To filter by a value, use a `Where` condition instead.

### Deprecated raw variants (kept for backward compatibility)

`Join`, `LeftJoin`, `RightJoin`, `InnerJoin` (raw `on string` parameter) remain functional
but are deprecated. Prefer `JoinOn` / `LeftJoinOn` / `JoinRaw` / `LeftJoinRaw`.

```go
// Deprecated — use JoinRaw or JoinOn instead
.Join("orders o", "o.user_id = u.id")
.LeftJoin("orders o", "o.user_id = u.id")
```

---

## MySQL 8 JSON helpers

```go
// JSON_EXTRACT as column expression
sqlbuilder.JSONExtract("data", "$.status")
// JSON_EXTRACT(`data`, ?)   args: ["$.status"]

// JSON_EXTRACT used as a SELECT column
sqlbuilder.Select("id", sqlbuilder.JSONExtract("data", "$.status")).From("orders")
// SELECT `id`, JSON_EXTRACT(`data`, ?) FROM `orders`

// JSON_UNQUOTE(JSON_EXTRACT(...))
sqlbuilder.JSONUnquote("data", "$.name")

// Equality condition
sqlbuilder.Select("id").From("orders").
    Where(sqlbuilder.JSONEq("data", "$.company_id", "ngus"))
// WHERE JSON_UNQUOTE(JSON_EXTRACT(`data`, ?)) = ?

// JSON_CONTAINS
sqlbuilder.Select("id").From("orders").
    Where(sqlbuilder.JSONContains("data", "$.tags", "urgent"))
// WHERE JSON_CONTAINS(JSON_EXTRACT(`data`, ?), JSON_QUOTE(?))
```

---

## Subqueries

```go
// WHERE IN subquery
sqlbuilder.Select("id", "name").
    From("users").
    Where(sqlbuilder.InSubquery("id",
        sqlbuilder.Select("user_id").From("orders").Where(sqlbuilder.Gt("total", 100)),
    ))

// WHERE EXISTS
sqlbuilder.Select("id", "name").
    From("users u").
    Where(sqlbuilder.Exists(
        sqlbuilder.Select("1").From("orders o").
            Where(sqlbuilder.RawCondition("o.user_id = u.id")),
    ))

// NOT EXISTS
sqlbuilder.NotExists(subquery)

// FROM subquery
sqlbuilder.Select("t.user_id", "t.total").
    FromSubquery(
        sqlbuilder.Select("user_id", sqlbuilder.Raw("SUM(total) AS total")).
            From("orders").GroupBy("user_id"),
        "t",
    )

// Subquery as column
sqlbuilder.Select(
    "u.id",
    "u.name",
    sqlbuilder.SubqueryColumn(
        sqlbuilder.Select(sqlbuilder.Raw("COUNT(*)")).From("orders o").
            Where(sqlbuilder.RawCondition("o.user_id = u.id")),
        "orders_count",
    ),
).From("users u")
```

Args from subqueries merge in SQL order: subquery args first, then outer query args.

---

## WITH / CTE

```go
// Single CTE
sqlbuilder.
    With("active_users",
        sqlbuilder.Select("id", "name").From("users").Where(sqlbuilder.Eq("status", "active")),
    ).
    Select("au.id", "au.name").
    From("active_users au")
// WITH `active_users` AS (
//     SELECT `id`, `name` FROM `users` WHERE `status` = ?
// )
// SELECT `au`.`id`, `au`.`name`
// FROM `active_users` AS `au`

// Multiple CTEs
sqlbuilder.
    With("active_users", q1).
    With("paid_orders", q2).
    Select("au.id", "po.total").
    From("active_users au").
    Join("paid_orders po", "po.user_id = au.id")
```

### WITH RECURSIVE

```go
sqlbuilder.
    WithRecursive("category_tree",
        sqlbuilder.RawQuery(`
            SELECT id, parent_id, name FROM categories WHERE id = ?
            UNION ALL
            SELECT c.id, c.parent_id, c.name
            FROM categories c
            INNER JOIN category_tree ct ON c.parent_id = ct.id
        `, 10),
    ).
    Select("*").
    From("category_tree")
```

CTE args always precede the main query args.

---

## Schema Builder / DDL Builder

The Schema Builder generates DDL SQL strings. It does **not** execute SQL.
All identifiers are validated and backtick-quoted. Identifiers cannot be parameterised in MySQL DDL.

### CREATE DATABASE / DROP DATABASE

```go
sqlbuilder.CreateDatabase("app_db").
    IfNotExists().
    CharacterSet("utf8mb4").
    Collate("utf8mb4_unicode_ci")
// CREATE DATABASE IF NOT EXISTS `app_db` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci

sqlbuilder.DropDatabase("app_db").IfExists()
// DROP DATABASE IF EXISTS `app_db`

sqlbuilder.UseDatabase("app_db")
// USE `app_db`
```

### CREATE TABLE

```go
sqlbuilder.CreateTable("users").
    IfNotExists().
    Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
    Column(sqlbuilder.Column("email", "VARCHAR(255)").NotNull()).
    Column(sqlbuilder.Column("created_at", "DATETIME").NotNull().DefaultRaw("CURRENT_TIMESTAMP")).
    PrimaryKey("id").
    UniqueIndex("uniq_users_email", "email").
    Engine("InnoDB").
    CharacterSet("utf8mb4").
    Collate("utf8mb4_unicode_ci")
```

Column modifiers: `NotNull()`, `Nullable()`, `AutoIncrement()`, `Default("val")`, `DefaultRaw("expr")`,
`Comment("text")`, `First()`, `After("col")`.

### DROP TABLE / TRUNCATE TABLE / RENAME TABLE

```go
sqlbuilder.DropTable("users").IfExists()
sqlbuilder.DropTable("old_users", "old_orders").IfExists()  // multi-table
sqlbuilder.TruncateTable("logs")
sqlbuilder.RenameTable("users_old", "users_archive")
sqlbuilder.RenameTables(
    sqlbuilder.TableRename("users_old", "users_archive"),
    sqlbuilder.TableRename("orders_old", "orders_archive"),
)
```

### ALTER TABLE

```go
// Add column
sqlbuilder.AlterTable("users").
    AddColumn(sqlbuilder.Column("phone", "VARCHAR(32)").Nullable())

// Drop column
sqlbuilder.AlterTable("users").DropColumn("phone")

// Modify column
sqlbuilder.AlterTable("users").
    ModifyColumn(sqlbuilder.Column("phone", "VARCHAR(64)").Nullable())

// Change column (rename + redefine)
sqlbuilder.AlterTable("users").
    ChangeColumn("old_phone", sqlbuilder.Column("phone", "VARCHAR(64)").Nullable())

// Multiple operations in one statement
sqlbuilder.AlterTable("users").
    DropColumn("phone").
    DropColumn("middle_name")
```

### Indexes

```go
sqlbuilder.AlterTable("users").AddIndex("idx_users_email", "email")
sqlbuilder.AlterTable("users").AddUniqueIndex("uniq_users_email", "email")
sqlbuilder.AlterTable("orders").AddIndex("idx_orders_user_status", "user_id", "status")
sqlbuilder.AlterTable("products").AddFullTextIndex("ft_products_name", "name", "description")
sqlbuilder.AlterTable("locations").AddSpatialIndex("sp_locations_point", "point")
sqlbuilder.AlterTable("users").DropIndex("idx_users_email")

// With prefix length
sqlbuilder.AlterTable("products").
    AddIndexColumns("idx_products_name", sqlbuilder.IndexColumn("name").Length(100))
```

### PRIMARY KEY and FOREIGN KEY

```go
fk := sqlbuilder.ForeignKey("fk_orders_user", []string{"user_id"}).
    References("users", []string{"id"}).
    OnDelete(sqlbuilder.Cascade).
    OnUpdate(sqlbuilder.Restrict)

sqlbuilder.CreateTable("orders").
    Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
    Column(sqlbuilder.Column("user_id", "BIGINT UNSIGNED").NotNull()).
    PrimaryKey("id").
    ForeignKey(fk)
```

Allowed referential actions: `Restrict`, `Cascade`, `SetNull`, `NoAction`, `SetDefault`.

---

## Migration Builder

`Migration` groups a set of `Up` (apply) and `Down` (rollback) SQL builders under a named identifier.

### What the Migration Builder provides

- `NewMigration(name string)` — creates a named migration.
- `Up(builders ...QueryBuilder)` / `Down(builders ...QueryBuilder)` — register steps.
- `BuildUp() ([]BuiltQuery, error)` / `BuildDown() ([]BuiltQuery, error)` — build all steps to SQL without executing.
- `RunUp(ctx, exec)` / `RunDown(ctx, exec)` — execute steps in order via an `Executor`. On error, execution stops; the error includes the step index.

### What the Migration Builder does NOT provide

- **No migration history table.** The library does not record which migrations have been applied.
- **No duplicate-run protection.** Running the same migration twice is your application's responsibility.
- **No state or version tracking.** There is no concept of a "current schema version".
- **No file discovery or ordering.** You must load and sequence `Migration` objects yourself.

If you need state tracking, implement a `schema_migrations` table and check it before calling `RunUp`.

```go
migration := sqlbuilder.NewMigration("20260607_120000_create_users").
    Up(
        sqlbuilder.CreateTable("users").
            IfNotExists().
            Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
            Column(sqlbuilder.Column("email", "VARCHAR(255)").NotNull()).
            PrimaryKey("id").
            UniqueIndex("uniq_users_email", "email"),
    ).
    Down(
        sqlbuilder.DropTable("users").IfExists(),
    )

// Build without executing — inspect the generated SQL
steps, err := migration.BuildUp()
for _, s := range steps {
    fmt.Println(s.SQL)
}

// Execute up
exec := sqlbuilder.NewExecutor(db)
if err := migration.RunUp(ctx, exec); err != nil {
    log.Fatal(err)
}

// Execute down (rollback)
if err := migration.RunDown(ctx, exec); err != nil {
    log.Fatal(err)
}
```

> **Warning — MySQL DDL and transactions.**
> Most DDL statements (CREATE TABLE, ALTER TABLE, DROP TABLE, TRUNCATE) cause an implicit COMMIT
> in MySQL. They cannot be rolled back inside a transaction. Always back up your database before
> running destructive migrations.

---

## Query Executor

```go
exec := sqlbuilder.NewExecutor(db)   // *sql.DB
exec := sqlbuilder.NewExecutor(tx)   // *sql.Tx

// Execute (INSERT, UPDATE, DELETE)
result, err := exec.ExecContext(ctx, query)

// Query multiple rows
rows, err := exec.QueryContext(ctx, query)
if err != nil { return err }
defer rows.Close()

// Single row — recommended: build error returned explicitly
row, err := exec.QueryRowContextErr(ctx, query)
if err != nil { return err }  // build error
var id int
if err := row.Scan(&id); err != nil { return err }  // DB / scan error

// Single row — deprecated: build error is silently masked
row := exec.QueryRowContext(ctx, query)

// Build only (no execution)
sql, args, err := sqlbuilder.ToSQL(query)
```

### QueryRowContextErr vs QueryRowContext

`QueryRowContext` (deprecated) masks build errors by executing a dummy query. The caller
cannot distinguish a build failure from a real database error when calling `row.Scan()`.

`QueryRowContextErr` returns the build error explicitly:
- `(nil, err)` — Build() failed; the database is never called.
- `(row, nil)` — Build() succeeded; call `row.Scan()` to read the result.

### Prepared Statements

```go
stmt, args, err := exec.PrepareContext(ctx, query)
if err != nil { return err }
defer stmt.Close()
rows, err := stmt.QueryContext(ctx, args...)

// One-shot helpers (prepare + execute + close statement)
rows, err := exec.PreparedQueryContext(ctx, query)
result, err := exec.PreparedExecContext(ctx, query)
```

Prepared statements are most useful when the same parameterised SQL is executed many times with different args. Both `db.QueryContext(ctx, sql, args...)` and `stmt.QueryContext(ctx, args...)` are equally safe against SQL injection.

---

## Debug SQL

> **WARNING: Debug SQL output must NEVER be executed as real SQL.**
> It is for human-readable logging only. Use it in development / local environments.
> In production, log the parameterised SQL and args count — not the values.

```go
sql, args, err := query.Build()

// Simple substitution (NOT for execution)
debug := sqlbuilder.DebugSQL(sql, args)
// SELECT `id`, `name` FROM `users` WHERE `email` = 'alex@example.com'

// From a builder directly
debug, err := sqlbuilder.Debug(query)

// Production-safe log (no values, only count)
log, err := sqlbuilder.SafeLogQuery(query)
// SafeLog{SQL: "SELECT ... WHERE `email` = ?", ArgsCount: 1}
```

### DebugConfig

```go
cfg := sqlbuilder.DebugConfig{
    MaxStringLength: 200,
    ShowBytes:       false,
    TimeFormat:      "2006-01-02 15:04:05",
    RedactFields:    []string{"password", "token", "access_token", "api_key"},
    RedactValue:     "[REDACTED]",
}
result := sqlbuilder.DebugSQLWithConfig(sql, args, cfg)
```

`DebugSQL` type handling:

| Type | Output |
|---|---|
| `nil` | `NULL` |
| `bool` | `TRUE` / `FALSE` |
| `int`, `int64`, etc. | `42` |
| `float32`, `float64` | `9.99` |
| `string` | `'value'` (single quotes escaped) |
| `time.Time` | `'2025-01-15 12:00:00'` |
| `[]byte` (ShowBytes=true) | `0xABCD` |
| `[]byte` (ShowBytes=false) | `'[bytes N]'` |

---

## Safe Logging

```go
// Production — safe, no values exposed
log, err := sqlbuilder.SafeLogQuery(query)
// {"sql": "SELECT ... WHERE `email` = ?", "args_count": 1}

// Development — redact sensitive fields
cfg := sqlbuilder.DebugConfig{
    RedactFields: []string{"password", "token"},
    RedactValue:  "[REDACTED]",
}
result := sqlbuilder.DebugSQLWithConfig(sql, args, cfg)
// WHERE `email` = 'alex@example.com' AND `password` = '[REDACTED]'
```

> Debug SQL can contain personal data (emails, names, etc.). Never log it in production.

---

## Security notes

### Values always use placeholders

```go
// Safe — value goes into args
Where(sqlbuilder.Eq("email", userEmail))
// `email` = ?   args: [userEmail]

// Never do this
sqlbuilder.RawCondition("email = '" + userEmail + "'")  // SQL injection risk
```

### Identifiers are validated and quoted

Table names, column names, index names, and aliases are validated against `^[A-Za-z_][A-Za-z0-9_]{0,63}$` and wrapped in backticks.

### Raw SQL is explicit and visible

```go
// Safe — known SQL fragment with parameterised values
sqlbuilder.RawCondition("JSON_EXTRACT(`data`, ?) = ?", "$.status", "active")

// Unsafe — DO NOT do this
sqlbuilder.RawCondition("name = '" + userInput + "'")
sqlbuilder.Raw("SELECT * FROM " + tableName)
```

### IN with empty slice returns an error

```go
_, _, err := builder.Where(sqlbuilder.In("id", []any{})).Build()
// err: sqlbuilder: IN requires at least one value
```

### LIMIT and OFFSET accept only non-negative integers

```go
.Limit(-1)   // error: sqlbuilder: LIMIT must be >= 0
.Offset(-1)  // error: sqlbuilder: OFFSET must be >= 0
```

### Column types use a whitelist

Supported types: `INT`, `BIGINT`, `TINYINT`, `SMALLINT`, `DECIMAL(p,s)`, `VARCHAR(n)`, `CHAR(n)`,
`TEXT`, `MEDIUMTEXT`, `LONGTEXT`, `JSON`, `DATETIME`, `TIMESTAMP`, `DATE`, `TIME`, `BOOLEAN`,
`ENUM(…)`, `SET(…)`, `FLOAT`, `DOUBLE`, `BLOB`, `LONGBLOB`, and more.

`ENUM` and `SET` values must use single-quoted syntax:

```go
sqlbuilder.Column("status", "ENUM('active','inactive','pending')")
sqlbuilder.Column("flags",  "SET('read','write','exec')")
```

For non-standard types use `RawType()`:

```go
sqlbuilder.Column("geom", sqlbuilder.RawType("GEOMETRY"))
// WARNING: RawType bypasses validation — only use for known, trusted type strings.
```

---

## Raw SQL / Expr usage

| API | Purpose |
|---|---|
| `Raw(sql, args…)` | Raw SQL fragment as a column expression |
| `Expr(sql, args…)` | Alias for Raw |
| `RawCondition(sql, args…)` | Raw SQL as a WHERE condition |
| `RawQuery(sql, args…)` | Full raw query builder |
| `RawType(typ)` | Bypass column type validation ⚠️ |

All Raw APIs accept `args` with `?` placeholders. Never concatenate user input into the `sql` argument.

---

## DDL safety notes

- `DROP DATABASE`, `DROP TABLE`, `DROP COLUMN`, `TRUNCATE TABLE` are destructive and irreversible.
- Always back up your database before running destructive DDL.
- Migrations with DDL cannot be rolled back with transactions in MySQL (DDL causes implicit COMMIT).
- Never use DebugSQL output as input to Execute.

---

## Difference between SQL Builder and Schema Builder

| | SQL Builder | Schema Builder |
|---|---|---|
| Statements | SELECT, INSERT, UPDATE, DELETE | CREATE, ALTER, DROP, RENAME, TRUNCATE |
| Args | Yes (`?` placeholders) | Usually empty (DDL doesn't support placeholders for identifiers) |
| Types | — | Whitelist-validated |
| Transactions | Fully transactional | Implicit COMMIT in MySQL for most DDL |

---

## Using context (timeout / cancel)

### SELECT with timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

rows, err := exec.QueryContext(ctx,
    sqlbuilder.Select("id", "name").
        From("users").
        Where(sqlbuilder.Eq("status", "active")),
)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        return err
    }
}
return rows.Err()
```

### Single row with timeout and explicit build-error check

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

row, err := exec.QueryRowContextErr(ctx,
    sqlbuilder.Select("id", "email").
        From("users").
        Where(sqlbuilder.Eq("id", userID)),
)
if err != nil {
    return err  // build error — DB was never called
}
var id int
var email string
if err := row.Scan(&id, &email); err != nil {
    return err  // DB / scan error
}
```

### INSERT / UPDATE / DELETE with cancel

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

result, err := exec.ExecContext(ctx,
    sqlbuilder.Update("users").
        Set("status", "inactive").
        Where(sqlbuilder.Eq("id", userID)),
)
if err != nil {
    return err
}
```

### Transaction with context

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()  // no-op after Commit

exec := sqlbuilder.NewExecutor(tx)

_, err = exec.ExecContext(ctx,
    sqlbuilder.Update("accounts").
        Set("balance", newBalance).
        Where(sqlbuilder.Eq("id", accountID)),
)
if err != nil {
    return err
}

return tx.Commit()
```

---

## High-load usage

### Connection pool

The builder does not manage database connections. Configure the pool on `*sql.DB` directly:

```go
db.SetMaxOpenConns(50)          // maximum open connections to DB
db.SetMaxIdleConns(25)          // keep up to 25 idle connections
db.SetConnMaxLifetime(5 * time.Minute)  // recycle connections after 5 minutes
db.SetConnMaxIdleTime(2 * time.Minute)  // close idle connections after 2 minutes
```

Tune these values based on your MySQL / MariaDB settings, workload, and infrastructure.
Start conservative, then raise `MaxOpenConns` based on observed `max_used_connections` in MySQL
and benchmark results. There is no universal "correct" value.

### Prepared statements

The builder generates SQL + args; it does not automatically use prepared statements.

- For one-off queries: `QueryContext` / `ExecContext` are sufficient and simpler.
- For the same parameterised query executed many times: use prepared statements to save
  the parse/plan step on each execution.
- One-shot prepare (prepare + execute + close) is often slower than a direct query.
- Always close statements to return connections to the pool.

```go
// Manual lifecycle — reuse stmt across many calls
stmt, args, err := exec.PrepareContext(ctx, query)
if err != nil { return err }
defer stmt.Close()

rows, err := stmt.QueryContext(ctx, args...)
```

### Logging

- Log the parameterised SQL template and args count, not the values:
  ```go
  log, _ := sqlbuilder.SafeLogQuery(query)
  logger.Info("query", "sql", log.SQL, "args", log.ArgsCount, "duration", elapsed)
  ```
- Do not log `DebugSQL` output in production — it may contain personal data.
- If you use `DebugSQLWithConfig`, always set `RedactFields` for sensitive columns
  (`password`, `token`, `api_key`, etc.).
- Include `duration`, `rows_affected`, `error`, and a trace/request ID in log entries.

### Transactions

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil { return err }
defer tx.Rollback()  // no-op after a successful Commit

exec := sqlbuilder.NewExecutor(tx)

_, err = exec.ExecContext(ctx, insertQuery)
if err != nil { return err }

_, err = exec.ExecContext(ctx, updateQuery)
if err != nil { return err }

return tx.Commit()
```

### Builder lifecycle

- **Create one builder per query / request / goroutine.** Builders are mutable; reusing
  them concurrently is a data race.
- **Do not store builders in global variables.** A builder accumulates state; sharing it
  across requests will corrupt SQL.
- **You can cache a built SQL string** if the query template is fully static
  (no dynamic columns, no dynamic WHERE), but args must always be assembled fresh per call.

---

## Compatibility

| Database | Version | Status | Notes |
|---|---|---:|---|
| MySQL | 8.x | **Supported** | Primary target |
| MySQL | 5.7 | Not guaranteed | JSON functions and some schema features require MySQL 8 |
| MariaDB | 10.x / 11.x | Not guaranteed | SQL syntax is largely compatible but JSON and DDL details differ; not tested |
| PostgreSQL | any | **Not supported** | Placeholder style (`$1`) and identifier quoting (`"`) differ from MySQL |
| SQLite | any | **Not supported** | Different SQL dialect |

**Notes:**

- The library is designed and tested exclusively against MySQL 8.
- MySQL 5.7 lacks `JSON_TABLE`, window functions, and several `JSON_*` functions used by the JSON helpers.
- MariaDB differs in JSON function signatures and some DDL syntax. If you use MariaDB, verify
  generated SQL manually before deploying.
- PostgreSQL and SQLite use different placeholder styles and quoting conventions that are
  incompatible with this builder's output.

---

## Limitations (v0.2.0)

- No migration history table or state tracking — `RunUp`/`RunDown` execute steps but do not record which migrations have been applied; implement tracking in your application layer.
- `DebugSQL` redaction is best-effort based on SQL text pattern matching.
- No UNION / INTERSECT / EXCEPT builder (use `RawQuery`).
- No INSERT … SELECT builder (use `RawQuery`).
- No per-query timeout configuration (use `context.WithTimeout`).
- No EXPLAIN wrapper.
- `QueryRowContext` is deprecated; use `QueryRowContextErr` to receive build errors explicitly.
- Raw JOIN variants (`Join`, `LeftJoin`, etc.) are deprecated; use `JoinOn` / `JoinRaw`.

---

## Examples

Runnable examples are in the [`examples/`](examples/) directory.
Each example is a standalone program that can be run with `go run`.

```
examples/
├── 01_select/         SELECT, WHERE, ORDER BY, LIMIT/OFFSET
├── 02_insert/         INSERT, bulk insert, ON DUPLICATE KEY UPDATE
├── 03_update/         UPDATE with Set, Where, Limit
├── 04_delete/         Safe DELETE with WHERE and LIMIT
├── 05_conditions/     All condition helpers (Eq, Like, In, Between, And/Or, …)
├── 06_join/           Safe typed JOIN (JoinOn/OnEq), LEFT JOIN, raw JOIN
├── 07_json/           MySQL 8 JSON helpers (JSONExtract, JSONEq, JSONContains)
├── 08_subqueries/     WHERE IN subquery, EXISTS, FROM subquery, SubqueryColumn
├── 09_cte/            WITH, multiple CTEs, WITH RECURSIVE
├── 10_schema/         DDL: CREATE TABLE, ALTER TABLE, indexes, FK, Migration
├── 11_transactions/   db.BeginTx, NewExecutor(tx), commit/rollback
├── 12_context_timeout/ context.WithTimeout per query
├── 13_debug_sql/      Debug, DebugSQLWithConfig, SafeLogQuery (no DB needed)
└── 14_high_load/      Pool config, prepared statements, db.Stats()
```

All runnable examples connect to MySQL via [`github.com/akula410/connect/v2`](https://github.com/akula410/connect):

```go
import (
    sqlbuilder "github.com/akula410/builder/v2"
    "github.com/akula410/connect/v2"
)

cfg := connect.DefaultConfig()
cfg.User = "root"
cfg.Password = os.Getenv("MYSQL_PASSWORD")
cfg.DBName = "builder_example"

db, err := connect.NewMySQLContext(ctx, cfg)
if err != nil { log.Fatal(err) }
defer db.Close()

exec := sqlbuilder.NewExecutor(db)
```

Run any example:

```bash
MYSQL_HOST=127.0.0.1 MYSQL_PORT=3306 \
MYSQL_USER=root MYSQL_PASSWORD=secret \
MYSQL_DATABASE=builder_example \
go run ./examples/01_select
```

See [`examples/README.md`](examples/README.md) for the full list and instructions.

---

## TODO / Roadmap

- Migration history table + state tracking (sequential runner is already provided via `RunUp`/`RunDown`)
- UNION / INTERSECT / EXCEPT builder
- INSERT … SELECT builder
- UPSERT helpers (beyond ON DUPLICATE KEY UPDATE)
- Window functions
- Multi-table UPDATE / DELETE
- EXPLAIN wrapper
- OpenTelemetry tracing integration
