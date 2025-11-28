package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// ============================================
// PAGE ACCOUNT ASSIGNMENTS API
// ============================================

// GetPageAssignments GET /api/pages/:id/assignments - Danh sách accounts của page
func (h *Handler) GetPageAssignments(w http.ResponseWriter, r *http.Request) {
	pageID := mux.Vars(r)["id"]

	assignments, err := h.store.GetAssignmentsByPage(pageID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch assignments: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, assignments)
}

// AssignPageToAccount POST /api/pages/:id/assign - Gán page vào account
func (h *Handler) AssignPageToAccount(w http.ResponseWriter, r *http.Request) {
	pageID := mux.Vars(r)["id"]

	var req struct {
		AccountID string `json:"account_id"`
		IsPrimary bool   `json:"is_primary"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AccountID == "" {
		respondError(w, http.StatusBadRequest, "account_id is required")
		return
	}

	// Check if account can accept more pages
	canAssign, err := h.store.CanAssignMorePages(req.AccountID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check account capacity: "+err.Error())
		return
	}

	if !canAssign {
		respondError(w, http.StatusBadRequest, "Account has reached maximum pages limit")
		return
	}

	if err := h.store.AssignPageToAccount(pageID, req.AccountID, req.IsPrimary); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to assign page: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Page assigned successfully"})
}

// UnassignPageFromAccount DELETE /api/pages/:id/assign/:accountId - Bỏ gán page
func (h *Handler) UnassignPageFromAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["id"]
	accountID := vars["accountId"]

	if err := h.store.UnassignPageFromAccount(pageID, accountID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to unassign page: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Page unassigned successfully"})
}

// SetPrimaryAccount PUT /api/pages/:id/primary - Đặt account làm primary
func (h *Handler) SetPrimaryAccount(w http.ResponseWriter, r *http.Request) {
	pageID := mux.Vars(r)["id"]

	var req struct {
		AccountID string `json:"account_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AccountID == "" {
		respondError(w, http.StatusBadRequest, "account_id is required")
		return
	}

	if err := h.store.SetPrimaryAccount(pageID, req.AccountID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to set primary account: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Primary account set successfully"})
}

// GetUnassignedPages GET /api/pages/unassigned - Danh sách pages chưa gán
func (h *Handler) GetUnassignedPages(w http.ResponseWriter, r *http.Request) {
	pages, err := h.store.GetUnassignedPages()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch unassigned pages: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pages)
}
