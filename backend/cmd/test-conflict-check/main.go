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
	fmt.Println("TEST CONFLICT CHECK API")
	fmt.Println("==============================================\n")

	// Lấy pages
	pages, err := store.GetPages()
	if err != nil || len(pages) == 0 {
		log.Fatal("No pages found")
	}

	// Lấy 2 pages
	pageIDs := []string{pages[0].ID}
	if len(pages) > 1 {
		pageIDs = append(pageIDs, pages[1].ID)
	}

	fmt.Printf("📄 Testing with %d pages:\n", len(pageIDs))
	for _, pageID := range pageIDs {
		page, _ := store.GetPageByID(pageID)
		if page != nil {
			fmt.Printf("   - %s\n", page.PageName)
		}
	}

	// Test 1: Check thời gian không có xung đột
	fmt.Println("\n🧪 Test 1: Thời gian không xung đột")
	futureTime := time.Now().Add(5 * time.Hour)
	testConflictCheck(pageIDs, futureTime, "Không xung đột")

	// Test 2: Tạo scheduled post và check lại
	fmt.Println("\n🧪 Test 2: Tạo bài và check xung đột")
	
	// Tạo post
	post := &db.Post{
		Content: "Test conflict check",
		Status:  "draft",
	}
	if err := store.CreatePost(post); err != nil {
		log.Fatal("Failed to create post:", err)
	}
	fmt.Printf("✅ Created post: %s\n", post.ID)

	// Tạo scheduled post cho page đầu tiên
	testTime := time.Now().Add(3 * time.Hour)
	account, _ := store.GetPrimaryAccountForPage(pageIDs[0])
	
	sp := &db.ScheduledPost{
		PostID:        post.ID,
		PageID:        pageIDs[0],
		ScheduledTime: testTime,
		Status:        "pending",
		MaxRetries:    3,
	}
	if account != nil {
		sp.AccountID = &account.ID
	}

	if err := store.CreateScheduledPost(sp); err != nil {
		log.Fatal("Failed to create scheduled post:", err)
	}
	fmt.Printf("✅ Created scheduled post at: %s\n", testTime.Format("2006-01-02 15:04:05"))

	// Check conflict
	testConflictCheck(pageIDs, testTime, "Có xung đột")

	// Test 3: Check với thời gian khác 1 phút
	fmt.Println("\n🧪 Test 3: Thời gian khác 1 phút")
	differentTime := testTime.Add(1 * time.Minute)
	testConflictCheck(pageIDs, differentTime, "Không xung đột")

	fmt.Println("\n✅ All tests completed!")
}

func testConflictCheck(pageIDs []string, scheduledTime time.Time, expected string) {
	reqBody := map[string]interface{}{
		"page_ids":       pageIDs,
		"scheduled_time": scheduledTime.Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post("http://localhost:8080/api/schedule/check-conflict", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ API call failed: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("❌ Status: %d, Response: %s", resp.StatusCode, string(body))
		return
	}

	var result struct {
		HasConflict     bool `json:"has_conflict"`
		ConflictPages   []struct {
			PageID   string `json:"page_id"`
			PageName string `json:"page_name"`
		} `json:"conflict_pages"`
		NoConflictPages []struct {
			PageID   string `json:"page_id"`
			PageName string `json:"page_name"`
		} `json:"no_conflict_pages"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ Parse error: %v", err)
		return
	}

	fmt.Printf("   Time: %s\n", scheduledTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Expected: %s\n", expected)
	fmt.Printf("   Has Conflict: %v\n", result.HasConflict)
	
	if len(result.ConflictPages) > 0 {
		fmt.Printf("   Conflict Pages:\n")
		for _, p := range result.ConflictPages {
			fmt.Printf("      - %s\n", p.PageName)
		}
	}
	
	if len(result.NoConflictPages) > 0 {
		fmt.Printf("   No Conflict Pages:\n")
		for _, p := range result.NoConflictPages {
			fmt.Printf("      - %s\n", p.PageName)
		}
	}

	// Verify
	if expected == "Có xung đột" && result.HasConflict {
		fmt.Println("   ✅ PASSED")
	} else if expected == "Không xung đột" && !result.HasConflict {
		fmt.Println("   ✅ PASSED")
	} else {
		fmt.Println("   ❌ FAILED")
	}
}
