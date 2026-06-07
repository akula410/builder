package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func BenchmarkSelectSimple(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id", "name", "email").
			From("users").
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectStar(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("*").From("users").Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectWhere(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id", "name", "email").
			From("users").
			Where(sqlbuilder.Eq("status", "active")).
			Where(sqlbuilder.Gt("age", 18)).
			Where(sqlbuilder.IsNotNull("email")).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectOrderByLimitOffset(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id", "name", "email").
			From("users").
			Where(sqlbuilder.Eq("status", "active")).
			OrderBy("created_at", sqlbuilder.Desc).
			Limit(20).
			Offset(40).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectJoinRaw(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("u.id", "u.name", "o.total").
			From("users u").
			LeftJoinRaw("orders o", "o.user_id = u.id").
			Where(sqlbuilder.Eq("u.status", "active")).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectJoinSafe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("u.id", "u.name", "o.total").
			From("users u").
			LeftJoinOn("orders o", sqlbuilder.OnEq("o.user_id", "u.id")).
			Where(sqlbuilder.Eq("u.status", "active")).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectComplexConditions(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id", "name").
			From("users").
			Where(sqlbuilder.And(
				sqlbuilder.Eq("status", "active"),
				sqlbuilder.Or(
					sqlbuilder.Gt("age", 18),
					sqlbuilder.IsNull("age"),
				),
			)).
			OrderBy("id", sqlbuilder.Desc).
			Limit(10).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectGroupByHaving(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("status", sqlbuilder.Raw("COUNT(*) AS cnt")).
			From("users").
			GroupBy("status").
			Having(sqlbuilder.Gt("cnt", 5)).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectDistinct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("status").
			Distinct().
			From("users").
			Where(sqlbuilder.IsNotNull("status")).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectInCondition(b *testing.B) {
	values := make([]any, 10)
	for i := range values {
		values[i] = i + 1
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.Select("id", "name").
			From("users").
			Where(sqlbuilder.In("id", values)).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}
