package cubrid

import (
	"database/sql/driver"
)

// stmt implements driver.Stmt.
type stmt struct {
	conn  *conn
	query string
}

// Close is a no-op because CUBRID statements are not prepared server-side
// until execution (we use PrepareAndExecute which combines both steps).
func (s *stmt) Close() error { return nil }

// NumInput returns -1 to indicate that the driver does not pre-validate
// the number of input parameters.
func (s *stmt) NumInput() int { return -1 }

// Exec executes a non-SELECT statement.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	sql, err := interpolateArgs(s.query, args)
	if err != nil {
		return nil, err
	}

	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	res, err := s.conn.execSQL(sql)
	if err != nil {
		return nil, err
	}

	// For INSERT statements, retrieve the last inserted ID.
	var lastInsertID int64
	if res.StatementType == StmtInsert {
		liReq := WriteGetLastInsertId(s.conn.casInfo)
		liResp, liErr := s.conn.sendAndRecv(liReq)
		if liErr == nil {
			if idStr, idErr := ParseGetLastInsertId(liResp); idErr == nil && idStr != "" {
				// Parse string to int64 (best-effort).
				var n int64
				for _, ch := range idStr {
					if ch >= '0' && ch <= '9' {
						n = n*10 + int64(ch-'0')
					}
				}
				lastInsertID = n
			}
		}
	}

	rowsAffected := int64(0)
	for _, info := range res.ResultInfos {
		rowsAffected += int64(info.ResultCount)
	}

	if s.conn.autoCommit {
		s.conn.autoCommit = true // restore default state
	}

	return &result{lastInsertID: lastInsertID, rowsAffected: rowsAffected}, nil
}

// Query executes a SELECT statement and returns the rows.
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	sql, err := interpolateArgs(s.query, args)
	if err != nil {
		return nil, err
	}

	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	res, err := s.conn.execSQL(sql)
	if err != nil {
		return nil, err
	}

	r := &rows{
		conn:          s.conn,
		queryHandle:   res.QueryHandle,
		columns:       res.Columns,
		stmtType:      res.StatementType,
		totalCount:    res.TotalTupleCount,
		fetchedCount:  len(res.Rows),
		bufferedRows:  res.Rows,
		bufferOffset:  0,
		exhausted:     len(res.Rows) == 0 || len(res.Rows) >= res.TotalTupleCount,
		fetchSize:     defaultFetchSize,
	}

	if res.StatementType != StmtSelect {
		// Non-SELECT: return empty rows immediately.
		r.exhausted = true
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
