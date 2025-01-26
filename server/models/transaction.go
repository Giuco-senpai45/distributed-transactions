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

	"github.com/lib/pq"
)

type TransactionData struct {
	ID        int
	CreatedAt time.Time
	Status    string
}

type Transaction struct {
	TransactionData
	records   []Record
	ctx       context.Context
	timestamp int64
}

var mu sync.Mutex

const operationDelay = 300 * time.Millisecond

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

	timestamp := time.Now().UnixNano()
	var id int

	stmt := "INSERT INTO transactions (status, timestamp) VALUES ($1, $2) RETURNING id"
	err := mvccConn.QueryRowContext(ctx, stmt, TxActive, timestamp).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	txData, err := GetTx(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction data: %v", err)
	}

	tx := &Transaction{
		TransactionData: *txData,
		ctx:             ctx,
		timestamp:       timestamp,
		records:         make([]Record, 0),
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
            SELECT *
            FROM ` + table

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

	var queryArgs []interface{}
	if len(args) > 0 {
		queryArgs = make([]interface{}, len(args)+1)
		queryArgs[0] = tx.ID
		copy(queryArgs[1:], args)
	} else {
		queryArgs = []interface{}{tx.ID}
	}

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
		for i, col := range cols {
			sv := reflect.Indirect(reflect.ValueOf(row[i])).Elem()
			switch col {
			case "tx_min":
				base.TxMin = int(sv.Int())
			case "tx_max":
				base.TxMax = int(sv.Int())
			case "tx_min_committed":
				base.TxMinCommitted = sv.Bool()
			case "tx_max_committed":
				base.TxMaxCommitted = sv.Bool()
			case "tx_min_rolled_back":
				base.TxMinRolledBack = sv.Bool()
			case "tx_max_rolled_back":
				base.TxMaxRolledBack = sv.Bool()
			}
		}

		visible := tx.IsRowVisible(&base)

		if visible {
			result := make(map[string]interface{})
			for i, col := range cols[6:] {
				sv := reflect.Indirect(reflect.ValueOf(row[i+6])).Elem()
				switch sv.Kind() {
				case reflect.Int64:
					result[col] = sv.Int()
				case reflect.Bool:
					result[col] = sv.Bool()
				case reflect.String:
					result[col] = sv.String()
				default:
					log.Error("Unknown type: %v", sv.Kind())
				}
			}
			results = append(results, result)
		}
	}

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
	time.Sleep(operationDelay)

	var id int
	seqName := table + "_id_seq"
	err := appConn.QueryRowContext(tx.ctx, fmt.Sprintf("SELECT nextval('%s')", seqName)).Scan(&id)
	if err != nil {
		return 0, err
	}

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

	stmt := "INSERT INTO " + table + " ("
	stmt += strings.Join(allFields, ", ")
	stmt += ") VALUES ("
	stmt += strings.Join(makeQueryParams(1, len(allFields)+1), ", ")
	stmt += ")"

	allValues := []interface{}{
		id,
		tx.ID,
		0,
		false,
		false,
		false,
		false,
	}
	allValues = append(allValues, values...)

	if _, err := appConn.ExecContext(tx.ctx, stmt, allValues...); err != nil {
		return 0, fmt.Errorf("error inserting: %v", err)
	}

	tx.records = append(tx.records, Record{
		Table:     table,
		ID:        id,
		Operation: OpInsert,
	})

	err = tx.Commit()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (tx *Transaction) Update(table string, id int, fields []string, values ...any) error {
	if tx.Status != TxActive {
		return fmt.Errorf("transaction %d is not active", tx.ID)
	}

	time.Sleep(operationDelay)

	if err := tx.acquireLock(table, id, WriteLock); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to acquire lock: %v", err)
	}

	time.Sleep(operationDelay)

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

	insertValues := []interface{}{
		id,
		tx.ID,
		0,
		false,
		false,
		false,
		false,
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

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (tx *Transaction) Delete(table string, id int) error {
	time.Sleep(operationDelay)

	if err := tx.acquireLock(table, id, WriteLock); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to acquire lock: %v", err)
	}

	time.Sleep(operationDelay)

	var exists bool
	err := appConn.QueryRowContext(tx.ctx,
		"SELECT EXISTS(SELECT 1 FROM "+table+" WHERE id=$1)", id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("record doesn't exist")
	}

	if err := tx.acquireLock(table, id, WriteLock); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to acquire lock: %v", err)
	}

	base, err := tx.selectRecord(table, id)
	if err != nil {
		return fmt.Errorf("failed to select record: %v", err)
	}

	if !tx.IsRowVisible(base) {
		return fmt.Errorf("transaction %v aborted due to concurrency", tx.ID)
	}

	updateStmt := `UPDATE ` + table + ` 
                   SET tx_max = $1, tx_max_rolled_back = FALSE
                   WHERE tx_min = $2 AND id = $3`
	if _, err := appConn.ExecContext(tx.ctx, updateStmt, tx.ID, base.TxMin, id); err != nil {
		return err
	}

	tx.records = append(tx.records, Record{
		Table:     table,
		ID:        id,
		Operation: OpDelete,
	})

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (tx *Transaction) Commit() error {
	defer func() {
		log.Debug("Deleting locks for transaction %d", tx.ID)
		_, err := mvccConn.ExecContext(tx.ctx,
			"DELETE FROM locks WHERE txid = $1",
			tx.ID)
		if err != nil {
			log.Error("Failed to delete locks: %v", err)
		}
	}()
	log.Info("Starting commit for transaction %d", tx.ID)

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

	_, err = mvccConn.ExecContext(tx.ctx,
		"DELETE FROM locks WHERE txid = $1",
		tx.ID)
	if err != nil {
		t.Rollback()
		return err
	}

	stmt := `UPDATE transactions SET status = $1 WHERE id = $2;`
	_, err = mvccConn.ExecContext(tx.ctx, stmt, TxCommitted, tx.ID)
	if err != nil {
		return err
	}

	tx.Status = TxCommitted
	return t.Commit()
}

func (tx *Transaction) Rollback() error {
	if tx == nil {
		return errors.New("attempt to rollback nil transaction")
	}
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
	// Get active transactions
	rows, err := mvccConn.QueryContext(ctx, `
        SELECT id, created_at, status 
        FROM transactions 
        WHERE status = 'active'`)
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

	tables := []string{"users", "accounts", "audit"}
	delCount := 0

	for _, table := range tables {
		rows, err = appConn.QueryContext(ctx, `
        WITH duplicates AS (
            SELECT id, 
                   tx_min,
                   tx_max,
                   row_number() OVER (
                       PARTITION BY id 
                       ORDER BY tx_max DESC
                   ) as rn
            FROM `+table+`
            WHERE tx_max_committed = true
        )
        SELECT d.id, d.tx_min, d.tx_max
        FROM duplicates d
        WHERE d.rn > 1`)

		if err != nil {
			return 0, err
		}

		toDelete := make([]int, 0)
		rowCount := 0

		for rows.Next() {
			rowCount++
			var txMin, txMax, id int
			if err := rows.Scan(&id, &txMin, &txMax); err != nil {
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
				toDelete = append(toDelete, id)
			}
		}

		log.Info("Found %d rows marked for deletion in table %s", rowCount, table)

		if len(toDelete) > 0 {
			query := fmt.Sprintf("DELETE FROM %s WHERE id = ANY($1)", table)
			result, err := appConn.ExecContext(ctx, query, pq.Array(toDelete))
			if err != nil {
				return 0, err
			}
			count, _ := result.RowsAffected()
			delCount += int(count)
		}
	}

	return delCount, nil
}

func (tx *Transaction) AcquireLock(table string, id int, lockType LockType) error {
	return tx.acquireLock(table, id, lockType)
}

func getTables(conn *sql.DB) ([]string, error) {
	rows, err := conn.Query(`
    SELECT tablename
    FROM pg_catalog.pg_tables
    WHERE schemaname != 'pg_catalog'
    AND schemaname != 'information_schema'
    AND tablename != 'schema_migrations';`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %v", err)
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %v", err)
		}
		tables = append(tables, tableName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %v", err)
	}

	if len(tables) == 0 {
		log.Debug("No tables found in database")
	} else {
		log.Debug("Found tables: %v", tables)
	}

	return tables, nil
}

func (tx *Transaction) checkDependencyCycle() error {
	// log.Debug("Starting dependency cycle check for transaction %d", tx.ID)
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
