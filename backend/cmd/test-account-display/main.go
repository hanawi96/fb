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
	fmt.Println("TEST ACCOUNT DISPLAY IN SCHEDULED POSTS")
	fmt.Println("==============================================\n")

	// Lấy pages
	pages, err := store.GetPages()
	if err != nil || len(pages) == 0 {
		log.Fatal("No pages found")
	}

	// Lấy 1 page đầu tiên
	page := pages[0]
	fmt.Printf("📄 Using page: %s\n", page.PageName)

	// Tạo test post
	post := &db.Post{
		Content: fmt.Sprintf("Test account display - %s", time.Now().Format("15:04:05")),
		Status:  "draft",
	}

	if err := store.CreatePost(post); err != nil {
		log.Fatal("Failed to create post:", err)
	}
	fmt.Printf("✅ Created post: %s\n", post.ID)

	// Schedule post
	smartScheduler := scheduler.NewSmartScheduler(store)
	nowVN := config.NowVN()
	today19h := time.Date(nowVN.Year(), nowVN.Month(), nowVN.Day(), 19, 0, 0, 0, config.VietnamTZ)

	req := scheduler.ScheduleRequest{
		PostID:        post.ID,
		PageIDs:       []string{page.ID},
		PreferredDate: today19h,
		UseTimeSlots:  true,
	}

	preview, err := smartScheduler.CalculateSchedule(req)
	if err != nil {
		log.Fatal("Failed to calculate schedule:", err)
	}

	fmt.Printf("\n📊 Schedule preview:\n")
	for _, result := range preview.Results {
		fmt.Printf("   Page: %s\n", result.PageName)
		fmt.Printf("   Account: %s (%s)\n", result.AccountName, result.AccountID)
		fmt.Printf("   Time: %s\n", config.ToVN(result.ScheduledTime).Format("2006-01-02 15:04:05"))
	}

	// Lưu vào database
	fmt.Println("\n💾 Saving to database...")
	for _, result := range preview.Results {
		if result.Error != nil {
			continue
		}

		sp := &db.ScheduledPost{
			PostID:        post.ID,
			PageID:        result.PageID,
			ScheduledTime: result.ScheduledTime.UTC(),
			Status:        "pending",
			MaxRetries:    3,
		}

		// Set account_id và time_slot_id
		if result.AccountID != "" {
			sp.AccountID = &result.AccountID
		}
		if result.TimeSlotID != "" {
			sp.TimeSlotID = &result.TimeSlotID
		}

		if err := store.CreateScheduledPost(sp); err != nil {
			log.Printf("Error saving: %v", err)
			continue
		}

		fmt.Printf("✅ Saved scheduled post: %s\n", sp.ID)
	}

	// Verify bằng cách lấy lại từ database
	fmt.Println("\n🔍 Verifying from database...")
	posts, err := store.GetScheduledPosts("", 5, 0)
	if err != nil {
		log.Fatal("Failed to get scheduled posts:", err)
	}

	found := false
	for _, sp := range posts {
		if sp.PostID == post.ID {
			found = true
			fmt.Printf("\n✅ Found scheduled post:\n")
			fmt.Printf("   ID: %s\n", sp.ID)
			fmt.Printf("   Page: %s\n", sp.Page.PageName)
			if sp.Account != nil {
				fmt.Printf("   Account: %s ✅\n", sp.Account.FbUserName)
				if sp.Account.ProfilePictureURL != "" {
					fmt.Printf("   Avatar: %s ✅\n", sp.Account.ProfilePictureURL)
				}
			} else {
				fmt.Printf("   Account: NULL ❌\n")
			}
			break
		}
	}

	if !found {
		fmt.Println("❌ Scheduled post not found in recent posts")
	}

	fmt.Println("\n✅ Test complete!")
}
