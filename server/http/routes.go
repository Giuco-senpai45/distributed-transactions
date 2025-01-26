package routes

import (
	"dt/controllers"
	"dt/models"
	"fmt"
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
	router.HandleFunc("POST /accounts/transfer", accountController.Transfer)

	router.HandleFunc("GET /audits/{id}", auditController.GetAudits)
	router.HandleFunc("POST /audits", auditController.CreateAudit)

	router.HandleFunc("POST /vacuum", func(w http.ResponseWriter, r *http.Request) {
		count, err := models.Vacuum(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Vacuumed %d records", count)
	})
}
