package models

import (
	"fmt"
)

type Path struct {
	ID        int    `json:"id"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// path.go

// Add these constants for path types
const (
	PathTypeDependency = "dependency"
	PathTypeLock       = "lock"
	DependencyTarget   = "target"
	DependencySource   = "source"
)

// Add new methods to Path struct
func NewDependencyPath(txID int, targetTxID int) *Path {
	return &Path{
		Path: fmt.Sprintf("root.dependencies.tx_%d.tx_%d", txID, targetTxID),
		Type: PathTypeDependency,
		Name: fmt.Sprintf("tx_%d_depends_on_tx_%d", txID, targetTxID),
	}
}

func NewLockPath(txID int, table string, recordID int) *Path {
	return &Path{
		Path: fmt.Sprintf("root.locks.tx_%d.%s_%d", txID, table, recordID),
		Type: PathTypeLock,
		Name: fmt.Sprintf("tx_%d_locks_%s_%d", txID, table, recordID),
	}
}

// Add to transaction.go
func (tx *Transaction) addDependency(targetTxID int) error {
	path := NewDependencyPath(tx.ID, targetTxID)

	_, err := mvccConn.ExecContext(tx.ctx, `
        INSERT INTO paths (path, type, name, dependency_type)
        VALUES (text2ltree($1), $2, $3, $4)`,
		path.Path, path.Type, path.Name, DependencyTarget)

	if err != nil {
		return fmt.Errorf("failed to add dependency: %v", err)
	}

	// Check for cycles after adding dependency
	if err := tx.checkDependencyCycle(); err != nil {
		// Remove the dependency if cycle is detected
		tx.cleanupDependencies()
		return err
	}

	return nil
}
