package api

import (
	"encoding/json"
	"fbscheduler/internal/db"
	"log"
	"net/http"
	"os"
)

type CallbackRequest struct {
	Code string `json:"code"`
}

type DebugPagesRequest struct {
	AccessToken string `json:"access_token"`
}

func (h *Handler) GetFacebookAuthURL(w http.ResponseWriter, r *http.Request) {
	redirectURI := os.Getenv("FACEBOOK_REDIRECT_URI")
	authURL := h.fbClient.GetAuthURL(redirectURI)
	
	respondJSON(w, http.StatusOK, map[string]string{
		"url": authURL,
	})
}

func (h *Handler) FacebookCallback(w http.ResponseWriter, r *http.Request) {
	var req CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	log.Printf("📥 Received callback with code: %s...", req.Code[:20])
	
	// Exchange code for access token
	redirectURI := os.Getenv("FACEBOOK_REDIRECT_URI")
	userToken, err := h.fbClient.ExchangeCodeForToken(req.Code, redirectURI)
	if err != nil {
		log.Printf("❌ Token exchange failed: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to exchange token: "+err.Error())
		return
	}
	
	log.Printf("✅ Got user access token: %s...", userToken[:20])
	
	// Get user's pages
	pages, err := h.fbClient.GetUserPages(userToken)
	if err != nil {
		log.Printf("❌ Failed to fetch pages: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch pages: "+err.Error())
		return
	}
	
	// Log the number of pages received
	log.Printf("📊 Received %d pages from Facebook", len(pages))
	for i, p := range pages {
		log.Printf("  Page %d: ID=%s, Name=%s, Category=%s", i+1, p.ID, p.Name, p.Category)
	}
	
	// Convert to response format
	responsePages := make([]map[string]interface{}, 0, len(pages))
	for _, pageInfo := range pages {
		pageData := map[string]interface{}{
			"page_id":             pageInfo.ID,
			"page_name":           pageInfo.Name,
			"access_token":        pageInfo.AccessToken,
			"category":            pageInfo.Category,
			"profile_picture_url": pageInfo.Picture.Data.URL,
		}
		
		responsePages = append(responsePages, pageData)
	}
	
	// Update access tokens for existing pages (tokens may have changed)
	existingPages, _ := h.store.GetPages()
	existingPageMap := make(map[string]bool)
	for _, p := range existingPages {
		existingPageMap[p.PageID] = true
	}
	
	for _, pageInfo := range pages {
		// Only update if page already exists in DB
		if existingPageMap[pageInfo.ID] {
			page := &db.Page{
				PageID:            pageInfo.ID,
				PageName:          pageInfo.Name,
				AccessToken:       pageInfo.AccessToken,
				Category:          pageInfo.Category,
				ProfilePictureURL: pageInfo.Picture.Data.URL,
			}
			
			if err := h.store.CreateOrUpdatePage(page); err != nil {
				log.Printf("⚠️ Warning: Failed to update token for page %s: %v", page.PageName, err)
			} else {
				log.Printf("🔄 Updated access token for existing page: %s", page.PageName)
			}
		}
	}
	
	log.Printf("✅ Returned %d pages to frontend for selection", len(responsePages))
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Successfully fetched pages",
		"pages":   responsePages,
		"count":   len(responsePages),
	})
}

// DebugPages - endpoint để test xem Facebook trả về bao nhiêu pages
func (h *Handler) DebugPages(w http.ResponseWriter, r *http.Request) {
	var req DebugPagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	log.Printf("🔍 Debug: Fetching pages with token: %s...", req.AccessToken[:20])
	
	pages, err := h.fbClient.GetUserPages(req.AccessToken)
	if err != nil {
		log.Printf("❌ Debug: Failed to fetch pages: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch pages: "+err.Error())
		return
	}
	
	log.Printf("🔍 Debug: Facebook returned %d pages", len(pages))
	for i, p := range pages {
		log.Printf("  Page %d: ID=%s, Name=%s", i+1, p.ID, p.Name)
	}
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"count": len(pages),
		"pages": pages,
	})
}

// SaveSelectedPages - Lưu các pages đã được user chọn
func (h *Handler) SaveSelectedPages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pages []struct {
			PageID            string `json:"page_id"`
			PageName          string `json:"page_name"`
			AccessToken       string `json:"access_token"`
			Category          string `json:"category"`
			ProfilePictureURL string `json:"profile_picture_url"`
		} `json:"pages"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	log.Printf("💾 Saving %d selected pages", len(req.Pages))
	
	// Get all existing pages
	existingPages, err := h.store.GetPages()
	if err != nil {
		log.Printf("❌ Failed to get existing pages: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get existing pages")
		return
	}
	
	// Create map of selected page IDs
	selectedPageIDs := make(map[string]bool)
	for _, p := range req.Pages {
		selectedPageIDs[p.PageID] = true
	}
	
	// Delete pages that are not selected
	for _, existingPage := range existingPages {
		if !selectedPageIDs[existingPage.PageID] {
			log.Printf("🗑️ Deleting unselected page: %s (ID: %s)", existingPage.PageName, existingPage.PageID)
			if err := h.store.DeletePage(existingPage.ID); err != nil {
				log.Printf("⚠️ Warning: Failed to delete page %s: %v", existingPage.PageName, err)
			}
		}
	}
	
	// Save or update selected pages
	savedPages := make([]db.Page, 0, len(req.Pages))
	for _, pageData := range req.Pages {
		page := &db.Page{
			PageID:            pageData.PageID,
			PageName:          pageData.PageName,
			AccessToken:       pageData.AccessToken,
			Category:          pageData.Category,
			ProfilePictureURL: pageData.ProfilePictureURL,
		}
		
		if err := h.store.CreateOrUpdatePage(page); err != nil {
			log.Printf("❌ Failed to save page %s: %v", page.PageName, err)
			respondError(w, http.StatusInternalServerError, "Failed to save page: "+err.Error())
			return
		}
		
		log.Printf("✅ Saved page: %s (ID: %s)", page.PageName, page.PageID)
		savedPages = append(savedPages, *page)
	}
	
	log.Printf("✅ Successfully saved %d pages", len(savedPages))
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Successfully saved selected pages",
		"pages":   savedPages,
		"count":   len(savedPages),
	})
}
