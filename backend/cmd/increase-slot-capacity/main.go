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

	fmt.Println("=== Tăng slot_capacity lên 5 bài/slot ===\n")
	
	result, err := db.Exec(`
		UPDATE page_time_slots 
		SET slot_capacity = 5
		WHERE slot_capacity < 5
	`)
	if err != nil {
		log.Fatal(err)
	}

	affected, _ := result.RowsAffected()
	fmt.Printf("✅ Đã cập nhật %d time slots\n", affected)
	
	// Kiểm tra lại
	fmt.Println("\n=== Kiểm tra lại capacity ===\n")
	
	rows, err := db.Query(`
		SELECT 
			p.page_name,
			pts.start_time,
			pts.end_time,
			pts.slot_capacity
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
		var pageName, startTime, endTime string
		var capacity int
		
		err := rows.Scan(&pageName, &startTime, &endTime, &capacity)
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Printf("%s: %s-%s (capacity: %d bài)\n", 
			pageName, startTime[11:16], endTime[11:16], capacity)
	}
}
