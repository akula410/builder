# builder — Examples

Practical, runnable examples for the `github.com/akula410/builder/v2` SQL builder.

Each example is a standalone `go run`-able program in its own subdirectory.  
Examples that require a MySQL connection skip gracefully when no environment variables are set.

---

## Prerequisites

- Go 1.26+
- MySQL 8.x (optional — build-only examples run without a database)

---

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `MYSQL_HOST` | `127.0.0.1` | MySQL server host |
| `MYSQL_PORT` | `3306` | MySQL server port |
| `MYSQL_USER` | `root` | Database user |
| `MYSQL_PASSWORD` | *(empty)* | Database password — **never logged** |
| `MYSQL_DATABASE` | `builder_example` | Database name |

---

## Running an example

```bash
# Build-only (no DB needed):
go run ./examples/13_debug_sql

# With a database:
MYSQL_HOST=127.0.0.1 \
MYSQL_PORT=3306 \
MYSQL_USER=root \
MYSQL_PASSWORD=secret \
MYSQL_DATABASE=builder_example \
go run ./examples/01_select
```

---

## Examples

| Directory | What it covers |
|---|---|
| `01_select/` | SELECT, WHERE, ORDER BY, LIMIT/OFFSET, QueryContext, rows.Scan |
| `02_insert/` | Single row, bulk insert, ON DUPLICATE KEY UPDATE |
| `03_update/` | Set, Where, Limit, ExecContext, RowsAffected |
| `04_delete/` | Safe DELETE with WHERE and LIMIT; danger of DELETE without WHERE |
| `05_conditions/` | All condition helpers: Eq, Neq, Gt, Like, In, IsNull, Between, And, Or, RawCondition |
| `06_join/` | Safe typed JOIN (JoinOn/OnEq), LEFT JOIN, compound ON, raw JOIN with warnings |
| `07_json/` | MySQL 8 JSON helpers: JSONExtract, JSONUnquote, JSONEq, JSONContains |
| `08_subqueries/` | WHERE IN subquery, EXISTS/NOT EXISTS, FROM subquery, SubqueryColumn |
| `09_cte/` | WITH, multiple CTEs, WITH RECURSIVE |
| `10_schema/` | CREATE TABLE, ALTER TABLE, indexes, foreign keys, DROP, TRUNCATE, RENAME, Migration |
| `11_transactions/` | db.BeginTx, NewExecutor(tx), commit/rollback pattern |
| `12_context_timeout/` | context.WithTimeout, QueryContext, ExecContext, QueryRowContextErr |
| `13_debug_sql/` | Debug, DebugSQL, DebugSQLWithConfig (redaction), SafeLogQuery — **no DB needed** |
| `14_high_load/` | Pool config, one *sql.DB per app, prepared statements, db.Stats() |

---

## Connection helper

All examples that require a database use the shared helper at
`examples/internal/db/db.go`.

The helper reads MySQL settings from environment variables, opens a `*sql.DB`
via `github.com/akula410/connect/v2`, and prints a friendly skip message when
no env vars are set — so examples never panic on missing configuration.

```go
// internal/db/db.go in brief:
db, err := connect.NewMySQLContext(ctx, cfg)   // connect/v2 opens *sql.DB
exec := sqlbuilder.NewExecutor(db)             // builder wraps it
```

---

## Security warnings

### DebugSQL is for logging only — never execute it

```go
debug, err := sqlbuilder.Debug(query)   // human-readable, values substituted
// DO NOT:  db.Exec(debug)              // SQL injection risk
```

Always use the parameterised query returned by `Build()` for execution:

```go
sql, args, err := query.Build()
rows, err := exec.QueryContext(ctx, query)   // safe — uses ? placeholders
```

### Raw / RawCondition — only for hardcoded expressions

```go
// Safe — the ON expression is a known, hardcoded column reference:
sqlbuilder.JoinRaw("orders o", "o.user_id = u.id")

// Safe — value uses a ? placeholder:
sqlbuilder.RawCondition("JSON_EXTRACT(`data`, ?) = ?", "$.status", "active")

// UNSAFE — never do this:
sqlbuilder.RawCondition("name = '" + userInput + "'")  // SQL injection!
sqlbuilder.Raw("SELECT * FROM " + tableName)           // SQL injection!
```

### JOIN ON raw expressions — never pass user input

```go
// JoinRaw / LeftJoinRaw accept raw ON strings embedded as-is into SQL.
// Only pass hardcoded, trusted column comparison expressions.
// To filter by a value, use a Where condition instead.
```

---

## Architecture notes

| Component | Responsibility |
|---|---|
| `github.com/akula410/builder/v2` | Builds parameterised SQL (`?` placeholders) |
| `github.com/akula410/connect/v2` | Opens `*sql.DB`, manages connection pool |
| `database/sql` | Executes SQL, manages transactions and rows |

`builder` does not open connections. `connect` does not build SQL.
They are deliberately separate concerns.
