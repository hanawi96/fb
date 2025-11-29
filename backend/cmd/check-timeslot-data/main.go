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

	fmt.Println("=== Checking page_time_slots data ===\n")
	
	rows, err := db.Query(`
		SELECT 
			pts.id,
			p.page_name,
			pts.start_time,
			pts.end_time,
			pts.slot_capacity,
			pts.is_active
		FROM page_time_slots pts
		LEFT JOIN pages p ON p.id = pts.page_id
		ORDER BY p.page_name, pts.start_time
		LIMIT 10
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, pageName, startTime, endTime string
		var capacity int
		var isActive bool
		
		err := rows.Scan(&id, &pageName, &startTime, &endTime, &capacity, &isActive)
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Printf("Page: %s\n", pageName)
		fmt.Printf("  ID: %s\n", id)
		fmt.Printf("  Start: %s\n", startTime)
		fmt.Printf("  End: %s\n", endTime)
		fmt.Printf("  Capacity: %d\n", capacity)
		fmt.Printf("  Active: %v\n\n", isActive)
	}
}
