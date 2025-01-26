package controllers

import (
	"context"
	"dt/models"
	"dt/services"
	"dt/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

type AuditController struct {
	service *services.AuditService
}

func NewAuditController(service *services.AuditService) *AuditController {
	return &AuditController{service: service}
}

func (c *AuditController) GetAudits(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	audits, err := c.service.GetAudits(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, audits)
}

func (c *AuditController) CreateAudit(w http.ResponseWriter, r *http.Request) {
	var audit models.Audit
	err := json.NewDecoder(r.Body).Decode(&audit)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = c.service.CreateAudit(context.Background(), &audit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, audit)
}
