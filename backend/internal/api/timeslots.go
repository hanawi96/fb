package api

import (
	"encoding/json"
	"net/http"

	"fbscheduler/internal/db"

	"github.com/gorilla/mux"
)

// ============================================
// TIME SLOTS API
// ============================================

// GetPageTimeSlots GET /api/pages/:id/timeslots - Danh sách khung giờ của page
func (h *Handler) GetPageTimeSlots(w http.ResponseWriter, r *http.Request) {
	pageID := mux.Vars(r)["id"]

	slots, err := h.store.GetTimeSlotsByPage(pageID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch time slots: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, slots)
}

// CreateTimeSlot POST /api/pages/:id/timeslots - Tạo khung giờ mới
func (h *Handler) CreateTimeSlot(w http.ResponseWriter, r *http.Request) {
	pageID := mux.Vars(r)["id"]

	var req struct {
		SlotName        string `json:"slot_name"`
		StartTime       string `json:"start_time"` // "13:00"
		EndTime         string `json:"end_time"`   // "15:00"
		DaysOfWeek      []int  `json:"days_of_week"`
		Priority        int    `json:"priority"`
		MaxPostsPerSlot int    `json:"max_posts_per_slot"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.StartTime == "" || req.EndTime == "" {
		respondError(w, http.StatusBadRequest, "start_time and end_time are required")
		return
	}

	// Normalize time format
	startTime := normalizeTimeFormat(req.StartTime)
	endTime := normalizeTimeFormat(req.EndTime)

	slot := &db.PageTimeSlot{
		PageID:          pageID,
		SlotName:        req.SlotName,
		StartTime:       startTime,
		EndTime:         endTime,
		DaysOfWeek:      req.DaysOfWeek,
		IsActive:        true,
		Priority:        req.Priority,
		MaxPostsPerSlot: 1,
	}

	if len(slot.DaysOfWeek) == 0 {
		slot.DaysOfWeek = []int{1, 2, 3, 4, 5, 6, 7}
	}
	if slot.Priority == 0 {
		slot.Priority = 5
	}
	if req.MaxPostsPerSlot > 0 {
		slot.MaxPostsPerSlot = req.MaxPostsPerSlot
	}

	if err := h.store.CreateTimeSlot(slot); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create time slot: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, slot)
}

// UpdateTimeSlot PUT /api/timeslots/:id - Cập nhật khung giờ
func (h *Handler) UpdateTimeSlot(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	slot, err := h.store.GetTimeSlotByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Time slot not found")
		return
	}

	var req struct {
		SlotName        string `json:"slot_name"`
		StartTime       string `json:"start_time"`
		EndTime         string `json:"end_time"`
		DaysOfWeek      []int  `json:"days_of_week"`
		IsActive        *bool  `json:"is_active"`
		Priority        int    `json:"priority"`
		MaxPostsPerSlot int    `json:"max_posts_per_slot"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SlotName != "" {
		slot.SlotName = req.SlotName
	}
	if req.StartTime != "" {
		slot.StartTime = normalizeTimeFormat(req.StartTime)
	}
	if req.EndTime != "" {
		slot.EndTime = normalizeTimeFormat(req.EndTime)
	}
	if len(req.DaysOfWeek) > 0 {
		slot.DaysOfWeek = req.DaysOfWeek
	}
	if req.IsActive != nil {
		slot.IsActive = *req.IsActive
	}
	if req.Priority > 0 {
		slot.Priority = req.Priority
	}
	if req.MaxPostsPerSlot > 0 {
		slot.MaxPostsPerSlot = req.MaxPostsPerSlot
	}

	if err := h.store.UpdateTimeSlot(slot); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update time slot: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, slot)
}

// DeleteTimeSlot DELETE /api/timeslots/:id - Xóa khung giờ
func (h *Handler) DeleteTimeSlot(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.store.DeleteTimeSlot(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete time slot: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Time slot deleted successfully"})
}

// normalizeTimeFormat chuyển "13:00" thành "13:00:00"
func normalizeTimeFormat(t string) string {
	if len(t) == 5 {
		return t + ":00"
	}
	return t
}
