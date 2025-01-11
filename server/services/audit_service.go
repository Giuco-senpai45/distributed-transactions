package services

import (
	"context"
	"dt/models"
	"fmt"
	"time"
)

type AuditService struct {
	mvccService *MVCCService
}

func NewAuditService(mvccService *MVCCService) *AuditService {
	return &AuditService{mvccService: mvccService}
}

func (as *AuditService) GetAudit(ctx context.Context, auditID int) (*models.Audit, error) {
	tx, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	results, err := tx.Where("audits", "id", auditID)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("audit not found")
	}

	audit := &models.Audit{
		ID:        int(results[0]["id"].(int64)),
		Timestamp: results[0]["timestamp"].(time.Time),
		Operation: results[0]["operation"].(string),
		UserID:    int(results[0]["user_id"].(int64)),
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return audit, nil
}

func (as *AuditService) CreateAudit(ctx context.Context, audit *models.Audit) error {
	tx, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Insert("audits", []string{"operation", "user_id", "timestamp"}, audit.Operation, audit.UserID, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
