package main

import (
	"database/sql"
	"encoding/json"
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

	fmt.Println("=== Test Auto Schedule với Slot Đầy ===\n")

	// Lấy danh sách pages từ database
	rows, err := database.Query("SELECT id, page_name FROM pages LIMIT 2")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var pageIDs []string
	var pageNames []string
	
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		pageIDs = append(pageIDs, id)
		pageNames = append(pageNames, name)
		fmt.Printf("Page: %s\n", name)
	}

	if len(pageIDs) == 0 {
		fmt.Println("❌ Không có page nào")
		return
	}

	// Tạo scheduling service
	schedulingService := scheduler.NewSchedulingService(store)

	// Test với ngày hôm nay
	preferredDate := config.NowVN()
	fmt.Printf("\nPreferred Date: %s\n\n", preferredDate.Format("2006-01-02"))

	// Tạo fake post ID
	postID := "test-post-" + time.Now().Format("20060102150405")

	// Calculate schedule
	preview, err := schedulingService.SchedulePostToPages(postID, pageIDs, preferredDate)
	if err != nil {
		log.Fatal("Error:", err)
	}

	// Hiển thị kết quả
	fmt.Println("=== Kết quả Schedule ===\n")
	fmt.Printf("Total Pages: %d\n", preview.TotalPages)
	fmt.Printf("Success: %d\n", preview.SuccessCount)
	fmt.Printf("Warning: %d\n", preview.WarningCount)
	fmt.Printf("Error: %d\n", preview.ErrorCount)
	fmt.Printf("Next Day: %d\n\n", preview.NextDayCount)

	// Hiển thị chi tiết từng page
	for i, result := range preview.Results {
		fmt.Printf("--- Result %d ---\n", i+1)
		fmt.Printf("Page: %s\n", result.PageName)
		
		if result.Error != nil {
			fmt.Printf("❌ Error: %s\n", result.Error.Error())
		} else {
			fmt.Printf("✅ Scheduled Time: %s\n", result.ScheduledTime.Format("2006-01-02 15:04:05"))
			if result.Warning != "" {
				fmt.Printf("⚠️  Warning: %s\n", result.Warning)
			}
			if result.AccountName != "" {
				fmt.Printf("Account: %s\n", result.AccountName)
			}
		}
		fmt.Println()
	}

	// Hiển thị JSON
	jsonData, _ := json.MarshalIndent(preview, "", "  ")
	fmt.Println("=== JSON Response ===")
	fmt.Println(string(jsonData))
}
