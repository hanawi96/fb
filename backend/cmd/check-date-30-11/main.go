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
	// Check ngày 30/11/2025
	date := time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC)
	dayOfWeek := int(date.Weekday())
	isoDayOfWeek := dayOfWeek
	if isoDayOfWeek == 0 {
		isoDayOfWeek = 7 // Sunday = 7 in ISO
	}
	
	fmt.Println("=== Kiểm tra ngày 30/11/2025 ===\n")
	fmt.Printf("Ngày: %s\n", date.Format("2006-01-02"))
	fmt.Printf("Thứ (Go): %d (%s)\n", dayOfWeek, date.Weekday())
	fmt.Printf("Thứ (ISO): %d\n\n", isoDayOfWeek)

	// Check database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:yendev96@localhost:5432/fbscheduler?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check slots của Ánh Lê có match ngày 30/11 không
	fmt.Println("=== Kiểm tra slots match với ngày 30/11 ===\n")

	query := `
		SELECT 
			pts.id,
			pts.start_time::text,
			pts.end_time::text,
			pts.days_of_week,
			EXTRACT(ISODOW FROM '2025-11-30'::date)::int as day_of_week,
			EXTRACT(ISODOW FROM '2025-11-30'::date)::int = ANY(pts.days_of_week) as is_match
		FROM page_time_slots pts
		JOIN pages p ON p.id = pts.page_id
		WHERE p.page_name = 'Ánh Lê'
			AND pts.is_active = true
		ORDER BY pts.start_time
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotID, startTime, endTime string
		var daysOfWeek []byte
		var dayOfWeek int
		var isMatch bool
		
		rows.Scan(&slotID, &startTime, &endTime, &daysOfWeek, &dayOfWeek, &isMatch)
		
		matchStatus := "❌ KHÔNG MATCH"
		if isMatch {
			matchStatus = "✅ MATCH"
		}
		
		fmt.Printf("%s\n", matchStatus)
		fmt.Printf("  Slot: %s-%s\n", startTime[:5], endTime[:5])
		fmt.Printf("  days_of_week: %s\n", string(daysOfWeek))
		fmt.Printf("  Ngày 30/11 là thứ: %d\n", dayOfWeek)
		fmt.Printf("  ID: %s\n\n", slotID[:8])
	}

	// Test query tìm slot cho ngày 30/11
	fmt.Println("=== Test query tìm slot ngày 30/11 ===\n")

	var pageID string
	db.QueryRow("SELECT id FROM pages WHERE page_name = 'Ánh Lê'").Scan(&pageID)

	query2 := `
		WITH date_check AS (
			SELECT '2025-11-30'::date as check_date
		),
		slot_availability AS (
			SELECT 
				pts.id as slot_id,
				dc.check_date,
				pts.start_time::text,
				pts.end_time::text,
				pts.slot_capacity,
				pts.days_of_week,
				COALESCE(COUNT(sp.id), 0) as used_count
			FROM page_time_slots pts
			CROSS JOIN date_check dc
			LEFT JOIN scheduled_posts sp 
				ON sp.time_slot_id = pts.id 
				AND DATE(sp.scheduled_time) = dc.check_date
				AND sp.status IN ('pending', 'processing')
			WHERE pts.page_id = $1
				AND pts.is_active = true
			GROUP BY pts.id, dc.check_date, pts.start_time, pts.end_time, 
					 pts.slot_capacity, pts.days_of_week
			HAVING COALESCE(COUNT(sp.id), 0) < pts.slot_capacity
			ORDER BY dc.check_date, pts.start_time
		)
		SELECT 
			slot_id,
			check_date,
			start_time,
			end_time,
			slot_capacity,
			used_count,
			EXTRACT(ISODOW FROM check_date)::int as day_of_week,
			EXTRACT(ISODOW FROM check_date)::int = ANY(days_of_week) as is_match
		FROM slot_availability
	`

	rows2, err := db.Query(query2, pageID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	count := 0
	for rows2.Next() {
		var slotID, startTime, endTime string
		var checkDate time.Time
		var capacity, usedCount, dayOfWeek int
		var isMatch bool
		
		rows2.Scan(&slotID, &checkDate, &startTime, &endTime, &capacity, &usedCount, &dayOfWeek, &isMatch)
		count++
		
		matchStatus := "❌"
		if isMatch {
			matchStatus = "✅"
		}
		
		fmt.Printf("%s Slot %s-%s | Used: %d/%d | Day: %d | Match: %v\n",
			matchStatus,
			startTime[:5],
			endTime[:5],
			usedCount,
			capacity,
			dayOfWeek,
			isMatch)
	}

	if count == 0 {
		fmt.Println("❌ Không tìm thấy slot nào!")
	}
}
