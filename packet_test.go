package cubrid

import (
	"testing"
	"time"
)

// ─── packetWriter ─────────────────────────────────────────────────────────────

func TestPacketWriter_writeByte(t *testing.T) {
	w := newPacketWriter()
	w.writeByte(0xAB)
	got := w.toBytes()
	if len(got) != 1 || got[0] != 0xAB {
		t.Fatalf("want [0xAB], got %v", got)
	}
}

func TestPacketWriter_writeShort(t *testing.T) {
	w := newPacketWriter()
	w.writeShort(0x0102)
	got := w.toBytes()
	want := []byte{0x01, 0x02}
	if string(got) != string(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestPacketWriter_writeInt(t *testing.T) {
	w := newPacketWriter()
	w.writeInt(0x01020304)
	got := w.toBytes()
	want := []byte{0x01, 0x02, 0x03, 0x04}
	if string(got) != string(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestPacketWriter_writeLong(t *testing.T) {
	w := newPacketWriter()
	w.writeLong(0x0102030405060708)
	got := w.toBytes()
	want := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if string(got) != string(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestPacketWriter_writeFixedString_short(t *testing.T) {
	w := newPacketWriter()
	w.writeFixedString("hi", 5)
	got := w.toBytes()
	if len(got) != 5 {
		t.Fatalf("want len 5, got %d", len(got))
	}
	if got[0] != 'h' || got[1] != 'i' || got[2] != 0 || got[3] != 0 || got[4] != 0 {
		t.Fatalf("unexpected bytes: %v", got)
	}
}

func TestPacketWriter_writeFixedString_truncate(t *testing.T) {
	w := newPacketWriter()
	w.writeFixedString("hello", 3)
	got := w.toBytes()
	if string(got) != "hel" {
		t.Fatalf("want 'hel', got %q", string(got))
	}
}

func TestPacketWriter_writeNullTermString(t *testing.T) {
	w := newPacketWriter()
	w.writeNullTermString("abc")
	got := w.toBytes()
	// int32(4) = [0,0,0,4] + "abc" + 0x00
	if len(got) != 4+4 {
		t.Fatalf("want len 8, got %d", len(got))
	}
	if got[3] != 4 { // low byte of length = 4 (3 chars + null)
		t.Fatalf("unexpected length prefix: %v", got[:4])
	}
	if string(got[4:7]) != "abc" || got[7] != 0x00 {
		t.Fatalf("unexpected string bytes: %v", got[4:])
	}
}

func TestPacketWriter_addInt(t *testing.T) {
	w := newPacketWriter()
	w.addInt(42)
	got := w.toBytes()
	// length prefix (4 bytes) + value (4 bytes)
	if len(got) != 8 {
		t.Fatalf("want len 8, got %d", len(got))
	}
	// length prefix = int32(4)
	if got[3] != 4 {
		t.Fatalf("unexpected length prefix: %v", got[:4])
	}
	// value = int32(42)
	if got[7] != 42 {
		t.Fatalf("unexpected value: %v", got[4:])
	}
}

func TestPacketWriter_addNull(t *testing.T) {
	w := newPacketWriter()
	w.addNull()
	got := w.toBytes()
	want := []byte{0, 0, 0, 0}
	if string(got) != string(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestPacketWriter_addDatetime(t *testing.T) {
	w := newPacketWriter()
	tm := time.Date(2024, 3, 15, 10, 30, 45, 500e6, time.UTC)
	w.addDatetime(tm)
	got := w.toBytes()
	// int32(14) + 7×int16
	if len(got) != 4+14 {
		t.Fatalf("want len 18, got %d", len(got))
	}
}

// ─── packetReader ─────────────────────────────────────────────────────────────

func TestPacketReader_roundtrip_int(t *testing.T) {
	w := newPacketWriter()
	w.writeInt(12345678)
	r := newPacketReader(w.toBytes())
	got := r.parseInt()
	if got != 12345678 {
		t.Fatalf("want 12345678, got %d", got)
	}
}

func TestPacketReader_roundtrip_long(t *testing.T) {
	w := newPacketWriter()
	w.writeLong(-9876543210)
	r := newPacketReader(w.toBytes())
	got := r.parseLong()
	if got != -9876543210 {
		t.Fatalf("want -9876543210, got %d", got)
	}
}

func TestPacketReader_roundtrip_float(t *testing.T) {
	w := newPacketWriter()
	w.writeFloat(3.14)
	r := newPacketReader(w.toBytes())
	got := r.parseFloat()
	if got < 3.13 || got > 3.15 {
		t.Fatalf("want ~3.14, got %f", got)
	}
}

func TestPacketReader_roundtrip_double(t *testing.T) {
	w := newPacketWriter()
	w.writeDouble(2.718281828)
	r := newPacketReader(w.toBytes())
	got := r.parseDouble()
	if got < 2.71 || got > 2.72 {
		t.Fatalf("want ~2.718, got %f", got)
	}
}

func TestPacketReader_parseNullTermString(t *testing.T) {
	// Build bytes: "hello\0"
	data := []byte{'h', 'e', 'l', 'l', 'o', 0x00}
	r := newPacketReader(data)
	got := r.parseNullTermString(6)
	if got != "hello" {
		t.Fatalf("want 'hello', got %q", got)
	}
}

func TestPacketReader_parseNullTermString_empty(t *testing.T) {
	r := newPacketReader([]byte{})
	got := r.parseNullTermString(0)
	if got != "" {
		t.Fatalf("want '', got %q", got)
	}
}

func TestPacketReader_parseDatetime_roundtrip(t *testing.T) {
	tm := time.Date(2024, 6, 15, 12, 30, 45, 500*int(time.Millisecond), time.UTC)
	w := newPacketWriter()
	w.writeShort(int16(tm.Year()))
	w.writeShort(int16(tm.Month()))
	w.writeShort(int16(tm.Day()))
	w.writeShort(int16(tm.Hour()))
	w.writeShort(int16(tm.Minute()))
	w.writeShort(int16(tm.Second()))
	w.writeShort(int16(tm.Nanosecond() / 1e6))

	r := newPacketReader(w.toBytes())
	got := r.parseDatetime()
	if got.Year() != 2024 || got.Month() != 6 || got.Day() != 15 {
		t.Fatalf("unexpected date: %v", got)
	}
	if got.Hour() != 12 || got.Minute() != 30 || got.Second() != 45 {
		t.Fatalf("unexpected time: %v", got)
	}
	if got.Nanosecond()/1e6 != 500 {
		t.Fatalf("unexpected ms: %d", got.Nanosecond()/1e6)
	}
}

func TestPacketReader_bytesRemaining(t *testing.T) {
	r := newPacketReader([]byte{1, 2, 3, 4, 5})
	r.parseInt() // consume 4 bytes
	if r.bytesRemaining() != 1 {
		t.Fatalf("want 1 remaining, got %d", r.bytesRemaining())
	}
}

// ─── Protocol header ──────────────────────────────────────────────────────────

func TestProtocolHeader_roundtrip(t *testing.T) {
	var cas [SizeCASInfo]byte
	cas[0] = 0x01
	cas[1] = 0x02
	cas[2] = 0x03
	cas[3] = 0x04

	hdr := buildProtocolHeader(1024, cas)
	if len(hdr) != SizeDataLength+SizeCASInfo {
		t.Fatalf("unexpected header length: %d", len(hdr))
	}

	length, gotCas := parseProtocolHeader(hdr)
	if length != 1024 {
		t.Fatalf("want length 1024, got %d", length)
	}
	if gotCas != cas {
		t.Fatalf("want cas %v, got %v", cas, gotCas)
	}
}
