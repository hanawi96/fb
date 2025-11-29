package api

import (
	"encoding/json"
	"fbscheduler/internal/db"
	"fmt"
	"io"
	"net/http"
	"time"

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

// PublishPost publishes a post immediately to selected pages
func (h *Handler) PublishPost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content   string   `json:"content"`
		MediaURLs []string `json:"media_urls"`
		MediaType string   `json:"media_type"`
		PageIDs   []string `json:"page_ids"`
		Privacy   string   `json:"privacy"`
		PostMode  string   `json:"post_mode"` // "album" | "individual"
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("❌ PublishPost: Failed to decode request body: %v\n", err)
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	fmt.Printf("📤 PublishPost request:\n")
	fmt.Printf("   Content: %s\n", req.Content)
	fmt.Printf("   MediaURLs: %v\n", req.MediaURLs)
	fmt.Printf("   MediaType: %s\n", req.MediaType)
	fmt.Printf("   PageIDs: %v\n", req.PageIDs)
	fmt.Printf("   Privacy: %s\n", req.Privacy)
	fmt.Printf("   PostMode: %s\n", req.PostMode)
	
	// Validate
	if req.Content == "" && len(req.MediaURLs) == 0 {
		fmt.Printf("❌ PublishPost: No content or media\n")
		respondError(w, http.StatusBadRequest, "Content or media is required")
		return
	}
	
	if len(req.PageIDs) == 0 {
		fmt.Printf("❌ PublishPost: No pages selected\n")
		respondError(w, http.StatusBadRequest, "At least one page is required")
		return
	}
	
	// Create post record
	post := &db.Post{
		Content:   req.Content,
		MediaURLs: req.MediaURLs,
		MediaType: req.MediaType,
		Status:    "published",
	}
	
	fmt.Printf("💾 Creating post record...\n")
	if err := h.store.CreatePost(post); err != nil {
		fmt.Printf("❌ Failed to create post: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}
	fmt.Printf("✅ Post created with ID: %s\n", post.ID)
	
	// Publish to each page concurrently
	fmt.Printf("⚡ Publishing to %d pages concurrently...\n", len(req.PageIDs))
	
	type publishResult struct {
		pageID       string
		pageName     string
		fbPostID     string
		err          error
		index        int
	}
	
	resultChan := make(chan publishResult, len(req.PageIDs))
	
	// Pre-download media if needed (for video/images)
	var mediaData [][]byte
	if len(req.MediaURLs) > 0 && req.MediaType == "video" {
		fmt.Printf("📥 Pre-downloading video to reuse across pages...\n")
		for _, mediaURL := range req.MediaURLs {
			resp, err := http.Get(mediaURL)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to download media: "+err.Error())
				return
			}
			data, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to read media: "+err.Error())
				return
			}
			mediaData = append(mediaData, data)
			fmt.Printf("✅ Downloaded %.2f MB\n", float64(len(data))/(1024*1024))
		}
	}
	
	// Post to all pages in parallel
	for i, pageID := range req.PageIDs {
		go func(idx int, pgID string, preloadedMedia [][]byte) {
			// Get page info
			page, err := h.store.GetPageByID(pgID)
			if err != nil || page == nil {
				resultChan <- publishResult{
					pageID: pgID,
					err:    fmt.Errorf("page not found"),
					index:  idx,
				}
				return
			}
			
			// Handle individual mode (each image = separate post)
			if req.PostMode == "individual" && len(req.MediaURLs) > 1 {
				fmt.Printf("📸 Individual mode: Posting %d images separately to %s\n", len(req.MediaURLs), page.PageName)
				var fbPostIDs []string
				for imgIdx, mediaURL := range req.MediaURLs {
					var singleFbPostID string
					var singleErr error
					
					if len(preloadedMedia) > imgIdx {
						singleFbPostID, singleErr = h.fbClient.PostToPageWithDataAndPlace(
							page.PageID,
							page.AccessToken,
							req.Content,
							[][]byte{preloadedMedia[imgIdx]},
							req.MediaType,
							"",
						)
					} else {
						singleFbPostID, singleErr = h.fbClient.PostToPageWithPlace(
							page.PageID,
							page.AccessToken,
							req.Content,
							[]string{mediaURL},
							req.MediaType,
							"",
						)
					}
					
					if singleErr != nil {
						err = fmt.Errorf("failed to post image %d: %v", imgIdx+1, singleErr)
						break
					}
					fbPostIDs = append(fbPostIDs, singleFbPostID)
				}
				
				resultChan <- publishResult{
					pageID:   pgID,
					pageName: page.PageName,
					fbPostID: fmt.Sprintf("%d posts: %v", len(fbPostIDs), fbPostIDs),
					err:      err,
					index:    idx,
				}
				return
			}
			
			// Album mode (default): Post all images in one post
			var fbPostID string
			if len(preloadedMedia) > 0 {
				fbPostID, err = h.fbClient.PostToPageWithDataAndPlace(
					page.PageID,
					page.AccessToken,
					req.Content,
					preloadedMedia,
					req.MediaType,
					"",
				)
			} else {
				fbPostID, err = h.fbClient.PostToPageWithPlace(
					page.PageID,
					page.AccessToken,
					req.Content,
					req.MediaURLs,
					req.MediaType,
					"",
				)
			}
			
			resultChan <- publishResult{
				pageID:   pgID,
				pageName: page.PageName,
				fbPostID: fbPostID,
				err:      err,
				index:    idx,
			}
		}(i, pageID, mediaData)
	}
	
	// Collect results
	publishResults := make([]publishResult, len(req.PageIDs))
	for i := 0; i < len(req.PageIDs); i++ {
		result := <-resultChan
		publishResults[result.index] = result
	}
	
	// Process results and create logs
	results := make([]map[string]interface{}, 0)
	hasError := false
	
	for _, result := range publishResults {
		logEntry := &db.PostLog{
			PostID: post.ID,
			PageID: result.pageID,
		}
		
		if result.err != nil {
			fmt.Printf("❌ Failed to post to page %s: %v\n", result.pageName, result.err)
			logEntry.Status = "failed"
			logEntry.ErrorMessage = result.err.Error()
			h.store.CreatePostLog(logEntry)
			
			// Tạo scheduled_post với status failed để hiển thị trong lịch đăng
			account, _ := h.store.GetPrimaryAccountForPage(result.pageID)
			scheduledPost := &db.ScheduledPost{
				PostID:        post.ID,
				PageID:        result.pageID,
				ScheduledTime: time.Now(),
				Status:        "failed",
				MaxRetries:    0,
			}
			if account != nil {
				scheduledPost.AccountID = &account.ID
			}
			h.store.CreateScheduledPost(scheduledPost)
			
			results = append(results, map[string]interface{}{
				"page_id":   result.pageID,
				"page_name": result.pageName,
				"status":    "failed",
				"error":     result.err.Error(),
			})
			hasError = true
		} else {
			fmt.Printf("✅ Successfully posted to page %s: %s\n", result.pageName, result.fbPostID)
			logEntry.Status = "success"
			logEntry.FacebookPostID = result.fbPostID
			h.store.CreatePostLog(logEntry)
			
			// Tạo scheduled_post với status completed để hiển thị trong lịch đăng
			account, _ := h.store.GetPrimaryAccountForPage(result.pageID)
			scheduledPost := &db.ScheduledPost{
				PostID:        post.ID,
				PageID:        result.pageID,
				ScheduledTime: time.Now(),
				Status:        "completed",
				MaxRetries:    0,
			}
			if account != nil {
				scheduledPost.AccountID = &account.ID
			}
			h.store.CreateScheduledPost(scheduledPost)
			
			results = append(results, map[string]interface{}{
				"page_id":         result.pageID,
				"page_name":       result.pageName,
				"status":          "success",
				"facebook_post_id": result.fbPostID,
			})
		}
	}
	
	response := map[string]interface{}{
		"post_id": post.ID,
		"results": results,
	}
	
	if hasError {
		fmt.Printf("⚠️ Post published with some errors\n")
		response["message"] = "Post published with some errors"
		respondJSON(w, http.StatusPartialContent, response)
	} else {
		fmt.Printf("✅ Post published successfully to all pages\n")
		response["message"] = "Post published successfully to all pages"
		respondJSON(w, http.StatusOK, response)
	}
}
