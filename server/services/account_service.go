package services

import (
	"context"
	"dt/utils/log"
	"fmt"
	"time"
)

type Account struct {
	ID      int `json:"id"`
	UserID  int `json:"user_id"`
	Balance int `json:"balance"`
}

type AccountService struct {
	mvccService *MVCCService
}

func NewAccountService(mvccService *MVCCService) *AccountService {
	return &AccountService{mvccService: mvccService}
}

func (as *AccountService) ListAccounts(ctx context.Context, userID int) (*[]Account, error) {
	log.Info("Service: ListAccounts called with userID=%d", userID)
	tx, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Use Transaction.Where for MVCC-aware query
	results, err := tx.Where("accounts", "user_id", userID)
	if err != nil {
		return nil, err
	}

	// Map results to Account structs
	accounts := make([]Account, 0, len(results))
	for _, result := range results {
		account := Account{
			ID:      int(result["id"].(int64)),
			UserID:  int(result["user_id"].(int64)),
			Balance: int(result["balance"].(int64)),
		}
		accounts = append(accounts, account)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &accounts, nil
}

func (as *AccountService) CreateAccount(ctx context.Context, userID int) (*Account, error) {
	log.Info("Service: CreateAccount called with userID=%d", userID)
	tx, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	id, err := tx.Insert("accounts", []string{"user_id", "balance"}, userID, 0)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:      id,
		UserID:  userID,
		Balance: 0,
	}, nil
}

func (as *AccountService) Deposit(ctx context.Context, accountID, amount int) (*Account, error) {
	tx, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get latest account data
	accounts, err := tx.Where("accounts", "id", accountID)
	if err != nil || len(accounts) == 0 {
		return nil, fmt.Errorf("account not found")
	}

	var acc Account
	acc.ID = accountID
	acc.UserID = int(accounts[0]["user_id"].(int64))
	acc.Balance = int(accounts[0]["balance"].(int64))
	newBalance := acc.Balance + amount

	// Update with all required fields
	err = tx.Update("accounts", accountID,
		[]string{"balance", "user_id"},
		newBalance, acc.UserID)
	if err != nil {
		return nil, err
	}

	// Create audit entry
	_, err = tx.Insert("audit", []string{"timestamp", "operation", "user_id"},
		time.Now(), "deposit", acc.UserID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	acc.Balance = newBalance
	return &acc, nil
}
