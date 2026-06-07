package sqlbuilder_test

import (
	"strings"
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

// FuzzSelectColumn exercises column identifier parsing through the public Select API.
// Any input must either return an error or produce SQL without unquoted dangerous tokens.
func FuzzSelectColumn(f *testing.F) {
	seeds := []string{
		"id",
		"user_id",
		"u.id",
		"orders.user_id",
		"*",
		"1",
		"",
		"users; DROP TABLE users",
		"user`name",
		"name desc",
		"таблица",
		"id--",
		"a b c",
		"col\x00name",
		strings.Repeat("a", 65),
		strings.Repeat("a", 64),
		"db.table.column",
		"/* comment */",
		"1=1",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		sql, _, err := sqlbuilder.Select(input).From("users").Build()
		if err != nil {
			return
		}
		if strings.Contains(sql, ";") {
			t.Errorf("generated SQL contains semicolon: %q", sql)
		}
		if strings.Contains(sql, "--") {
			t.Errorf("generated SQL contains SQL comment: %q", sql)
		}
		if strings.Contains(sql, "/*") {
			t.Errorf("generated SQL contains block comment: %q", sql)
		}
	})
}

// FuzzFromTable exercises table expression parsing through the public From API.
func FuzzFromTable(f *testing.F) {
	seeds := []string{
		"users",
		"users u",
		"users AS u",
		"schema.table",
		"",
		"users; DROP TABLE users",
		"user`name",
		"users orders extra",
		"a b c d",
		strings.Repeat("t", 65),
		strings.Repeat("t", 64),
		"table AS alias extra",
		"таблица",
		"t\x00able",
		"table--",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		sql, _, err := sqlbuilder.Select("id").From(input).Build()
		if err != nil {
			return
		}
		if strings.Contains(sql, ";") {
			t.Errorf("generated SQL contains semicolon: %q", sql)
		}
		if strings.Contains(sql, "--") {
			t.Errorf("generated SQL contains SQL comment: %q", sql)
		}
	})
}

// FuzzInsertTable exercises table identifier quoting through InsertInto.
func FuzzInsertTable(f *testing.F) {
	seeds := []string{
		"users",
		"orders",
		"",
		"users; DROP TABLE users",
		"user`name",
		"a b",
		strings.Repeat("t", 65),
		"таблица",
		"users--",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		sql, _, err := sqlbuilder.InsertInto(input).
			Columns("id").
			Rows([]any{1}).
			Build()
		if err != nil {
			return
		}
		if strings.Contains(sql, ";") {
			t.Errorf("generated SQL contains semicolon: %q", sql)
		}
	})
}

// FuzzOnConditionLeftColumn exercises OnCondition column identifier validation.
func FuzzOnConditionLeftColumn(f *testing.F) {
	seeds := []string{
		"o.user_id",
		"u.id",
		"users.id",
		"",
		"a; DROP TABLE a",
		"col`name",
		"a b",
		strings.Repeat("c", 65),
		strings.Repeat("c", 64),
		"db.table.column",
		"col--",
		"таблица.поле",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		sql, _, err := sqlbuilder.Select("id").
			From("users u").
			JoinOn("orders o", sqlbuilder.OnEq(input, "u.id")).
			Build()
		if err != nil {
			return
		}
		if strings.Contains(sql, ";") {
			t.Errorf("generated SQL contains semicolon: %q", sql)
		}
		if strings.Contains(sql, "--") {
			t.Errorf("generated SQL contains SQL comment: %q", sql)
		}
		if strings.Contains(sql, "/*") {
			t.Errorf("generated SQL contains block comment: %q", sql)
		}
	})
}

// FuzzWhereConditionField exercises condition field identifier validation.
func FuzzWhereConditionField(f *testing.F) {
	seeds := []string{
		"id",
		"u.id",
		"status",
		"",
		"id; DROP TABLE users",
		"id--",
		"field`name",
		strings.Repeat("f", 65),
		"таблица",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		sql, _, err := sqlbuilder.Select("id").
			From("users").
			Where(sqlbuilder.Eq(input, "value")).
			Build()
		if err != nil {
			return
		}
		if strings.Contains(sql, ";") {
			t.Errorf("generated SQL contains semicolon: %q", sql)
		}
		if strings.Contains(sql, "--") {
			t.Errorf("generated SQL contains SQL comment: %q", sql)
		}
	})
}

// FuzzUpdateTable exercises UPDATE table identifier validation.
func FuzzUpdateTable(f *testing.F) {
	seeds := []string{
		"users",
		"",
		"users; DROP TABLE users",
		"user`name",
		"a b",
		strings.Repeat("t", 65),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		sql, _, err := sqlbuilder.Update(input).
			Set("status", "active").
			Build()
		if err != nil {
			return
		}
		if strings.Contains(sql, ";") {
			t.Errorf("generated SQL contains semicolon: %q", sql)
		}
	})
}

// FuzzDeleteTable exercises DELETE FROM table identifier validation.
func FuzzDeleteTable(f *testing.F) {
	seeds := []string{
		"users",
		"",
		"users; DROP TABLE users",
		"user`name",
		"a b",
		strings.Repeat("t", 65),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		sql, _, err := sqlbuilder.DeleteFrom(input).Build()
		if err != nil {
			return
		}
		if strings.Contains(sql, ";") {
			t.Errorf("generated SQL contains semicolon: %q", sql)
		}
	})
}
