//go:build integration

package cubrid_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/cubrid-labs/cubrid-go"
)

// testDSN returns the DSN from the CUBRID_DSN environment variable.
// Run integration tests with:
//
//	CUBRID_DSN=cubrid://dba:@localhost:33000/demodb go test -tags integration ./...
func testDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("CUBRID_DSN")
	if dsn == "" {
		t.Skip("CUBRID_DSN not set; skipping integration test")
	}
	return dsn
}

func TestIntegration_Ping(t *testing.T) {
	db, err := sql.Open("cubrid", testDSN(t))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestIntegration_QueryOne(t *testing.T) {
	db, err := sql.Open("cubrid", testDSN(t))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var n int
	if err := db.QueryRow("SELECT 1 FROM db_root").Scan(&n); err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if n != 1 {
		t.Errorf("want 1, got %d", n)
	}
}

func TestIntegration_PreparedStmt(t *testing.T) {
	db, err := sql.Open("cubrid", testDSN(t))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	// Create a temp table.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS _go_test_tmp (id INTEGER, val VARCHAR(64))`); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	defer db.Exec(`DROP TABLE IF EXISTS _go_test_tmp`)

	// Insert with prepared stmt.
	if _, err := db.Exec(`INSERT INTO _go_test_tmp VALUES (?, ?)`, 1, "hello"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	// Query back.
	var id int
	var val string
	if err := db.QueryRow(`SELECT id, val FROM _go_test_tmp WHERE id = ?`, 1).Scan(&id, &val); err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if id != 1 || val != "hello" {
		t.Errorf("want (1, 'hello'), got (%d, %q)", id, val)
	}
}

func TestIntegration_Transaction_Commit(t *testing.T) {
	db, err := sql.Open("cubrid", testDSN(t)+"?autocommit=false")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	db.Exec(`CREATE TABLE IF NOT EXISTS _go_test_tx (id INTEGER)`)
	defer db.Exec(`DROP TABLE IF EXISTS _go_test_tx`)

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if _, err := tx.Exec(`INSERT INTO _go_test_tx VALUES (?)`, 42); err != nil {
		tx.Rollback()
		t.Fatalf("INSERT: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM _go_test_tx WHERE id = ?`, 42).Scan(&count)
	if count != 1 {
		t.Errorf("want 1 row after commit, got %d", count)
	}
}

func TestIntegration_Transaction_Rollback(t *testing.T) {
	db, err := sql.Open("cubrid", testDSN(t)+"?autocommit=false")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	db.Exec(`CREATE TABLE IF NOT EXISTS _go_test_rb (id INTEGER)`)
	defer db.Exec(`DROP TABLE IF EXISTS _go_test_rb`)

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	tx.Exec(`INSERT INTO _go_test_rb VALUES (?)`, 99)
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM _go_test_rb`).Scan(&count)
	if count != 0 {
		t.Errorf("want 0 rows after rollback, got %d", count)
	}
}

func TestIntegration_ServerVersion(t *testing.T) {
	dsn := testDSN(t)
	db, err := sql.Open("cubrid", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	// Use a raw connection to call ServerVersion.
	rawConn, err := db.Driver().Open(dsn)
	if err != nil {
		t.Fatalf("driver.Open: %v", err)
	}
	defer rawConn.Close()

	type versionGetter interface {
		ServerVersion() (string, error)
	}
	if vg, ok := rawConn.(versionGetter); ok {
		ver, err := vg.ServerVersion()
		if err != nil {
			t.Fatalf("ServerVersion: %v", err)
		}
		if ver == "" {
			t.Error("ServerVersion returned empty string")
		}
		t.Logf("CUBRID server version: %s", ver)
	} else {
		t.Skip("conn does not implement ServerVersion")
	}
}
