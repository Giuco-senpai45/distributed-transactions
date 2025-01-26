package services

import (
	"context"
	"dt/models"
	"dt/utils/log"
	"fmt"
)

type UserService struct {
	mvccService *MVCCService
}

func NewUserService(mvccService *MVCCService) *UserService {
	return &UserService{mvccService: mvccService}
}

func (us *UserService) GetUser(ctx context.Context, userID int) (*models.User, error) {
	tx, err := us.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	results, err := tx.Where("users", "id", userID)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	user := &models.User{
		ID:       int(results[0]["id"].(int64)),
		Username: results[0]["username"].(string),
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserService) CreateUser(ctx context.Context, user *models.User) error {
	log.Debug("Creating user with data: %v", user)
	tx, err := us.mvccService.OpenTx(ctx)
	if err != nil {
		log.Error("Error opening transaction: %v", err)
		tx.Rollback()
		return err
	}

	// Check if username exists
	existing, err := tx.Where("users", "username", user.Username)
	if err != nil {
		log.Error("Error checking username: %v", err)
		tx.Rollback()
		return err
	}
	if len(existing) > 0 {
		tx.Rollback()
		return fmt.Errorf("username already exists")
	}

	id, err := tx.Insert("users", []string{"username"}, user.Username)
	if err != nil {
		log.Error("Error inserting user: %v", err)
		tx.Rollback()
		return err
	}
	user.ID = id

	err = tx.Commit()
	if err != nil {
		log.Error("Error committing transaction: %v", err)
		return err
	}

	return nil
}

func (us *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	log.Debug("Service: Getting user by username: %v", username)
	tx, err := us.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, err
	}

	results, err := tx.Where("users", "username", username)
	if err != nil {
		log.Error("Error querying where user: %v", err)
		tx.Rollback()
		return nil, err
	}

	if len(results) == 0 {
		log.Error("User not found")
		tx.Rollback()
		return nil, fmt.Errorf("user not found")
	}

	user := &models.User{
		ID:       int(results[0]["id"].(int64)),
		Username: results[0]["username"].(string),
	}
	log.Debug("Service: Found user: %v", user)

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return user, nil
}

func (us *UserService) ListUsers(ctx context.Context) ([]models.User, error) {
	tx, err := us.mvccService.OpenTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Empty where clause for all records
	results, err := tx.Where("users", "")
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0, len(results))
	for _, result := range results {
		users = append(users, models.User{
			ID:       int(result["id"].(int64)),
			Username: result["username"].(string),
		})
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return users, nil
}
