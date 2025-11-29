package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
	fmt.Println("CHECK ACCOUNT ASSIGNMENT IN SCHEDULED POSTS")
	fmt.Println("==============================================\n")

	// 1. Kiểm tra scheduled_posts có account_id không
	fmt.Println("📊 Checking scheduled_posts table...")
	query := `
		SELECT 
			sp.id, 
			sp.page_id, 
			sp.account_id,
			pg.page_name,
			fa.fb_user_name
		FROM scheduled_posts sp
		JOIN pages pg ON pg.id = sp.page_id
		LEFT JOIN facebook_accounts fa ON fa.id = sp.account_id
		ORDER BY sp.created_at DESC
		LIMIT 10
	`

	rows, err := database.Query(query)
	if err != nil {
		log.Fatal("Query error:", err)
	}
	defer rows.Close()

	hasAccount := 0
	noAccount := 0

	fmt.Println("\nRecent scheduled posts:")
	fmt.Println("----------------------------------------")
	for rows.Next() {
		var id, pageID string
		var accountID sql.NullString
		var pageName string
		var accountName sql.NullString

		if err := rows.Scan(&id, &pageID, &accountID, &pageName, &accountName); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		if accountID.Valid {
			hasAccount++
			fmt.Printf("✅ Post %s | Page: %s | Account: %s\n", 
				id[:8], pageName, accountName.String)
		} else {
			noAccount++
			fmt.Printf("❌ Post %s | Page: %s | Account: NULL\n", 
				id[:8], pageName)
		}
	}

	fmt.Println("\n📊 Summary:")
	fmt.Printf("   - Posts with account: %d\n", hasAccount)
	fmt.Printf("   - Posts without account: %d\n", noAccount)

	// 2. Kiểm tra page_account_assignments
	fmt.Println("\n📊 Checking page_account_assignments...")
	query2 := `
		SELECT 
			pg.page_name,
			fa.fb_user_name,
			paa.is_primary
		FROM page_account_assignments paa
		JOIN pages pg ON pg.id = paa.page_id
		JOIN facebook_accounts fa ON fa.id = paa.account_id
		ORDER BY pg.page_name
	`

	rows2, err := database.Query(query2)
	if err != nil {
		log.Fatal("Query error:", err)
	}
	defer rows2.Close()

	fmt.Println("\nPage-Account assignments:")
	fmt.Println("----------------------------------------")
	count := 0
	for rows2.Next() {
		var pageName, accountName string
		var isPrimary bool

		if err := rows2.Scan(&pageName, &accountName, &isPrimary); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		count++
		primaryMark := ""
		if isPrimary {
			primaryMark = " (PRIMARY)"
		}
		fmt.Printf("📄 %s → 👤 %s%s\n", pageName, accountName, primaryMark)
	}

	if count == 0 {
		fmt.Println("⚠️  No page-account assignments found!")
	}

	// 3. Đề xuất fix
	if noAccount > 0 {
		fmt.Println("\n💡 Solution:")
		fmt.Println("   Scheduled posts không có account_id vì:")
		fmt.Println("   1. Khi tạo scheduled post, account_id không được set")
		fmt.Println("   2. Cần update logic để auto-assign account khi schedule")
		fmt.Println("\n   Để fix, chạy:")
		fmt.Println("   UPDATE scheduled_posts sp")
		fmt.Println("   SET account_id = (")
		fmt.Println("       SELECT paa.account_id")
		fmt.Println("       FROM page_account_assignments paa")
		fmt.Println("       WHERE paa.page_id = sp.page_id")
		fmt.Println("       AND paa.is_primary = true")
		fmt.Println("       LIMIT 1")
		fmt.Println("   )")
		fmt.Println("   WHERE sp.account_id IS NULL;")
	}

	// 4. Tự động fix nếu user muốn
	if noAccount > 0 {
		fmt.Print("\n❓ Bạn có muốn tự động fix không? (y/n): ")
		var answer string
		fmt.Scanln(&answer)

		if answer == "y" || answer == "Y" {
			fmt.Println("\n🔧 Fixing...")
			updateQuery := `
				UPDATE scheduled_posts sp
				SET account_id = (
					SELECT paa.account_id
					FROM page_account_assignments paa
					WHERE paa.page_id = sp.page_id
					AND paa.is_primary = true
					LIMIT 1
				)
				WHERE sp.account_id IS NULL
			`

			result, err := database.Exec(updateQuery)
			if err != nil {
				log.Fatal("Update error:", err)
			}

			affected, _ := result.RowsAffected()
			fmt.Printf("✅ Updated %d scheduled posts\n", affected)

			// Verify
			fmt.Println("\n🔍 Verifying...")
			posts, err := store.GetScheduledPosts("", 5, 0)
			if err != nil {
				log.Fatal("Verify error:", err)
			}

			for _, post := range posts {
				if post.Account != nil {
					fmt.Printf("✅ Post %s now has account: %s\n", 
						post.ID[:8], post.Account.FbUserName)
				}
			}
		}
	}

	fmt.Println("\n✅ Check complete!")
}
