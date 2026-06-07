package builder

// RawExpr holds a raw SQL fragment with optional args.
// Use Raw() or Expr() to create one.
//
// WARNING: Never pass user input directly into the sql string.
// Only use Raw for known SQL fragments. Values should still pass through args.
type RawExpr struct {
	sql  string
	args []any
}

// Build implements QueryBuilder.
func (r RawExpr) Build() (string, []any, error) {
	return r.sql, r.args, nil
}

// SQL returns the raw SQL string.
func (r RawExpr) SQL() string { return r.sql }

// Args returns the raw args slice.
func (r RawExpr) Args() []any { return r.args }

// Raw creates a raw SQL fragment embedded as-is in the query.
//
// WARNING: Never pass user input directly in sql.
// Values must still be passed through args with ? placeholders.
//
// Safe usage:
//
//	Raw("SUM(total) AS total")
//	Raw("JSON_EXTRACT(`data`, ?) = ?", "$.status", "active")
//
// Unsafe usage (DO NOT DO THIS):
//
//	Raw("name = '" + userInput + "'")  // SQL injection risk
func Raw(sql string, args ...any) RawExpr {
	return RawExpr{sql: sql, args: args}
}

// Expr is an alias for Raw.
//
// WARNING: Same security constraints as Raw apply.
func Expr(sql string, args ...any) RawExpr {
	return RawExpr{sql: sql, args: args}
}

// rawQueryBuilder wraps a full raw query.
type rawQueryBuilder struct {
	sql  string
	args []any
}

// Build implements QueryBuilder.
func (r *rawQueryBuilder) Build() (string, []any, error) {
	return r.sql, r.args, nil
}

// RawQuery creates a full raw query builder.
//
// WARNING: Never pass user input directly in sql.
// Values must still be passed through args with ? placeholders.
//
// Safe usage:
//
//	RawQuery("SELECT 1 FROM dual")
//	RawQuery("SELECT id FROM users WHERE id = ?", userID)
//
// Unsafe usage (DO NOT DO THIS):
//
//	RawQuery("SELECT * FROM " + tableName)  // SQL injection risk
func RawQuery(sql string, args ...any) *rawQueryBuilder {
	return &rawQueryBuilder{sql: sql, args: args}
}
