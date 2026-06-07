package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func BenchmarkUpdateWithWhere(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Update("users").
			Set("status", "inactive").
			Set("updated_at", "2026-06-07").
			Where(sqlbuilder.Eq("id", 42)).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateMultipleWhere(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Update("users").
			Set("status", "inactive").
			Where(sqlbuilder.Eq("status", "active")).
			Where(sqlbuilder.Lt("last_login", "2024-01-01")).
			Limit(100).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeleteWithWhere(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.DeleteFrom("users").
			Where(sqlbuilder.Eq("id", 42)).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeleteMultipleWhere(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.DeleteFrom("sessions").
			Where(sqlbuilder.Eq("user_id", 42)).
			Where(sqlbuilder.Lt("expires_at", "2026-01-01")).
			Limit(50).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONConditionEq(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id").
			From("orders").
			Where(sqlbuilder.JSONEq("data", "$.status", "active")).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONConditionContains(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id").
			From("orders").
			Where(sqlbuilder.JSONContains("data", "$.tags", "urgent")).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONExtractColumn(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id", sqlbuilder.JSONExtract("data", "$.status")).
			From("orders").
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDebugSQL(b *testing.B) {
	sql, args, err := sqlbuilder.Select("id", "name", "email").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		Where(sqlbuilder.Gt("age", 18)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sqlbuilder.DebugSQL(sql, args)
	}
}

func BenchmarkDebugSQLWithConfig(b *testing.B) {
	sql, args, err := sqlbuilder.Select("id", "name", "email").
		From("users").
		Where(sqlbuilder.Eq("status", "active")).
		Where(sqlbuilder.Eq("password", "secret")).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	cfg := sqlbuilder.DebugConfig{
		MaxStringLength: 100,
		RedactFields:    []string{"password", "token"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sqlbuilder.DebugSQLWithConfig(sql, args, cfg)
	}
}

func BenchmarkSafeJoinOnEq(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("u.id", "u.name").
			From("users u").
			LeftJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
			Where(sqlbuilder.Eq("u.status", "active")).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSafeJoinOnAnd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("u.id", "u.name").
			From("users u").
			LeftJoinOn("orders o", sqlbuilder.OnAnd(
				sqlbuilder.OnEq("o.user_id", "u.id"),
				sqlbuilder.OnEq("o.status", "u.status"),
			)).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConditionAnd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id").
			From("users").
			Where(sqlbuilder.And(
				sqlbuilder.Eq("status", "active"),
				sqlbuilder.Gt("age", 18),
				sqlbuilder.IsNotNull("email"),
			)).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConditionOr(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id").
			From("users").
			Where(sqlbuilder.Or(
				sqlbuilder.Eq("role", "admin"),
				sqlbuilder.Eq("role", "moderator"),
				sqlbuilder.Eq("role", "superuser"),
			)).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}
