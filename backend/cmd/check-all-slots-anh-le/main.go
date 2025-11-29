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

	fmt.Println("=== TẤT CẢ slots của Ánh Lê ===\n")

	query := `
		SELECT 
			pts.id,
			pts.start_time::text,
			pts.end_time::text,
			pts.days_of_week,
			pts.is_active,
			pts.slot_capacity
		FROM page_time_slots pts
		JOIN pages p ON p.id = pts.page_id
		WHERE p.page_name = 'Ánh Lê'
		ORDER BY pts.start_time
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var slotID, startTime, endTime string
		var daysOfWeek []byte
		var isActive bool
		var capacity int
		
		rows.Scan(&slotID, &startTime, &endTime, &daysOfWeek, &isActive, &capacity)
		count++
		
		activeStatus := "✅"
		if !isActive {
			activeStatus = "❌"
		}
		
		fmt.Printf("%d. %s %s-%s | days: %s | cap: %d | active: %v | ID: %s\n",
			count,
			activeStatus,
			startTime[:5],
			endTime[:5],
			string(daysOfWeek),
			capacity,
			isActive,
			slotID[:8])
	}

	fmt.Printf("\n✅ Tổng: %d slots\n", count)
}
