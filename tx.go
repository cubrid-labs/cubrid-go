package cubrid

// tx implements database/sql/driver.Tx.
type tx struct {
	conn *conn
}

// Commit commits the current transaction.
func (t *tx) Commit() error {
	t.conn.mu.Lock()
	defer t.conn.mu.Unlock()

	req := writeEndTran(TxCommit, t.conn.casInfo)
	resp, err := t.conn.sendAndRecv(req)
	if err != nil {
		return err
	}
	if err = parseSimpleResponse(resp); err != nil {
		return err
	}
	t.conn.autoCommit = true
	return nil
}

// Rollback aborts the current transaction.
func (t *tx) Rollback() error {
	t.conn.mu.Lock()
	defer t.conn.mu.Unlock()

	req := writeEndTran(TxRollback, t.conn.casInfo)
	resp, err := t.conn.sendAndRecv(req)
	if err != nil {
		return err
	}
	if err = parseSimpleResponse(resp); err != nil {
		return err
	}
	t.conn.autoCommit = true
	return nil
}
