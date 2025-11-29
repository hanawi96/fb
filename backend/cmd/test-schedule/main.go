package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

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

	fmt.Println("=== TEST AUTOMATIC SCHEDULING ===\n")

	// Bước 1: Lấy danh sách pages
	fmt.Println("📋 Bước 1: Lấy danh sách pages...")
	var pageIDs []string
	rows, err := db.Query("SELECT id FROM pages LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		pageIDs = append(pageIDs, id)
	}
	rows.Close()
	fmt.Printf("   ✅ Tìm thấy %d pages\n\n", len(pageIDs))

	if len(pageIDs) == 0 {
		fmt.Println("❌ Không có page nào để test!")
		return
	}

	// Bước 2: Tạo post mới
	fmt.Println("📝 Bước 2: Tạo post test...")
	var postID string
	err = db.QueryRow(`
		INSERT INTO posts (content, media_type, status)
		VALUES ('Test post - ' || NOW(), 'text', 'draft')
		RETURNING id
	`).Scan(&postID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   ✅ Post ID: %s\n\n", postID)

	// Bước 3: Gọi API schedule
	fmt.Println("🚀 Bước 3: Gọi API schedule...")
	
	// Thời gian: 19:00 hôm nay (Vietnam time)
	now := time.Now()
	scheduledTime := time.Date(now.Year(), now.Month(), now.Day(), 19, 0, 0, 0, time.FixedZone("Asia/Ho_Chi_Minh", 7*3600))
	
	requestBody := map[string]interface{}{
		"post_id":        postID,
		"page_ids":       pageIDs,
		"scheduled_time": scheduledTime.Format(time.RFC3339),
	}
	
	jsonData, _ := json.Marshal(requestBody)
	resp, err := http.Post("http://localhost:8080/api/schedule", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("API call failed:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		fmt.Printf("   ❌ API failed with status: %d\n", resp.StatusCode)
		return
	}
	fmt.Println("   ✅ API call successful\n")

	// Đợi 1 giây để database update
	time.Sleep(1 * time.Second)

	// Bước 4: Kiểm tra database
	fmt.Println("🔍 Bước 4: Kiểm tra database...\n")

	rows2, err := db.Query(`
		SELECT 
			sp.id,
			p.page_name,
			sp.scheduled_time,
			sp.time_slot_id,
			pts.slot_capacity,
			sp.status
		FROM scheduled_posts sp
		LEFT JOIN pages p ON p.id = sp.page_id
		LEFT JOIN page_time_slots pts ON pts.id = sp.time_slot_id
		WHERE sp.post_id = $1
		ORDER BY p.page_name
	`, postID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	withSlot := 0
	withoutSlot := 0
	
	fmt.Println("📊 Kết quả:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	
	for rows2.Next() {
		var id, pageName, status string
		var scheduledTime time.Time
		var timeSlotID sql.NullString
		var capacity sql.NullInt64

		rows2.Scan(&id, &pageName, &scheduledTime, &timeSlotID, &capacity, &status)

		slotInfo := "❌ NO SLOT"
		if timeSlotID.Valid {
			slotInfo = fmt.Sprintf("✅ Slot (capacity: %d)", capacity.Int64)
			withSlot++
		} else {
			withoutSlot++
		}

		fmt.Printf("%-40s | %s | %s\n", 
			pageName, 
			scheduledTime.Format("15:04"), 
			slotInfo)
	}

	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("\n📈 Tổng kết:\n")
	fmt.Printf("   ✅ Có time_slot_id: %d bài\n", withSlot)
	fmt.Printf("   ❌ KHÔNG có time_slot_id: %d bài\n", withoutSlot)

	// Bước 5: Kiểm tra giới hạn slot
	fmt.Println("\n🔍 Bước 5: Kiểm tra giới hạn slot...\n")

	rows3, err := db.Query(`
		SELECT 
			p.page_name,
			pts.start_time,
			pts.end_time,
			pts.slot_capacity,
			COUNT(sp.id) as current_count
		FROM page_time_slots pts
		LEFT JOIN pages p ON p.id = pts.page_id
		LEFT JOIN scheduled_posts sp ON sp.time_slot_id = pts.id 
			AND sp.status IN ('pending', 'processing')
			AND DATE(sp.scheduled_time) = CURRENT_DATE
		WHERE pts.id IN (
			SELECT DISTINCT time_slot_id 
			FROM scheduled_posts 
			WHERE post_id = $1 AND time_slot_id IS NOT NULL
		)
		GROUP BY p.page_name, pts.start_time, pts.end_time, pts.slot_capacity
		ORDER BY p.page_name
	`, postID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows3.Close()

	fmt.Println("📊 Tình trạng slots:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	
	hasOverflow := false
	for rows3.Next() {
		var pageName, startTime, endTime string
		var capacity, currentCount int

		rows3.Scan(&pageName, &startTime, &endTime, &capacity, &currentCount)

		status := "✅ OK"
		if currentCount > capacity {
			status = "❌ VƯỢT QUÁ!"
			hasOverflow = true
		} else if currentCount == capacity {
			status = "⚠️ ĐẦY"
		}

		fmt.Printf("%-40s | %s-%s | %d/%d bài | %s\n",
			pageName,
			startTime[:5],
			endTime[:5],
			currentCount,
			capacity,
			status)
	}
	fmt.Println("─────────────────────────────────────────────────────────────")

	// Kết luận
	fmt.Println("\n" + "═══════════════════════════════════════════════════════════════")
	if withSlot == len(pageIDs) && !hasOverflow {
		fmt.Println("✅ TEST PASSED!")
		fmt.Println("   - Tất cả bài đều có time_slot_id")
		fmt.Println("   - Không có slot nào vượt quá capacity")
	} else {
		fmt.Println("❌ TEST FAILED!")
		if withoutSlot > 0 {
			fmt.Printf("   - %d bài KHÔNG có time_slot_id\n", withoutSlot)
		}
		if hasOverflow {
			fmt.Println("   - Có slot vượt quá capacity")
		}
	}
	fmt.Println("═══════════════════════════════════════════════════════════════")
}
