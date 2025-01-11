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

func (c *AuditController) GetAudit(w http.ResponseWriter, r *http.Request) {
	auditID, err := strconv.Atoi(r.URL.Query().Get("audit_id"))
	if err != nil {
		http.Error(w, "Invalid audit ID", http.StatusBadRequest)
		return
	}

	audit, err := c.service.GetAudit(context.Background(), auditID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, audit)
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
