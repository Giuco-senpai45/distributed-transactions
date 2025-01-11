package models

import (
	"context"
	"database/sql"
	errs "dt/utils/errors"
	"dt/utils/log"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TransactionData struct {
	ID        int
	CreatedAt time.Time
	Status    string
}

type Transaction struct {
	TransactionData
	records []Record
	ctx     context.Context
}

var mu sync.Mutex

// fetches transaction data
func GetTx(ctx context.Context, txID int) (*TransactionData, error) {
	row := mvccConn.QueryRowContext(ctx, "SELECT id, created_at, status FROM transactions WHERE id = $1", txID)

	t := &TransactionData{}
	err := row.Scan(&t.ID, &t.CreatedAt, &t.Status)
	errs.ErrorCheck(err)

	return t, nil
}

// create new transaction (insert into table)
func OpenTx(ctx context.Context) (*Transaction, error) {
	mu.Lock()
	defer mu.Unlock()

	var id int
	stmt := "INSERT INTO transactions (status) VALUES ($1) RETURNING id"
	err := mvccConn.QueryRowContext(ctx, stmt, TxActive).Scan(&id)
	errs.ErrorCheck(err)

	txData, err := GetTx(ctx, id)
	errs.ErrorCheck(err)

	tx := &Transaction{
		TransactionData: *txData,
		ctx:             ctx,
	}
	return tx, nil
}

func (tx *Transaction) IsRowVisible(row *RecordData) bool {
	if row.TxMin > tx.ID || row.TxMinRolledBack {
		// row inserted after current transaction & insert aborted
		return false
	}
	if row.TxMin < tx.ID && !row.TxMinCommitted {
		// row created by a concurrent transaction
		return false
	}
	if row.TxMax < tx.ID && row.TxMaxCommitted {
		// row deleted by a concurrent transaction
		return false
	}
	if row.TxMax == tx.ID {
		// row deleted by current transaction
		return false
	}
	return true
}

// select specified record from table, check visibility and return record data
func (tx *Transaction) selectRecord(table string, id int, data ...any) (*RecordData, error) {
	log.Debug("Selecting record from %s with id %d", table, id)

	rows, err := appConn.QueryContext(tx.ctx, "SELECT * FROM "+table+" WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, fmt.Errorf("record not found")
	}

	values := make([]interface{}, len(cols))
	for i := range values {
		values[i] = new(interface{})
	}

	err = rows.Scan(values...)
	if err != nil {
		return nil, err
	}

	base := &RecordData{}
	for i, col := range cols {
		val := *(values[i].(*interface{}))
		switch col {
		case "tx_min":
			base.TxMin = int(val.(int64))
		case "tx_max":
			base.TxMax = int(val.(int64))
		case "tx_min_committed":
			base.TxMinCommitted = val.(bool)
		case "tx_max_committed":
			base.TxMaxCommitted = val.(bool)
		case "tx_min_rolled_back":
			base.TxMinRolledBack = val.(bool)
		case "tx_max_rolled_back":
			base.TxMaxRolledBack = val.(bool)
		}
	}

	log.Debug("Selected record metadata: %+v", base)
	return base, nil
}

func (tx *Transaction) Where(table string, where string, args ...any) ([]map[string]interface{}, error) {
	baseQuery := `
        WITH latest_versions AS (
            SELECT DISTINCT ON (id) *
            FROM ` + table

	// Handle WHERE clause
	if where != "" {
		baseQuery += ` WHERE ` + where + ` = $2 AND `
	} else {
		baseQuery += ` WHERE `
	}

	baseQuery += `
            tx_min_committed = true
            AND NOT tx_min_rolled_back
            AND (tx_max = 0 OR (tx_max > $1 AND NOT tx_max_committed))
            ORDER BY id, tx_min DESC
        )
        SELECT * FROM latest_versions`

	// Prepare args
	var queryArgs []interface{}
	if len(args) > 0 {
		queryArgs = make([]interface{}, len(args)+1)
		queryArgs[0] = tx.ID
		copy(queryArgs[1:], args)
	} else {
		queryArgs = []interface{}{tx.ID}
	}

	log.Debug("Where query: %s", baseQuery)
	log.Debug("Where args: %v", queryArgs)

	// Execute query
	rows, err := appConn.QueryContext(tx.ctx, baseQuery, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		row := make([]interface{}, len(cols))
		for i := range row {
			row[i] = new(interface{})
		}

		err := rows.Scan(row...)
		if err != nil {
			return nil, err
		}

		base := RecordData{}
		log.Debug("Processing row with columns: %v", cols)
		for i, col := range cols {
			sv := reflect.Indirect(reflect.ValueOf(row[i])).Elem()
			switch col {
			case "tx_min":
				base.TxMin = int(sv.Int())
				log.Debug("tx_min: %v", base.TxMin)
			case "tx_max":
				base.TxMax = int(sv.Int())
				log.Debug("tx_max: %v", base.TxMax)
			case "tx_min_committed":
				base.TxMinCommitted = sv.Bool()
				log.Debug("tx_min_committed: %v", base.TxMinCommitted)
			case "tx_max_committed":
				base.TxMaxCommitted = sv.Bool()
				log.Debug("tx_max_committed: %v", base.TxMaxCommitted)
			case "tx_min_rolled_back":
				base.TxMinRolledBack = sv.Bool()
				log.Debug("tx_min_rolled_back: %v", base.TxMinRolledBack)
			case "tx_max_rolled_back":
				base.TxMaxRolledBack = sv.Bool()
				log.Debug("tx_max_rolled_back: %v", base.TxMaxRolledBack)
			}
		}

		visible := tx.IsRowVisible(&base)
		log.Debug("Row visibility for tx %d: %v", tx.ID, visible)

		if visible {
			result := make(map[string]interface{})
			for i, col := range cols[6:] {
				sv := reflect.Indirect(reflect.ValueOf(row[i+6])).Elem()
				result[col] = sv.Interface()
			}
			results = append(results, result)
		}
	}

	log.Debug("Query returned %d visible results", len(results))
	return results, nil
}

func (tx *Transaction) Select(table string, id int, data ...any) *Transaction {
	_, err := tx.selectRecord(table, id, data...)
	errs.ErrorCheck(err)

	return tx
}

func makeQueryParams(min, max int) []string {
	a := make([]string, max-min)
	for i := range a {
		a[i] = fmt.Sprintf("$%d", i+min)
	}
	return a
}

// insert new record into table, return record id
func (tx *Transaction) Insert(table string, fields []string, values ...any) (int, error) {
	var id int
	seqName := table + "_id_seq"
	err := appConn.QueryRowContext(tx.ctx, fmt.Sprintf("SELECT nextval('%s')", seqName)).Scan(&id)
	if err != nil {
		return 0, err
	}

	// All required columns for MVCC
	allFields := []string{
		"id",
		"tx_min",
		"tx_max",
		"tx_min_committed",
		"tx_max_committed",
		"tx_min_rolled_back",
		"tx_max_rolled_back",
	}
	allFields = append(allFields, fields...)

	// Build INSERT statement
	stmt := "INSERT INTO " + table + " ("
	stmt += strings.Join(allFields, ", ")
	stmt += ") VALUES ("
	stmt += strings.Join(makeQueryParams(1, len(allFields)+1), ", ")
	stmt += ")"

	// Prepare all values including MVCC metadata
	allValues := []interface{}{
		id,    // id from sequence
		tx.ID, // tx_min
		0,     // tx_max
		false, // tx_min_committed
		false, // tx_max_committed
		false, // tx_min_rolled_back
		false, // tx_max_rolled_back
	}
	allValues = append(allValues, values...)

	log.Debug("Insert statement: %s", stmt)
	log.Debug("Insert values: %v", allValues)

	if _, err := appConn.ExecContext(tx.ctx, stmt, allValues...); err != nil {
		return 0, fmt.Errorf("error inserting: %v", err)
	}

	tx.records = append(tx.records, Record{
		Table:     table,
		ID:        id,
		Operation: OpInsert,
	})

	return id, nil
}

func (tx *Transaction) Update(table string, id int, fields []string, values ...any) error {
	// Get current version
	stmt := `SELECT tx_min 
            FROM ` + table + ` 
            WHERE id = $1 
            AND tx_max = 0 
            ORDER BY tx_min DESC 
            LIMIT 1`

	var currentTxMin int
	err := appConn.QueryRowContext(tx.ctx, stmt, id).Scan(&currentTxMin)
	if err != nil {
		return err
	}

	// Mark current version as ended
	updateStmt := `UPDATE ` + table + ` 
                   SET tx_max = $1, tx_max_committed = FALSE 
                   WHERE id = $2 AND tx_min = $3 AND tx_max = 0`
	if _, err := appConn.ExecContext(tx.ctx, updateStmt, tx.ID, id, currentTxMin); err != nil {
		return err
	}

	// Insert new version with same ID
	stmt = "INSERT INTO " + table + " ("
	f := append([]string{
		"id",
		"tx_min",
		"tx_max",
		"tx_min_committed",
		"tx_max_committed",
		"tx_min_rolled_back",
		"tx_max_rolled_back",
	}, fields...)
	stmt += strings.Join(f, ", ")
	stmt += ") VALUES ("
	params := makeQueryParams(1, len(f)+1)
	stmt += strings.Join(params, ", ")
	stmt += ")"

	// Prepare values with MVCC metadata
	insertValues := []interface{}{
		id,    // same id
		tx.ID, // new tx_min
		0,     // tx_max (0 means active)
		false, // tx_min_committed
		false, // tx_max_committed
		false, // tx_min_rolled_back
		false, // tx_max_rolled_back
	}
	insertValues = append(insertValues, values...)

	if _, err := appConn.ExecContext(tx.ctx, stmt, insertValues...); err != nil {
		return err
	}

	tx.records = append(tx.records, Record{
		Table:     table,
		ID:        id,
		Operation: OpUpdate,
	})

	return nil
}

func (tx *Transaction) Delete(table string, id int) error {
	base, err := tx.selectRecord(table, id)
	if err != nil {
		return fmt.Errorf("failed to select record: %v", err)
	}

	if !tx.IsRowVisible(base) {
		return fmt.Errorf("transaction %v aborted due to concurrency", tx.ID)
	}

	mu.Lock()
	defer mu.Unlock()

	// Check existing locks
	var lockID int
	row := mvccConn.QueryRowContext(tx.ctx,
		"SELECT id FROM locks WHERE record_table=$1 AND record_id=$2",
		table, id)

	if err := row.Scan(&lockID); err == nil {
		return fmt.Errorf("record already locked")
	}

	// Add resource dependency first
	if err := tx.addResourceDependency(table, id); err != nil {
		return fmt.Errorf("failed to add dependency: %v", err)
	}

	// Insert lock with correct parameter order
	stmt := "INSERT INTO locks(record_table, record_id, txid) VALUES($1, $2, $3) RETURNING id"
	if err := mvccConn.QueryRowContext(tx.ctx, stmt, table, id, tx.ID).Scan(&lockID); err != nil {
		log.Error("Failed to insert lock: %v", err)
		return err
	}

	// Add lock to zookeeper
	if err := tx.addResourceLock(table, id); err != nil {
		return fmt.Errorf("failed to add lock: %v", err)
	}

	// Mark record for deletion
	updateStmt := `UPDATE ` + table + ` SET tx_max = $1, tx_max_rolled_back = FALSE 
                   WHERE tx_min = $2 AND id = $3`
	if _, err := appConn.ExecContext(tx.ctx, updateStmt, tx.ID, base.TxMin, id); err != nil {
		log.Error("Failed to mark for deletion: %v", err)
		return err
	}

	// Record the operation
	tx.records = append(tx.records, Record{
		Table:     table,
		ID:        id,
		Operation: OpDelete,
		LockID:    lockID,
	})

	log.Debug("Successfully marked record %d in table %s for deletion", id, table)
	return nil
}

func (tx *Transaction) Commit() error {
	if err := tx.checkDependencyCycle(); err != nil {
		return err
	}

	t, err := mvccConn.BeginTx(tx.ctx, nil)
	if err != nil {
		return err
	}

	for _, r := range tx.records {
		switch r.Operation {
		case OpDelete:
			stmt := `UPDATE ` + r.Table + ` 
                     SET tx_max_committed = TRUE 
                     WHERE id = $1 AND tx_max = $2`
			_, err := appConn.ExecContext(tx.ctx, stmt, r.ID, tx.ID)
			if err != nil {
				t.Rollback()
				return err
			}
		case OpInsert, OpUpdate:
			stmt := `UPDATE ` + r.Table + ` 
                     SET tx_min_committed = TRUE 
                     WHERE id = $1 AND tx_min = $2`
			_, err := appConn.ExecContext(tx.ctx, stmt, r.ID, tx.ID)
			if err != nil {
				t.Rollback()
				return err
			}
		}
	}

	if err := tx.cleanupDependencies(); err != nil {
		t.Rollback()
		return err
	}

	tx.Status = TxCommitted
	return t.Commit()
}

func (tx *Transaction) Rollback() error {
	if tx.Status == TxCommitted {
		return nil
	}

	log.Debug("Rolling back transaction %d", tx.ID)

	t, err := mvccConn.BeginTx(tx.ctx, nil)
	if err != nil {
		return err
	}

	for _, r := range tx.records {
		switch r.Operation {
		case OpDelete:
			stmt := `UPDATE ` + r.Table + ` SET tx_max_rolled_back = TRUE WHERE tx_max = $1 AND id = $2;`
			_, err := appConn.ExecContext(tx.ctx, stmt, tx.ID, r.ID)
			if err != nil {
				return err
			}

			stmt = `DELETE FROM locks WHERE id = $1;`
			_, err = mvccConn.ExecContext(tx.ctx, stmt, r.LockID)
			if err != nil {
				return err
			}
		case OpInsert:
			stmt := `UPDATE ` + r.Table + ` SET tx_min_rolled_back = TRUE WHERE tx_min = $1 AND id = $2;`
			_, err := appConn.ExecContext(tx.ctx, stmt, tx.ID, r.ID)
			if err != nil {
				return err
			}
		}
	}

	stmt := `UPDATE transactions SET status = $1 WHERE id = $2;`
	_, err = mvccConn.ExecContext(tx.ctx, stmt, TxRolledBack, tx.ID)
	if err != nil {
		return err
	}

	tx.Status = TxRolledBack
	return t.Commit()
}

func Vacuum(ctx context.Context) (int, error) {
	rows, err := mvccConn.QueryContext(ctx, "SELECT id FROM transactions WHERE status = 'active'")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var activeTxs []TransactionData
	for rows.Next() {
		var tx TransactionData
		if err := rows.Scan(&tx.ID, &tx.CreatedAt, &tx.Status); err != nil {
			return 0, err
		}
		activeTxs = append(activeTxs, tx)
	}

	tables, err := getTables(appConn)
	if err != nil {
		return 0, err
	}

	// fetch rows marked for deletion (inserted but rolled back, committed delete)
	delCount := 0
	for _, table := range tables {
		rows, err = appConn.QueryContext(ctx, "SELECT tx_min, tx_max, id FROM"+table+" WHERE tx_max != 0 AND tx_max_committed = TRUE OR tx_min_rolled_back = TRUE")
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		for rows.Next() {
			var txMin, txMax, id int
			if err := rows.Scan(&txMin, &txMax, &id); err != nil {
				return 0, err
			}

			canBeDeleted := true
			for _, tx := range activeTxs {
				if tx.ID > txMin && tx.ID < txMax {
					canBeDeleted = false
					break
				}
			}

			if canBeDeleted {
				_, err := appConn.ExecContext(ctx, "DELETE FROM "+table+" WHERE tx_min=$1 AND tx_max=$2 AND id=$3", txMin, txMax, id)
				if err != nil {
					return 0, err
				}
				delCount++
			}
		}
	}

	if delCount > 0 {
		log.Info("Vacuumed %d records", delCount)
	}

	return delCount, nil
}

func getTables(conn *sql.DB) ([]string, error) {
	rows, err := conn.Query(`
	SELECT tablename
	FROM pg_catalog.pg_tables
	WHERE schemaname != 'pg_catalog'
	AND schemaname != 'information_schema'
	AND tablename != 'schema_migrations';`)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			return []string{}, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func (tx *Transaction) addResourceDependency(table string, id int) error {
	// Create dependency path: root.dependencies.tx_{txid}.{table}_{id}
	depPath := fmt.Sprintf("root.dependencies.tx_%d.%s_%d", tx.ID, table, id)

	_, err := mvccConn.ExecContext(tx.ctx, `
        INSERT INTO paths (path, type, name, dependency_type)
        VALUES ($1, 'dependency', $2, 'target')`,
		depPath, fmt.Sprintf("%s_%d", table, id),
	)

	return err
}

func (tx *Transaction) addResourceLock(table string, id int) error {
	// Create lock path: root.locks.tx_{txid}.{table}_{id}
	lockPath := fmt.Sprintf("root.locks.tx_%d.%s_%d", tx.ID, table, id)

	_, err := mvccConn.ExecContext(tx.ctx, `
        INSERT INTO paths (path, type, name, dependency_type)
        VALUES ($1, 'lock', $2, 'source')`,
		lockPath, fmt.Sprintf("%s_%d", table, id),
	)

	return err
}

func (tx *Transaction) checkDependencyCycle() error {
	log.Debug("Starting dependency cycle check for transaction %d", tx.ID)
	rows, err := mvccConn.QueryContext(tx.ctx, `
        WITH RECURSIVE dependency_chain AS (
            SELECT path, 1 as depth
            FROM paths
            WHERE path <@ text2ltree('root.dependencies.tx_' || $1::text)
            AND dependency_type = 'target'
            
            UNION ALL
            
            SELECT p.path, dc.depth + 1
            FROM paths p
            JOIN dependency_chain dc ON p.path <@ dc.path
            WHERE p.dependency_type = 'target'
            AND p.path != dc.path
        )
        SELECT EXISTS (
            SELECT 1 FROM dependency_chain
            WHERE depth > (
                SELECT COUNT(DISTINCT id) 
                FROM transactions 
                WHERE status = 'active'
            )
        ) as has_cycle;`,
		tx.ID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var hasCycle bool
	if rows.Next() {
		err = rows.Scan(&hasCycle)
		if err != nil {
			return err
		}
	}

	if hasCycle {
		return errors.New("dependency cycle detected")
	}
	return nil
}

func (tx *Transaction) cleanupDependencies() error {
	log.Debug("Cleaning up dependencies for transaction %d", tx.ID)
	_, err := mvccConn.ExecContext(tx.ctx, `
        DELETE FROM paths 
        WHERE path <@ text2ltree('root.dependencies.tx_' || $1::text) 
        OR path <@ text2ltree('root.locks.tx_' || $1::text)`,
		strconv.Itoa(tx.ID),
	)
	return err
}
