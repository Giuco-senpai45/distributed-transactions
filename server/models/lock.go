package models

import (
	"fmt"
	"time"
)

type LockType int

const (
	ReadLock LockType = iota
	WriteLock
)

type Lock struct {
	Table string
	ID    int
	Type  LockType
	TxID  int
}

func (tx *Transaction) acquireLock(table string, id int, lockType LockType) error {
	// First check if lock exists
	var existingLock Lock
	err := mvccConn.QueryRowContext(tx.ctx, `
        SELECT record_table, record_id, txid 
        FROM locks 
        WHERE record_table = $1 AND record_id = $2`,
		table, id).Scan(&existingLock.Table, &existingLock.ID, &existingLock.TxID)

	if err == nil {
		// Lock exists - add dependency and wait
		if err := tx.addDependency(existingLock.TxID); err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100)
		return tx.acquireLock(table, id, lockType)
	}

	// Create new lock
	_, err = mvccConn.ExecContext(tx.ctx, `
        INSERT INTO locks (record_table, record_id, txid, shared)
        VALUES ($1, $2, $3, $4)`,
		table, id, tx.ID, lockType == ReadLock)

	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}

	// Create lock path for tracking
	path := NewLockPath(tx.ID, table, id)
	_, err = mvccConn.ExecContext(tx.ctx, `
        INSERT INTO paths (path, type, name, dependency_type)
        VALUES (text2ltree($1), $2, $3, 'lock')`,
		path.Path, path.Type, path.Name)

	return nil
}
