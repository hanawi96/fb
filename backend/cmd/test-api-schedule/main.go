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
	fmt.Println("TEST API SCHEDULE - FULL FLOW")
	fmt.Println("==============================================\n")

	// 1. Lấy pages
	pages, err := store.GetPages()
	if err != nil || len(pages) == 0 {
		log.Fatal("No pages found")
	}

	// Lấy 2 pages đầu
	pageIDs := []string{pages[0].ID}
	if len(pages) > 1 {
		pageIDs = append(pageIDs, pages[1].ID)
	}

	fmt.Printf("📄 Using %d pages:\n", len(pageIDs))
	for _, pageID := range pageIDs {
		page, _ := store.GetPageByID(pageID)
		if page != nil {
			fmt.Printf("   - %s\n", page.PageName)
			
			// Kiểm tra account
			account, _ := store.GetPrimaryAccountForPage(pageID)
			if account != nil {
				fmt.Printf("     Account: %s ✅\n", account.FbUserName)
			} else {
				fmt.Printf("     Account: NULL ❌\n")
			}
		}
	}

	// 2. Tạo post
	fmt.Println("\n📝 Creating post...")
	post := &db.Post{
		Content: fmt.Sprintf("Test API schedule - %s", time.Now().Format("15:04:05")),
		Status:  "draft",
	}

	if err := store.CreatePost(post); err != nil {
		log.Fatal("Failed to create post:", err)
	}
	fmt.Printf("✅ Created post: %s\n", post.ID)

	// 3. Call API schedule
	fmt.Println("\n🚀 Calling API /api/schedule...")
	scheduledTime := time.Now().Add(2 * time.Hour)

	reqBody := map[string]interface{}{
		"post_id":        post.ID,
		"page_ids":       pageIDs,
		"scheduled_time": scheduledTime.Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post("http://localhost:8080/api/schedule", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("API call failed:", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   Status: %d\n", resp.StatusCode)
	fmt.Printf("   Response: %s\n", string(body))

	if resp.StatusCode != 201 {
		log.Fatal("API returned error")
	}

	// Parse response
	var apiResp struct {
		Message   string `json:"message"`
		Scheduled []struct {
			ID        string `json:"id"`
			PostID    string `json:"post_id"`
			PageID    string `json:"page_id"`
			AccountID *string `json:"account_id"`
		} `json:"scheduled"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Fatal("Failed to parse response:", err)
	}

	fmt.Printf("\n✅ API created %d scheduled posts\n", len(apiResp.Scheduled))

	// 4. Kiểm tra database
	fmt.Println("\n🔍 Verifying database...")
	time.Sleep(500 * time.Millisecond) // Wait a bit

	for i, scheduled := range apiResp.Scheduled {
		fmt.Printf("\n--- Scheduled Post #%d ---\n", i+1)
		fmt.Printf("   ID: %s\n", scheduled.ID)

		// Query trực tiếp từ database
		query := `
			SELECT 
				sp.id, sp.post_id, sp.page_id, sp.account_id,
				pg.page_name,
				fa.fb_user_name, fa.profile_picture_url
			FROM scheduled_posts sp
			JOIN pages pg ON pg.id = sp.page_id
			LEFT JOIN facebook_accounts fa ON fa.id = sp.account_id
			WHERE sp.id = $1
		`

		var id, postID, pageID string
		var accountID *string
		var pageName string
		var accountName, accountPicture *string

		err := database.QueryRow(query, scheduled.ID).Scan(
			&id, &postID, &pageID, &accountID,
			&pageName,
			&accountName, &accountPicture,
		)

		if err != nil {
			fmt.Printf("   ❌ Error querying: %v\n", err)
			continue
		}

		fmt.Printf("   Page: %s\n", pageName)

		if accountID != nil && *accountID != "" {
			fmt.Printf("   ✅ Account ID in DB: %s\n", *accountID)
			if accountName != nil {
				fmt.Printf("   ✅ Account Name: %s\n", *accountName)
			}
			if accountPicture != nil && *accountPicture != "" {
				fmt.Printf("   ✅ Avatar URL: %s\n", *accountPicture)
			}
		} else {
			fmt.Printf("   ❌ Account ID: NULL\n")
			fmt.Printf("   ⚠️  PROBLEM: account_id not saved to database!\n")
		}

		// So sánh với API response
		if scheduled.AccountID != nil && *scheduled.AccountID != "" {
			fmt.Printf("   API Response Account ID: %s\n", *scheduled.AccountID)
			if accountID == nil || *accountID != *scheduled.AccountID {
				fmt.Printf("   ❌ MISMATCH: API says %s but DB has %v\n", *scheduled.AccountID, accountID)
			}
		} else {
			fmt.Printf("   ⚠️  API Response Account ID: NULL\n")
		}
	}

	// 5. Test GetScheduledPosts API
	fmt.Println("\n🔍 Testing GET /api/schedule...")
	resp2, err := http.Get("http://localhost:8080/api/schedule?limit=5")
	if err != nil {
		log.Fatal("GET API failed:", err)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	
	var getPosts []struct {
		ID      string `json:"id"`
		PostID  string `json:"post_id"`
		Account *struct {
			ID              string `json:"id"`
			FbUserName      string `json:"fb_user_name"`
			ProfilePictureURL string `json:"profile_picture_url"`
		} `json:"account"`
	}

	if err := json.Unmarshal(body2, &getPosts); err != nil {
		log.Fatal("Failed to parse GET response:", err)
	}

	fmt.Printf("   Found %d posts from GET API\n", len(getPosts))

	// Tìm post vừa tạo
	for _, p := range getPosts {
		if p.PostID == post.ID {
			fmt.Printf("\n   ✅ Found our post in GET response:\n")
			fmt.Printf("      ID: %s\n", p.ID)
			if p.Account != nil {
				fmt.Printf("      ✅ Account: %s\n", p.Account.FbUserName)
				if p.Account.ProfilePictureURL != "" {
					fmt.Printf("      ✅ Avatar: %s\n", p.Account.ProfilePictureURL)
				}
			} else {
				fmt.Printf("      ❌ Account: NULL in GET response\n")
			}
		}
	}

	fmt.Println("\n==============================================")
	fmt.Println("SUMMARY")
	fmt.Println("==============================================")
	
	allGood := true
	for _, scheduled := range apiResp.Scheduled {
		query := `SELECT account_id FROM scheduled_posts WHERE id = $1`
		var accountID *string
		database.QueryRow(query, scheduled.ID).Scan(&accountID)
		
		if accountID == nil || *accountID == "" {
			allGood = false
			fmt.Printf("❌ Post %s: account_id is NULL\n", scheduled.ID[:8])
		} else {
			fmt.Printf("✅ Post %s: account_id = %s\n", scheduled.ID[:8], *accountID)
		}
	}

	if allGood {
		fmt.Println("\n✅ ALL TESTS PASSED!")
	} else {
		fmt.Println("\n❌ TESTS FAILED - account_id not properly saved")
	}
}
