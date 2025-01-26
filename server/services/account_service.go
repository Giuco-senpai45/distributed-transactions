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

type TransferResult struct {
	FromAccount *Account `json:"from_account"`
	ToAccount   *Account `json:"to_account"`
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

	log.Debug("Raw query results: %+v", results)

	accounts := make([]Account, 0, len(results))
	for _, result := range results {
		log.Debug("Processing result: %+v", result)

		var account Account
		account.ID = int(result["id"].(int64))
		account.UserID = int(result["user_id"].(int64))

		if bal, ok := result["balance"].(int64); ok {
			account.Balance = int(bal)
		} else {
			log.Error("Invalid balance type: %T", result["balance"])
			account.Balance = 0
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

func (as *AccountService) Transfer(ctx context.Context, fromAccountID, toAccountID, amount int) (*TransferResult, error) {
	// First transaction - deduct from source
	tx1, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx1 open failed: %v", err)
	}
	defer tx1.Rollback()

	fromAccounts, err := tx1.Where("accounts", "id", fromAccountID)
	if err != nil || len(fromAccounts) == 0 {
		return nil, fmt.Errorf("source account not found")
	}

	fromAcc := Account{
		ID:      fromAccountID,
		UserID:  int(fromAccounts[0]["user_id"].(int64)),
		Balance: int(fromAccounts[0]["balance"].(int64)),
	}

	if fromAcc.Balance < amount {
		return nil, fmt.Errorf("insufficient balance")
	}

	if err = tx1.Update("accounts", fromAccountID,
		[]string{"balance", "user_id"},
		fromAcc.Balance-amount, fromAcc.UserID); err != nil {
		return nil, fmt.Errorf("source update failed: %v", err)
	}

	if err = tx1.Commit(); err != nil {
		return nil, fmt.Errorf("tx1 commit failed: %v", err)
	}

	// Second transaction - add to destination
	tx2, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx2 open failed: %v", err)
	}
	defer tx2.Rollback()

	toAccounts, err := tx2.Where("accounts", "id", toAccountID)
	if err != nil || len(toAccounts) == 0 {
		return nil, fmt.Errorf("destination account not found")
	}

	toAcc := Account{
		ID:      toAccountID,
		UserID:  int(toAccounts[0]["user_id"].(int64)),
		Balance: int(toAccounts[0]["balance"].(int64)),
	}

	if err = tx2.Update("accounts", toAccountID,
		[]string{"balance", "user_id"},
		toAcc.Balance+amount, toAcc.UserID); err != nil {
		return nil, fmt.Errorf("destination update failed: %v", err)
	}

	if err = tx2.Commit(); err != nil {
		return nil, fmt.Errorf("tx2 commit failed: %v", err)
	}

	// Third transaction - audit entry
	tx3, err := as.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx3 open failed: %v", err)
	}
	defer tx3.Rollback()

	if _, err = tx3.Insert("audit", []string{"timestamp", "operation", "user_id"},
		time.Now(), "transfer", fromAcc.UserID); err != nil {
		return nil, fmt.Errorf("audit creation failed: %v", err)
	}

	if err = tx3.Commit(); err != nil {
		return nil, fmt.Errorf("tx3 commit failed: %v", err)
	}

	fromAcc.Balance -= amount
	toAcc.Balance += amount

	return &TransferResult{
		FromAccount: &fromAcc,
		ToAccount:   &toAcc,
	}, nil
}
