package models

import "database/sql"

var appConn *sql.DB
var mvccConn *sql.DB

func New(appPool, mvccPool *sql.DB) {
	appConn = appPool
	mvccConn = mvccPool
}

const (
	TxActive     = "active"
	TxCommitted  = "committed"
	TxRolledBack = "rolled_back"
)

const (
	OpInsert = "insert"
	OpUpdate = "update"
	OpDelete = "delete"
)

type Models struct {
	User        User
	Account     Account
	Record      Record
	Path        Path
	Transaction Transaction
}
