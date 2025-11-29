package api

import (
	"encoding/json"
	"fbscheduler/internal/db"
	"log"
	"net/http"
	"sort"
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

	// Chuẩn hóa về UTC để so sánh chính xác
	scheduledUTC := req.ScheduledTime.UTC()
	nowUTC := time.Now().UTC()

	log.Printf("📅 Schedule request: scheduled=%v, now=%v, is_future=%v",
		scheduledUTC.Format("2006-01-02 15:04:05"),
		nowUTC.Format("2006-01-02 15:04:05"),
		scheduledUTC.After(nowUTC))

	if scheduledUTC.Before(nowUTC) {
		respondError(w, http.StatusBadRequest, "scheduled_time must be in the future")
		return
	}
	
	// Create scheduled posts for each page
	// Lưu thời gian ở UTC
	var scheduled []db.ScheduledPost
	for _, pageID := range req.PageIDs {
		sp := &db.ScheduledPost{
			PostID:        req.PostID,
			PageID:        pageID,
			ScheduledTime: scheduledUTC, // Luôn lưu UTC
			Status:        "pending",
			MaxRetries:    3,
		}
		
		// Tìm time_slot_id phù hợp với thời gian đã chọn
		timeSlotID, err := h.findMatchingTimeSlot(pageID, scheduledUTC)
		if err == nil && timeSlotID != "" {
			sp.TimeSlotID = &timeSlotID
		}
		
		// Tự động assign account cho page (lấy primary account)
		account, err := h.store.GetPrimaryAccountForPage(pageID)
		if err == nil && account != nil {
			sp.AccountID = &account.ID
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

// TestScheduleNow POST /api/schedule/:id/test - Test đăng ngay 1 scheduled post (DEV ONLY)
// findMatchingTimeSlot tìm time_slot_id phù hợp với thời gian đã chọn
// Nếu slot đầy, tự động tìm slot tiếp theo còn chỗ
func (h *Handler) findMatchingTimeSlot(pageID string, scheduledTime time.Time) (string, error) {
	// Lấy tất cả time slots của page
	slots, err := h.store.GetTimeSlotsByPage(pageID)
	if err != nil || len(slots) == 0 {
		return "", err
	}

	// Chuyển sang Vietnam timezone để so sánh
	scheduledVN := scheduledTime.In(time.FixedZone("Asia/Ho_Chi_Minh", 7*3600))
	dayOfWeek := int(scheduledVN.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7 // Sunday = 7
	}

	// Lọc slots theo ngày trong tuần và sắp xếp theo thời gian
	var validSlots []db.PageTimeSlot
	for _, slot := range slots {
		for _, day := range slot.DaysOfWeek {
			if day == dayOfWeek {
				validSlots = append(validSlots, slot)
				break
			}
		}
	}

	if len(validSlots) == 0 {
		return "", nil
	}

	// Sort slots theo start_time
	sort.Slice(validSlots, func(i, j int) bool {
		return validSlots[i].StartTime < validSlots[j].StartTime
	})

	// Tìm slot đầu tiên còn chỗ
	for _, slot := range validSlots {
		available, err := h.store.IsSlotAvailable(slot.ID, scheduledTime)
		if err == nil && available {
			return slot.ID, nil
		}
	}

	// Nếu tất cả slots đều đầy, trả về empty
	return "", nil
}

func (h *Handler) TestScheduleNow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// Cập nhật scheduled_time về ngay bây giờ để scheduler pick up
	query := `UPDATE scheduled_posts SET scheduled_time = NOW(), status = 'pending' WHERE id = $1`
	_, err := h.db.Exec(query, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update: "+err.Error())
		return
	}
	
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Scheduled post updated to NOW. Scheduler will pick it up in ~30 seconds.",
		"id":      id,
	})
}
