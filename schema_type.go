package builder

import (
	"fmt"
	"regexp"
	"strings"
)

// ColumnTypeDef represents a MySQL column type.
// Create via a plain string in Column() (validated) or RawType() (unsafe, bypasses validation).
type ColumnTypeDef struct {
	sql string
	raw bool
}

// String returns the SQL representation of the type.
func (t ColumnTypeDef) String() string { return t.sql }

// RawType creates an unvalidated column type for non-standard or unsupported types.
//
// WARNING: RawType bypasses whitelist validation. Only use for known, trusted type strings.
// Never pass user input to RawType.
func RawType(typ string) ColumnTypeDef {
	return ColumnTypeDef{sql: typ, raw: true}
}

// validTypePatterns is the whitelist of supported MySQL 8 column types.
var validTypePatterns = []*regexp.Regexp{
	// Integer types (with optional display width and UNSIGNED/ZEROFILL)
	regexp.MustCompile(`(?i)^(TINY|SMALL|MEDIUM|BIG)?INT(\(\d+\))?(\s+UNSIGNED)?(\s+ZEROFILL)?$`),
	// Fixed-point
	regexp.MustCompile(`(?i)^DECIMAL(\(\d+\s*,\s*\d+\))?(\s+UNSIGNED)?$`),
	regexp.MustCompile(`(?i)^NUMERIC(\(\d+\s*,\s*\d+\))?(\s+UNSIGNED)?$`),
	// Floating-point
	regexp.MustCompile(`(?i)^(FLOAT|DOUBLE|REAL)(\(\d+\s*,\s*\d+\))?(\s+UNSIGNED)?$`),
	// String types
	regexp.MustCompile(`(?i)^(VAR)?CHAR\(\d+\)$`),
	regexp.MustCompile(`(?i)^(TINY|MEDIUM|LONG)?TEXT$`),
	regexp.MustCompile(`(?i)^TINYTEXT$`),
	// Binary string types
	regexp.MustCompile(`(?i)^(VAR)?BINARY(\(\d+\))?$`),
	regexp.MustCompile(`(?i)^(TINY|MEDIUM|LONG)?BLOB$`),
	// Date/time types
	regexp.MustCompile(`(?i)^DATE$`),
	regexp.MustCompile(`(?i)^TIME(\(\d\))?$`),
	regexp.MustCompile(`(?i)^DATETIME(\(\d\))?$`),
	regexp.MustCompile(`(?i)^TIMESTAMP(\(\d\))?$`),
	regexp.MustCompile(`(?i)^YEAR(\(4\))?$`),
	// Boolean
	regexp.MustCompile(`(?i)^(BOOL|BOOLEAN)$`),
	// JSON
	regexp.MustCompile(`(?i)^JSON$`),
	// Bit
	regexp.MustCompile(`(?i)^BIT(\(\d+\))?$`),
	// ENUM and SET: values must be single-quoted; '' is the only allowed escape inside a value.
	// Examples: ENUM('a','b'), SET('x','y','z')
	regexp.MustCompile(`(?i)^ENUM\('(?:[^']|'')*'(?:\s*,\s*'(?:[^']|'')*')*\)$`),
	regexp.MustCompile(`(?i)^SET\('(?:[^']|'')*'(?:\s*,\s*'(?:[^']|'')*')*\)$`),
	// Spatial types
	regexp.MustCompile(`(?i)^(GEOMETRY|POINT|LINESTRING|POLYGON|MULTI(POINT|LINESTRING|POLYGON)|GEOMETRYCOLLECTION)$`),
}

// validateColumnType returns an error if typ is not a recognised MySQL 8 column type.
// Type comparison is case-insensitive. Pass RawType() to bypass.
func validateColumnType(typ string) error {
	t := strings.TrimSpace(typ)
	if t == "" {
		return ErrEmptyColumnType
	}
	for _, re := range validTypePatterns {
		if re.MatchString(t) {
			return nil
		}
	}
	return fmt.Errorf("%w: %q (use RawType() to bypass validation)", ErrInvalidColumnType, typ)
}

// normalizeColumnType uppercases the base keyword while preserving numeric parameters.
func normalizeColumnType(typ string) string {
	return strings.ToUpper(strings.TrimSpace(typ))
}
