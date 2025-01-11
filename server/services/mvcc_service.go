package services

import (
	"context"
	"database/sql"
	"dt/models"
)

type MVCCService struct {
	mvccConn *sql.DB
	appConn  *sql.DB

	transaction []*models.Transaction
}

func NewMVCCService(mvccConn, appConn *sql.DB) *MVCCService {
	models.New(appConn, mvccConn)
	return &MVCCService{
		mvccConn: mvccConn,
		appConn:  appConn,
	}
}

func (mvccs *MVCCService) OpenTx(ctx context.Context) (*models.Transaction, error) {
	tx, err := models.OpenTx(ctx)
	if err != nil {
		return nil, err
	}

	mvccs.transaction = append(mvccs.transaction, tx)

	return tx, nil
}

func (mvccs *MVCCService) Vacuum() (int, error) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	return models.Vacuum(ctx)
}

func (mvccs *MVCCService) Cleanup() {
	// commit all open transactions
	for _, tx := range mvccs.transaction {
		tx.Commit()
	}
	mvccs.Vacuum()
}
