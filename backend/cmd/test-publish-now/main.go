package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"fbscheduler/internal/db"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Connect DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	database, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	store := db.NewStore(database)

	fmt.Println("==============================================")
	fmt.Println("TEST PUBLISH NOW - SCHEDULE DISPLAY")
	fmt.Println("==============================================\n")

	// Lấy pages
	pages, err := store.GetPages()
	if err != nil || len(pages) == 0 {
		log.Fatal("No pages found")
	}

	pageIDs := []string{pages[0].ID}
	fmt.Printf("📄 Using page: %s\n", pages[0].PageName)

	// Check account
	account, _ := store.GetPrimaryAccountForPage(pageIDs[0])
	if account != nil {
		fmt.Printf("👤 Account: %s\n", account.FbUserName)
	}

	// Call API publish (đăng ngay)
	fmt.Println("\n🚀 Calling API /api/posts/publish (đăng ngay)...")

	reqBody := map[string]interface{}{
		"content":   fmt.Sprintf("Test publish now - %s", time.Now().Format("15:04:05")),
		"media_urls": []string{},
		"media_type": "",
		"page_ids":   pageIDs,
		"privacy":    "public",
		"post_mode":  "album",
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post("http://localhost:8080/api/posts/publish", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("API call failed:", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   Status: %d\n", resp.StatusCode)
	fmt.Printf("   Response: %s\n", string(body))

	if resp.StatusCode != 200 && resp.StatusCode != 207 {
		log.Fatal("API returned error")
	}

	// Parse response
	var apiResp struct {
		PostID  string `json:"post_id"`
		Message string `json:"message"`
		Results []struct {
			PageID   string `json:"page_id"`
			PageName string `json:"page_name"`
			Status   string `json:"status"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Fatal("Failed to parse response:", err)
	}

	fmt.Printf("\n✅ Post ID: %s\n", apiResp.PostID)

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Kiểm tra scheduled_posts
	fmt.Println("\n🔍 Checking scheduled_posts table...")
	query := `
		SELECT 
			sp.id, sp.status, sp.scheduled_time,
			pg.page_name,
			fa.fb_user_name
		FROM scheduled_posts sp
		JOIN pages pg ON pg.id = sp.page_id
		LEFT JOIN facebook_accounts fa ON fa.id = sp.account_id
		WHERE sp.post_id = $1
	`

	rows, err := database.Query(query, apiResp.PostID)
	if err != nil {
		log.Fatal("Query error:", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var id, status string
		var scheduledTime time.Time
		var pageName string
		var accountName *string

		if err := rows.Scan(&id, &status, &scheduledTime, &pageName, &accountName); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		fmt.Printf("\n✅ Found in scheduled_posts:\n")
		fmt.Printf("   ID: %s\n", id)
		fmt.Printf("   Page: %s\n", pageName)
		fmt.Printf("   Status: %s\n", status)
		fmt.Printf("   Time: %s\n", scheduledTime.Format("2006-01-02 15:04:05"))
		if accountName != nil {
			fmt.Printf("   Account: %s ✅\n", *accountName)
		} else {
			fmt.Printf("   Account: NULL ❌\n")
		}
	}

	if !found {
		fmt.Println("\n❌ NOT FOUND in scheduled_posts!")
		fmt.Println("   This means the post won't show in schedule page")
	}

	// Test GET /api/schedule
	fmt.Println("\n🔍 Testing GET /api/schedule...")
	resp2, err := http.Get("http://localhost:8080/api/schedule?limit=10")
	if err != nil {
		log.Fatal("GET API failed:", err)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)

	var getPosts []struct {
		ID     string `json:"id"`
		PostID string `json:"post_id"`
		Status string `json:"status"`
		Page   *struct {
			PageName string `json:"page_name"`
		} `json:"page"`
		Account *struct {
			FbUserName string `json:"fb_user_name"`
		} `json:"account"`
	}

	if err := json.Unmarshal(body2, &getPosts); err != nil {
		log.Fatal("Failed to parse GET response:", err)
	}

	foundInAPI := false
	for _, p := range getPosts {
		if p.PostID == apiResp.PostID {
			foundInAPI = true
			fmt.Printf("\n✅ Found in GET /api/schedule:\n")
			fmt.Printf("   ID: %s\n", p.ID)
			fmt.Printf("   Status: %s\n", p.Status)
			if p.Page != nil {
				fmt.Printf("   Page: %s\n", p.Page.PageName)
			}
			if p.Account != nil {
				fmt.Printf("   Account: %s ✅\n", p.Account.FbUserName)
			} else {
				fmt.Printf("   Account: NULL ❌\n")
			}
			break
		}
	}

	if !foundInAPI {
		fmt.Println("\n❌ NOT FOUND in GET /api/schedule response!")
	}

	fmt.Println("\n==============================================")
	if found && foundInAPI {
		fmt.Println("✅ TEST PASSED!")
		fmt.Println("   Post published now appears in schedule page")
	} else {
		fmt.Println("❌ TEST FAILED!")
		if !found {
			fmt.Println("   - Not in scheduled_posts table")
		}
		if !foundInAPI {
			fmt.Println("   - Not in API response")
		}
	}
	fmt.Println("==============================================")
}
