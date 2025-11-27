package facebook

import (
	"encoding/json"
	"fmt"
	"io"
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
func (c *Client) PostToPage(pageID, accessToken, message string, imageURLs []string) (string, error) {
	if len(imageURLs) == 0 {
		// Text-only post
		return c.postTextOnly(pageID, accessToken, message)
	} else if len(imageURLs) == 1 {
		// Single image post
		return c.postWithSingleImage(pageID, accessToken, message, imageURLs[0])
	} else {
		// Multiple images post
		return c.postWithMultipleImages(pageID, accessToken, message, imageURLs)
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

func (c *Client) postWithSingleImage(pageID, accessToken, message, imageURL string) (string, error) {
	data := url.Values{}
	data.Set("message", message)
	data.Set("url", imageURL)
	data.Set("access_token", accessToken)
	
	resp, err := c.httpClient.PostForm(fmt.Sprintf("%s/%s/photos", GraphAPIURL, pageID), data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	return c.parsePostResponse(resp)
}

func (c *Client) postWithMultipleImages(pageID, accessToken, message string, imageURLs []string) (string, error) {
	// Step 1: Upload images and get media IDs
	var mediaIDs []string
	for _, imageURL := range imageURLs {
		mediaID, err := c.uploadPhoto(pageID, accessToken, imageURL)
		if err != nil {
			return "", fmt.Errorf("failed to upload image: %w", err)
		}
		mediaIDs = append(mediaIDs, mediaID)
	}
	
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

func (c *Client) uploadPhoto(pageID, accessToken, imageURL string) (string, error) {
	data := url.Values{}
	data.Set("url", imageURL)
	data.Set("published", "false")
	data.Set("access_token", accessToken)
	
	resp, err := c.httpClient.PostForm(fmt.Sprintf("%s/%s/photos", GraphAPIURL, pageID), data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	var result struct {
		ID string `json:"id"`
	}
	
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse upload response: %s", string(body))
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
