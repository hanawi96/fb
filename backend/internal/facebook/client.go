package facebook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	GraphAPIURL = "https://graph.facebook.com/v18.0"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAuthURL generates Facebook OAuth URL
func (c *Client) GetAuthURL(redirectURI string) string {
	params := url.Values{}
	params.Add("client_id", os.Getenv("FACEBOOK_APP_ID"))
	params.Add("redirect_uri", redirectURI)
	params.Add("scope", "pages_show_list,pages_read_engagement,pages_manage_posts,pages_manage_metadata,business_management")
	params.Add("response_type", "code")
	params.Add("auth_type", "rerequest") // Force Facebook to show permission dialog again
	params.Add("display", "popup")
	
	authURL := fmt.Sprintf("https://www.facebook.com/v18.0/dialog/oauth?%s", params.Encode())
	fmt.Printf("🔗 Generated Auth URL: %s\n", authURL)
	
	return authURL
}

// ExchangeCodeForToken exchanges authorization code for access token
func (c *Client) ExchangeCodeForToken(code, redirectURI string) (string, error) {
	params := url.Values{}
	params.Add("client_id", os.Getenv("FACEBOOK_APP_ID"))
	params.Add("client_secret", os.Getenv("FACEBOOK_APP_SECRET"))
	params.Add("redirect_uri", redirectURI)
	params.Add("code", code)
	
	apiURL := fmt.Sprintf("%s/oauth/access_token?%s", GraphAPIURL, params.Encode())
	fmt.Printf("Exchanging code for token: %s\n", apiURL)
	
	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	fmt.Printf("Token exchange response: %s\n", string(body))
	
	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	
	if result.Error != nil {
		return "", fmt.Errorf("facebook token exchange error: %s (code: %d)", result.Error.Message, result.Error.Code)
	}
	
	if result.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}
	
	fmt.Printf("Successfully got access token: %s...\n", result.AccessToken[:20])
	
	return result.AccessToken, nil
}

// GetUserPages retrieves all pages managed by the user
func (c *Client) GetUserPages(userAccessToken string) ([]PageInfo, error) {
	allPages := []PageInfo{}
	url := fmt.Sprintf("%s/me/accounts?access_token=%s&fields=id,name,access_token,category,picture&limit=100", 
		GraphAPIURL, userAccessToken)
	
	for url != "" {
		resp, err := c.httpClient.Get(url)
		if err != nil {
			return nil, err
		}
		
		// Read response body for logging
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		
		fmt.Printf("Facebook API Response: %s\n", string(body))
		
		var result struct {
			Data []PageInfo `json:"data"`
			Paging *struct {
				Next string `json:"next"`
			} `json:"paging"`
			Error *struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		
		if result.Error != nil {
			return nil, fmt.Errorf("facebook API error: %s (code: %d)", result.Error.Message, result.Error.Code)
		}
		
		allPages = append(allPages, result.Data...)
		
		// Check if there's a next page
		if result.Paging != nil && result.Paging.Next != "" {
			url = result.Paging.Next
		} else {
			url = ""
		}
	}
	
	return allPages, nil
}

// PostToPage posts content to a Facebook page
func (c *Client) PostToPage(pageID, accessToken, message string, mediaURLs []string, mediaType string) (string, error) {
	if len(mediaURLs) == 0 {
		// Text-only post
		return c.postTextOnly(pageID, accessToken, message)
	}
	
	// Check if it's video
	if mediaType == "video" {
		if len(mediaURLs) > 1 {
			return "", fmt.Errorf("facebook only supports 1 video per post")
		}
		return c.postWithVideo(pageID, accessToken, message, mediaURLs[0])
	}
	
	// Image posts
	if len(mediaURLs) == 1 {
		return c.postWithSingleImage(pageID, accessToken, message, mediaURLs[0])
	} else {
		return c.postWithMultipleImages(pageID, accessToken, message, mediaURLs)
	}
}

func (c *Client) postTextOnly(pageID, accessToken, message string) (string, error) {
	data := url.Values{}
	data.Set("message", message)
	data.Set("access_token", accessToken)
	
	resp, err := c.httpClient.PostForm(fmt.Sprintf("%s/%s/feed", GraphAPIURL, pageID), data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	return c.parsePostResponse(resp)
}

// PostToPageWithData posts with preloaded media data (optimized for multiple pages)
func (c *Client) PostToPageWithData(pageID, accessToken, message string, mediaData [][]byte, mediaType string) (string, error) {
	if mediaType == "video" && len(mediaData) > 0 {
		return c.postVideoWithData(pageID, accessToken, message, mediaData[0])
	}
	// Fallback to URL-based upload
	return "", fmt.Errorf("preloaded data only supported for video")
}

func (c *Client) postVideoWithData(pageID, accessToken, message string, videoData []byte) (string, error) {
	fmt.Printf("📹 Uploading video to Facebook (%.2f MB)...\n", float64(len(videoData))/(1024*1024))
	
	// Upload to Facebook using multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add video file
	part, err := writer.CreateFormFile("source", "video.mp4")
	if err != nil {
		return "", err
	}
	part.Write(videoData)
	
	// Add other fields
	writer.WriteField("description", message)
	writer.WriteField("access_token", accessToken)
	writer.Close()
	
	// Post to Facebook
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/videos", GraphAPIURL, pageID), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	fbResp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer fbResp.Body.Close()
	
	return c.parsePostResponse(fbResp)
}

func (c *Client) postWithVideo(pageID, accessToken, message, videoURL string) (string, error) {
	fmt.Printf("📹 Downloading and uploading video...\n")
	
	// Download video from local URL
	resp, err := c.httpClient.Get(videoURL)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()
	
	// Read video data
	videoData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read video: %w", err)
	}
	
	return c.postVideoWithData(pageID, accessToken, message, videoData)
}

func (c *Client) postWithSingleImage(pageID, accessToken, message, imageURL string) (string, error) {
	// Download image from local URL
	resp, err := c.httpClient.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()
	
	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}
	
	// Upload to Facebook using multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add image file
	part, err := writer.CreateFormFile("source", "image.jpg")
	if err != nil {
		return "", err
	}
	part.Write(imageData)
	
	// Add other fields
	writer.WriteField("message", message)
	writer.WriteField("access_token", accessToken)
	writer.Close()
	
	// Post to Facebook
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/photos", GraphAPIURL, pageID), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	fbResp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer fbResp.Body.Close()
	
	return c.parsePostResponse(fbResp)
}

func (c *Client) postWithMultipleImages(pageID, accessToken, message string, imageURLs []string) (string, error) {
	fmt.Printf("⚡ Uploading %d images concurrently...\n", len(imageURLs))
	startTime := time.Now()
	
	// Step 1: Upload images concurrently
	type uploadResult struct {
		mediaID string
		index   int
		err     error
	}
	
	resultChan := make(chan uploadResult, len(imageURLs))
	
	// Upload all images in parallel
	for i, imageURL := range imageURLs {
		go func(idx int, url string) {
			mediaID, err := c.uploadPhoto(pageID, accessToken, url, "")
			resultChan <- uploadResult{mediaID: mediaID, index: idx, err: err}
		}(i, imageURL)
	}
	
	// Collect results in order
	results := make([]uploadResult, len(imageURLs))
	for i := 0; i < len(imageURLs); i++ {
		result := <-resultChan
		results[result.index] = result
	}
	
	// Check for errors and build mediaIDs array
	var mediaIDs []string
	for i, result := range results {
		if result.err != nil {
			return "", fmt.Errorf("failed to upload image %d: %w", i+1, result.err)
		}
		mediaIDs = append(mediaIDs, result.mediaID)
	}
	
	uploadDuration := time.Since(startTime)
	fmt.Printf("✅ Uploaded %d images in %.2f seconds (%.2f seconds per image)\n", 
		len(imageURLs), uploadDuration.Seconds(), uploadDuration.Seconds()/float64(len(imageURLs)))
	
	// Step 2: Create post with multiple images
	data := url.Values{}
	data.Set("message", message)
	for _, mediaID := range mediaIDs {
		data.Add("attached_media[]", fmt.Sprintf(`{"media_fbid":"%s"}`, mediaID))
	}
	data.Set("access_token", accessToken)
	
	resp, err := c.httpClient.PostForm(fmt.Sprintf("%s/%s/feed", GraphAPIURL, pageID), data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	return c.parsePostResponse(resp)
}

// ImageWithCaption represents an image with optional caption
type ImageWithCaption struct {
	URL     string `json:"url"`
	Caption string `json:"caption,omitempty"`
}

// PostAlbumWithCaptions creates an album and uploads photos with individual captions
func (c *Client) PostAlbumWithCaptions(pageID, accessToken, albumName, albumMessage string, images []ImageWithCaption) (string, error) {
	fmt.Printf("📸 Creating album with %d photos (each with caption)...\n", len(images))
	
	// Step 1: Create album
	albumID, err := c.createAlbum(pageID, accessToken, albumName, albumMessage)
	if err != nil {
		return "", fmt.Errorf("failed to create album: %w", err)
	}
	fmt.Printf("✅ Album created: %s\n", albumID)
	
	// Step 2: Upload photos concurrently
	type uploadResult struct {
		photoID string
		index   int
		err     error
	}
	
	resultChan := make(chan uploadResult, len(images))
	
	for i, img := range images {
		go func(idx int, image ImageWithCaption) {
			photoID, err := c.uploadPhotoToAlbum(albumID, accessToken, image.URL, image.Caption)
			resultChan <- uploadResult{photoID: photoID, index: idx, err: err}
		}(i, img)
	}
	
	// Collect results
	results := make([]uploadResult, len(images))
	for i := 0; i < len(images); i++ {
		result := <-resultChan
		results[result.index] = result
	}
	
	// Check for errors
	for i, result := range results {
		if result.err != nil {
			return "", fmt.Errorf("failed to upload photo %d: %w", i+1, result.err)
		}
		fmt.Printf("✅ Photo %d uploaded with caption\n", i+1)
	}
	
	fmt.Printf("✅ Album posted successfully: %s\n", albumID)
	return albumID, nil
}

func (c *Client) createAlbum(pageID, accessToken, name, message string) (string, error) {
	data := url.Values{}
	data.Set("name", name)
	data.Set("message", message)
	data.Set("access_token", accessToken)
	
	resp, err := c.httpClient.PostForm(fmt.Sprintf("%s/%s/albums", GraphAPIURL, pageID), data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	bodyBytes, _ := io.ReadAll(resp.Body)
	
	var result struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse album response: %s", string(bodyBytes))
	}
	
	if result.Error != nil {
		return "", fmt.Errorf("facebook error: %s (code: %d)", result.Error.Message, result.Error.Code)
	}
	
	return result.ID, nil
}

func (c *Client) uploadPhotoToAlbum(albumID, accessToken, imageURL, caption string) (string, error) {
	// Download image
	resp, err := c.httpClient.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()
	
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}
	
	// Upload to album
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("source", "image.jpg")
	if err != nil {
		return "", err
	}
	part.Write(imageData)
	
	writer.WriteField("message", caption)
	writer.WriteField("published", "true")
	writer.WriteField("access_token", accessToken)
	writer.Close()
	
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/photos", GraphAPIURL, albumID), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	fbResp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer fbResp.Body.Close()
	
	bodyBytes, _ := io.ReadAll(fbResp.Body)
	
	var result struct {
		ID    string `json:"id"`
		PostID string `json:"post_id"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse photo response: %s", string(bodyBytes))
	}
	
	if result.Error != nil {
		return "", fmt.Errorf("facebook error: %s", result.Error.Message)
	}
	
	if result.PostID != "" {
		return result.PostID, nil
	}
	return result.ID, nil
}

// PostToPageWithCaptions posts multiple images with individual captions
func (c *Client) PostToPageWithCaptions(pageID, accessToken, message string, images []ImageWithCaption, mediaType string) (string, error) {
	if len(images) == 0 {
		return c.postTextOnly(pageID, accessToken, message)
	}
	
	if len(images) == 1 {
		return c.postWithSingleImage(pageID, accessToken, message, images[0].URL)
	}
	
	fmt.Printf("⚡ Uploading %d images with captions concurrently...\n", len(images))
	startTime := time.Now()
	
	// Step 1: Upload images concurrently with captions
	type uploadResult struct {
		mediaID string
		caption string
		index   int
		err     error
	}
	
	resultChan := make(chan uploadResult, len(images))
	
	// Upload all images in parallel
	for i, img := range images {
		go func(idx int, image ImageWithCaption) {
			mediaID, err := c.uploadPhoto(pageID, accessToken, image.URL, image.Caption)
			resultChan <- uploadResult{
				mediaID: mediaID,
				caption: image.Caption,
				index:   idx,
				err:     err,
			}
		}(i, img)
	}
	
	// Collect results in order
	results := make([]uploadResult, len(images))
	for i := 0; i < len(images); i++ {
		result := <-resultChan
		results[result.index] = result
	}
	
	// Check for errors and build attached_media array
	// NOTE: Facebook API does NOT support "description" field in attached_media[]
	// Captions can only be used in Individual mode (separate posts)
	var attachedMedia []string
	for i, result := range results {
		if result.err != nil {
			return "", fmt.Errorf("failed to upload image %d: %w", i+1, result.err)
		}
		
		// Build attached_media JSON - only media_fbid is supported
		mediaJSON := map[string]string{
			"media_fbid": result.mediaID,
		}
		jsonBytes, err := json.Marshal(mediaJSON)
		if err != nil {
			return "", fmt.Errorf("failed to marshal media JSON: %w", err)
		}
		attachedMedia = append(attachedMedia, string(jsonBytes))
	}
	
	uploadDuration := time.Since(startTime)
	fmt.Printf("✅ Uploaded %d images with captions in %.2f seconds\n", 
		len(images), uploadDuration.Seconds())
	
	// Step 2: Create post with multiple images and captions
	data := url.Values{}
	data.Set("message", message)
	
	fmt.Printf("📝 Building attached_media array:\n")
	for i, media := range attachedMedia {
		fmt.Printf("   [%d] %s\n", i, media)
		data.Add("attached_media[]", media)
	}
	data.Set("access_token", accessToken)
	
	fmt.Printf("🚀 Posting to Facebook API: %s/%s/feed\n", GraphAPIURL, pageID)
	fmt.Printf("   Message: %s\n", message)
	fmt.Printf("   Attached media count: %d\n", len(attachedMedia))
	fmt.Printf("⚠️  NOTE: Facebook does NOT support individual captions in album mode\n")
	fmt.Printf("   Use Individual mode if you need different captions per image\n")
	
	resp, err := c.httpClient.PostForm(fmt.Sprintf("%s/%s/feed", GraphAPIURL, pageID), data)
	if err != nil {
		fmt.Printf("❌ Facebook API request failed: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()
	
	// Read response for debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("📥 Facebook API Response (status %d): %s\n", resp.StatusCode, string(bodyBytes))
	
	// Parse response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("facebook API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}
	
	var result struct {
		ID    string `json:"id"`
		PostID string `json:"post_id"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	if result.Error != nil {
		return "", fmt.Errorf("facebook API error: %s (code: %d)", result.Error.Message, result.Error.Code)
	}
	
	if result.PostID != "" {
		fmt.Printf("✅ Post created successfully: %s\n", result.PostID)
		return result.PostID, nil
	}
	fmt.Printf("✅ Post created successfully: %s\n", result.ID)
	return result.ID, nil
}

func (c *Client) uploadPhoto(pageID, accessToken, imageURL, caption string) (string, error) {
	// Download image from local URL
	resp, err := c.httpClient.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()
	
	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}
	
	// Upload to Facebook using multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add image file
	part, err := writer.CreateFormFile("source", "image.jpg")
	if err != nil {
		return "", err
	}
	part.Write(imageData)
	
	// Add other fields
	writer.WriteField("published", "false")
	writer.WriteField("access_token", accessToken)
	
	// NOTE: Caption is NOT used here because:
	// - For unpublished photos, caption is ignored by Facebook
	// - For album mode, only the post message is shown
	// - For individual mode, we post each image separately with its own message
	
	writer.Close()
	
	// Post to Facebook
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/photos", GraphAPIURL, pageID), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	fbResp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer fbResp.Body.Close()
	
	var result struct {
		ID string `json:"id"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	
	bodyBytes, _ := io.ReadAll(fbResp.Body)
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse upload response: %s", string(bodyBytes))
	}
	
	if result.Error != nil {
		return "", fmt.Errorf("facebook upload error: %s", result.Error.Message)
	}
	
	return result.ID, nil
}

func (c *Client) parsePostResponse(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("facebook API error: %s", string(body))
	}
	
	var result struct {
		ID    string `json:"id"`
		PostID string `json:"post_id"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	
	if result.PostID != "" {
		return result.PostID, nil
	}
	return result.ID, nil
}

// PageInfo represents a Facebook page
type PageInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
	Category    string `json:"category"`
	Picture     struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}
