package api

import (
	"encoding/json"
	"net/http"

	"fbscheduler/internal/config"
	"fbscheduler/internal/scheduler"
)

// ============================================
// SCHEDULE PREVIEW API
// ============================================

// PreviewSchedule POST /api/schedule/preview - Preview trước khi schedule
func (h *Handler) PreviewSchedule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PostID        string   `json:"post_id"`
		PageIDs       []string `json:"page_ids"`
		PreferredDate string   `json:"preferred_date"` // "2024-01-15"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.PageIDs) == 0 {
		respondError(w, http.StatusBadRequest, "page_ids is required")
		return
	}

	// Parse date using Vietnam timezone
	preferredDate := config.NowVN()
	if req.PreferredDate != "" {
		parsed, err := config.ParseDateVN(req.PreferredDate)
		if err == nil {
			preferredDate = parsed
		}
	}

	// Create scheduling service
	schedulingService := scheduler.NewSchedulingService(h.store)

	// Get preview
	preview, err := schedulingService.PreviewSchedule(req.PostID, req.PageIDs, preferredDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate schedule: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, preview)
}

// ScheduleWithPreview POST /api/schedule/smart - Schedule với smart algorithm
func (h *Handler) ScheduleWithPreview(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PostID        string   `json:"post_id"`
		PageIDs       []string `json:"page_ids"`
		PreferredDate string   `json:"preferred_date"`
		Confirm       bool     `json:"confirm"` // true = tạo schedule luôn, false = chỉ preview
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PostID == "" || len(req.PageIDs) == 0 {
		respondError(w, http.StatusBadRequest, "post_id and page_ids are required")
		return
	}

	// Parse date using Vietnam timezone
	preferredDate := config.NowVN()
	if req.PreferredDate != "" {
		parsed, err := config.ParseDateVN(req.PreferredDate)
		if err == nil {
			preferredDate = parsed
		}
	}

	// Create scheduling service
	schedulingService := scheduler.NewSchedulingService(h.store)

	// Calculate schedule
	preview, err := schedulingService.SchedulePostToPages(req.PostID, req.PageIDs, preferredDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate schedule: "+err.Error())
		return
	}

	// If confirm, create scheduled posts
	if req.Confirm {
		if err := schedulingService.ConfirmSchedule(req.PostID, preview.Results); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create schedule: "+err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"message":       "Schedule created successfully",
			"preview":       preview,
			"scheduled":     true,
			"success_count": preview.SuccessCount,
		})
		return
	}

	// Return preview only
	respondJSON(w, http.StatusOK, preview)
}

// GetScheduleStats GET /api/schedule/stats - Thống kê schedule
func (h *Handler) GetScheduleStats(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")

	// Parse date using Vietnam timezone
	date := config.NowVN()
	if dateStr != "" {
		parsed, err := config.ParseDateVN(dateStr)
		if err == nil {
			date = parsed
		}
	}

	schedulingService := scheduler.NewSchedulingService(h.store)
	stats, err := schedulingService.GetScheduleStats(date)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
