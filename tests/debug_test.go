package sqlbuilder_test

import (
	"testing"
	"time"

	sqlbuilder "github.com/akula410/builder/v2"
)

func TestDebugSQL_String(t *testing.T) {
	result := sqlbuilder.DebugSQL("WHERE `email` = ?", []any{"alex@example.com"})
	assertEqual(t, "WHERE `email` = 'alex@example.com'", result)
}

func TestDebugSQL_EscapeSingleQuote(t *testing.T) {
	result := sqlbuilder.DebugSQL("WHERE `name` = ?", []any{"O'Brien"})
	assertEqual(t, "WHERE `name` = 'O''Brien'", result)
}

func TestDebugSQL_Nil(t *testing.T) {
	result := sqlbuilder.DebugSQL("WHERE `deleted_at` = ?", []any{nil})
	assertEqual(t, "WHERE `deleted_at` = NULL", result)
}

func TestDebugSQL_Bool_True(t *testing.T) {
	result := sqlbuilder.DebugSQL("WHERE `active` = ?", []any{true})
	assertEqual(t, "WHERE `active` = TRUE", result)
}

func TestDebugSQL_Bool_False(t *testing.T) {
	result := sqlbuilder.DebugSQL("WHERE `active` = ?", []any{false})
	assertEqual(t, "WHERE `active` = FALSE", result)
}

func TestDebugSQL_Int(t *testing.T) {
	result := sqlbuilder.DebugSQL("WHERE `id` = ?", []any{42})
	assertEqual(t, "WHERE `id` = 42", result)
}

func TestDebugSQL_Int64(t *testing.T) {
	result := sqlbuilder.DebugSQL("WHERE `id` = ?", []any{int64(999)})
	assertEqual(t, "WHERE `id` = 999", result)
}

func TestDebugSQL_Float(t *testing.T) {
	result := sqlbuilder.DebugSQL("WHERE `price` > ?", []any{float64(9.99)})
	assertContains(t, result, "9.99")
}

func TestDebugSQL_Time(t *testing.T) {
	tm := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := sqlbuilder.DebugSQL("WHERE `created_at` = ?", []any{tm})
	assertEqual(t, "WHERE `created_at` = '2025-01-15 12:00:00'", result)
}

func TestDebugSQL_Bytes_ShowBytes(t *testing.T) {
	cfg := sqlbuilder.DebugConfig{ShowBytes: true, RedactValue: "[REDACTED]"}
	result := sqlbuilder.DebugSQLWithConfig("WHERE `data` = ?", []any{[]byte{0xAB, 0xCD}}, cfg)
	assertContains(t, result, "0x")
}

func TestDebugSQL_Bytes_Hide(t *testing.T) {
	result := sqlbuilder.DebugSQL("WHERE `data` = ?", []any{[]byte{0xAB, 0xCD}})
	assertContains(t, result, "[bytes 2]")
}

func TestDebugSQL_MaxStringLength(t *testing.T) {
	cfg := sqlbuilder.DebugConfig{MaxStringLength: 5, RedactValue: "[REDACTED]"}
	result := sqlbuilder.DebugSQLWithConfig("WHERE `x` = ?", []any{"hello world"}, cfg)
	assertContains(t, result, "hello...")
}

func TestDebugSQL_RedactField(t *testing.T) {
	cfg := sqlbuilder.DebugConfig{
		RedactFields: []string{"password"},
		RedactValue:  "[REDACTED]",
	}
	result := sqlbuilder.DebugSQLWithConfig(
		"WHERE `password` = ?",
		[]any{"secret123"},
		cfg,
	)
	assertContains(t, result, "[REDACTED]")
	if containsStr(result, "secret123") {
		t.Fatal("sensitive value not redacted")
	}
}

func TestDebugSQL_MultipleArgs(t *testing.T) {
	sql := "WHERE `email` = ? AND `created_at` > ?"
	args := []any{"alex@example.com", "2025-01-01"}
	result := sqlbuilder.DebugSQL(sql, args)
	assertContains(t, result, "'alex@example.com'")
	assertContains(t, result, "'2025-01-01'")
}

func TestDebug_QueryBuilder(t *testing.T) {
	q := sqlbuilder.Select("id", "name").
		From("users").
		Where(sqlbuilder.Eq("email", "alex@example.com"))

	result, err := sqlbuilder.Debug(q)
	assertNoErr(t, err)
	assertContains(t, result, "'alex@example.com'")
}

func TestSafeLogQuery(t *testing.T) {
	q := sqlbuilder.Select("id").From("users").Where(sqlbuilder.Eq("email", "x@x.com"))
	log, err := sqlbuilder.SafeLogQuery(q)
	assertNoErr(t, err)
	if log.ArgsCount != 1 {
		t.Fatalf("expected ArgsCount=1, got %d", log.ArgsCount)
	}
	assertContains(t, log.SQL, "?")
	// The actual value must NOT appear in the safe log SQL
	if containsStr(log.SQL, "x@x.com") {
		t.Fatal("value leaked into safe log")
	}
}
