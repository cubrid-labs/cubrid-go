package cubrid

import "fmt"

// CubridError is the base error type returned by the CUBRID driver.
type CubridError struct {
	Code    int
	Message string
}

func (e *CubridError) Error() string {
	return fmt.Sprintf("cubrid: [%d] %s", e.Code, e.Message)
}

// IntegrityError is raised for constraint violations
// (unique key, foreign key, etc.).
type IntegrityError struct{ CubridError }

// ProgrammingError is raised for SQL syntax errors or missing objects.
type ProgrammingError struct{ CubridError }

// OperationalError is raised for network or server-side operational issues.
type OperationalError struct{ CubridError }

// newError inspects the message and returns the most specific error type.
func newError(code int32, message string) error {
	base := CubridError{Code: int(code), Message: message}
	msg := message
	for _, kw := range []string{"unique", "duplicate", "foreign key", "constraint violation"} {
		if containsCI(msg, kw) {
			return &IntegrityError{base}
		}
	}
	for _, kw := range []string{"syntax", "unknown class", "does not exist", "not found"} {
		if containsCI(msg, kw) {
			return &ProgrammingError{base}
		}
	}
	return &base
}

func containsCI(s, sub string) bool {
	return len(s) >= len(sub) && indexCI(s, sub) >= 0
}

func indexCI(s, sub string) int {
	sLow := toLower(s)
	subLow := toLower(sub)
	for i := 0; i <= len(sLow)-len(subLow); i++ {
		if sLow[i:i+len(subLow)] == subLow {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
