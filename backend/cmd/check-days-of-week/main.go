package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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

	fmt.Println("=== Kiểm tra days_of_week ===\n")

	// Ngày 29/11/2025 là thứ mấy?
	date := time.Date(2025, 11, 29, 0, 0, 0, 0, time.UTC)
	dayOfWeek := int(date.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7 // Sunday = 7
	}
	
	fmt.Printf("Ngày 29/11/2025 là thứ: %d (%s)\n\n", dayOfWeek, date.Weekday())

	query := `
		SELECT 
			pts.id,
			pts.start_time::text,
			pts.end_time::text,
			pts.days_of_week,
			$1 = ANY(pts.days_of_week) as is_match
		FROM page_time_slots pts
		JOIN pages p ON p.id = pts.page_id
		WHERE p.page_name = 'Ánh Lê'
			AND pts.is_active = true
		ORDER BY pts.start_time
	`

	rows, err := db.Query(query, dayOfWeek)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotID, startTime, endTime string
		var daysOfWeek []byte
		var isMatch bool
		
		rows.Scan(&slotID, &startTime, &endTime, &daysOfWeek, &isMatch)
		
		matchStatus := "❌"
		if isMatch {
			matchStatus = "✅"
		}
		
		fmt.Printf("%s Slot %s-%s | days_of_week: %s | Match: %v | ID: %s\n",
			matchStatus,
			startTime[:5],
			endTime[:5],
			string(daysOfWeek),
			isMatch,
			slotID[:8])
	}
}
