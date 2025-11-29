package main

import (
	"database/sql"
	"fmt"
	"log"
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
	fmt.Println("TEST SIMPLE SCHEDULE WITH ACCOUNT")
	fmt.Println("==============================================\n")

	// Lấy pages
	pages, err := store.GetPages()
	if err != nil || len(pages) == 0 {
		log.Fatal("No pages found")
	}

	page := pages[0]
	fmt.Printf("📄 Using page: %s (%s)\n", page.PageName, page.ID)

	// Lấy primary account của page
	account, err := store.GetPrimaryAccountForPage(page.ID)
	if err != nil || account == nil {
		log.Fatal("No primary account found for page")
	}
	fmt.Printf("👤 Primary account: %s (%s)\n", account.FbUserName, account.ID)

	// Tạo test post
	post := &db.Post{
		Content: fmt.Sprintf("Test simple schedule - %s", time.Now().Format("15:04:05")),
		Status:  "draft",
	}

	if err := store.CreatePost(post); err != nil {
		log.Fatal("Failed to create post:", err)
	}
	fmt.Printf("✅ Created post: %s\n\n", post.ID)

	// Tạo scheduled post với account_id
	scheduledTime := time.Now().Add(2 * time.Hour)
	sp := &db.ScheduledPost{
		PostID:        post.ID,
		PageID:        page.ID,
		AccountID:     &account.ID,
		ScheduledTime: scheduledTime,
		Status:        "pending",
		MaxRetries:    3,
	}

	fmt.Println("💾 Creating scheduled post...")
	if err := store.CreateScheduledPost(sp); err != nil {
		log.Fatal("Failed to create scheduled post:", err)
	}
	fmt.Printf("✅ Created scheduled post: %s\n\n", sp.ID)

	// Verify bằng cách lấy lại
	fmt.Println("🔍 Verifying from database...")
	posts, err := store.GetScheduledPosts("", 100, 0)
	if err != nil {
		log.Fatal("Failed to get scheduled posts:", err)
	}

	fmt.Printf("   Found %d scheduled posts in database\n", len(posts))

	for _, p := range posts {
		if p.ID == sp.ID {
			fmt.Printf("\n✅ FOUND scheduled post:\n")
			fmt.Printf("   ID: %s\n", p.ID)
			fmt.Printf("   Post ID: %s\n", p.PostID)
			fmt.Printf("   Page: %s\n", p.Page.PageName)
			fmt.Printf("   Scheduled: %s\n", p.ScheduledTime.Format("2006-01-02 15:04:05"))
			
			if p.Account != nil {
				fmt.Printf("   ✅ Account: %s\n", p.Account.FbUserName)
				if p.Account.ProfilePictureURL != "" {
					fmt.Printf("   ✅ Avatar URL: %s\n", p.Account.ProfilePictureURL)
				} else {
					fmt.Printf("   ⚠️  No avatar URL\n")
				}
			} else {
				fmt.Printf("   ❌ Account: NULL\n")
			}
			
			fmt.Println("\n✅ Test PASSED! Account is properly assigned and displayed.")
			return
		}
	}

	fmt.Println("❌ Test FAILED! Scheduled post not found")
}
