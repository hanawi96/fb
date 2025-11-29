package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
	fmt.Println("VERIFY PUBLISH NOW IN DATABASE")
	fmt.Println("==============================================\n")

	// Lấy 10 scheduled posts gần nhất với status completed
	fmt.Println("📊 Recent completed posts (from publish now):")
	fmt.Println("----------------------------------------------")

	query := `
		SELECT 
			sp.id,
			sp.post_id,
			sp.page_id,
			sp.account_id,
			sp.scheduled_time,
			sp.status,
			sp.created_at,
			pg.page_name,
			fa.fb_user_name,
			fa.profile_picture_url,
			p.content
		FROM scheduled_posts sp
		JOIN pages pg ON pg.id = sp.page_id
		LEFT JOIN facebook_accounts fa ON fa.id = sp.account_id
		JOIN posts p ON p.id = sp.post_id
		WHERE sp.status = 'completed'
		ORDER BY sp.created_at DESC
		LIMIT 10
	`

	rows, err := database.Query(query)
	if err != nil {
		log.Fatal("Query error:", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, postID, pageID string
		var accountID *string
		var scheduledTime, createdAt string
		var status, pageName string
		var accountName, accountPicture, content *string

		err := rows.Scan(
			&id, &postID, &pageID, &accountID,
			&scheduledTime, &status, &createdAt,
			&pageName, &accountName, &accountPicture, &content,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		count++
		fmt.Printf("\n%d. Post ID: %s\n", count, postID[:8])
		fmt.Printf("   Scheduled Post ID: %s\n", id[:8])
		fmt.Printf("   Page: %s\n", pageName)
		
		if accountID != nil && *accountID != "" {
			fmt.Printf("   ✅ Account ID: %s\n", (*accountID)[:8])
			if accountName != nil {
				fmt.Printf("   ✅ Account Name: %s\n", *accountName)
			}
			if accountPicture != nil && *accountPicture != "" {
				fmt.Printf("   ✅ Has Avatar: Yes\n")
			}
		} else {
			fmt.Printf("   ❌ Account: NULL\n")
		}

		fmt.Printf("   Status: %s\n", status)
		fmt.Printf("   Time: %s\n", scheduledTime)
		
		if content != nil {
			contentPreview := *content
			if len(contentPreview) > 50 {
				contentPreview = contentPreview[:50] + "..."
			}
			fmt.Printf("   Content: %s\n", contentPreview)
		}
	}

	if count == 0 {
		fmt.Println("⚠️  No completed posts found")
		fmt.Println("   Try publishing a post with 'Đăng ngay' first")
	}

	// Thống kê
	fmt.Println("\n==============================================")
	fmt.Println("STATISTICS")
	fmt.Println("==============================================")

	// Đếm posts theo status
	statsQuery := `
		SELECT 
			status,
			COUNT(*) as count,
			COUNT(account_id) as with_account
		FROM scheduled_posts
		GROUP BY status
		ORDER BY status
	`

	rows2, err := database.Query(statsQuery)
	if err != nil {
		log.Fatal("Stats query error:", err)
	}
	defer rows2.Close()

	fmt.Println("\nPosts by status:")
	for rows2.Next() {
		var status string
		var count, withAccount int

		if err := rows2.Scan(&status, &count, &withAccount); err != nil {
			continue
		}

		percentage := float64(withAccount) / float64(count) * 100
		fmt.Printf("   %s: %d posts (%d with account = %.1f%%)\n", 
			status, count, withAccount, percentage)
	}

	// Kiểm tra posts từ hôm nay
	fmt.Println("\n📅 Today's posts:")
	todayQuery := `
		SELECT 
			status,
			COUNT(*) as count
		FROM scheduled_posts
		WHERE DATE(created_at) = CURRENT_DATE
		GROUP BY status
	`

	rows3, err := database.Query(todayQuery)
	if err != nil {
		log.Fatal("Today query error:", err)
	}
	defer rows3.Close()

	for rows3.Next() {
		var status string
		var count int

		if err := rows3.Scan(&status, &count); err != nil {
			continue
		}

		fmt.Printf("   %s: %d\n", status, count)
	}

	// Test GetScheduledPosts API
	fmt.Println("\n==============================================")
	fmt.Println("TEST GetScheduledPosts API")
	fmt.Println("==============================================")

	posts, err := store.GetScheduledPosts("completed", 5, 0)
	if err != nil {
		log.Fatal("GetScheduledPosts error:", err)
	}

	fmt.Printf("\nFound %d completed posts via API:\n", len(posts))
	for i, post := range posts {
		fmt.Printf("\n%d. ID: %s\n", i+1, post.ID[:8])
		if post.Page != nil {
			fmt.Printf("   Page: %s\n", post.Page.PageName)
		}
		if post.Account != nil {
			fmt.Printf("   ✅ Account: %s\n", post.Account.FbUserName)
			if post.Account.ProfilePictureURL != "" {
				fmt.Printf("   ✅ Has Avatar\n")
			}
		} else {
			fmt.Printf("   ❌ Account: NULL\n")
		}
		
		timeVN := config.ToVN(post.ScheduledTime)
		fmt.Printf("   Time: %s (VN)\n", timeVN.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n✅ Verification complete!")
}
