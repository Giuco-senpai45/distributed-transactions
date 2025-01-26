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

func (as *AuditService) GetAudits(ctx context.Context, userID int) ([]*models.Audit, error) {
	tx, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open transaction: %v", err)
	}
	defer tx.Rollback()

	results, err := tx.SelectByColumn("audit", "user_id", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch audits: %v", err)
	}

	audits := make([]*models.Audit, 0, len(results))
	for _, result := range results {
		var timestamp time.Time
		if ts, ok := result["timestamp"].(time.Time); ok {
			timestamp = ts
		} else {
			timestamp = time.Now()
		}

		audit := &models.Audit{
			ID:        int(result["id"].(int64)),
			Timestamp: timestamp,
			Operation: result["operation"].(string),
			UserID:    int(result["user_id"].(int64)),
		}
		audits = append(audits, audit)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %v", err)
	}

	return audits, nil
}

func (as *AuditService) CreateAudit(ctx context.Context, audit *models.Audit) error {
	tx, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Insert("audit", []string{"operation", "user_id", "timestamp"}, audit.Operation, audit.UserID, time.Now())
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
