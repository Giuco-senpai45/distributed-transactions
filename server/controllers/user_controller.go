package controllers

import (
	"context"
	"dt/models"
	"dt/services"
	"dt/utils"
	"dt/utils/log"
	"encoding/json"
	"net/http"
	"strconv"
)

type UserController struct {
	service *services.UserService
}

func NewUserController(service *services.UserService) *UserController {
	return &UserController{service: service}
}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	log.Info("Getting user by ID")
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := c.service.GetUser(context.Background(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Info("Creating user")
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = c.service.CreateUser(context.Background(), &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, user)
}

func (c *UserController) Login(w http.ResponseWriter, r *http.Request) {
	log.Info("Logging in user")
	var credentials struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Debug("Controller: Logging in user with username: %s", credentials.Username)

	user, err := c.service.GetUserByUsername(context.Background(), credentials.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (c *UserController) ListUsers(w http.ResponseWriter, r *http.Request) {
	log.Info("Listing users")
	users, err := c.service.ListUsers(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, users)
}
