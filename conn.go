package cubrid

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const defaultFetchSize = 100

// conn is a single TCP connection to the CUBRID CAS broker.
// It implements driver.Conn, driver.ConnBeginTx, and driver.Pinger.
type conn struct {
	mu         sync.Mutex
	socket     net.Conn
	host       string
	port       int
	database   string
	user       string
	password   string
	timeout    time.Duration
	casInfo    [SizeCASInfo]byte
	protoVer   int
	autoCommit bool
	closed     bool
}

// connect performs the two-step CUBRID broker handshake.
func (c *conn) connect() error {
	brokerAddr := fmt.Sprintf("%s:%d", c.host, c.port)
	brokerConn, err := net.DialTimeout("tcp", brokerAddr, c.timeout)
	if err != nil {
		return &OperationalError{CubridError{Code: -1, Message: fmt.Sprintf("dial broker %s: %v", brokerAddr, err)}}
	}
	defer brokerConn.Close()

	if c.timeout > 0 {
		brokerConn.SetDeadline(time.Now().Add(c.timeout))
	}

	// Step 1: send ClientInfoExchange (10 bytes, no framing header).
	if _, err = brokerConn.Write(WriteClientInfoExchange()); err != nil {
		return err
	}

	// Step 2: receive redirected port (4 bytes).
	portBuf := make([]byte, 4)
	if _, err = io.ReadFull(brokerConn, portBuf); err != nil {
		return err
	}
	newPort := int(int32(binary.BigEndian.Uint32(portBuf)))
	if newPort < 0 {
		return &OperationalError{CubridError{Code: newPort, Message: "broker rejected connection"}}
	}

	// Step 3: connect to the CAS port returned by the broker.
	casAddr := fmt.Sprintf("%s:%d", c.host, newPort)
	c.socket, err = net.DialTimeout("tcp", casAddr, c.timeout)
	if err != nil {
		return &OperationalError{CubridError{Code: -1, Message: fmt.Sprintf("dial CAS %s: %v", casAddr, err)}}
	}

	if c.timeout > 0 {
		c.socket.SetDeadline(time.Now().Add(c.timeout))
	}

	// Step 4: send OpenDatabase request (628 bytes, no framing header).
	if _, err = c.socket.Write(WriteOpenDatabase(c.database, c.user, c.password)); err != nil {
		return err
	}

	// Step 5: receive the framed OpenDatabase response.
	data, err := c.recv()
	if err != nil {
		return err
	}
	res, err := ParseOpenDatabase(data)
	if err != nil {
		return err
	}
	c.casInfo = res.CASInfo
	c.protoVer = res.ProtocolVersion

	// Clear deadline after successful connect.
	c.socket.SetDeadline(time.Time{})
	return nil
}

// send writes all bytes to the socket.
func (c *conn) send(data []byte) error {
	_, err := c.socket.Write(data)
	return err
}

// recv reads a length-framed response from the socket.
// It reads a 4-byte DATA_LENGTH header, then reads exactly that many bytes.
func (c *conn) recv() ([]byte, error) {
	lenBuf := make([]byte, SizeDataLength)
	if _, err := io.ReadFull(c.socket, lenBuf); err != nil {
		return nil, err
	}
	dataLen := int(binary.BigEndian.Uint32(lenBuf))

	data := make([]byte, dataLen)
	if _, err := io.ReadFull(c.socket, data); err != nil {
		return nil, err
	}
	return data, nil
}

// sendAndRecv sends data and returns the framed response.
func (c *conn) sendAndRecv(data []byte) ([]byte, error) {
	if err := c.send(data); err != nil {
		return nil, err
	}
	return c.recv()
}

// execSQL runs sql (with pre-interpolated arguments) against the server.
func (c *conn) execSQL(sql string) (*PrepareAndExecuteResult, error) {
	req := WritePrepareAndExecute(sql, c.autoCommit, c.casInfo)
	resp, err := c.sendAndRecv(req)
	if err != nil {
		return nil, err
	}
	res, err := ParsePrepareAndExecute(resp, c.protoVer)
	if err != nil {
		return nil, err
	}
	// Keep CAS info updated from response header.
	if len(resp) >= SizeCASInfo {
		copy(c.casInfo[:], resp[:SizeCASInfo])
	}
	return res, nil
}

// closeQueryHandle releases the server-side query handle.
func (c *conn) closeQueryHandle(qh int) {
	req := WriteCloseReqHandle(qh, c.casInfo)
	resp, err := c.sendAndRecv(req)
	if err != nil {
		return
	}
	_ = ParseSimpleResponse(resp)
}

// ─── driver.Conn ─────────────────────────────────────────────────────────────

// Prepare returns a prepared statement.
func (c *conn) Prepare(query string) (driver.Stmt, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, driver.ErrBadConn
	}
	return &stmt{conn: c, query: query}, nil
}

// Close closes the connection.
func (c *conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	if c.socket != nil {
		req := WriteConClose(c.casInfo)
		c.socket.Write(req) // best-effort
		c.socket.Close()
	}
	return nil
}

// Begin starts a transaction (auto-commit off).
func (c *conn) Begin() (driver.Tx, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, driver.ErrBadConn
	}
	c.autoCommit = false
	return &tx{conn: c}, nil
}

// Ping verifies the connection is alive.
func (c *conn) Ping() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return driver.ErrBadConn
	}
	req := WriteGetDbVersion(true, c.casInfo)
	resp, err := c.sendAndRecv(req)
	if err != nil {
		return driver.ErrBadConn
	}
	_, err = ParseGetDbVersion(resp)
	return err
}

// ServerVersion returns the CUBRID engine version string.
func (c *conn) ServerVersion() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	req := WriteGetDbVersion(true, c.casInfo)
	resp, err := c.sendAndRecv(req)
	if err != nil {
		return "", err
	}
	return ParseGetDbVersion(resp)
}
