package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 10MB per file)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		fmt.Printf("❌ Parse multipart form error: %v\n", err)
		respondError(w, http.StatusBadRequest, "File too large")
		return
	}
	
	file, header, err := r.FormFile("image")
	if err != nil {
		fmt.Printf("❌ FormFile error: %v\n", err)
		respondError(w, http.StatusBadRequest, "No file uploaded: "+err.Error())
		return
	}
	defer file.Close()
	
	// Validate file type
	contentType := header.Header.Get("Content-Type")
	fmt.Printf("📤 Uploading file: %s, Content-Type: %s\n", header.Filename, contentType)
	
	// Accept common image formats
	validTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"video/mp4":       true,
		"video/quicktime": true,
	}
	
	if !validTypes[contentType] {
		fmt.Printf("❌ Invalid content type: %s\n", contentType)
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid file type: %s. Allowed: JPEG, PNG, GIF, WebP, MP4", contentType))
		return
	}
	
	// Create uploads directory if not exists
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create uploads directory")
		return
	}
	
	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)
	filepath := filepath.Join(uploadsDir, filename)
	
	// Save file
	dst, err := os.Create(filepath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()
	
	if _, err := io.Copy(dst, file); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	
	// Return public URL
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}
	publicURL := fmt.Sprintf("%s/uploads/%s", backendURL, filename)
	
	fmt.Printf("✅ File uploaded successfully: %s\n", publicURL)
	
	respondJSON(w, http.StatusOK, map[string]string{
		"url": publicURL,
	})
}
