package routes

import (
	"dt/controllers"
	"net/http"
)

func RegisterRoutes(router *http.ServeMux, userController *controllers.UserController, accountController *controllers.AccountController, auditController *controllers.AuditController) {
	router.HandleFunc("GET /users", userController.ListUsers)
	router.HandleFunc("GET /users/{id}", userController.GetUser)
	router.HandleFunc("POST /users", userController.CreateUser)
	router.HandleFunc("POST /users/login", userController.Login)

	router.HandleFunc("GET /accounts/{id}", accountController.ListAccounts)
	router.HandleFunc("POST /accounts", accountController.CreateAccount)
	router.HandleFunc("PATCH /accounts", accountController.Deposit)

	router.HandleFunc("GET /audits", auditController.GetAudit)
	router.HandleFunc("POST /audits", auditController.CreateAudit)
}
