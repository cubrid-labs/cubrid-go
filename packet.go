package cubrid

import (
	"encoding/binary"
	"math"
	"time"
)

// buildProtocolHeader creates an 8-byte protocol header:
// 4 bytes DATA_LENGTH (big-endian) + 4 bytes CAS_INFO.
func buildProtocolHeader(dataLength int, casInfo [SizeCASInfo]byte) []byte {
	buf := make([]byte, SizeDataLength+SizeCASInfo)
	binary.BigEndian.PutUint32(buf[:SizeDataLength], uint32(dataLength))
	copy(buf[SizeDataLength:], casInfo[:])
	return buf
}

// parseProtocolHeader parses an 8-byte protocol header.
func parseProtocolHeader(data []byte) (int, [SizeCASInfo]byte) {
	length := int(binary.BigEndian.Uint32(data[:SizeDataLength]))
	var cas [SizeCASInfo]byte
	copy(cas[:], data[SizeDataLength:SizeDataLength+SizeCASInfo])
	return length, cas
}

// packetWriter serializes values into a big-endian byte buffer.
type packetWriter struct {
	buf []byte
}

func newPacketWriter() *packetWriter {
	return &packetWriter{buf: make([]byte, 0, 64)}
}

func (w *packetWriter) writeByte(v byte) {
	w.buf = append(w.buf, v)
}

func (w *packetWriter) writeShort(v int16) {
	b := make([]byte, SizeShort)
	binary.BigEndian.PutUint16(b, uint16(v))
	w.buf = append(w.buf, b...)
}

func (w *packetWriter) writeInt(v int32) {
	b := make([]byte, SizeInt)
	binary.BigEndian.PutUint32(b, uint32(v))
	w.buf = append(w.buf, b...)
}

func (w *packetWriter) writeLong(v int64) {
	b := make([]byte, SizeLong)
	binary.BigEndian.PutUint64(b, uint64(v))
	w.buf = append(w.buf, b...)
}

func (w *packetWriter) writeFloat(v float32) {
	b := make([]byte, SizeFloat)
	binary.BigEndian.PutUint32(b, math.Float32bits(v))
	w.buf = append(w.buf, b...)
}

func (w *packetWriter) writeDouble(v float64) {
	b := make([]byte, SizeDouble)
	binary.BigEndian.PutUint64(b, math.Float64bits(v))
	w.buf = append(w.buf, b...)
}

func (w *packetWriter) writeRawBytes(v []byte) {
	w.buf = append(w.buf, v...)
}

func (w *packetWriter) writeFiller(count int) {
	for i := 0; i < count; i++ {
		w.buf = append(w.buf, 0x00)
	}
}

// writeFixedString writes a UTF-8 string padded/truncated to exactly length bytes.
func (w *packetWriter) writeFixedString(s string, length int) {
	encoded := []byte(s)
	if len(encoded) > length {
		encoded = encoded[:length]
	}
	w.buf = append(w.buf, encoded...)
	if len(encoded) < length {
		w.writeFiller(length - len(encoded))
	}
}

// writeNullTermString writes a length-prefixed null-terminated UTF-8 string.
// Format: int32(len+1) + bytes + 0x00
func (w *packetWriter) writeNullTermString(s string) {
	encoded := []byte(s)
	w.writeInt(int32(len(encoded) + 1))
	w.buf = append(w.buf, encoded...)
	w.buf = append(w.buf, 0x00)
}

// addByte writes int32(1) + byte.
func (w *packetWriter) addByte(v byte) {
	w.writeInt(SizeByte)
	w.writeByte(v)
}

// addShort writes int32(2) + int16.
func (w *packetWriter) addShort(v int16) {
	w.writeInt(SizeShort)
	w.writeShort(v)
}

// addInt writes int32(4) + int32.
func (w *packetWriter) addInt(v int32) {
	w.writeInt(SizeInt)
	w.writeInt(v)
}

// addLong writes int32(8) + int64.
func (w *packetWriter) addLong(v int64) {
	w.writeInt(SizeLong)
	w.writeLong(v)
}

// addFloat writes int32(4) + float32.
func (w *packetWriter) addFloat(v float32) {
	w.writeInt(SizeFloat)
	w.writeFloat(v)
}

// addDouble writes int32(8) + float64.
func (w *packetWriter) addDouble(v float64) {
	w.writeInt(SizeDouble)
	w.writeDouble(v)
}

// addBytes writes int32(len) + raw bytes.
func (w *packetWriter) addBytes(v []byte) {
	w.writeInt(int32(len(v)))
	w.buf = append(w.buf, v...)
}

// addNull writes a zero-length marker.
func (w *packetWriter) addNull() {
	w.writeInt(0)
}

// addDatetime writes a length-prefixed datetime value (7 x int16 = 14 bytes).
func (w *packetWriter) addDatetime(t time.Time) {
	w.writeInt(SizeDatetime)
	w.writeShort(int16(t.Year()))
	w.writeShort(int16(t.Month()))
	w.writeShort(int16(t.Day()))
	w.writeShort(int16(t.Hour()))
	w.writeShort(int16(t.Minute()))
	w.writeShort(int16(t.Second()))
	w.writeShort(int16(t.Nanosecond() / 1e6))
}

// addDate writes a datetime with zeroed time fields.
func (w *packetWriter) addDate(t time.Time) {
	w.writeInt(SizeDatetime)
	w.writeShort(int16(t.Year()))
	w.writeShort(int16(t.Month()))
	w.writeShort(int16(t.Day()))
	w.writeShort(0) // hour
	w.writeShort(0) // minute
	w.writeShort(0) // second
	w.writeShort(0) // millisecond
}

// addCacheTime writes a length-prefixed cache time (two zero int32s).
func (w *packetWriter) addCacheTime() {
	w.writeInt(SizeLong)
	w.writeInt(0) // sec
	w.writeInt(0) // usec
}

func (w *packetWriter) toBytes() []byte { return w.buf }

// packetReader deserializes big-endian binary data.
type packetReader struct {
	buf    []byte
	offset int
}

func newPacketReader(data []byte) *packetReader {
	return &packetReader{buf: data}
}

func (r *packetReader) parseByte() byte {
	v := r.buf[r.offset]
	r.offset += SizeByte
	return v
}

func (r *packetReader) parseShort() int16 {
	v := int16(binary.BigEndian.Uint16(r.buf[r.offset:]))
	r.offset += SizeShort
	return v
}

func (r *packetReader) parseInt() int32 {
	v := int32(binary.BigEndian.Uint32(r.buf[r.offset:]))
	r.offset += SizeInt
	return v
}

func (r *packetReader) parseLong() int64 {
	v := int64(binary.BigEndian.Uint64(r.buf[r.offset:]))
	r.offset += SizeLong
	return v
}

func (r *packetReader) parseFloat() float32 {
	v := math.Float32frombits(binary.BigEndian.Uint32(r.buf[r.offset:]))
	r.offset += SizeFloat
	return v
}

func (r *packetReader) parseDouble() float64 {
	v := math.Float64frombits(binary.BigEndian.Uint64(r.buf[r.offset:]))
	r.offset += SizeDouble
	return v
}

func (r *packetReader) parseRawBytes(count int) []byte {
	v := make([]byte, count)
	copy(v, r.buf[r.offset:r.offset+count])
	r.offset += count
	return v
}

// parseNullTermString reads length bytes and strips the trailing null byte.
func (r *packetReader) parseNullTermString(length int) string {
	if length <= 0 {
		return ""
	}
	data := r.parseRawBytes(length)
	if len(data) > 0 && data[len(data)-1] == 0x00 {
		data = data[:len(data)-1]
	}
	return string(data)
}

func (r *packetReader) parseDate() time.Time {
	year := int(r.parseShort())
	month := int(r.parseShort())
	day := int(r.parseShort())
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func (r *packetReader) parseTime() time.Time {
	hour := int(r.parseShort())
	min := int(r.parseShort())
	sec := int(r.parseShort())
	return time.Date(0, 1, 1, hour, min, sec, 0, time.UTC)
}

func (r *packetReader) parseTimestamp() time.Time {
	year := int(r.parseShort())
	month := int(r.parseShort())
	day := int(r.parseShort())
	hour := int(r.parseShort())
	min := int(r.parseShort())
	sec := int(r.parseShort())
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
}

func (r *packetReader) parseDatetime() time.Time {
	year := int(r.parseShort())
	month := int(r.parseShort())
	day := int(r.parseShort())
	hour := int(r.parseShort())
	min := int(r.parseShort())
	sec := int(r.parseShort())
	ms := int(r.parseShort())
	return time.Date(year, time.Month(month), day, hour, min, sec, ms*1e6, time.UTC)
}

func (r *packetReader) bytesRemaining() int {
	return len(r.buf) - r.offset
}

// readError reads an error response body: int32 code + null-term message.
func (r *packetReader) readError(responseLength int) (int32, string) {
	code := r.parseInt()
	msgSize := responseLength - SizeInt
	msg := r.parseNullTermString(msgSize)
	return code, msg
}
