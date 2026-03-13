package cubrid

import (
	"testing"
	"time"
)

func TestParseDSN_basic(t *testing.T) {
	cfg, err := parseDSN("cubrid://dba:secret@localhost:33000/demodb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.host != "localhost" {
		t.Errorf("host: want 'localhost', got %q", cfg.host)
	}
	if cfg.port != 33000 {
		t.Errorf("port: want 33000, got %d", cfg.port)
	}
	if cfg.database != "demodb" {
		t.Errorf("database: want 'demodb', got %q", cfg.database)
	}
	if cfg.user != "dba" {
		t.Errorf("user: want 'dba', got %q", cfg.user)
	}
	if cfg.password != "secret" {
		t.Errorf("password: want 'secret', got %q", cfg.password)
	}
	if !cfg.autoCommit {
		t.Error("autocommit: want true (default)")
	}
}

func TestParseDSN_emptyPassword(t *testing.T) {
	cfg, err := parseDSN("cubrid://dba:@localhost:33000/demodb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.user != "dba" {
		t.Errorf("user: want 'dba', got %q", cfg.user)
	}
	if cfg.password != "" {
		t.Errorf("password: want '', got %q", cfg.password)
	}
}

func TestParseDSN_defaults(t *testing.T) {
	cfg, err := parseDSN("cubrid://localhost/demodb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.port != 33000 {
		t.Errorf("port default: want 33000, got %d", cfg.port)
	}
	if cfg.timeout != 30*time.Second {
		t.Errorf("timeout default: want 30s, got %v", cfg.timeout)
	}
}

func TestParseDSN_queryParams(t *testing.T) {
	cfg, err := parseDSN("cubrid://dba:@localhost:33000/demodb?autocommit=false&timeout=10s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.autoCommit {
		t.Error("autocommit: want false, got true")
	}
	if cfg.timeout != 10*time.Second {
		t.Errorf("timeout: want 10s, got %v", cfg.timeout)
	}
}

func TestParseDSN_wrongScheme(t *testing.T) {
	_, err := parseDSN("mysql://dba:@localhost:33000/demodb")
	if err == nil {
		t.Fatal("expected error for wrong scheme")
	}
}

func TestParseDSN_missingDatabase(t *testing.T) {
	_, err := parseDSN("cubrid://dba:@localhost:33000/")
	if err == nil {
		t.Fatal("expected error for missing database")
	}
}

func TestParseDSN_invalidPort(t *testing.T) {
	_, err := parseDSN("cubrid://dba:@localhost:abc/demodb")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestParseDSN_customPort(t *testing.T) {
	cfg, err := parseDSN("cubrid://dba:@myhost:12345/testdb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.port != 12345 {
		t.Errorf("port: want 12345, got %d", cfg.port)
	}
	if cfg.host != "myhost" {
		t.Errorf("host: want 'myhost', got %q", cfg.host)
	}
}
