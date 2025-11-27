package api

import (
	"encoding/json"
	"fbscheduler/internal/db"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type ScheduleRequest struct {
	PostID        string    `json:"post_id"`
	PageIDs       []string  `json:"page_ids"`
	ScheduledTime time.Time `json:"scheduled_time"`
}

func (h *Handler) SchedulePost(w http.ResponseWriter, r *http.Request) {
	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate
	if req.PostID == "" || len(req.PageIDs) == 0 {
		respondError(w, http.StatusBadRequest, "post_id and page_ids are required")
		return
	}
	
	if req.ScheduledTime.Before(time.Now()) {
		respondError(w, http.StatusBadRequest, "scheduled_time must be in the future")
		return
	}
	
	// Create scheduled posts for each page
	var scheduled []db.ScheduledPost
	for _, pageID := range req.PageIDs {
		sp := &db.ScheduledPost{
			PostID:        req.PostID,
			PageID:        pageID,
			ScheduledTime: req.ScheduledTime,
			Status:        "pending",
			MaxRetries:    3,
		}
		
		if err := h.store.CreateScheduledPost(sp); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to schedule post: "+err.Error())
			return
		}
		
		scheduled = append(scheduled, *sp)
	}
	
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":   "Post scheduled successfully",
		"scheduled": scheduled,
	})
}

func (h *Handler) GetScheduledPosts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := getQueryInt(r, "limit", 50)
	offset := getQueryInt(r, "offset", 0)
	
	posts, err := h.store.GetScheduledPosts(status, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch scheduled posts")
		return
	}
	
	respondJSON(w, http.StatusOK, posts)
}

func (h *Handler) DeleteScheduledPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if err := h.store.DeleteScheduledPost(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete scheduled post")
		return
	}
	
	respondJSON(w, http.StatusOK, map[string]string{"message": "Scheduled post deleted"})
}

func (h *Handler) RetryScheduledPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if err := h.store.UpdateScheduledPostStatus(id, "pending"); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retry post")
		return
	}
	
	respondJSON(w, http.StatusOK, map[string]string{"message": "Post queued for retry"})
}
