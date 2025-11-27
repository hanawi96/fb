package api

import "net/http"

func (h *Handler) GetPostLogs(w http.ResponseWriter, r *http.Request) {
	limit := getQueryInt(r, "limit", 50)
	offset := getQueryInt(r, "offset", 0)
	
	logs, err := h.store.GetPostLogs(limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch logs")
		return
	}
	
	respondJSON(w, http.StatusOK, logs)
}
