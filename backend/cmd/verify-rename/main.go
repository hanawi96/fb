package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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

	fmt.Println("=== Verifying column rename ===\n")

	// Check if slot_capacity exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.columns 
			WHERE table_name = 'page_time_slots' 
			AND column_name = 'slot_capacity'
		)
	`).Scan(&exists)

	if err != nil {
		log.Fatal(err)
	}

	if exists {
		fmt.Println("✅ Column 'slot_capacity' EXISTS!")
		
		// Show sample data
		rows, err := db.Query(`
			SELECT 
				p.page_name,
				pts.start_time,
				pts.end_time,
				pts.slot_capacity
			FROM page_time_slots pts
			LEFT JOIN pages p ON p.id = pts.page_id
			ORDER BY p.page_name, pts.start_time
			LIMIT 5
		`)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		fmt.Println("\nSample data:")
		for rows.Next() {
			var pageName, startTime, endTime string
			var capacity int
			rows.Scan(&pageName, &startTime, &endTime, &capacity)
			fmt.Printf("  %s: %s-%s (capacity: %d bài)\n", 
				pageName, startTime[:5], endTime[:5], capacity)
		}
	} else {
		fmt.Println("❌ Column 'slot_capacity' NOT FOUND!")
		fmt.Println("   Run: go run cmd/rename-column/main.go")
	}

	// Check if old column still exists
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.columns 
			WHERE table_name = 'page_time_slots' 
			AND column_name = 'max_posts_per_slot'
		)
	`).Scan(&exists)

	if err == nil && exists {
		fmt.Println("\n⚠️ Old column 'max_posts_per_slot' still exists!")
	} else {
		fmt.Println("\n✅ Old column 'max_posts_per_slot' removed")
	}
}
