package cubrid

import (
	"database/sql/driver"
	"testing"
)

var testCASInfo = [SizeCASInfo]byte{0x01, 0x00, 0x00, 0x00}

// ─── writeClientInfoExchange ──────────────────────────────────────────────────

func TestWriteClientInfoExchange_length(t *testing.T) {
	pkt := writeClientInfoExchange()
	if len(pkt) != 10 {
		t.Fatalf("want 10 bytes, got %d", len(pkt))
	}
}

func TestWriteClientInfoExchange_magic(t *testing.T) {
	pkt := writeClientInfoExchange()
	if string(pkt[:5]) != MagicString {
		t.Errorf("want magic %q, got %q", MagicString, string(pkt[:5]))
	}
	if pkt[5] != ClientJDBC {
		t.Errorf("want ClientJDBC %d, got %d", ClientJDBC, pkt[5])
	}
}

// ─── writeOpenDatabase ────────────────────────────────────────────────────────

func TestWriteOpenDatabase_length(t *testing.T) {
	pkt := writeOpenDatabase("demodb", "dba", "")
	// 32 + 32 + 32 + 512 + 20 = 628
	if len(pkt) != 628 {
		t.Fatalf("want 628 bytes, got %d", len(pkt))
	}
}

func TestWriteOpenDatabase_fields(t *testing.T) {
	pkt := writeOpenDatabase("demodb", "dba", "pass")
	if string(pkt[:6]) != "demodb" {
		t.Errorf("database prefix: want 'demodb', got %q", string(pkt[:6]))
	}
	if string(pkt[32:35]) != "dba" {
		t.Errorf("user prefix: want 'dba', got %q", string(pkt[32:35]))
	}
	if string(pkt[64:68]) != "pass" {
		t.Errorf("password prefix: want 'pass', got %q", string(pkt[64:68]))
	}
}

// ─── writeConClose ────────────────────────────────────────────────────────────

func TestWriteConClose_structure(t *testing.T) {
	pkt := writeConClose(testCASInfo)
	// header (8) + FuncConClose (1)
	if len(pkt) != 9 {
		t.Fatalf("want 9 bytes, got %d", len(pkt))
	}
	if pkt[8] != FuncConClose {
		t.Errorf("want FuncConClose %d, got %d", FuncConClose, pkt[8])
	}
}

// ─── writeEndTran ─────────────────────────────────────────────────────────────

func TestWriteEndTran_commit(t *testing.T) {
	pkt := writeEndTran(TxCommit, testCASInfo)
	// header (8) + func (1) + addByte: int32(1)+byte = 6 bytes total after header
	if pkt[8] != FuncEndTran {
		t.Errorf("want FuncEndTran %d, got %d", FuncEndTran, pkt[8])
	}
	// addByte writes int32(1) + byte
	if pkt[len(pkt)-1] != TxCommit {
		t.Errorf("want TxCommit %d, got %d", TxCommit, pkt[len(pkt)-1])
	}
}

func TestWriteEndTran_rollback(t *testing.T) {
	pkt := writeEndTran(TxRollback, testCASInfo)
	if pkt[len(pkt)-1] != TxRollback {
		t.Errorf("want TxRollback %d, got %d", TxRollback, pkt[len(pkt)-1])
	}
}

// ─── writeGetDbVersion ────────────────────────────────────────────────────────

func TestWriteGetDbVersion_structure(t *testing.T) {
	pkt := writeGetDbVersion(true, testCASInfo)
	if pkt[8] != FuncGetDbVersion {
		t.Errorf("want FuncGetDbVersion %d, got %d", FuncGetDbVersion, pkt[8])
	}
}

// ─── writeCloseReqHandle ─────────────────────────────────────────────────────

func TestWriteCloseReqHandle_structure(t *testing.T) {
	pkt := writeCloseReqHandle(42, testCASInfo)
	if pkt[8] != FuncCloseReqHandle {
		t.Errorf("want FuncCloseReqHandle %d, got %d", FuncCloseReqHandle, pkt[8])
	}
}

// ─── writePrepare ─────────────────────────────────────────────────────────────

func TestWritePrepare_structure(t *testing.T) {
	pkt := writePrepare("SELECT 1", true, testCASInfo)
	if pkt[8] != FuncPrepare {
		t.Errorf("want FuncPrepare %d, got %d", FuncPrepare, pkt[8])
	}
	// payload: FuncPrepare(1) + nullTermString("SELECT 1") + addByte(flag) + addByte(autocommit)
	// nullTermString = int32(9) + "SELECT 1" + 0x00 = 4+9 = 13 bytes
	if len(pkt) < 9 {
		t.Fatal("packet too short")
	}
}

// ─── writeFetch ───────────────────────────────────────────────────────────────

func TestWriteFetch_structure(t *testing.T) {
	pkt := writeFetch(5, 100, 50, testCASInfo)
	if pkt[8] != FuncFetch {
		t.Errorf("want FuncFetch %d, got %d", FuncFetch, pkt[8])
	}
}

// ─── encodeBindParams ─────────────────────────────────────────────────────────

func TestEncodeBindParams_nil(t *testing.T) {
	result := encodeBindParams(nil)
	if result != nil {
		t.Errorf("want nil for empty args, got %v", result)
	}
}

func TestEncodeBindParams_int64(t *testing.T) {
	result := encodeBindParams([]driver.Value{int64(42)})
	// int32(size) + TypeBigInt + int64 = 4 + 1 + 8 = 13 bytes
	if len(result) != 13 {
		t.Fatalf("want 13 bytes for int64 bind, got %d", len(result))
	}
	if result[4] != TypeBigInt {
		t.Errorf("want TypeBigInt %d, got %d", TypeBigInt, result[4])
	}
}

func TestEncodeBindParams_string(t *testing.T) {
	result := encodeBindParams([]driver.Value{"hello"})
	// int32(7) + TypeString + "hello" + 0x00 = 4 + 1 + 5 + 1 = 11 bytes
	if len(result) != 11 {
		t.Fatalf("want 11 bytes for 'hello' bind, got %d: %v", len(result), result)
	}
	if result[4] != TypeString {
		t.Errorf("want TypeString %d, got %d", TypeString, result[4])
	}
}

func TestEncodeBindParams_null(t *testing.T) {
	result := encodeBindParams([]driver.Value{nil})
	// int32(0) = 4 bytes
	if len(result) != 4 {
		t.Fatalf("want 4 bytes for NULL bind, got %d", len(result))
	}
	for _, b := range result {
		if b != 0 {
			t.Errorf("want all zero bytes for NULL, got %v", result)
			break
		}
	}
}

// ─── parseSimpleResponse ──────────────────────────────────────────────────────

func TestParseSimpleResponse_ok(t *testing.T) {
	// CASInfo(4) + int32(0) = success
	data := make([]byte, SizeCASInfo+SizeInt)
	err := parseSimpleResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseSimpleResponse_error(t *testing.T) {
	w := newPacketWriter()
	w.writeFiller(SizeCASInfo) // CASInfo
	w.writeInt(-111)           // error code
	w.writeInt(-111)           // duplicate (readError reads code again)
	// error message
	msg := "syntax error"
	w.writeRawBytes([]byte(msg))
	w.writeByte(0x00)

	err := parseSimpleResponse(w.toBytes())
	if err == nil {
		t.Fatal("expected error")
	}
}

// ─── rows.ColumnTypeDatabaseTypeName ─────────────────────────────────────────

func TestColumnTypeDatabaseTypeName(t *testing.T) {
	cases := []struct {
		colType int
		want    string
	}{
		{TypeChar, "CHAR"},
		{TypeString, "VARCHAR"},
		{TypeShort, "SMALLINT"},
		{TypeInt, "INTEGER"},
		{TypeBigInt, "BIGINT"},
		{TypeFloat, "FLOAT"},
		{TypeDouble, "DOUBLE"},
		{TypeDate, "DATE"},
		{TypeTime, "TIME"},
		{TypeDatetime, "DATETIME"},
		{TypeTimestamp, "TIMESTAMP"},
		{TypeBlob, "BLOB"},
		{TypeClob, "CLOB"},
		{TypeNull, "NULL"},
		{999, "UNKNOWN"},
	}

	r := &rows{}
	for _, tc := range cases {
		r.columns = []columnMetaData{{ColumnType: tc.colType}}
		got := r.ColumnTypeDatabaseTypeName(0)
		if got != tc.want {
			t.Errorf("type %d: want %q, got %q", tc.colType, tc.want, got)
		}
	}
}
