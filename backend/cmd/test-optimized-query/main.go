package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"fbscheduler/internal/db"

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

	fmt.Println("=== Test Optimized Query ===\n")

	// Lấy page đầu tiên
	var pageID, pageName string
	err = database.QueryRow("SELECT id, page_name FROM pages LIMIT 1").Scan(&pageID, &pageName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Testing with page: %s\n", pageName)
	fmt.Printf("Page ID: %s\n\n", pageID)

	// Test query
	startDate := time.Now()
	fmt.Printf("Start Date: %s\n\n", startDate.Format("2006-01-02"))

	result, err := store.FindNextAvailableSlot(pageID, startDate, 30)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else if result == nil {
		fmt.Println("❌ No slot found")
	} else {
		fmt.Println("✅ Found slot:")
		fmt.Printf("  Slot ID: %s\n", result.SlotID)
		fmt.Printf("  Date: %s\n", result.Date.Format("2006-01-02"))
		fmt.Printf("  Time: %s - %s\n", result.StartTime, result.EndTime)
		fmt.Printf("  Capacity: %d\n", result.Capacity)
		fmt.Printf("  Used: %d\n", result.UsedCount)
	}
}
