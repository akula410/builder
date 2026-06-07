package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func BenchmarkInsertSingleRowMap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.InsertInto("users").
			Values(map[string]any{
				"name":  "Alex",
				"email": "alex@example.com",
			}).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsertSingleRowColumns(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.InsertInto("users").
			Columns("name", "email", "status").
			Rows([]any{"Alex", "alex@example.com", "active"}).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsertBulk10(b *testing.B) {
	rows := make([][]any, 10)
	for i := range rows {
		rows[i] = []any{"name", "email@example.com", "active"}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.InsertInto("users").
			Columns("name", "email", "status").
			Rows(rows...).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsertBulk100(b *testing.B) {
	rows := make([][]any, 100)
	for i := range rows {
		rows[i] = []any{"name", "email@example.com", "active"}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.InsertInto("users").
			Columns("name", "email", "status").
			Rows(rows...).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsertBulk1000(b *testing.B) {
	rows := make([][]any, 1000)
	for i := range rows {
		rows[i] = []any{"name", "email@example.com", "active"}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.InsertInto("users").
			Columns("name", "email", "status").
			Rows(rows...).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsertOnDuplicateKeyUpdate(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := sqlbuilder.InsertInto("users").
			Values(map[string]any{
				"id":    1,
				"name":  "Alex",
				"email": "alex@example.com",
			}).
			OnDuplicateKeyUpdate(map[string]any{
				"name":  "Alex Updated",
				"email": "new@example.com",
			}).
			Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}
