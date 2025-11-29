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

	_ "github.com/lib/pq"
)

const (
	baseURL = "http://localhost:8080"
)

type CreatePostRequest struct {
	Content string   `json:"content"`
	Status  string   `json:"status"`
	PageIDs []string `json:"page_ids"`
}

type ScheduleRequest struct {
	PostID        string   `json:"post_id"`
	PageIDs       []string `json:"page_ids"`
	PreferredDate string   `json:"preferred_date"`
	Confirm       bool     `json:"confirm"`
}

type ScheduleResponse struct {
	Message      string `json:"message"`
	Scheduled    bool   `json:"scheduled"`
	SuccessCount int    `json:"success_count"`
	Preview      struct {
		Results       []ScheduleResult `json:"Results"`
		TotalPages    int              `json:"TotalPages"`
		SuccessCount  int              `json:"SuccessCount"`
		WarningCount  int              `json:"WarningCount"`
		ErrorCount    int              `json:"ErrorCount"`
		NextDayCount  int              `json:"NextDayCount"`
	} `json:"preview"`
}

type ScheduleResult struct {
	PageID        string    `json:"PageID"`
	PageName      string    `json:"PageName"`
	AccountID     string    `json:"AccountID"`
	AccountName   string    `json:"AccountName"`
	TimeSlotID    string    `json:"TimeSlotID"`
	ScheduledTime time.Time `json:"ScheduledTime"`
	RandomOffset  int       `json:"RandomOffset"`
	Warning       string    `json:"Warning"`
	Error         *string   `json:"Error"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:yendev96@localhost:5432/fbscheduler?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║          API AUTO SCHEDULE END-TO-END TEST                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Lấy danh sách pages
	pageIDs, pageNames := getPages(db)
	if len(pageIDs) == 0 {
		log.Fatal("❌ Không có pages trong database")
	}

	fmt.Printf("📄 Tìm thấy %d pages:\n", len(pageIDs))
	for i, name := range pageNames {
		fmt.Printf("   %d. %s\n", i+1, name)
	}
	fmt.Println()

	// Test 1: Tạo bài và schedule cho 1 page
	fmt.Println("🧪 TEST 1: Schedule 1 bài cho 1 page")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testSinglePageSchedule(db, pageIDs[:1], pageNames[:1])
	fmt.Println()

	// Test 2: Tạo bài và schedule cho nhiều pages
	fmt.Println("🧪 TEST 2: Schedule 1 bài cho nhiều pages")
	fmt.Println("─────────────────────────────────────────────────────────────")
	numPages := 3
	if len(pageIDs) < 3 {
		numPages = len(pageIDs)
	}
	testMultiplePageSchedule(db, pageIDs[:numPages], pageNames[:numPages])
	fmt.Println()

	// Test 3: Schedule liên tiếp 5 bài
	fmt.Println("🧪 TEST 3: Schedule liên tiếp 5 bài cho 1 page")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testConsecutiveSchedule(db, pageIDs[:1], pageNames[:1], 5)
	fmt.Println()

	// Test 4: Kiểm tra database
	fmt.Println("🧪 TEST 4: Kiểm tra Database")
	fmt.Println("─────────────────────────────────────────────────────────────")
	checkDatabase(db)
	fmt.Println()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  ALL TESTS COMPLETED                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

func getPages(db *sql.DB) ([]string, []string) {
	rows, err := db.Query("SELECT id, page_name FROM pages ORDER BY page_name LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var ids, names []string
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		ids = append(ids, id)
		names = append(names, name)
	}
	return ids, names
}

func testSinglePageSchedule(db *sql.DB, pageIDs, pageNames []string) {
	// Bước 1: Tạo post
	content := fmt.Sprintf("Test Auto Schedule - Single Page - %s", time.Now().Format("15:04:05"))
	postID := createPost(content, pageIDs)
	if postID == "" {
		fmt.Println("❌ Tạo post thất bại")
		return
	}
	fmt.Printf("✅ Tạo post thành công: %s\n", postID[:8])

	// Bước 2: Schedule với auto
	start := time.Now()
	response := schedulePost(postID, pageIDs, time.Now().Format("2006-01-02"), true)
	duration := time.Since(start)

	if response == nil {
		fmt.Println("❌ Schedule thất bại")
		return
	}

	fmt.Printf("✅ Schedule thành công trong %v\n", duration)
	fmt.Printf("   Success: %d, Warning: %d, Error: %d\n",
		response.Preview.SuccessCount,
		response.Preview.WarningCount,
		response.Preview.ErrorCount)

	for _, result := range response.Preview.Results {
		if result.Error == nil {
			fmt.Printf("   📅 %s: %s %s\n",
				result.PageName,
				result.ScheduledTime.Format("2006-01-02"),
				result.ScheduledTime.Format("15:04:05"))
			if result.Warning != "" {
				fmt.Printf("      ⚠️  %s\n", result.Warning)
			}
		} else {
			fmt.Printf("   ❌ %s: %s\n", result.PageName, *result.Error)
		}
	}

	// Kiểm tra database
	verifyScheduledPost(db, postID, pageIDs[0])
}

func testMultiplePageSchedule(db *sql.DB, pageIDs, pageNames []string) {
	content := fmt.Sprintf("Test Auto Schedule - Multiple Pages - %s", time.Now().Format("15:04:05"))
	postID := createPost(content, pageIDs)
	if postID == "" {
		fmt.Println("❌ Tạo post thất bại")
		return
	}
	fmt.Printf("✅ Tạo post cho %d pages: %s\n", len(pageIDs), postID[:8])

	start := time.Now()
	response := schedulePost(postID, pageIDs, time.Now().Format("2006-01-02"), true)
	duration := time.Since(start)

	if response == nil {
		fmt.Println("❌ Schedule thất bại")
		return
	}

	fmt.Printf("✅ Schedule %d pages trong %v\n", len(pageIDs), duration)
	fmt.Printf("   Success: %d, Warning: %d, Error: %d\n",
		response.Preview.SuccessCount,
		response.Preview.WarningCount,
		response.Preview.ErrorCount)

	// Hiển thị kết quả từng page
	for _, result := range response.Preview.Results {
		if result.Error == nil {
			fmt.Printf("   📅 %s: %s %s\n",
				result.PageName,
				result.ScheduledTime.Format("2006-01-02"),
				result.ScheduledTime.Format("15:04:05"))
		}
	}

	// Verify tất cả pages
	for _, pageID := range pageIDs {
		verifyScheduledPost(db, postID, pageID)
	}
}

func testConsecutiveSchedule(db *sql.DB, pageIDs, pageNames []string, count int) {
	totalDuration := time.Duration(0)
	successCount := 0

	for i := 1; i <= count; i++ {
		content := fmt.Sprintf("Test Consecutive #%d - %s", i, time.Now().Format("15:04:05"))
		postID := createPost(content, pageIDs)
		if postID == "" {
			fmt.Printf("❌ Bài %d: Tạo post thất bại\n", i)
			continue
		}

		start := time.Now()
		response := schedulePost(postID, pageIDs, time.Now().Format("2006-01-02"), true)
		duration := time.Since(start)
		totalDuration += duration

		if response != nil && response.Preview.SuccessCount > 0 {
			result := response.Preview.Results[0]
			fmt.Printf("✅ Bài %d: %s %s (%v)\n",
				i,
				result.ScheduledTime.Format("2006-01-02"),
				result.ScheduledTime.Format("15:04:05"),
				duration)
			successCount++
		} else {
			fmt.Printf("❌ Bài %d: Schedule thất bại\n", i)
		}

		// Delay nhỏ để tránh race condition
		time.Sleep(10 * time.Millisecond)
	}

	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(successCount)
		fmt.Printf("\n📊 Thành công: %d/%d bài\n", successCount, count)
		fmt.Printf("📊 Trung bình: %v/bài\n", avgDuration)
		fmt.Printf("📊 Tổng thời gian: %v\n", totalDuration)
	}
}

func createPost(content string, pageIDs []string) string {
	reqBody := CreatePostRequest{
		Content: content,
		Status:  "draft",
		PageIDs: pageIDs,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(baseURL+"/api/posts", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating post: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to create post: %s", string(body))
		return ""
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if post, ok := result["post"].(map[string]interface{}); ok {
		if id, ok := post["id"].(string); ok {
			return id
		}
	}
	return ""
}

func schedulePost(postID string, pageIDs []string, preferredDate string, confirm bool) *ScheduleResponse {
	reqBody := ScheduleRequest{
		PostID:        postID,
		PageIDs:       pageIDs,
		PreferredDate: preferredDate,
		Confirm:       confirm,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(baseURL+"/api/schedule/smart", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error scheduling post: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to schedule: %s", string(body))
		return nil
	}

	var result ScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil
	}

	return &result
}

func verifyScheduledPost(db *sql.DB, postID, pageID string) {
	query := `
		SELECT id, scheduled_time, status, time_slot_id
		FROM scheduled_posts
		WHERE post_id = $1 AND page_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var id, status, timeSlotID string
	var scheduledTime time.Time
	err := db.QueryRow(query, postID, pageID).Scan(&id, &scheduledTime, &status, &timeSlotID)
	
	if err != nil {
		fmt.Printf("   ⚠️  Không tìm thấy scheduled_post trong DB\n")
		return
	}

	fmt.Printf("   ✅ DB: %s | %s | slot: %s\n",
		scheduledTime.Format("2006-01-02 15:04"),
		status,
		timeSlotID[:8])
}

func checkDatabase(db *sql.DB) {
	// Đếm số bài pending
	var pendingCount int
	db.QueryRow(`
		SELECT COUNT(*) FROM scheduled_posts 
		WHERE status = 'pending' AND scheduled_time > NOW()
	`).Scan(&pendingCount)
	fmt.Printf("📊 Pending posts: %d\n", pendingCount)

	// Đếm số bài hôm nay
	var todayCount int
	db.QueryRow(`
		SELECT COUNT(*) FROM scheduled_posts 
		WHERE DATE(scheduled_time) = CURRENT_DATE 
		AND status IN ('pending', 'processing')
	`).Scan(&todayCount)
	fmt.Printf("📊 Posts hôm nay: %d\n", todayCount)

	// Hiển thị 5 bài gần nhất
	fmt.Println("\n📅 5 bài gần nhất:")
	rows, _ := db.Query(`
		SELECT 
			sp.scheduled_time,
			p.page_name,
			po.content,
			sp.status
		FROM scheduled_posts sp
		JOIN pages p ON p.id = sp.page_id
		JOIN posts po ON po.id = sp.post_id
		WHERE sp.scheduled_time > NOW()
		ORDER BY sp.scheduled_time
		LIMIT 5
	`)
	defer rows.Close()

	for rows.Next() {
		var scheduledTime time.Time
		var pageName, content, status string
		rows.Scan(&scheduledTime, &pageName, &content, &status)
		
		// Truncate content
		if len(content) > 40 {
			content = content[:40] + "..."
		}
		
		fmt.Printf("   %s | %s | %s\n",
			scheduledTime.Format("2006-01-02 15:04"),
			pageName,
			content)
	}
}
