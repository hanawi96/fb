package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"fbscheduler/internal/config"
	"fbscheduler/internal/db"
	"fbscheduler/internal/scheduler"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:yendev96@localhost:5432/fbscheduler?sslmode=disable"
	}

	database, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	store := db.NewStore(database)

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║     COMPREHENSIVE SCHEDULING TEST SUITE                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Test 1: Kiểm tra indexes
	fmt.Println("📊 TEST 1: Kiểm tra Database Indexes")
	fmt.Println("─────────────────────────────────────────────────────────────")
	checkIndexes(database)
	fmt.Println()

	// Test 2: Kiểm tra slot availability
	fmt.Println("🎯 TEST 2: Kiểm tra Slot Availability")
	fmt.Println("─────────────────────────────────────────────────────────────")
	checkSlotAvailability(database)
	fmt.Println()

	// Test 3: Test query performance
	fmt.Println("⚡ TEST 3: Test Query Performance")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testQueryPerformance(store)
	fmt.Println()

	// Test 4: Test trường hợp slot đầy
	fmt.Println("🔴 TEST 4: Slot Đầy - Tìm Slot Tiếp Theo")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testFullSlot(store)
	fmt.Println()

	// Test 5: Test trường hợp có slot trống ở giữa
	fmt.Println("🟢 TEST 5: Slot Trống Ở Giữa (User Xóa Bài)")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testGapInSchedule(store, database)
	fmt.Println()

	// Test 6: Test schedule nhiều pages cùng lúc
	fmt.Println("📦 TEST 6: Schedule Nhiều Pages Cùng Lúc")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testMultiplePages(store, database)
	fmt.Println()

	// Test 7: Test schedule liên tiếp
	fmt.Println("🔄 TEST 7: Schedule Liên Tiếp (10 Bài)")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testConsecutiveScheduling(store, database)
	fmt.Println()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    TEST COMPLETED                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

func checkIndexes(db *sql.DB) {
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			indexdef
		FROM pg_indexes
		WHERE schemaname = 'public'
			AND (
				indexname LIKE 'idx_scheduled_posts%' 
				OR indexname LIKE 'idx_page_time_slots%'
			)
		ORDER BY tablename, indexname
	`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var schema, table, indexName, indexDef string
		rows.Scan(&schema, &table, &indexName, &indexDef)
		fmt.Printf("✅ %s.%s\n", table, indexName)
		count++
	}

	if count == 0 {
		fmt.Println("⚠️  Không tìm thấy indexes! Cần chạy migration 007")
	} else {
		fmt.Printf("\n✅ Tổng cộng: %d indexes\n", count)
	}
}

func checkSlotAvailability(db *sql.DB) {
	query := `
		SELECT 
			p.page_name,
			pts.start_time::text,
			pts.end_time::text,
			pts.slot_capacity,
			COUNT(sp.id) as used_count,
			pts.slot_capacity - COUNT(sp.id) as available
		FROM page_time_slots pts
		LEFT JOIN pages p ON p.id = pts.page_id
		LEFT JOIN scheduled_posts sp 
			ON sp.time_slot_id = pts.id 
			AND DATE(sp.scheduled_time) = CURRENT_DATE
			AND sp.status IN ('pending', 'processing')
		WHERE pts.is_active = true
		GROUP BY p.page_name, pts.start_time, pts.end_time, pts.slot_capacity
		ORDER BY p.page_name, pts.start_time
		LIMIT 10
	`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pageName, startTime, endTime string
		var capacity, used, available int
		rows.Scan(&pageName, &startTime, &endTime, &capacity, &used, &available)

		status := "✅"
		if available == 0 {
			status = "🔴"
		} else if available < capacity/2 {
			status = "🟡"
		}

		fmt.Printf("%s %s | %s-%s | %d/%d (còn %d)\n",
			status, pageName, startTime[:5], endTime[:5], used, capacity, available)
	}
}

func testQueryPerformance(store *db.Store) {
	// Lấy page đầu tiên
	var pageID string
	err := store.DB().QueryRow("SELECT id FROM pages LIMIT 1").Scan(&pageID)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	startDate := config.NowVN()

	// Test 1: Single query
	start := time.Now()
	result, err := store.FindNextAvailableSlot(pageID, startDate, 30)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
	} else if result == nil {
		fmt.Printf("⚠️  No slot found\n")
	} else {
		fmt.Printf("✅ Query completed in %v\n", duration)
		fmt.Printf("   Found: %s at %s\n", result.Date.Format("2006-01-02"), result.StartTime[:5])
	}

	// Test 2: Batch query
	var pageIDs []string
	rows, _ := store.DB().Query("SELECT id FROM pages LIMIT 5")
	for rows.Next() {
		var id string
		rows.Scan(&id)
		pageIDs = append(pageIDs, id)
	}
	rows.Close()

	start = time.Now()
	results, err := store.FindNextAvailableSlotsForPages(pageIDs, startDate, 30)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("❌ Batch query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Batch query for %d pages completed in %v\n", len(pageIDs), duration)
		fmt.Printf("   Found slots for %d pages\n", len(results))
	}
}

func testFullSlot(store *db.Store) {
	schedulingService := scheduler.NewSchedulingService(store)

	// Lấy 1 page
	var pageID string
	err := store.DB().QueryRow("SELECT id FROM pages LIMIT 1").Scan(&pageID)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	preferredDate := config.NowVN()
	postID := fmt.Sprintf("test-full-%d", time.Now().Unix())

	preview, err := schedulingService.SchedulePostToPages(postID, []string{pageID}, preferredDate)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if preview.SuccessCount > 0 {
		result := preview.Results[0]
		fmt.Printf("✅ Tìm được slot: %s %s\n",
			result.ScheduledTime.Format("2006-01-02"),
			result.ScheduledTime.Format("15:04"))
		if result.Warning != "" {
			fmt.Printf("   ⚠️  %s\n", result.Warning)
		}
	} else {
		fmt.Printf("❌ Không tìm được slot\n")
	}
}

func testGapInSchedule(store *db.Store, db *sql.DB) {
	// Tạo gap bằng cách xóa 1 bài ở giữa
	var spID string
	err := db.QueryRow(`
		SELECT id FROM scheduled_posts 
		WHERE status = 'pending' 
		AND scheduled_time > NOW()
		ORDER BY scheduled_time 
		LIMIT 1 OFFSET 1
	`).Scan(&spID)

	if err != nil {
		fmt.Println("⚠️  Không có bài để xóa, skip test")
		return
	}

	// Xóa bài để tạo gap
	_, err = db.Exec("DELETE FROM scheduled_posts WHERE id = $1", spID)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Đã tạo gap (xóa bài %s)\n", spID[:8])

	// Test xem có tìm được gap không
	var pageID string
	db.QueryRow("SELECT id FROM pages LIMIT 1").Scan(&pageID)

	result, err := store.FindNextAvailableSlot(pageID, config.NowVN(), 30)
	if err == nil && result != nil {
		fmt.Printf("✅ Hệ thống tìm được gap: %s %s\n",
			result.Date.Format("2006-01-02"),
			result.StartTime[:5])
	} else {
		fmt.Println("❌ Không tìm được gap")
	}
}

func testMultiplePages(store *db.Store, db *sql.DB) {
	schedulingService := scheduler.NewSchedulingService(store)

	// Lấy 3 pages
	var pageIDs []string
	rows, _ := db.Query("SELECT id FROM pages LIMIT 3")
	for rows.Next() {
		var id string
		rows.Scan(&id)
		pageIDs = append(pageIDs, id)
	}
	rows.Close()

	if len(pageIDs) == 0 {
		fmt.Println("⚠️  Không có pages")
		return
	}

	preferredDate := config.NowVN()
	postID := fmt.Sprintf("test-multi-%d", time.Now().Unix())

	start := time.Now()
	preview, err := schedulingService.SchedulePostToPages(postID, pageIDs, preferredDate)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Schedule %d pages trong %v\n", len(pageIDs), duration)
	fmt.Printf("   Success: %d, Warning: %d, Error: %d\n",
		preview.SuccessCount, preview.WarningCount, preview.ErrorCount)

	for _, result := range preview.Results {
		if result.Error == nil {
			fmt.Printf("   - %s: %s %s\n",
				result.PageName,
				result.ScheduledTime.Format("2006-01-02"),
				result.ScheduledTime.Format("15:04"))
		}
	}
}

func testConsecutiveScheduling(store *db.Store, db *sql.DB) {
	schedulingService := scheduler.NewSchedulingService(store)

	var pageID string
	err := db.QueryRow("SELECT id FROM pages LIMIT 1").Scan(&pageID)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	preferredDate := config.NowVN()
	totalDuration := time.Duration(0)

	fmt.Println("Scheduling 10 bài liên tiếp...")

	for i := 1; i <= 10; i++ {
		postID := fmt.Sprintf("test-consecutive-%d-%d", time.Now().Unix(), i)

		start := time.Now()
		preview, err := schedulingService.SchedulePostToPages(postID, []string{pageID}, preferredDate)
		duration := time.Since(start)
		totalDuration += duration

		if err != nil || preview.SuccessCount == 0 {
			fmt.Printf("❌ Bài %d: Failed\n", i)
			break
		}

		result := preview.Results[0]
		fmt.Printf("✅ Bài %2d: %s %s (%v)\n",
			i,
			result.ScheduledTime.Format("2006-01-02"),
			result.ScheduledTime.Format("15:04"),
			duration)
	}

	avgDuration := totalDuration / 10
	fmt.Printf("\n📊 Trung bình: %v/bài\n", avgDuration)
	fmt.Printf("📊 Tổng thời gian: %v\n", totalDuration)
}
