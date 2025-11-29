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

	fmt.Println("=== Kiểm tra lịch của Ánh Lê ngày 29/11 ===\n")

	query := `
		SELECT 
			sp.scheduled_time,
			pts.start_time::text,
			pts.end_time::text,
			sp.status,
			pts.slot_capacity,
			sp.time_slot_id
		FROM scheduled_posts sp
		JOIN pages p ON p.id = sp.page_id
		LEFT JOIN page_time_slots pts ON pts.id = sp.time_slot_id
		WHERE p.page_name = 'Ánh Lê' 
			AND DATE(sp.scheduled_time) = '2025-11-29'
		ORDER BY sp.scheduled_time
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var scheduledTime, startTime, endTime, status, timeSlotID string
		var capacity int
		
		rows.Scan(&scheduledTime, &startTime, &endTime, &status, &capacity, &timeSlotID)
		count++
		
		fmt.Printf("%d. %s | Slot: %s-%s (cap:%d) | Status: %s | SlotID: %s\n",
			count,
			scheduledTime[:16],
			startTime[:5],
			endTime[:5],
			capacity,
			status,
			timeSlotID[:8])
	}

	if count == 0 {
		fmt.Println("❌ Không có bài nào được schedule")
	} else {
		fmt.Printf("\n✅ Tổng: %d bài\n", count)
	}

	// Kiểm tra slot availability
	fmt.Println("\n=== Kiểm tra slot availability ===\n")

	query2 := `
		SELECT 
			pts.id,
			pts.start_time::text,
			pts.end_time::text,
			pts.slot_capacity,
			COUNT(sp.id) as used_count
		FROM page_time_slots pts
		JOIN pages p ON p.id = pts.page_id
		LEFT JOIN scheduled_posts sp 
			ON sp.time_slot_id = pts.id 
			AND DATE(sp.scheduled_time) = '2025-11-29'
			AND sp.status IN ('pending', 'processing')
		WHERE p.page_name = 'Ánh Lê'
			AND pts.is_active = true
		GROUP BY pts.id, pts.start_time, pts.end_time, pts.slot_capacity
		ORDER BY pts.start_time
	`

	rows2, _ := db.Query(query2)
	defer rows2.Close()

	for rows2.Next() {
		var slotID, startTime, endTime string
		var capacity, usedCount int
		
		rows2.Scan(&slotID, &startTime, &endTime, &capacity, &usedCount)
		
		status := "✅ TRỐNG"
		if usedCount >= capacity {
			status = "🔴 ĐẦY"
		}
		
		fmt.Printf("%s Slot %s-%s: %d/%d | ID: %s\n",
			status,
			startTime[:5],
			endTime[:5],
			usedCount,
			capacity,
			slotID[:8])
	}
}
