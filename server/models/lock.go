package models

import (
	"dt/utils/log"
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
	// Add logging to debug
	log.Debug("Attempting to acquire lock for table: %s, id: %d, tx: %d", table, id, tx.ID)

	// First check if lock exists
	var existingLock Lock
	err := mvccConn.QueryRowContext(tx.ctx, `
        SELECT record_table, record_id, txid 
        FROM locks 
        WHERE record_table = $1 AND record_id = $2`,
		table, id).Scan(&existingLock.Table, &existingLock.ID, &existingLock.TxID)

	if err == nil {
		// Lock exists - add dependency and wait
		log.Debug("Lock exists, adding dependency from tx %d to tx %d", tx.ID, existingLock.TxID)
		if err := tx.addDependency(existingLock.TxID); err != nil {
			return err
		}
		time.Sleep(operationDelay) // Use the configured delay
		return tx.acquireLock(table, id, lockType)
	}

	// Create new lock
	log.Debug("Creating new lock for tx %d", tx.ID)
	_, err = mvccConn.ExecContext(tx.ctx, `
        INSERT INTO locks (record_table, record_id, txid, shared)
        VALUES ($1, $2, $3, $4)`,
		table, id, tx.ID, lockType == ReadLock)

	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}

	return nil
}
