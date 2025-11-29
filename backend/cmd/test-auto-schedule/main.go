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
	fmt.Println("TEST AUTO SCHEDULE - SMART SCHEDULING")
	fmt.Println("==============================================\n")

	// Bước 1: Tạo test data
	fmt.Println("📝 Bước 1: Tạo test data...")
	testData := setupTestData(store)
	if testData == nil {
		log.Fatal("Failed to setup test data")
	}

	// Bước 2: Tạo time slots (19h-20h và 21h-22h)
	fmt.Println("\n⏰ Bước 2: Tạo time slots...")
	createTimeSlots(store, testData)

	// Bước 3: Test schedule nhiều bài
	fmt.Println("\n🚀 Bước 3: Test schedule nhiều bài...")
	testMultipleSchedules(store, testData)

	// Bước 4: Kiểm tra database
	fmt.Println("\n🔍 Bước 4: Kiểm tra database...")
	verifyDatabase(store, testData)

	fmt.Println("\n✅ Test hoàn tất!")
}

type TestData struct {
	AccountID string
	PageIDs   []string
	PostIDs   []string
}

func setupTestData(store *db.Store) *TestData {
	// Lấy pages thực tế từ database
	pages, err := store.GetPages()
	if err != nil || len(pages) == 0 {
		log.Fatal("❌ Không có pages trong database. Vui lòng thêm pages trước.")
	}

	// Lấy tối đa 3 pages đầu tiên
	maxPages := 3
	if len(pages) < maxPages {
		maxPages = len(pages)
	}

	var pageIDs []string
	for i := 0; i < maxPages; i++ {
		pageIDs = append(pageIDs, pages[i].ID)
		fmt.Printf("📄 Using page: %s (%s)\n", pages[i].PageName, pages[i].ID)
	}

	// Lấy account từ page đầu tiên (nếu có)
	accountID := ""
	assignments, err := store.GetAssignmentsByPage(pageIDs[0])
	if err == nil && len(assignments) > 0 {
		accountID = assignments[0].AccountID
		fmt.Printf("👤 Using account: %s\n", assignments[0].Account.FbUserName)
	}

	// Tạo 5 test posts
	postIDs := createTestPosts(store, 5)
	if len(postIDs) == 0 {
		return nil
	}

	fmt.Printf("✅ Using: %d pages, created %d posts\n", len(pageIDs), len(postIDs))

	return &TestData{
		AccountID: accountID,
		PageIDs:   pageIDs,
		PostIDs:   postIDs,
	}
}



func createTestPosts(store *db.Store, count int) []string {
	var postIDs []string

	for i := 1; i <= count; i++ {
		post := &db.Post{
			Content: fmt.Sprintf("Test post #%d - Auto schedule test at %s", i, time.Now().Format("15:04:05")),
			Status:  "draft",
		}

		if err := store.CreatePost(post); err != nil {
			log.Printf("Error creating post %d: %v", i, err)
			continue
		}

		postIDs = append(postIDs, post.ID)
		fmt.Printf("✅ Post %d created: %s\n", i, post.ID)
	}

	return postIDs
}

func createTimeSlots(store *db.Store, testData *TestData) {
	// Kiểm tra và tạo 2 khung giờ cho mỗi page: 19h-20h và 21h-22h
	// Mỗi khung giờ chỉ cho phép 1 bài/page (slot_capacity = 1)

	for _, pageID := range testData.PageIDs {
		// Lấy page name
		page, _ := store.GetPageByID(pageID)
		pageName := pageID
		if page != nil {
			pageName = page.PageName
		}

		// Kiểm tra slots hiện có
		existingSlots, err := store.GetTimeSlotsByPage(pageID)
		if err == nil && len(existingSlots) >= 2 {
			fmt.Printf("ℹ️  Page '%s' đã có %d slots, bỏ qua tạo mới\n", pageName, len(existingSlots))
			continue
		}

		// Slot 1: 19h-20h
		slot1 := &db.PageTimeSlot{
			PageID:       pageID,
			SlotName:     "Test Evening Slot 1",
			StartTime:    "19:00:00",
			EndTime:      "20:00:00",
			DaysOfWeek:   []int{1, 2, 3, 4, 5, 6, 7}, // Tất cả các ngày
			SlotCapacity: 1,                           // Chỉ cho phép 1 bài/slot
			IsActive:     true,
			Priority:     1,
		}

		if err := store.CreateTimeSlot(slot1); err != nil {
			log.Printf("Error creating slot 1 for page '%s': %v", pageName, err)
		} else {
			fmt.Printf("✅ Slot 1 (19h-20h) created for '%s' (capacity: 1)\n", pageName)
		}

		// Slot 2: 21h-22h
		slot2 := &db.PageTimeSlot{
			PageID:       pageID,
			SlotName:     "Test Evening Slot 2",
			StartTime:    "21:00:00",
			EndTime:      "22:00:00",
			DaysOfWeek:   []int{1, 2, 3, 4, 5, 6, 7},
			SlotCapacity: 1, // Chỉ cho phép 1 bài/slot
			IsActive:     true,
			Priority:     2,
		}

		if err := store.CreateTimeSlot(slot2); err != nil {
			log.Printf("Error creating slot 2 for page '%s': %v", pageName, err)
		} else {
			fmt.Printf("✅ Slot 2 (21h-22h) created for '%s' (capacity: 1)\n", pageName)
		}
	}
}

func testMultipleSchedules(store *db.Store, testData *TestData) {
	smartScheduler := scheduler.NewSmartScheduler(store)

	// Lấy ngày hôm nay lúc 19h (Vietnam time)
	nowVN := config.NowVN()
	today19h := time.Date(nowVN.Year(), nowVN.Month(), nowVN.Day(), 19, 0, 0, 0, config.VietnamTZ)

	fmt.Printf("\n📅 Preferred date: %s (Vietnam time)\n", today19h.Format("2006-01-02 15:04:05"))

	// Test schedule từng bài
	for i, postID := range testData.PostIDs {
		fmt.Printf("\n--- Schedule Post #%d (ID: %s) ---\n", i+1, postID)

		req := scheduler.ScheduleRequest{
			PostID:        postID,
			PageIDs:       testData.PageIDs,
			PreferredDate: today19h,
			UseTimeSlots:  true,
		}

		preview, err := smartScheduler.CalculateSchedule(req)
		if err != nil {
			log.Printf("❌ Error calculating schedule: %v", err)
			continue
		}

		// Hiển thị preview
		fmt.Printf("\n📊 Preview:\n")
		fmt.Printf("   Total pages: %d\n", preview.TotalPages)
		fmt.Printf("   Success: %d, Warning: %d, Error: %d\n", 
			preview.SuccessCount, preview.WarningCount, preview.ErrorCount)
		fmt.Printf("   Next day count: %d\n\n", preview.NextDayCount)

		// Hiển thị chi tiết từng page
		for _, result := range preview.Results {
			timeVN := config.ToVN(result.ScheduledTime)
			fmt.Printf("   📄 Page: %s\n", result.PageName)
			fmt.Printf("      Time: %s (VN)\n", timeVN.Format("2006-01-02 15:04:05"))
			fmt.Printf("      Slot: %s\n", result.TimeSlotID)
			if result.Warning != "" {
				fmt.Printf("      ⚠️  Warning: %s\n", result.Warning)
			}
			if result.Error != nil {
				fmt.Printf("      ❌ Error: %v\n", result.Error)
			}
			fmt.Println()
		}

		// Lưu vào database
		fmt.Println("💾 Saving to database...")
		for _, result := range preview.Results {
			if result.Error != nil {
				continue
			}

			sp := &db.ScheduledPost{
				PostID:        postID,
				PageID:        result.PageID,
				ScheduledTime: result.ScheduledTime.UTC(), // Lưu UTC
				Status:        "pending",
				MaxRetries:    3,
			}

			if result.TimeSlotID != "" {
				sp.TimeSlotID = &result.TimeSlotID
			}

			if err := store.CreateScheduledPost(sp); err != nil {
				log.Printf("   ❌ Error saving for page %s: %v", result.PageName, err)
			} else {
				fmt.Printf("   ✅ Saved for page %s\n", result.PageName)
			}
		}

		// Delay giữa các bài để dễ quan sát
		time.Sleep(500 * time.Millisecond)
	}
}

func verifyDatabase(store *db.Store, testData *TestData) {
	fmt.Println("\n==============================================")
	fmt.Println("DATABASE VERIFICATION")
	fmt.Println("==============================================\n")

	// Lấy tất cả scheduled posts trực tiếp từ database
	query := `
		SELECT sp.id, sp.post_id, sp.page_id, sp.scheduled_time, sp.time_slot_id
		FROM scheduled_posts sp
		ORDER BY sp.scheduled_time
	`
	
	rows, err := store.DB().Query(query)
	if err != nil {
		log.Printf("❌ Error fetching scheduled posts: %v", err)
		return
	}
	defer rows.Close()

	var posts []struct {
		ID            string
		PostID        string
		PageID        string
		ScheduledTime time.Time
		TimeSlotID    *string
	}

	for rows.Next() {
		var p struct {
			ID            string
			PostID        string
			PageID        string
			ScheduledTime time.Time
			TimeSlotID    *string
		}
		if err := rows.Scan(&p.ID, &p.PostID, &p.PageID, &p.ScheduledTime, &p.TimeSlotID); err != nil {
			log.Printf("❌ Error scanning row: %v", err)
			continue
		}
		posts = append(posts, p)
	}

	// Nhóm theo ngày và slot
	type SlotKey struct {
		PageID string
		Date   string
		Slot   string
	}

	slotMap := make(map[SlotKey][]struct {
		ID            string
		PostID        string
		PageID        string
		ScheduledTime time.Time
		TimeSlotID    *string
	})

	for _, post := range posts {
		timeVN := config.ToVN(post.ScheduledTime)
		date := timeVN.Format("2006-01-02")
		hour := timeVN.Hour()

		slotName := ""
		if hour >= 19 && hour < 20 {
			slotName = "19h-20h"
		} else if hour >= 21 && hour < 22 {
			slotName = "21h-22h"
		} else {
			slotName = fmt.Sprintf("%dh", hour)
		}

		key := SlotKey{
			PageID: post.PageID,
			Date:   date,
			Slot:   slotName,
		}

		slotMap[key] = append(slotMap[key], post)
	}

	// Hiển thị kết quả
	fmt.Printf("📊 Total scheduled posts: %d\n\n", len(posts))

	// Nhóm theo page
	pageMap := make(map[string][]struct {
		ID            string
		PostID        string
		PageID        string
		ScheduledTime time.Time
		TimeSlotID    *string
	})
	for _, post := range posts {
		pageMap[post.PageID] = append(pageMap[post.PageID], post)
	}

	for pageID, pagePosts := range pageMap {
		// Lấy page name
		page, _ := store.GetPageByID(pageID)
		pageName := pageID
		if page != nil {
			pageName = page.PageName
		}

		fmt.Printf("📄 Page: %s\n", pageName)
		fmt.Printf("   Total posts: %d\n\n", len(pagePosts))

		// Sắp xếp theo thời gian
		sortedPosts := pagePosts
		for i := 0; i < len(sortedPosts); i++ {
			for j := i + 1; j < len(sortedPosts); j++ {
				if sortedPosts[i].ScheduledTime.After(sortedPosts[j].ScheduledTime) {
					sortedPosts[i], sortedPosts[j] = sortedPosts[j], sortedPosts[i]
				}
			}
		}

		for idx, post := range sortedPosts {
			timeVN := config.ToVN(post.ScheduledTime)
			postIDShort := post.PostID
			if len(postIDShort) > 8 {
				postIDShort = postIDShort[:8]
			}
			fmt.Printf("   %d. %s (VN) - Post: %s\n", 
				idx+1, 
				timeVN.Format("2006-01-02 15:04:05"),
				postIDShort)
		}
		fmt.Println()
	}

	// Kiểm tra vi phạm quy tắc CHỈ cho test data
	fmt.Println("🔍 Checking rules for test data...")
	violations := 0

	// Lọc chỉ test pages
	testPageMap := make(map[string]bool)
	for _, pageID := range testData.PageIDs {
		testPageMap[pageID] = true
	}

	for key, posts := range slotMap {
		// Chỉ kiểm tra test pages
		if !testPageMap[key.PageID] {
			continue
		}

		if len(posts) > 1 {
			violations++
			
			// Lấy page name
			page, _ := store.GetPageByID(key.PageID)
			pageName := key.PageID
			if page != nil {
				pageName = page.PageName
			}
			
			fmt.Printf("   ⚠️  VIOLATION: %s, Date %s, Slot %s has %d posts (max: 1)\n",
				pageName, key.Date, key.Slot, len(posts))
			
			for _, post := range posts {
				timeVN := config.ToVN(post.ScheduledTime)
				postIDShort := post.PostID
				if len(postIDShort) > 8 {
					postIDShort = postIDShort[:8]
				}
				fmt.Printf("      - %s: %s\n", postIDShort, timeVN.Format("15:04:05"))
			}
		}
	}

	if violations == 0 {
		fmt.Println("   ✅ No violations found! All slots respect slot_capacity = 1")
		fmt.Println("   ✅ Logic hoạt động đúng:")
		fmt.Println("      - Mỗi page chỉ đăng 1 bài/slot")
		fmt.Println("      - Slot đầy thì chuyển sang slot tiếp theo")
		fmt.Println("      - Hết slot trong ngày thì chuyển sang ngày mai")
	} else {
		fmt.Printf("   ❌ Found %d violations!\n", violations)
	}

	// Tổng kết
	fmt.Println("\n📊 Summary:")
	fmt.Printf("   - Total test posts: %d\n", len(testData.PostIDs))
	fmt.Printf("   - Total test pages: %d\n", len(testData.PageIDs))
	fmt.Printf("   - Expected scheduled posts: %d\n", len(testData.PostIDs)*len(testData.PageIDs))
	
	testPostCount := 0
	for _, pagePosts := range pageMap {
		if testPageMap[pagePosts[0].PageID] {
			testPostCount += len(pagePosts)
		}
	}
	fmt.Printf("   - Actual scheduled posts: %d\n", testPostCount)
	
	if testPostCount == len(testData.PostIDs)*len(testData.PageIDs) {
		fmt.Println("   ✅ All posts scheduled successfully!")
	}
}
