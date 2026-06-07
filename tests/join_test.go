package sqlbuilder_test

import (
	"errors"
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func TestJoinOn_SimpleEq(t *testing.T) {
	sql, args, err := sqlbuilder.Select("u.id", "u.name").
		From("users u").
		JoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "INNER JOIN `orders` AS `o` ON `o`.`user_id` = `u`.`id`")
	assertArgs(t, nil, args)
}

func TestLeftJoinOn_SimpleEq(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		LeftJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "LEFT JOIN `orders` AS `o` ON `o`.`user_id` = `u`.`id`")
}

func TestRightJoinOn(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		RightJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "RIGHT JOIN `orders` AS `o` ON `o`.`user_id` = `u`.`id`")
}

func TestInnerJoinOn(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		InnerJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "INNER JOIN `orders` AS `o` ON `o`.`user_id` = `u`.`id`")
}

func TestJoinOn_InvalidLeftColumn_Space(t *testing.T) {
	_, _, err := sqlbuilder.Select("id").
		From("users u").
		JoinOn("orders o", sqlbuilder.OnEq("o.user id", "u.id")).
		Build()
	assertErr(t, err)
	if !errors.Is(err, sqlbuilder.ErrInvalidIdent) {
		t.Fatalf("expected ErrInvalidIdent, got: %v", err)
	}
}

func TestJoinOn_InvalidRightColumn_Semicolon(t *testing.T) {
	_, _, err := sqlbuilder.Select("id").
		From("users u").
		JoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id; DROP TABLE users")).
		Build()
	assertErr(t, err)
}

func TestJoinOn_InvalidColumn_EmptyString(t *testing.T) {
	_, _, err := sqlbuilder.Select("id").
		From("users u").
		JoinOn("orders o", sqlbuilder.OnEq("", "u.id")).
		Build()
	assertErr(t, err)
	if !errors.Is(err, sqlbuilder.ErrInvalidIdent) {
		t.Fatalf("expected ErrInvalidIdent, got: %v", err)
	}
}

func TestJoinOn_DottedIdentifiers_Quoted(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		LeftJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "`o`.`user_id` = `u`.`id`")
}

func TestOnAnd_GroupsWithAnd(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		JoinOn("orders o", sqlbuilder.OnAnd(
			sqlbuilder.OnEq("o.user_id", "u.id"),
			sqlbuilder.OnEq("o.status", "u.status"),
		)).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "(`o`.`user_id` = `u`.`id` AND `o`.`status` = `u`.`status`)")
}

func TestOnOr_GroupsWithOr(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		JoinOn("orders o", sqlbuilder.OnOr(
			sqlbuilder.OnEq("o.user_id", "u.id"),
			sqlbuilder.OnEq("o.ref_id", "u.id"),
		)).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "(`o`.`user_id` = `u`.`id` OR `o`.`ref_id` = `u`.`id`)")
}

func TestOnNe(t *testing.T) {
	sql, _, err := sqlbuilder.Select("a.id").
		From("a_table a").
		JoinOn("b_table b", sqlbuilder.OnNe("a.status", "b.status")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "`a`.`status` != `b`.`status`")
}

func TestOnGt_OnLt(t *testing.T) {
	sql, _, err := sqlbuilder.Select("a.id").
		From("a_table a").
		JoinOn("b_table b", sqlbuilder.OnAnd(
			sqlbuilder.OnGt("a.score", "b.min_score"),
			sqlbuilder.OnLt("a.score", "b.max_score"),
		)).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "`a`.`score` > `b`.`min_score`")
	assertContains(t, sql, "`a`.`score` < `b`.`max_score`")
}

func TestOnGte_OnLte(t *testing.T) {
	sql, _, err := sqlbuilder.Select("a.id").
		From("a_table a").
		JoinOn("b_table b", sqlbuilder.OnAnd(
			sqlbuilder.OnGte("a.score", "b.min_score"),
			sqlbuilder.OnLte("a.score", "b.max_score"),
		)).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "`a`.`score` >= `b`.`min_score`")
	assertContains(t, sql, "`a`.`score` <= `b`.`max_score`")
}

func TestOnAnd_SingleCondition_NoParens(t *testing.T) {
	sql, _, err := sqlbuilder.Select("a.id").
		From("a_table a").
		JoinOn("b_table b", sqlbuilder.OnAnd(sqlbuilder.OnEq("a.id", "b.id"))).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "`a`.`id` = `b`.`id`")
}

func TestOnOr_SingleCondition_NoParens(t *testing.T) {
	sql, _, err := sqlbuilder.Select("a.id").
		From("a_table a").
		JoinOn("b_table b", sqlbuilder.OnOr(sqlbuilder.OnEq("a.id", "b.id"))).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "`a`.`id` = `b`.`id`")
}

func TestJoinRaw_StillWorks(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		JoinRaw("orders o", "o.user_id = u.id").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "INNER JOIN `orders` AS `o` ON o.user_id = u.id")
}

func TestLeftJoinRaw_StillWorks(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		LeftJoinRaw("orders o", "o.user_id = u.id").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "LEFT JOIN `orders` AS `o` ON o.user_id = u.id")
}

func TestRightJoinRaw_StillWorks(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		RightJoinRaw("orders o", "o.user_id = u.id").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "RIGHT JOIN `orders` AS `o` ON o.user_id = u.id")
}

func TestInnerJoinRaw_StillWorks(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		InnerJoinRaw("orders o", "o.user_id = u.id").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "INNER JOIN `orders` AS `o` ON o.user_id = u.id")
}

func TestOldJoin_BackwardCompat(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		Join("orders o", "o.user_id = u.id").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "INNER JOIN `orders` AS `o` ON o.user_id = u.id")
}

func TestOldLeftJoin_BackwardCompat(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id").
		From("users u").
		LeftJoin("orders o", "o.user_id = u.id").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "LEFT JOIN `orders` AS `o` ON o.user_id = u.id")
}

func TestJoinOn_MultipleJoins(t *testing.T) {
	sql, _, err := sqlbuilder.Select("u.id", "o.total", "p.name").
		From("users u").
		LeftJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
		LeftJoinOn("products p", sqlbuilder.OnEq("p.id", "o.product_id")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "LEFT JOIN `orders` AS `o` ON `o`.`user_id` = `u`.`id`")
	assertContains(t, sql, "LEFT JOIN `products` AS `p` ON `p`.`id` = `o`.`product_id`")
}

func TestJoinOn_WithWhereArgsOrder(t *testing.T) {
	sql, args, err := sqlbuilder.Select("u.id").
		From("users u").
		LeftJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
		Where(sqlbuilder.Eq("u.status", "active")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "LEFT JOIN `orders` AS `o` ON `o`.`user_id` = `u`.`id`")
	assertContains(t, sql, "WHERE `u`.`status` = ?")
	assertArgs(t, []any{"active"}, args)
}

func TestJoinOn_InvalidTable(t *testing.T) {
	_, _, err := sqlbuilder.Select("id").
		From("users").
		JoinOn("bad table name here", sqlbuilder.OnEq("a.id", "b.id")).
		Build()
	assertErr(t, err)
}

func TestJoinOn_InvalidColumnThreePartDotted(t *testing.T) {
	// db.table.column is not supported — only table.column
	_, _, err := sqlbuilder.Select("id").
		From("users u").
		JoinOn("orders o", sqlbuilder.OnEq("db.table.column", "u.id")).
		Build()
	assertErr(t, err)
}
