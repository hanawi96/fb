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

	"fbscheduler/internal/config"
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
	fmt.Println("TEST SCHEDULE TIMEZONE")
	fmt.Println("==============================================\n")

	// Lấy pages
	pages, err := store.GetPages()
	if err != nil || len(pages) == 0 {
		log.Fatal("No pages found")
	}

	pageIDs := []string{pages[0].ID}
	fmt.Printf("📄 Using page: %s\n", pages[0].PageName)

	// Tạo post
	post := &db.Post{
		Content: fmt.Sprintf("Test timezone - %s", time.Now().Format("15:04:05")),
		Status:  "draft",
	}

	if err := store.CreatePost(post); err != nil {
		log.Fatal("Failed to create post:", err)
	}
	fmt.Printf("✅ Created post: %s\n", post.ID)

	// Test case: Schedule vào 2 giờ sau (Vietnam time)
	vietnamLoc := time.FixedZone("Asia/Ho_Chi_Minh", 7*3600)
	nowVN := time.Now().In(vietnamLoc)
	scheduledTimeVN := nowVN.Add(2 * time.Hour)
	
	fmt.Println("\n📅 Test Case:")
	fmt.Printf("   User chọn: %s (Vietnam time)\n", scheduledTimeVN.Format("2006-01-02 15:04:05"))
	
	fmt.Printf("   Vietnam time: %s\n", scheduledTimeVN.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("   UTC time: %s\n", scheduledTimeVN.UTC().Format("2006-01-02 15:04:05 MST"))

	// Call API với RFC3339 format
	fmt.Println("\n🚀 Calling API /api/schedule...")
	
	reqBody := map[string]interface{}{
		"post_id":        post.ID,
		"page_ids":       pageIDs,
		"scheduled_time": scheduledTimeVN.Format(time.RFC3339),
	}

	fmt.Printf("   Request scheduled_time: %s\n", scheduledTimeVN.Format(time.RFC3339))

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post("http://localhost:8080/api/schedule", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("API call failed:", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   Status: %d\n", resp.StatusCode)

	if resp.StatusCode != 201 {
		fmt.Printf("   Response: %s\n", string(body))
		log.Fatal("API returned error")
	}

	// Parse response
	var apiResp struct {
		Message   string `json:"message"`
		Scheduled []struct {
			ID            string    `json:"id"`
			ScheduledTime time.Time `json:"scheduled_time"`
		} `json:"scheduled"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Fatal("Failed to parse response:", err)
	}

	if len(apiResp.Scheduled) == 0 {
		log.Fatal("No scheduled posts in response")
	}

	scheduledPost := apiResp.Scheduled[0]
	fmt.Printf("\n✅ API Response:\n")
	fmt.Printf("   Scheduled Post ID: %s\n", scheduledPost.ID)
	fmt.Printf("   Scheduled Time (from API): %s\n", scheduledPost.ScheduledTime.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("   Scheduled Time (UTC): %s\n", scheduledPost.ScheduledTime.UTC().Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("   Scheduled Time (VN): %s\n", config.ToVN(scheduledPost.ScheduledTime).Format("2006-01-02 15:04:05 MST"))

	// Kiểm tra database
	fmt.Println("\n🔍 Checking database...")
	time.Sleep(500 * time.Millisecond)

	query := `
		SELECT 
			id, scheduled_time, 
			scheduled_time AT TIME ZONE 'UTC' as utc_time,
			scheduled_time AT TIME ZONE 'Asia/Ho_Chi_Minh' as vn_time
		FROM scheduled_posts
		WHERE id = $1
	`

	var id string
	var scheduledTimeDB, utcTime, vnTime time.Time

	err = database.QueryRow(query, scheduledPost.ID).Scan(&id, &scheduledTimeDB, &utcTime, &vnTime)
	if err != nil {
		log.Fatal("Query error:", err)
	}

	fmt.Printf("\n📊 Database values:\n")
	fmt.Printf("   scheduled_time (raw): %s\n", scheduledTimeDB.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("   scheduled_time (UTC): %s\n", utcTime.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("   scheduled_time (VN): %s\n", vnTime.Format("2006-01-02 15:04:05 MST"))

	// So sánh
	fmt.Println("\n==============================================")
	fmt.Println("COMPARISON")
	fmt.Println("==============================================")

	fmt.Printf("\n1. User Input (VN):\n")
	fmt.Printf("   %s\n", scheduledTimeVN.Format("2006-01-02 15:04:05"))

	fmt.Printf("\n2. API Response (VN):\n")
	fmt.Printf("   %s\n", config.ToVN(scheduledPost.ScheduledTime).Format("2006-01-02 15:04:05"))

	fmt.Printf("\n3. Database (VN):\n")
	fmt.Printf("   %s\n", vnTime.Format("2006-01-02 15:04:05"))

	// Kiểm tra GET API
	fmt.Println("\n🔍 Testing GET /api/schedule...")
	resp2, err := http.Get("http://localhost:8080/api/schedule?limit=5")
	if err != nil {
		log.Fatal("GET API failed:", err)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)

	var getPosts []struct {
		ID            string    `json:"id"`
		ScheduledTime time.Time `json:"scheduled_time"`
	}

	if err := json.Unmarshal(body2, &getPosts); err != nil {
		log.Fatal("Failed to parse GET response:", err)
	}

	for _, p := range getPosts {
		if p.ID == scheduledPost.ID {
			fmt.Printf("\n4. GET API Response (VN):\n")
			fmt.Printf("   %s\n", config.ToVN(p.ScheduledTime).Format("2006-01-02 15:04:05"))
			break
		}
	}

	// Kết luận
	fmt.Println("\n==============================================")
	fmt.Println("RESULT")
	fmt.Println("==============================================")

	expectedVN := scheduledTimeVN.Format("2006-01-02 15:04")
	actualVN := vnTime.Format("2006-01-02 15:04")

	if expectedVN == actualVN {
		fmt.Println("✅ PASSED: Timezone is correct!")
		fmt.Printf("   Expected: %s\n", expectedVN)
		fmt.Printf("   Actual: %s\n", actualVN)
	} else {
		fmt.Println("❌ FAILED: Timezone mismatch!")
		fmt.Printf("   Expected: %s\n", expectedVN)
		fmt.Printf("   Actual: %s\n", actualVN)
		
		diff := vnTime.Sub(scheduledTimeVN)
		fmt.Printf("   Difference: %v\n", diff)
	}
}
