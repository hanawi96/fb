package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lib/pq"
)

// CheckConflictRequest request để check xung đột thời gian
type CheckConflictRequest struct {
	PageIDs       []string  `json:"page_ids"`
	ScheduledTime time.Time `json:"scheduled_time"`
}

// ConflictPage thông tin page bị xung đột
type ConflictPage struct {
	PageID   string `json:"page_id"`
	PageName string `json:"page_name"`
}

// CheckConflictResponse response của API check conflict
type CheckConflictResponse struct {
	HasConflict    bool           `json:"has_conflict"`
	ConflictPages  []ConflictPage `json:"conflict_pages"`
	NoConflictPages []ConflictPage `json:"no_conflict_pages"`
}

// CheckScheduleConflict kiểm tra xung đột thời gian đăng bài
func (h *Handler) CheckScheduleConflict(w http.ResponseWriter, r *http.Request) {
	var req CheckConflictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate
	if len(req.PageIDs) == 0 {
		respondError(w, http.StatusBadRequest, "page_ids is required")
		return
	}

	// Chuẩn hóa thời gian về UTC, làm tròn đến phút
	scheduledUTC := req.ScheduledTime.UTC().Truncate(time.Minute)

	// Query database để tìm các bài đã lên lịch trùng thời gian
	query := `
		SELECT DISTINCT
			sp.page_id,
			pg.page_name
		FROM scheduled_posts sp
		JOIN pages pg ON pg.id = sp.page_id
		WHERE sp.page_id = ANY($1)
			AND DATE_TRUNC('minute', sp.scheduled_time) = $2
			AND sp.status IN ('pending', 'processing')
	`

	rows, err := h.db.Query(query, pq.Array(req.PageIDs), scheduledUTC)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	// Map để track pages có xung đột
	conflictMap := make(map[string]string) // pageID -> pageName

	for rows.Next() {
		var pageID, pageName string
		if err := rows.Scan(&pageID, &pageName); err != nil {
			continue
		}
		conflictMap[pageID] = pageName
	}

	// Tạo response
	response := CheckConflictResponse{
		HasConflict:     len(conflictMap) > 0,
		ConflictPages:   make([]ConflictPage, 0),
		NoConflictPages: make([]ConflictPage, 0),
	}

	// Phân loại pages
	for _, pageID := range req.PageIDs {
		page, err := h.store.GetPageByID(pageID)
		if err != nil || page == nil {
			continue
		}

		if pageName, exists := conflictMap[pageID]; exists {
			response.ConflictPages = append(response.ConflictPages, ConflictPage{
				PageID:   pageID,
				PageName: pageName,
			})
		} else {
			response.NoConflictPages = append(response.NoConflictPages, ConflictPage{
				PageID:   pageID,
				PageName: page.PageName,
			})
		}
	}

	respondJSON(w, http.StatusOK, response)
}
