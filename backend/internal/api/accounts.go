package api

import (
	"encoding/json"
	"net/http"

	"fbscheduler/internal/db"

	"github.com/gorilla/mux"
)

// ============================================
// FACEBOOK ACCOUNTS API
// ============================================

// GetAccounts GET /api/accounts - Danh sách tất cả nick
func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.store.GetAllAccounts()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch accounts: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, accounts)
}

// GetAccount GET /api/accounts/:id - Chi tiết 1 nick
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	account, err := h.store.GetAccountByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Account not found")
		return
	}

	respondJSON(w, http.StatusOK, account)
}

// GetAccountPages GET /api/accounts/:id/pages - Danh sách pages của nick
func (h *Handler) GetAccountPages(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	assignments, err := h.store.GetAssignmentsByAccount(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch pages: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, assignments)
}

// CreateAccount POST /api/accounts - Tạo nick mới (sau OAuth)
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FbUserID       string `json:"fb_user_id"`
		FbUserName     string `json:"fb_user_name"`
		AccessToken    string `json:"access_token"`
		TokenExpiresAt string `json:"token_expires_at"`
		MaxPages       int    `json:"max_pages"`
		MaxPostsPerDay int    `json:"max_posts_per_day"`
		Notes          string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.FbUserID == "" || req.AccessToken == "" {
		respondError(w, http.StatusBadRequest, "fb_user_id and access_token are required")
		return
	}

	// Check if account already exists
	existing, _ := h.store.GetAccountByFbUserID(req.FbUserID)
	if existing != nil {
		// Update existing account
		existing.FbUserName = req.FbUserName
		existing.AccessToken = req.AccessToken
		if req.MaxPages > 0 {
			existing.MaxPages = req.MaxPages
		}
		if req.MaxPostsPerDay > 0 {
			existing.MaxPostsPerDay = req.MaxPostsPerDay
		}
		existing.Notes = req.Notes
		existing.Status = "active"

		if err := h.store.UpdateAccount(existing); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update account: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, existing)
		return
	}

	// Create new account
	account := &db.FacebookAccount{
		FbUserID:       req.FbUserID,
		FbUserName:     req.FbUserName,
		AccessToken:    req.AccessToken,
		MaxPages:       5,
		MaxPostsPerDay: 20,
		Notes:          req.Notes,
		Status:         "active",
	}

	if req.MaxPages > 0 {
		account.MaxPages = req.MaxPages
	}
	if req.MaxPostsPerDay > 0 {
		account.MaxPostsPerDay = req.MaxPostsPerDay
	}

	if err := h.store.CreateAccount(account); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create account: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, account)
}

// UpdateAccount PUT /api/accounts/:id - Cập nhật nick
func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	account, err := h.store.GetAccountByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Account not found")
		return
	}

	var req struct {
		FbUserName     string `json:"fb_user_name"`
		MaxPages       int    `json:"max_pages"`
		MaxPostsPerDay int    `json:"max_posts_per_day"`
		Status         string `json:"status"`
		Notes          string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.FbUserName != "" {
		account.FbUserName = req.FbUserName
	}
	if req.MaxPages > 0 {
		account.MaxPages = req.MaxPages
	}
	if req.MaxPostsPerDay > 0 {
		account.MaxPostsPerDay = req.MaxPostsPerDay
	}
	if req.Status != "" {
		account.Status = req.Status
	}
	if req.Notes != "" {
		account.Notes = req.Notes
	}

	if err := h.store.UpdateAccount(account); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update account: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, account)
}

// DeleteAccount DELETE /api/accounts/:id - Xóa nick
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.store.DeleteAccount(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete account: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Account deleted successfully"})
}

// RefreshAccountToken POST /api/accounts/:id/refresh - Refresh token
func (h *Handler) RefreshAccountToken(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	account, err := h.store.GetAccountByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Account not found")
		return
	}

	var req struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AccessToken == "" {
		respondError(w, http.StatusBadRequest, "access_token is required")
		return
	}

	account.AccessToken = req.AccessToken
	account.Status = "active"

	if err := h.store.UpdateAccount(account); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update token: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Token refreshed successfully"})
}
