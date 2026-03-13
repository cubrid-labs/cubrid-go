package cubrid

import (
	"testing"
)

func TestNewError_integrity(t *testing.T) {
	cases := []string{
		"unique key violation",
		"duplicate entry",
		"foreign key constraint",
		"constraint violation occurred",
	}
	for _, msg := range cases {
		err := newError(-1, msg)
		if _, ok := err.(*IntegrityError); !ok {
			t.Errorf("message %q: want *IntegrityError, got %T", msg, err)
		}
	}
}

func TestNewError_programming(t *testing.T) {
	cases := []string{
		"syntax error in SQL",
		"unknown class 'foo'",
		"table does not exist",
		"column not found",
	}
	for _, msg := range cases {
		err := newError(-1, msg)
		if _, ok := err.(*ProgrammingError); !ok {
			t.Errorf("message %q: want *ProgrammingError, got %T", msg, err)
		}
	}
}

func TestNewError_generic(t *testing.T) {
	err := newError(-940, "network timeout")
	if _, ok := err.(*CubridError); !ok {
		t.Errorf("want *CubridError, got %T", err)
	}
}

func TestCubridError_Error(t *testing.T) {
	err := &CubridError{Code: -123, Message: "test error"}
	want := "cubrid: [-123] test error"
	if err.Error() != want {
		t.Errorf("want %q, got %q", want, err.Error())
	}
}

func TestNewError_caseInsensitive(t *testing.T) {
	err := newError(-1, "UNIQUE KEY VIOLATION")
	if _, ok := err.(*IntegrityError); !ok {
		t.Errorf("want *IntegrityError for uppercase msg, got %T", err)
	}
}
