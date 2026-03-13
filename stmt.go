package cubrid

import (
	"database/sql/driver"
)

// stmt is a server-side prepared statement obtained via conn.Prepare().
// It holds the query handle returned by FC=2 (PREPARE) and reuses it
// across multiple Exec / Query calls via FC=3 (EXECUTE).
type stmt struct {
	conn        *conn
	query       string // original SQL (kept for Explain / debugging)
	queryHandle int    // server-side handle from FC=2
	stmtType    int    // statement type (SELECT, INSERT, …)
	bindCount   int    // number of ? placeholders
	columns     []columnMetaData
	closed      bool
}

// Close releases the server-side prepared-statement handle (FC=6).
func (s *stmt) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	if s.queryHandle > 0 {
		s.conn.mu.Lock()
		defer s.conn.mu.Unlock()
		s.conn.closeQueryHandle(s.queryHandle)
		s.queryHandle = 0
	}
	return nil
}

// NumInput returns the number of ? placeholders the server found during Prepare.
// database/sql uses this to validate argument counts before calling Exec/Query.
func (s *stmt) NumInput() int { return s.bindCount }

// Exec executes a non-SELECT (or any) statement with the given arguments.
// Parameters are sent as typed wire values (FC=3), never interpolated into SQL.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	if s.closed || s.conn.closed {
		return nil, driver.ErrBadConn
	}

	req := writeExecute(s.queryHandle, s.stmtType, args, s.conn.autoCommit, s.conn.casInfo)
	resp, err := s.conn.sendAndRecv(req)
	if err != nil {
		return nil, err
	}
	res, err := parseExecute(resp, s.columns, s.stmtType, s.conn.protoVer)
	if err != nil {
		return nil, err
	}

	// Retrieve last insert ID for INSERT statements.
	var lastID int64
	if s.stmtType == StmtInsert {
		lastID = s.conn.fetchLastInsertID()
	}

	var affected int64
	for _, info := range res.ResultInfos {
		affected += int64(info.ResultCount)
	}

	return &result{lastInsertID: lastID, rowsAffected: affected}, nil
}

// Query executes the statement and returns the result rows.
// Parameters are sent as typed wire values (FC=3), never interpolated into SQL.
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	if s.closed || s.conn.closed {
		return nil, driver.ErrBadConn
	}

	req := writeExecute(s.queryHandle, s.stmtType, args, s.conn.autoCommit, s.conn.casInfo)
	resp, err := s.conn.sendAndRecv(req)
	if err != nil {
		return nil, err
	}
	res, err := parseExecute(resp, s.columns, s.stmtType, s.conn.protoVer)
	if err != nil {
		return nil, err
	}

	r := &rows{
		conn:         s.conn,
		queryHandle:  s.queryHandle,
		columns:      s.columns,
		stmtType:     s.stmtType,
		totalCount:   res.TotalTupleCount,
		fetchedCount: len(res.Rows),
		bufferedRows: res.Rows,
		bufferOffset: 0,
		exhausted:    res.TupleCount == 0 || len(res.Rows) >= res.TotalTupleCount,
		fetchSize:    defaultFetchSize,
		// Rows does NOT own the handle; stmt.Close() sends FC=6.
		closeHandle: false,
	}

	return r, nil
}

// result implements driver.Result.
type result struct {
	lastInsertID int64
	rowsAffected int64
}

func (r *result) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r *result) RowsAffected() (int64, error) { return r.rowsAffected, nil }
