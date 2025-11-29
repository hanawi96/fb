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
	
	// Get Facebook user info
	fbUser, err := h.fbClient.GetUserInfo(userToken)
	fbUserID := "unknown"
	fbUserName := "Unknown User"
	fbPictureURL := ""
	if err != nil {
		log.Printf("⚠️ Could not get user info: %v", err)
	} else {
		fbUserID = fbUser.ID
		fbUserName = fbUser.Name
		fbPictureURL = fbUser.PictureURL
	}
	
	// Create or update facebook_account
	account, err := h.store.GetAccountByFbUserID(fbUserID)
	if err != nil || account == nil {
		// Create new account
		account = &db.FacebookAccount{
			FbUserID:          fbUserID,
			FbUserName:        fbUserName,
			ProfilePictureURL: fbPictureURL,
			AccessToken:       userToken,
			MaxPages:          5,
			MaxPostsPerDay:    20,
			Status:            "active",
		}
		if err := h.store.CreateAccount(account); err != nil {
			log.Printf("⚠️ Failed to create account: %v", err)
		} else {
			log.Printf("✅ Created new Facebook account: %s (%s)", account.FbUserName, account.FbUserID)
		}
	} else {
		// Update existing account token
		account.AccessToken = userToken
		account.FbUserName = fbUserName
		account.ProfilePictureURL = fbPictureURL
		account.Status = "active"
		if err := h.store.UpdateAccount(account); err != nil {
			log.Printf("⚠️ Failed to update account: %v", err)
		} else {
			log.Printf("🔄 Updated Facebook account: %s", account.FbUserName)
		}
	}
	
	// Get user's pages
	pages, err := h.fbClient.GetUserPages(userToken)
	if err != nil {
		log.Printf("❌ Failed to fetch pages: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch pages: "+err.Error())
		return
	}
	
	log.Printf("📊 Received %d pages from Facebook", len(pages))
	
	// Convert to response format - include account_id and tasks
	responsePages := make([]map[string]interface{}, 0, len(pages))
	for _, pageInfo := range pages {
		// Check if page has CREATE_CONTENT permission
		hasCreateContent := false
		for _, task := range pageInfo.Tasks {
			if task == "CREATE_CONTENT" {
				hasCreateContent = true
				break
			}
		}
		
		pageData := map[string]interface{}{
			"page_id":             pageInfo.ID,
			"page_name":           pageInfo.Name,
			"access_token":        pageInfo.AccessToken,
			"category":            pageInfo.Category,
			"profile_picture_url": pageInfo.Picture.Data.URL,
			"account_id":          account.ID,
			"account_name":        account.FbUserName,
			"tasks":               pageInfo.Tasks,
			"can_post":            hasCreateContent,
		}
		responsePages = append(responsePages, pageData)
	}
	
	// Update access tokens for existing pages
	existingPages, _ := h.store.GetPages()
	existingPageMap := make(map[string]bool)
	for _, p := range existingPages {
		existingPageMap[p.PageID] = true
	}
	
	for _, pageInfo := range pages {
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
			}
		}
	}
	
	log.Printf("✅ Returned %d pages to frontend for selection", len(responsePages))
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Successfully fetched pages",
		"pages":        responsePages,
		"count":        len(responsePages),
		"account_id":   account.ID,
		"account_name": account.FbUserName,
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

// SaveSelectedPages - Lưu các pages đã được user chọn và gán vào account
func (h *Handler) SaveSelectedPages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID         string   `json:"account_id"`
		UnselectedPageIDs []string `json:"unselected_page_ids"`
		Pages             []struct {
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

	log.Printf("💾 Saving %d selected pages for account %s", len(req.Pages), req.AccountID)

	// Xóa các pages bị bỏ chọn
	if len(req.UnselectedPageIDs) > 0 {
		log.Printf("🗑️ Removing %d unselected pages", len(req.UnselectedPageIDs))
		for _, pageID := range req.UnselectedPageIDs {
			if err := h.store.DeletePageByPageID(pageID); err != nil {
				log.Printf("⚠️ Failed to delete page %s: %v", pageID, err)
			} else {
				log.Printf("🗑️ Deleted page: %s", pageID)
			}
		}
	}

	// Save or update selected pages and assign to account
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

		// Assign page to account if account_id provided
		if req.AccountID != "" && page.ID != "" {
			if err := h.store.AssignPageToAccount(page.ID, req.AccountID, true); err != nil {
				log.Printf("⚠️ Failed to assign page %s to account: %v", page.PageName, err)
			} else {
				log.Printf("🔗 Assigned page %s to account %s", page.PageName, req.AccountID)
			}
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
