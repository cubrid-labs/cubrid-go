package cubrid

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// interpolateArgs replaces `?` placeholders in sql with the supplied
// driver.Value arguments, returning a complete SQL string ready to send.
// This mirrors pycubrid's client-side parameter binding.
func interpolateArgs(sql string, args []driver.Value) (string, error) {
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
			formatted, err := formatValue(args[i])
			if err != nil {
				return "", err
			}
			sb.WriteString(formatted)
		}
	}
	return sb.String(), nil
}

// formatValue converts a driver.Value into a CUBRID SQL literal.
func formatValue(v driver.Value) (string, error) {
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
		return "'" + val.Format("2006-01-02 15:04:05.000") + "'", nil
	default:
		return "", fmt.Errorf("cubrid: unsupported value type %T", v)
	}
}

// escapeString escapes single quotes and backslashes for CUBRID string literals.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `''`)
	return s
}

// hexEncode returns the hex encoding of b (lowercase).
func hexEncode(b []byte) string {
	const hx = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hx[c>>4]
		out[i*2+1] = hx[c&0x0f]
	}
	return string(out)
}

// namedValueToValue converts []driver.NamedValue to []driver.Value.
func namedValueToValue(named []driver.NamedValue) []driver.Value {
	out := make([]driver.Value, len(named))
	for i, n := range named {
		out[i] = n.Value
	}
	return out
}
