package controllers

import (
	"context"
	"dt/services"
	"dt/utils"
	"dt/utils/log"
	"encoding/json"
	"net/http"
	"strconv"
)

type AccountController struct {
	service *services.AccountService
}

func NewAccountController(service *services.AccountService) *AccountController {
	return &AccountController{service: service}
}

func (c *AccountController) ListAccounts(w http.ResponseWriter, r *http.Request) {
	log.Info("Controller: ListAccounts")
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	accounts, err := c.service.ListAccounts(context.Background(), userID)
	log.Info("Accounts: %v", accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, accounts)
}

func (c *AccountController) CreateAccount(w http.ResponseWriter, r *http.Request) {
	log.Info("Controller: CreateAccount")
	var req struct {
		UserID int `json:"user_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	account, err := c.service.CreateAccount(context.Background(), req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, account)
}

func (c *AccountController) Deposit(w http.ResponseWriter, r *http.Request) {
	log.Info("Controller: Deposit")
	var req struct {
		AccountID int `json:"account_id"`
		Amount    int `json:"amount"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	account, err := c.service.Deposit(context.Background(), req.AccountID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, account)
}
