package builder

import "strings"

// mergeArgs concatenates multiple arg slices into one.
func mergeArgs(slices ...[]any) []any {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	if total == 0 {
		return nil
	}
	result := make([]any, 0, total)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// placeholders returns n comma-separated "?" placeholders.
func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = "?"
	}
	return strings.Join(parts, ", ")
}
