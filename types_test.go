package cubrid

import (
	"database/sql/driver"
	"testing"
	"time"
)

func TestInterpolateArgsSkipsQuestionMarkInSingleQuotedString(t *testing.T) {
	sql := "SELECT * FROM t WHERE name = '?' AND id = ?"
	got, err := InterpolateArgs(sql, []driver.Value{int64(7)})
	if err != nil {
		t.Fatalf("InterpolateArgs returned error: %v", err)
	}

	want := "SELECT * FROM t WHERE name = '?' AND id = 7"
	if got != want {
		t.Fatalf("unexpected SQL\nwant: %s\n got: %s", want, got)
	}
}

func TestInterpolateArgsSkipsQuestionMarkInBlockComment(t *testing.T) {
	sql := "SELECT 1 /* ? */ WHERE id = ?"
	got, err := InterpolateArgs(sql, []driver.Value{int64(42)})
	if err != nil {
		t.Fatalf("InterpolateArgs returned error: %v", err)
	}

	want := "SELECT 1 /* ? */ WHERE id = 42"
	if got != want {
		t.Fatalf("unexpected SQL\nwant: %s\n got: %s", want, got)
	}
}

func TestInterpolateArgsSkipsQuestionMarkInLineComment(t *testing.T) {
	sql := "SELECT 1 -- ? \nWHERE id = ?"
	got, err := InterpolateArgs(sql, []driver.Value{int64(5)})
	if err != nil {
		t.Fatalf("InterpolateArgs returned error: %v", err)
	}

	want := "SELECT 1 -- ? \nWHERE id = 5"
	if got != want {
		t.Fatalf("unexpected SQL\nwant: %s\n got: %s", want, got)
	}
}

func TestInterpolateArgsHandlesEscapedSingleQuote(t *testing.T) {
	sql := "SELECT * FROM t WHERE name = 'it''s?' AND id = ?"
	got, err := InterpolateArgs(sql, []driver.Value{int64(9)})
	if err != nil {
		t.Fatalf("InterpolateArgs returned error: %v", err)
	}

	want := "SELECT * FROM t WHERE name = 'it''s?' AND id = 9"
	if got != want {
		t.Fatalf("unexpected SQL\nwant: %s\n got: %s", want, got)
	}
}

func TestInterpolateArgsWithNoPlaceholders(t *testing.T) {
	sql := "SELECT 1"
	got, err := InterpolateArgs(sql, nil)
	if err != nil {
		t.Fatalf("InterpolateArgs returned error: %v", err)
	}

	if got != sql {
		t.Fatalf("unexpected SQL\nwant: %s\n got: %s", sql, got)
	}
}

func TestInterpolateArgsMultiplePlaceholders(t *testing.T) {
	sql := "SELECT * FROM t WHERE a = ? AND b = ?"
	got, err := InterpolateArgs(sql, []driver.Value{int64(1), "x"})
	if err != nil {
		t.Fatalf("InterpolateArgs returned error: %v", err)
	}

	want := "SELECT * FROM t WHERE a = 1 AND b = 'x'"
	if got != want {
		t.Fatalf("unexpected SQL\nwant: %s\n got: %s", want, got)
	}
}

func TestFormatValueNormalizesTimeToUTC(t *testing.T) {
	loc := time.FixedZone("KST", 9*60*60)
	input := time.Date(2024, 1, 15, 9, 30, 0, 123000000, loc)

	got, err := FormatValue(input)
	if err != nil {
		t.Fatalf("FormatValue returned error: %v", err)
	}

	want := "DATETIME'2024-01-15 00:30:00.123'"
	if got != want {
		t.Fatalf("unexpected DATETIME literal\nwant: %s\n got: %s", want, got)
	}
}
