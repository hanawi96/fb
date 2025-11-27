package api

import (
	"database/sql"
	"encoding/json"
	"fbscheduler/internal/db"
	"fbscheduler/internal/facebook"
	"net/http"
	"strconv"
)

type Handler struct {
	store    *db.Store
	fbClient *facebook.Client
	db       *sql.DB
}

func NewHandler(store *db.Store, database *sql.DB) *Handler {
	return &Handler{
		store:    store,
		fbClient: facebook.NewClient(),
		db:       database,
	}
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func getQueryInt(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}
