package builder

import (
	"fmt"
	"strings"
)

// OnCondition represents a type-safe column-to-column comparison for JOIN ON clauses.
// Column references are validated and quoted with backticks; no value placeholders are used.
//
// Use OnEq, OnNe, OnGt, OnGte, OnLt, OnLte, OnAnd, OnOr to construct ON conditions.
// To join on a raw expression use JoinRaw / LeftJoinRaw / RightJoinRaw / InnerJoinRaw.
type OnCondition interface {
	buildOnCondition() (string, error)
}

// onCmpCond is a validated column-to-column comparison: leftCol OP rightCol.
type onCmpCond struct {
	left  string
	op    string
	right string
}

func (c onCmpCond) buildOnCondition() (string, error) {
	l, err := parseColumnExpr(c.left)
	if err != nil {
		return "", fmt.Errorf("ON left column: %w", err)
	}
	r, err := parseColumnExpr(c.right)
	if err != nil {
		return "", fmt.Errorf("ON right column: %w", err)
	}
	return l + " " + c.op + " " + r, nil
}

// OnEq returns a safe ON condition: leftCol = rightCol.
func OnEq(leftCol, rightCol string) OnCondition { return onCmpCond{leftCol, "=", rightCol} }

// OnNe returns a safe ON condition: leftCol != rightCol.
func OnNe(leftCol, rightCol string) OnCondition { return onCmpCond{leftCol, "!=", rightCol} }

// OnGt returns a safe ON condition: leftCol > rightCol.
func OnGt(leftCol, rightCol string) OnCondition { return onCmpCond{leftCol, ">", rightCol} }

// OnGte returns a safe ON condition: leftCol >= rightCol.
func OnGte(leftCol, rightCol string) OnCondition { return onCmpCond{leftCol, ">=", rightCol} }

// OnLt returns a safe ON condition: leftCol < rightCol.
func OnLt(leftCol, rightCol string) OnCondition { return onCmpCond{leftCol, "<", rightCol} }

// OnLte returns a safe ON condition: leftCol <= rightCol.
func OnLte(leftCol, rightCol string) OnCondition { return onCmpCond{leftCol, "<=", rightCol} }

// onAndCond groups ON conditions with AND.
type onAndCond struct{ conds []OnCondition }

func (c onAndCond) buildOnCondition() (string, error) {
	return joinOnConditions(c.conds, " AND ")
}

// OnAnd groups ON conditions with AND, wrapping in parentheses when more than one.
func OnAnd(conds ...OnCondition) OnCondition { return onAndCond{conds} }

// onOrCond groups ON conditions with OR.
type onOrCond struct{ conds []OnCondition }

func (c onOrCond) buildOnCondition() (string, error) {
	return joinOnConditions(c.conds, " OR ")
}

// OnOr groups ON conditions with OR, wrapping in parentheses when more than one.
func OnOr(conds ...OnCondition) OnCondition { return onOrCond{conds} }

func joinOnConditions(conds []OnCondition, sep string) (string, error) {
	if len(conds) == 0 {
		return "", fmt.Errorf("sqlbuilder: ON condition group requires at least one condition")
	}
	parts := make([]string, 0, len(conds))
	for _, c := range conds {
		s, err := c.buildOnCondition()
		if err != nil {
			return "", err
		}
		if s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	return "(" + strings.Join(parts, sep) + ")", nil
}
