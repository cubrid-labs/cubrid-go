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
	placeholders := findBindPlaceholders(sql)
	if len(placeholders) != len(args) {
		return "", fmt.Errorf(
			"cubrid: expected %d bind args, got %d",
			len(placeholders), len(args),
		)
	}
	if len(placeholders) == 0 {
		return sql, nil
	}

	var sb strings.Builder
	prev := 0
	for i, pos := range placeholders {
		sb.WriteString(sql[prev:pos])
		formatted, err := FormatValue(args[i])
		if err != nil {
			return "", err
		}
		sb.WriteString(formatted)
		prev = pos + 1
	}
	sb.WriteString(sql[prev:])
	return sb.String(), nil
}

func findBindPlaceholders(sql string) []int {
	const (
		scanNormal = iota
		scanSingleQuote
		scanBlockComment
		scanLineComment
	)

	state := scanNormal
	positions := make([]int, 0, 8)

	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		switch state {
		case scanNormal:
			switch ch {
			case '\'':
				state = scanSingleQuote
			case '/':
				if i+1 < len(sql) && sql[i+1] == '*' {
					state = scanBlockComment
					i++
				}
			case '-':
				if i+1 < len(sql) && sql[i+1] == '-' {
					state = scanLineComment
					i++
				}
			case '?':
				positions = append(positions, i)
			}
		case scanSingleQuote:
			if ch == '\'' {
				if i+1 < len(sql) && sql[i+1] == '\'' {
					i++
					continue
				}
				state = scanNormal
			}
		case scanBlockComment:
			if ch == '*' && i+1 < len(sql) && sql[i+1] == '/' {
				state = scanNormal
				i++
			}
		case scanLineComment:
			if ch == '\n' {
				state = scanNormal
			}
		}
	}

	return positions
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
		val = val.UTC()
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
