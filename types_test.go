package cubrid

import (
	"database/sql/driver"
	"testing"
	"time"
)

func TestFormatValue_nil(t *testing.T) {
	got, err := formatValue(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "NULL" {
		t.Errorf("want NULL, got %q", got)
	}
}

func TestFormatValue_bool(t *testing.T) {
	got, _ := formatValue(true)
	if got != "1" {
		t.Errorf("true: want '1', got %q", got)
	}
	got, _ = formatValue(false)
	if got != "0" {
		t.Errorf("false: want '0', got %q", got)
	}
}

func TestFormatValue_int64(t *testing.T) {
	got, _ := formatValue(int64(42))
	if got != "42" {
		t.Errorf("want '42', got %q", got)
	}
	got, _ = formatValue(int64(-100))
	if got != "-100" {
		t.Errorf("want '-100', got %q", got)
	}
}

func TestFormatValue_float64(t *testing.T) {
	got, _ := formatValue(float64(3.14))
	if got == "" {
		t.Error("want non-empty float string")
	}
}

func TestFormatValue_string(t *testing.T) {
	got, _ := formatValue("hello")
	if got != "'hello'" {
		t.Errorf("want \"'hello'\", got %q", got)
	}
}

func TestFormatValue_string_escape(t *testing.T) {
	got, _ := formatValue("it's a test")
	if got != "'it''s a test'" {
		t.Errorf("unexpected escaping: %q", got)
	}
}

func TestFormatValue_bytes(t *testing.T) {
	got, _ := formatValue([]byte{0xDE, 0xAD, 0xBE, 0xEF})
	if got != "X'deadbeef'" {
		t.Errorf("want \"X'deadbeef'\", got %q", got)
	}
}

func TestFormatValue_time(t *testing.T) {
	tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	got, err := formatValue(tm)
	if err != nil {
		t.Fatal(err)
	}
	if got != "DATETIME'2024-01-15 10:30:00.000'" {
		t.Errorf("unexpected datetime: %q", got)
	}
}

func TestInterpolateArgs_noArgs(t *testing.T) {
	got, err := interpolateArgs("SELECT 1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "SELECT 1" {
		t.Errorf("want 'SELECT 1', got %q", got)
	}
}

func TestInterpolateArgs_single(t *testing.T) {
	got, err := interpolateArgs("SELECT * FROM t WHERE id = ?", []driver.Value{int64(5)})
	if err != nil {
		t.Fatal(err)
	}
	want := "SELECT * FROM t WHERE id = 5"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestInterpolateArgs_multiple(t *testing.T) {
	got, err := interpolateArgs(
		"INSERT INTO t (a, b) VALUES (?, ?)",
		[]driver.Value{int64(1), "hello"},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := "INSERT INTO t (a, b) VALUES (1, 'hello')"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestInterpolateArgs_mismatch(t *testing.T) {
	_, err := interpolateArgs("SELECT ?", []driver.Value{int64(1), int64(2)})
	if err == nil {
		t.Fatal("expected error for arg count mismatch")
	}
}
