package cubrid

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// InterpolateArgs replaces `?` placeholders with formatted argument literals.
// This is used only for logging / GORM's Explain method — NOT for actual
// query execution, which sends typed bind parameters over the wire (FC=3).
func InterpolateArgs(sql string, args []driver.Value) (string, error) {
	if len(args) == 0 {
		return sql, nil
	}

	parts := strings.Split(sql, "?")
	if len(parts)-1 != len(args) {
		return "", fmt.Errorf(
			"cubrid: expected %d bind args, got %d",
			len(parts)-1, len(args),
		)
	}

	var sb strings.Builder
	for i, part := range parts {
		sb.WriteString(part)
		if i < len(args) {
			formatted, err := FormatValue(args[i])
			if err != nil {
				return "", err
			}
			sb.WriteString(formatted)
		}
	}
	return sb.String(), nil
}

// FormatValue converts a driver.Value to a CUBRID SQL literal string.
// Used by InterpolateArgs and the GORM dialector's Explain method.
func FormatValue(v driver.Value) (string, error) {
	if v == nil {
		return "NULL", nil
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "1", nil
		}
		return "0", nil
	case int64:
		return strconv.FormatInt(val, 10), nil
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64), nil
	case string:
		return "'" + escapeString(val) + "'", nil
	case []byte:
		return "X'" + hexEncode(val) + "'", nil
	case time.Time:
		ms := val.Nanosecond() / 1e6
		return fmt.Sprintf("DATETIME'%s.%03d'", val.Format("2006-01-02 15:04:05"), ms), nil
	default:
		return "", fmt.Errorf("cubrid: unsupported value type %T", v)
	}
}

// namedValueToValue converts []driver.NamedValue to []driver.Value.
func namedValueToValue(named []driver.NamedValue) []driver.Value {
	out := make([]driver.Value, len(named))
	for i, n := range named {
		out[i] = n.Value
	}
	return out
}

// escapeString escapes single quotes and backslashes for CUBRID string literals.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `''`)
	return s
}

// hexEncode returns the lowercase hex encoding of b.
func hexEncode(b []byte) string {
	const hx = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hx[c>>4]
		out[i*2+1] = hx[c&0x0f]
	}
	return string(out)
}
