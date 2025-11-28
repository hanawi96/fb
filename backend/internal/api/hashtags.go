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

type HashtagSet struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Hashtags  string    `json:"hashtags"`
	CreatedAt time.Time `json:"created_at"`
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

// GetSavedHashtags returns user's saved hashtag sets
func (h *Handler) GetSavedHashtags(w http.ResponseWriter, r *http.Request) {
	userID := 1

	rows, err := h.db.Query(`
		SELECT id, user_id, name, hashtags, created_at
		FROM hashtag_sets
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		http.Error(w, "Failed to fetch hashtag sets", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var hashtagSets []HashtagSet
	for rows.Next() {
		var hs HashtagSet
		err := rows.Scan(&hs.ID, &hs.UserID, &hs.Name, &hs.Hashtags, &hs.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan hashtag set", http.StatusInternalServerError)
			return
		}
		hashtagSets = append(hashtagSets, hs)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hashtagSets)
}

// SaveHashtags saves a hashtag set for user
func (h *Handler) SaveHashtags(w http.ResponseWriter, r *http.Request) {
	userID := 1

	var req struct {
		Name     string `json:"name"`
		Hashtags string `json:"hashtags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec(`
		INSERT INTO hashtag_sets (user_id, name, hashtags, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, name) DO UPDATE
		SET hashtags = $3, created_at = $4
	`, userID, req.Name, req.Hashtags, time.Now())

	if err != nil {
		http.Error(w, "Failed to save hashtag set", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Hashtag set saved successfully"})
}

// DeleteSavedHashtag deletes a saved hashtag set
func (h *Handler) DeleteSavedHashtag(w http.ResponseWriter, r *http.Request) {
	userID := 1

	hashtagID := r.URL.Query().Get("id")
	if hashtagID == "" {
		http.Error(w, "Hashtag set ID is required", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec(`
		DELETE FROM hashtag_sets
		WHERE id = $1 AND user_id = $2
	`, hashtagID, userID)

	if err != nil {
		http.Error(w, "Failed to delete hashtag set", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Hashtag set deleted successfully"})
}
