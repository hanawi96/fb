package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type HashtagSearchResult struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MediaCount int64 `json:"media_count"`
}

type HashtagSearchResponse struct {
	Data []HashtagSearchResult `json:"data"`
}

type SavedHashtag struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Hashtag    string    `json:"name"`
	MediaCount int64     `json:"media_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// SearchHashtags searches for hashtags using Instagram Graph API
func (h *Handler) SearchHashtags(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Return mock data for now (Instagram API integration can be added later)
	mockResults := h.getMockHashtagResults(query)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockResults)
}

// Mock data for development/fallback
func (h *Handler) getMockHashtagResults(query string) HashtagSearchResponse {
	mockData := map[string][]HashtagSearchResult{
		"marketing": {
			{ID: "1", Name: "marketing", MediaCount: 15234567},
			{ID: "2", Name: "marketingdigital", MediaCount: 8765432},
			{ID: "3", Name: "marketingtips", MediaCount: 3456789},
			{ID: "4", Name: "marketingstrategy", MediaCount: 2345678},
		},
		"business": {
			{ID: "5", Name: "business", MediaCount: 25678901},
			{ID: "6", Name: "businessman", MediaCount: 12345678},
			{ID: "7", Name: "businessowner", MediaCount: 5678901},
		},
		"vietnam": {
			{ID: "8", Name: "vietnam", MediaCount: 18765432},
			{ID: "9", Name: "vietnamtravel", MediaCount: 9876543},
			{ID: "10", Name: "vietnamfood", MediaCount: 4567890},
		},
	}

	// Return mock data based on query
	for key, results := range mockData {
		if len(query) > 0 && len(key) >= len(query) && key[:len(query)] == query {
			return HashtagSearchResponse{Data: results}
		}
	}

	// Default mock results
	return HashtagSearchResponse{
		Data: []HashtagSearchResult{
			{ID: "default1", Name: query, MediaCount: 1234567},
			{ID: "default2", Name: query + "s", MediaCount: 567890},
			{ID: "default3", Name: query + "daily", MediaCount: 234567},
		},
	}
}

// GetSavedHashtags returns user's saved hashtags
func (h *Handler) GetSavedHashtags(w http.ResponseWriter, r *http.Request) {
	// For now, use a default user_id of 1 (can be enhanced with proper auth later)
	userID := 1

	rows, err := h.db.Query(`
		SELECT id, user_id, hashtag, media_count, created_at
		FROM saved_hashtags
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		http.Error(w, "Failed to fetch saved hashtags", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var savedHashtags []SavedHashtag
	for rows.Next() {
		var h SavedHashtag
		err := rows.Scan(&h.ID, &h.UserID, &h.Hashtag, &h.MediaCount, &h.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan hashtag", http.StatusInternalServerError)
			return
		}
		savedHashtags = append(savedHashtags, h)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(savedHashtags)
}

// SaveHashtags saves hashtags for user
func (h *Handler) SaveHashtags(w http.ResponseWriter, r *http.Request) {
	// For now, use a default user_id of 1 (can be enhanced with proper auth later)
	userID := 1

	var req struct {
		Hashtags []struct {
			Name       string `json:"name"`
			MediaCount int64  `json:"media_count"`
		} `json:"hashtags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Save each hashtag
	for _, hashtag := range req.Hashtags {
		_, err := h.db.Exec(`
			INSERT INTO saved_hashtags (user_id, hashtag, media_count, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, hashtag) DO UPDATE
			SET media_count = $3, created_at = $4
		`, userID, hashtag.Name, hashtag.MediaCount, time.Now())

		if err != nil {
			http.Error(w, "Failed to save hashtag", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Hashtags saved successfully"})
}

// DeleteSavedHashtag deletes a saved hashtag
func (h *Handler) DeleteSavedHashtag(w http.ResponseWriter, r *http.Request) {
	// For now, use a default user_id of 1 (can be enhanced with proper auth later)
	userID := 1

	hashtagID := r.URL.Query().Get("id")
	if hashtagID == "" {
		http.Error(w, "Hashtag ID is required", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec(`
		DELETE FROM saved_hashtags
		WHERE id = $1 AND user_id = $2
	`, hashtagID, userID)

	if err != nil {
		http.Error(w, "Failed to delete hashtag", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Hashtag deleted successfully"})
}
