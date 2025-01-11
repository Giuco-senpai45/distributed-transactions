package models

type Record struct {
	Table     string
	ID        int
	Operation string
	LockID    int
}

type RecordData struct {
	TxMin           int  `json:"-"` // transaction that created the row version
	TxMax           int  `json:"-"` // transaction that deleted the row version (set after UPDATE or DELETE)
	TxMinCommitted  bool `json:"-"` // txMin committed (row is now considered by other transactions)
	TxMaxCommitted  bool `json:"-"` // txMax committed
	TxMinRolledBack bool `json:"-"` // txMin rolled back (row is not considered by other transactions)
	TxMaxRolledBack bool `json:"-"` // txMax rolled back
}
