package builder

// JSONExtract returns a raw SQL expression for JSON_EXTRACT(`col`, path).
// path is a JSON path string like "$.status" and is passed as a placeholder arg.
//
// Example:
//
//	JSONExtract("data", "$.company_id")
//	→ SQL: JSON_EXTRACT(`data`, ?)   args: ["$.company_id"]
func JSONExtract(col, path string) RawExpr {
	q, err := quoteIdent(col)
	if err != nil {
		return RawExpr{sql: "/* invalid column: " + col + " */"}
	}
	return RawExpr{sql: "JSON_EXTRACT(" + q + ", ?)", args: []any{path}}
}

// JSONUnquote returns a raw SQL expression for JSON_UNQUOTE(JSON_EXTRACT(`col`, path)).
//
// Example:
//
//	JSONUnquote("data", "$.name")
//	→ SQL: JSON_UNQUOTE(JSON_EXTRACT(`data`, ?))   args: ["$.name"]
func JSONUnquote(col, path string) RawExpr {
	q, err := quoteIdent(col)
	if err != nil {
		return RawExpr{sql: "/* invalid column: " + col + " */"}
	}
	return RawExpr{sql: "JSON_UNQUOTE(JSON_EXTRACT(" + q + ", ?))", args: []any{path}}
}

// jsonEqCond implements a JSON equality condition using JSON_UNQUOTE(JSON_EXTRACT(...)) = ?.
type jsonEqCond struct {
	col   string
	path  string
	value any
}

func (c jsonEqCond) buildCondition() (string, []any, error) {
	q, err := quoteIdent(c.col)
	if err != nil {
		return "", nil, err
	}
	sql := "JSON_UNQUOTE(JSON_EXTRACT(" + q + ", ?)) = ?"
	return sql, []any{c.path, c.value}, nil
}

// JSONEq returns a condition that checks JSON_UNQUOTE(JSON_EXTRACT(`col`, path)) = value.
//
// Example:
//
//	JSONEq("data", "$.company_id", "ngus")
//	→ WHERE JSON_UNQUOTE(JSON_EXTRACT(`data`, ?)) = ?   args: ["$.company_id", "ngus"]
func JSONEq(col, path string, value any) Condition {
	return jsonEqCond{col: col, path: path, value: value}
}

// jsonContainsCond implements JSON_CONTAINS(JSON_EXTRACT(`col`, path), JSON_QUOTE(?)).
type jsonContainsCond struct {
	col   string
	path  string
	value any
}

func (c jsonContainsCond) buildCondition() (string, []any, error) {
	q, err := quoteIdent(c.col)
	if err != nil {
		return "", nil, err
	}
	sql := "JSON_CONTAINS(JSON_EXTRACT(" + q + ", ?), JSON_QUOTE(?))"
	return sql, []any{c.path, c.value}, nil
}

// JSONContains returns a condition using JSON_CONTAINS to check if a JSON array/object
// contains a scalar value at the given path.
//
// Example:
//
//	JSONContains("data", "$.tags", "urgent")
//	→ JSON_CONTAINS(JSON_EXTRACT(`data`, ?), JSON_QUOTE(?))   args: ["$.tags", "urgent"]
func JSONContains(col, path string, value any) Condition {
	return jsonContainsCond{col: col, path: path, value: value}
}
