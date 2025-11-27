package api

import (
	"encoding/json"
	"fbscheduler/internal/db"
	"net/http"

	"github.com/gorilla/mux"
)

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var post db.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if post.Status == "" {
		post.Status = "draft"
	}
	
	if err := h.store.CreatePost(&post); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}
	
	respondJSON(w, http.StatusCreated, post)
}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	limit := getQueryInt(r, "limit", 20)
	offset := getQueryInt(r, "offset", 0)
	
	posts, err := h.store.GetPosts(limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch posts")
		return
	}
	
	respondJSON(w, http.StatusOK, posts)
}

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	post, err := h.store.GetPostByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch post")
		return
	}
	
	if post == nil {
		respondError(w, http.StatusNotFound, "Post not found")
		return
	}
	
	respondJSON(w, http.StatusOK, post)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	var post db.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	post.ID = id
	if err := h.store.UpdatePost(&post); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update post")
		return
	}
	
	respondJSON(w, http.StatusOK, post)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if err := h.store.DeletePost(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete post")
		return
	}
	
	respondJSON(w, http.StatusOK, map[string]string{"message": "Post deleted"})
}
