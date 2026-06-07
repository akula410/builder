package builder

import (
	"fmt"
	"strings"
)

// SortDir represents sort direction in ORDER BY.
type SortDir string

const (
	Asc  SortDir = "ASC"
	Desc SortDir = "DESC"
)

// colExpr is a resolved SELECT column expression.
type colExpr struct {
	sql  string
	args []any
}

// subqueryColumn is a subquery used as a SELECT column with an alias.
type subqueryColumn struct {
	sub   QueryBuilder
	alias string
}

// SubqueryColumn wraps a subquery as a named SELECT column.
//
// Example:
//
//	SubqueryColumn(
//	    Select(Raw("COUNT(*)")).From("orders o").Where(RawCondition("o.user_id = u.id")),
//	    "orders_count",
//	)
func SubqueryColumn(sub QueryBuilder, alias string) subqueryColumn {
	return subqueryColumn{sub: sub, alias: alias}
}

// joinEntry holds a JOIN clause.
type joinEntry struct {
	kind  string // "INNER JOIN", "LEFT JOIN", "RIGHT JOIN"
	table string // quoted table expression
	on    string // raw ON condition
}

// orderEntry holds an ORDER BY entry.
type orderEntry struct {
	expr string // quoted column or raw expression
	dir  SortDir
}

// SelectBuilder builds SELECT queries.
// Zero value is not usable; create via Select() or With(...).Select().
type SelectBuilder struct {
	colExprs  []colExpr
	distinct  bool
	fromExpr  string
	fromSub   QueryBuilder
	fromAlias string
	joins     []joinEntry
	wheres    []Condition
	groupBys  []string
	havings   []Condition
	orderBys  []orderEntry
	limitVal  int64
	offsetVal int64
	hasLimit  bool
	hasOffset bool
	cteList   []cteEntry
	recursive bool
	err       error
}

// Select starts a SELECT query builder.
// Columns can be strings ("id", "u.name", "*") or RawExpr / subqueryColumn values.
func Select(columns ...any) *SelectBuilder {
	b := &SelectBuilder{}
	b.addColumns(columns)
	return b
}

func (b *SelectBuilder) addColumns(columns []any) {
	for _, col := range columns {
		switch v := col.(type) {
		case string:
			q, err := parseColumnExpr(v)
			if err != nil {
				b.setErr(err)
				return
			}
			b.colExprs = append(b.colExprs, colExpr{sql: q})
		case RawExpr:
			b.colExprs = append(b.colExprs, colExpr{sql: v.sql, args: v.args})
		case subqueryColumn:
			sql, args, err := buildSubqueryColumn(v)
			if err != nil {
				b.setErr(err)
				return
			}
			b.colExprs = append(b.colExprs, colExpr{sql: sql, args: args})
		default:
			b.setErr(fmt.Errorf("sqlbuilder: unsupported column type %T", col))
			return
		}
	}
}

func buildSubqueryColumn(sc subqueryColumn) (string, []any, error) {
	subSQL, subArgs, err := sc.sub.Build()
	if err != nil {
		return "", nil, fmt.Errorf("subquery column: %w", err)
	}
	qa, err := quoteIdent(sc.alias)
	if err != nil {
		return "", nil, fmt.Errorf("subquery column alias: %w", err)
	}
	return "(\n    " + subSQL + "\n) AS " + qa, subArgs, nil
}

func (b *SelectBuilder) setErr(err error) {
	if b.err == nil {
		b.err = err
	}
}

// Distinct adds the DISTINCT modifier.
func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.distinct = true
	return b
}

// From sets the FROM table (with optional alias).
//
// Supported formats:
//
//	From("users")
//	From("users u")
//	From("users AS u")
func (b *SelectBuilder) From(tableExpr string) *SelectBuilder {
	q, err := parseTableExpr(tableExpr)
	if err != nil {
		b.setErr(err)
		return b
	}
	b.fromExpr = q
	return b
}

// FromSubquery sets the FROM clause to a subquery with a required alias.
//
// Example:
//
//	FromSubquery(Select("user_id", Raw("SUM(total) AS total")).From("orders").GroupBy("user_id"), "t")
func (b *SelectBuilder) FromSubquery(sub QueryBuilder, alias string) *SelectBuilder {
	if alias == "" {
		b.setErr(fmt.Errorf("sqlbuilder: FROM subquery requires an alias"))
		return b
	}
	b.fromSub = sub
	b.fromAlias = alias
	return b
}

// Where adds a WHERE condition. Multiple calls are AND-ed together.
func (b *SelectBuilder) Where(cond Condition) *SelectBuilder {
	b.wheres = append(b.wheres, cond)
	return b
}

// addJoin appends a join clause.
func (b *SelectBuilder) addJoin(kind, tableExpr, on string) *SelectBuilder {
	q, err := parseTableExpr(tableExpr)
	if err != nil {
		b.setErr(err)
		return b
	}
	b.joins = append(b.joins, joinEntry{kind: kind, table: q, on: on})
	return b
}

// Join adds an INNER JOIN.
//
// WARNING: The on parameter is embedded into SQL as-is without validation or escaping.
// Only pass known, trusted column comparison expressions (e.g. "o.user_id = u.id").
// Never pass user-controlled input as the on parameter — it is a SQL injection risk.
func (b *SelectBuilder) Join(tableExpr, on string) *SelectBuilder {
	return b.addJoin("INNER JOIN", tableExpr, on)
}

// LeftJoin adds a LEFT JOIN.
//
// WARNING: The on parameter is embedded into SQL as-is without validation or escaping.
// Only pass known, trusted column comparison expressions.
// Never pass user-controlled input as the on parameter — it is a SQL injection risk.
func (b *SelectBuilder) LeftJoin(tableExpr, on string) *SelectBuilder {
	return b.addJoin("LEFT JOIN", tableExpr, on)
}

// RightJoin adds a RIGHT JOIN.
//
// WARNING: The on parameter is embedded into SQL as-is without validation or escaping.
// Only pass known, trusted column comparison expressions.
// Never pass user-controlled input as the on parameter — it is a SQL injection risk.
func (b *SelectBuilder) RightJoin(tableExpr, on string) *SelectBuilder {
	return b.addJoin("RIGHT JOIN", tableExpr, on)
}

// InnerJoin adds an INNER JOIN (alias for Join).
//
// WARNING: The on parameter is embedded into SQL as-is without validation or escaping.
// Only pass known, trusted column comparison expressions.
// Never pass user-controlled input as the on parameter — it is a SQL injection risk.
func (b *SelectBuilder) InnerJoin(tableExpr, on string) *SelectBuilder {
	return b.addJoin("INNER JOIN", tableExpr, on)
}

// GroupBy adds GROUP BY columns.
func (b *SelectBuilder) GroupBy(fields ...string) *SelectBuilder {
	for _, f := range fields {
		q, err := parseColumnExpr(f)
		if err != nil {
			b.setErr(err)
			return b
		}
		b.groupBys = append(b.groupBys, q)
	}
	return b
}

// Having adds a HAVING condition. Multiple calls are AND-ed together.
func (b *SelectBuilder) Having(cond Condition) *SelectBuilder {
	b.havings = append(b.havings, cond)
	return b
}

// OrderBy adds an ORDER BY expression.
//
// Examples:
//
//	OrderBy("id", Desc)
//	OrderBy("created_at", Asc)
//	OrderBy("id")  // defaults to ASC
func (b *SelectBuilder) OrderBy(field string, dir ...SortDir) *SelectBuilder {
	q, err := parseColumnExpr(field)
	if err != nil {
		b.setErr(err)
		return b
	}
	d := Asc
	if len(dir) > 0 {
		d = dir[0]
	}
	b.orderBys = append(b.orderBys, orderEntry{expr: q, dir: d})
	return b
}

// Limit sets the LIMIT value. Must be >= 0.
func (b *SelectBuilder) Limit(n int64) *SelectBuilder {
	if n < 0 {
		b.setErr(ErrNegativeLimit)
		return b
	}
	b.limitVal = n
	b.hasLimit = true
	return b
}

// Offset sets the OFFSET value. Must be >= 0.
func (b *SelectBuilder) Offset(n int64) *SelectBuilder {
	if n < 0 {
		b.setErr(ErrNegativeOffset)
		return b
	}
	b.offsetVal = n
	b.hasOffset = true
	return b
}

// Build assembles the SELECT query and returns SQL, args, and any error.
func (b *SelectBuilder) Build() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}

	var parts []string
	var args []any

	// 1. CTE prefix
	if len(b.cteList) > 0 {
		cteSQL, cteArgs, err := buildCTEClause(b.cteList, b.recursive)
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, cteSQL)
		args = append(args, cteArgs...)
	}

	// 2. SELECT clause
	selSQL, selArgs, err := b.buildSelectClause()
	if err != nil {
		return "", nil, err
	}
	parts = append(parts, selSQL)
	args = append(args, selArgs...)

	// 3. FROM clause
	fromSQL, fromArgs, err := b.buildFromClause()
	if err != nil {
		return "", nil, err
	}
	parts = append(parts, fromSQL)
	args = append(args, fromArgs...)

	// 4. JOIN clauses
	for _, j := range b.joins {
		parts = append(parts, j.kind+" "+j.table+" ON "+j.on)
	}

	// 5. WHERE clause
	if len(b.wheres) > 0 {
		whereSQL, whereArgs, err := buildConditions(b.wheres)
		if err != nil {
			return "", nil, err
		}
		if whereSQL != "" {
			parts = append(parts, "WHERE "+whereSQL)
			args = append(args, whereArgs...)
		}
	}

	// 6. GROUP BY clause
	if len(b.groupBys) > 0 {
		parts = append(parts, "GROUP BY "+strings.Join(b.groupBys, ", "))
	}

	// 7. HAVING clause
	if len(b.havings) > 0 {
		havingSQL, havingArgs, err := buildConditions(b.havings)
		if err != nil {
			return "", nil, err
		}
		if havingSQL != "" {
			parts = append(parts, "HAVING "+havingSQL)
			args = append(args, havingArgs...)
		}
	}

	// 8. ORDER BY clause
	if len(b.orderBys) > 0 {
		obs := make([]string, len(b.orderBys))
		for i, o := range b.orderBys {
			obs[i] = o.expr + " " + string(o.dir)
		}
		parts = append(parts, "ORDER BY "+strings.Join(obs, ", "))
	}

	// 9. LIMIT / OFFSET
	if b.hasLimit {
		parts = append(parts, fmt.Sprintf("LIMIT %d", b.limitVal))
	}
	if b.hasOffset {
		parts = append(parts, fmt.Sprintf("OFFSET %d", b.offsetVal))
	}

	return strings.Join(parts, "\n"), args, nil
}

func (b *SelectBuilder) buildSelectClause() (string, []any, error) {
	kw := "SELECT"
	if b.distinct {
		kw = "SELECT DISTINCT"
	}

	if len(b.colExprs) == 0 {
		return kw + " *", nil, nil
	}

	cols := make([]string, len(b.colExprs))
	var args []any
	for i, ce := range b.colExprs {
		cols[i] = ce.sql
		args = append(args, ce.args...)
	}
	return kw + " " + strings.Join(cols, ", "), args, nil
}

func (b *SelectBuilder) buildFromClause() (string, []any, error) {
	if b.fromSub != nil {
		subSQL, subArgs, err := b.fromSub.Build()
		if err != nil {
			return "", nil, fmt.Errorf("FROM subquery: %w", err)
		}
		qa, err := quoteIdent(b.fromAlias)
		if err != nil {
			return "", nil, fmt.Errorf("FROM alias: %w", err)
		}
		return "FROM (\n    " + subSQL + "\n) AS " + qa, subArgs, nil
	}
	if b.fromExpr == "" {
		return "", nil, ErrEmptyFrom
	}
	return "FROM " + b.fromExpr, nil, nil
}
