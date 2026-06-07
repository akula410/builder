package sqlbuilder_test

import (
	"strings"
	"testing"
)

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertEqual(t *testing.T, want, got string) {
	t.Helper()
	if want != got {
		t.Fatalf("SQL mismatch:\nwant: %q\n got: %q", want, got)
	}
}

func assertContains(t *testing.T, sql, substr string) {
	t.Helper()
	if !strings.Contains(sql, substr) {
		t.Fatalf("SQL does not contain %q:\n%s", substr, sql)
	}
}

func assertArgs(t *testing.T, want, got []any) {
	t.Helper()
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if len(want) != len(got) {
		t.Fatalf("args length mismatch: want %d (%v), got %d (%v)", len(want), want, len(got), got)
	}
	for i := range want {
		if want[i] != got[i] {
			t.Fatalf("args[%d] mismatch: want %v, got %v", i, want[i], got[i])
		}
	}
}
