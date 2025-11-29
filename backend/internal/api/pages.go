package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (h *Handler) GetPages(w http.ResponseWriter, r *http.Request) {
	pages, err := h.store.GetPagesWithAccount()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch pages: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pages)
}

func (h *Handler) DeletePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.store.DeletePage(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete page: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Page deleted successfully"})
}

func (h *Handler) TogglePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.store.TogglePage(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to toggle page: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Page status updated successfully"})
}
