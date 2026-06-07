package builder

import (
	"fmt"
	"strings"
)

// Condition is implemented by all WHERE/HAVING condition types.
type Condition interface {
	buildCondition() (string, []any, error)
}

// cmpCond is a simple comparison: field OP value.
type cmpCond struct {
	field string
	op    string
	value any
}

func (c cmpCond) buildCondition() (string, []any, error) {
	col, err := parseColumnExpr(c.field)
	if err != nil {
		return "", nil, err
	}
	return col + " " + c.op + " ?", []any{c.value}, nil
}

// Eq returns a field = value condition.
func Eq(field string, value any) Condition { return cmpCond{field, "=", value} }

// Neq returns a field != value condition.
func Neq(field string, value any) Condition { return cmpCond{field, "!=", value} }

// Gt returns a field > value condition.
func Gt(field string, value any) Condition { return cmpCond{field, ">", value} }

// Gte returns a field >= value condition.
func Gte(field string, value any) Condition { return cmpCond{field, ">=", value} }

// Lt returns a field < value condition.
func Lt(field string, value any) Condition { return cmpCond{field, "<", value} }

// Lte returns a field <= value condition.
func Lte(field string, value any) Condition { return cmpCond{field, "<=", value} }

// likeCond implements LIKE and NOT LIKE.
type likeCond struct {
	field   string
	pattern any
	not     bool
}

func (c likeCond) buildCondition() (string, []any, error) {
	col, err := parseColumnExpr(c.field)
	if err != nil {
		return "", nil, err
	}
	op := "LIKE"
	if c.not {
		op = "NOT LIKE"
	}
	return col + " " + op + " ?", []any{c.pattern}, nil
}

// Like returns a field LIKE pattern condition.
func Like(field string, pattern any) Condition { return likeCond{field, pattern, false} }

// NotLike returns a field NOT LIKE pattern condition.
func NotLike(field string, pattern any) Condition { return likeCond{field, pattern, true} }

// inCond implements IN and NOT IN.
type inCond struct {
	field  string
	values []any
	not    bool
}

func (c inCond) buildCondition() (string, []any, error) {
	if len(c.values) == 0 {
		return "", nil, ErrEmptyIN
	}
	col, err := parseColumnExpr(c.field)
	if err != nil {
		return "", nil, err
	}
	op := "IN"
	if c.not {
		op = "NOT IN"
	}
	return col + " " + op + " (" + placeholders(len(c.values)) + ")", c.values, nil
}

// In returns a field IN (values...) condition.
// Returns an error if values is empty.
func In(field string, values []any) Condition { return inCond{field, values, false} }

// NotIn returns a field NOT IN (values...) condition.
// Returns an error if values is empty.
func NotIn(field string, values []any) Condition { return inCond{field, values, true} }

// nullCond implements IS NULL and IS NOT NULL.
type nullCond struct {
	field  string
	isNull bool
}

func (c nullCond) buildCondition() (string, []any, error) {
	col, err := parseColumnExpr(c.field)
	if err != nil {
		return "", nil, err
	}
	if c.isNull {
		return col + " IS NULL", nil, nil
	}
	return col + " IS NOT NULL", nil, nil
}

// IsNull returns a field IS NULL condition.
func IsNull(field string) Condition { return nullCond{field, true} }

// IsNotNull returns a field IS NOT NULL condition.
func IsNotNull(field string) Condition { return nullCond{field, false} }

// betweenCond implements BETWEEN.
type betweenCond struct {
	field string
	from  any
	to    any
}

func (c betweenCond) buildCondition() (string, []any, error) {
	col, err := parseColumnExpr(c.field)
	if err != nil {
		return "", nil, err
	}
	return col + " BETWEEN ? AND ?", []any{c.from, c.to}, nil
}

// Between returns a field BETWEEN from AND to condition.
func Between(field string, from, to any) Condition { return betweenCond{field, from, to} }

// andCond wraps multiple conditions with AND.
type andCond struct{ conditions []Condition }

func (c andCond) buildCondition() (string, []any, error) {
	return joinConditions(c.conditions, " AND ")
}

// And groups conditions with AND, wrapping the result in parentheses.
func And(conditions ...Condition) Condition { return andCond{conditions} }

// orCond wraps multiple conditions with OR.
type orCond struct{ conditions []Condition }

func (c orCond) buildCondition() (string, []any, error) {
	return joinConditions(c.conditions, " OR ")
}

// Or groups conditions with OR, wrapping the result in parentheses.
func Or(conditions ...Condition) Condition { return orCond{conditions} }

func joinConditions(conditions []Condition, sep string) (string, []any, error) {
	if len(conditions) == 0 {
		return "", nil, nil
	}
	parts := make([]string, 0, len(conditions))
	var args []any
	for _, cond := range conditions {
		s, a, err := cond.buildCondition()
		if err != nil {
			return "", nil, err
		}
		if s != "" {
			parts = append(parts, s)
			args = append(args, a...)
		}
	}
	if len(parts) == 0 {
		return "", nil, nil
	}
	if len(parts) == 1 {
		return parts[0], args, nil
	}
	return "(" + strings.Join(parts, sep) + ")", args, nil
}

// rawCond wraps a raw SQL fragment as a Condition.
type rawCond struct {
	sql  string
	args []any
}

func (c rawCond) buildCondition() (string, []any, error) { return c.sql, c.args, nil }

// RawCondition creates a raw SQL condition.
//
// WARNING: Never pass user input directly in sql.
// Values must still be passed through args with ? placeholders.
//
// Safe usage:
//
//	RawCondition("o.user_id = u.id")
//	RawCondition("JSON_EXTRACT(`data`, ?) = ?", "$.status", "active")
//
// Unsafe usage (DO NOT DO THIS):
//
//	RawCondition("name = '" + userInput + "'")  // SQL injection risk
func RawCondition(sql string, args ...any) Condition { return rawCond{sql, args} }

// inSubqueryCond implements field IN (subquery).
type inSubqueryCond struct {
	field    string
	subquery QueryBuilder
	not      bool
}

func (c inSubqueryCond) buildCondition() (string, []any, error) {
	col, err := parseColumnExpr(c.field)
	if err != nil {
		return "", nil, err
	}
	subSQL, subArgs, err := c.subquery.Build()
	if err != nil {
		return "", nil, fmt.Errorf("IN subquery: %w", err)
	}
	op := "IN"
	if c.not {
		op = "NOT IN"
	}
	return col + " " + op + " (\n    " + subSQL + "\n)", subArgs, nil
}

// InSubquery returns a field IN (subquery) condition.
func InSubquery(field string, sub QueryBuilder) Condition {
	return inSubqueryCond{field, sub, false}
}

// NotInSubquery returns a field NOT IN (subquery) condition.
func NotInSubquery(field string, sub QueryBuilder) Condition {
	return inSubqueryCond{field, sub, true}
}

// existsCond implements EXISTS (subquery).
type existsCond struct {
	subquery QueryBuilder
	not      bool
}

func (c existsCond) buildCondition() (string, []any, error) {
	subSQL, subArgs, err := c.subquery.Build()
	if err != nil {
		return "", nil, fmt.Errorf("EXISTS subquery: %w", err)
	}
	kw := "EXISTS"
	if c.not {
		kw = "NOT EXISTS"
	}
	return kw + " (\n    " + subSQL + "\n)", subArgs, nil
}

// Exists returns an EXISTS (subquery) condition.
func Exists(sub QueryBuilder) Condition { return existsCond{sub, false} }

// NotExists returns a NOT EXISTS (subquery) condition.
func NotExists(sub QueryBuilder) Condition { return existsCond{sub, true} }

// buildConditions joins a slice of conditions with AND (no outer parens).
func buildConditions(conds []Condition) (string, []any, error) {
	if len(conds) == 0 {
		return "", nil, nil
	}
	parts := make([]string, 0, len(conds))
	var args []any
	for _, c := range conds {
		s, a, err := c.buildCondition()
		if err != nil {
			return "", nil, err
		}
		if s != "" {
			parts = append(parts, s)
			args = append(args, a...)
		}
	}
	return strings.Join(parts, " AND "), args, nil
}
