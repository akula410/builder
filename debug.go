package builder

// DebugSQL and related helpers are FOR LOGGING ONLY.
//
// WARNING: The output of DebugSQL must NEVER be executed as SQL.
// It substitutes placeholder values directly into the query string
// for human-readable display. The result is NOT safe to execute.
//
// In production, log the parameterised SQL and args_count only:
//
//	log.Printf("sql=%s args_count=%d", sql, len(args))
//
// Only use debug output in local/development environments.

import (
	"fmt"
	"strings"
	"time"
)

// DebugConfig controls how DebugSQL formats values.
type DebugConfig struct {
	// MaxStringLength truncates strings longer than this value. 0 = no limit.
	MaxStringLength int
	// ShowBytes controls whether []byte values are shown as hex or "[bytes N]".
	ShowBytes bool
	// TimeFormat is the format string for time.Time values.
	// Defaults to "2006-01-02 15:04:05".
	TimeFormat string
	// RedactFields is a list of lowercase column names whose values are replaced with RedactValue.
	// Matching is done against the previous placeholder token in the SQL, best-effort.
	RedactFields []string
	// RedactValue is the replacement for redacted fields. Defaults to "[REDACTED]".
	RedactValue string
}

var defaultDebugConfig = DebugConfig{
	MaxStringLength: 0,
	ShowBytes:       false,
	TimeFormat:      "2006-01-02 15:04:05",
	RedactValue:     "[REDACTED]",
}

// DebugSQL substitutes args into sql for human-readable display.
//
// WARNING: The result must never be executed as SQL.
// Use the default config. For custom config use DebugSQLWithConfig.
func DebugSQL(sql string, args []any) string {
	return DebugSQLWithConfig(sql, args, defaultDebugConfig)
}

// DebugSQLWithConfig substitutes args into sql using the provided config.
//
// WARNING: The result must never be executed as SQL.
func DebugSQLWithConfig(query string, args []any, cfg DebugConfig) string {
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = defaultDebugConfig.TimeFormat
	}
	if cfg.RedactValue == "" {
		cfg.RedactValue = defaultDebugConfig.RedactValue
	}

	redactSet := make(map[string]bool, len(cfg.RedactFields))
	for _, f := range cfg.RedactFields {
		redactSet[strings.ToLower(f)] = true
	}

	argIdx := 0
	var buf strings.Builder
	buf.Grow(len(query) + len(args)*8)

	for i := 0; i < len(query); i++ {
		if query[i] != '?' {
			buf.WriteByte(query[i])
			continue
		}
		if argIdx >= len(args) {
			buf.WriteByte('?')
			continue
		}

		arg := args[argIdx]
		argIdx++

		// Best-effort redaction: check if the token before '?' contains a known field name.
		if len(redactSet) > 0 {
			preceding := strings.ToLower(extractPrecedingToken(query[:i]))
			if redactSet[preceding] {
				buf.WriteString("'" + cfg.RedactValue + "'")
				continue
			}
		}

		buf.WriteString(formatDebugArg(arg, cfg))
	}

	return buf.String()
}

// formatDebugArg formats a single arg value for debug display.
func formatDebugArg(arg any, cfg DebugConfig) string {
	if arg == nil {
		return "NULL"
	}
	switch v := arg.(type) {
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case int:
		return fmt.Sprintf("%d", v)
	case int8:
		return fmt.Sprintf("%d", v)
	case int16:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case uint:
		return fmt.Sprintf("%d", v)
	case uint8:
		return fmt.Sprintf("%d", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return fmt.Sprintf("%g", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case []byte:
		if cfg.ShowBytes {
			return fmt.Sprintf("0x%X", v)
		}
		return fmt.Sprintf("'[bytes %d]'", len(v))
	case time.Time:
		return "'" + v.Format(cfg.TimeFormat) + "'"
	case string:
		s := v
		if cfg.MaxStringLength > 0 && len(s) > cfg.MaxStringLength {
			s = s[:cfg.MaxStringLength] + "..."
		}
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	default:
		return "'" + strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''") + "'"
	}
}

// extractPrecedingToken extracts the last identifier before the current position,
// used for best-effort field-name detection in redaction.
// It strips trailing whitespace and comparison operators (=, !=, <, >, etc.)
// then looks for the last backtick-quoted name.
func extractPrecedingToken(sql string) string {
	s := strings.TrimRight(sql, " \t\n")
	// Strip trailing comparison operator tokens so "col` = " → "col`"
	s = strings.TrimRight(s, " \t\n=!<>")
	s = strings.TrimRight(s, " \t\n")
	// Find the last backtick-quoted identifier.
	if len(s) > 0 && s[len(s)-1] == '`' {
		start := strings.LastIndex(s[:len(s)-1], "`")
		if start >= 0 {
			return s[start+1 : len(s)-1]
		}
	}
	// Fall back to last space-delimited token.
	idx := strings.LastIndexAny(s, " \t\n(,")
	if idx >= 0 {
		return s[idx+1:]
	}
	return s
}

// Debug builds a QueryBuilder and returns a debug string.
//
// WARNING: The result must never be executed as SQL.
func Debug(b QueryBuilder) (string, error) {
	sql, args, err := b.Build()
	if err != nil {
		return "", err
	}
	return DebugSQL(sql, args), nil
}

// SafeLog returns a struct suitable for structured logging in production.
// It does NOT include arg values — only the SQL template and arg count.
type SafeLog struct {
	SQL       string `json:"sql"`
	ArgsCount int    `json:"args_count"`
}

// SafeLogQuery builds a QueryBuilder and returns a SafeLog for production logging.
func SafeLogQuery(b QueryBuilder) (SafeLog, error) {
	sql, args, err := b.Build()
	if err != nil {
		return SafeLog{}, err
	}
	return SafeLog{SQL: sql, ArgsCount: len(args)}, nil
}
