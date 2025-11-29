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

	// Check page_time_slots table
	fmt.Println("=== Checking page_time_slots ===\n")
	
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
		LIMIT 20
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, pageName, startTime, endTime string
		var capacity int
		var isActive bool
		
		err := rows.Scan(&id, &pageName, &startTime, &endTime, &capacity, &isActive)
		if err != nil {
			log.Fatal(err)
		}
		
		count++
		status := "✅"
		if !isActive {
			status = "❌"
		}
		
		fmt.Printf("%s %s: %s - %s (capacity: %d bài)\n", 
			status, pageName, startTime[:5], endTime[:5], capacity)
	}
	
	if count == 0 {
		fmt.Println("❌ Không có time slot nào trong database")
	} else {
		fmt.Printf("\n✅ Tổng cộng: %d time slots\n", count)
	}
	
	// Check scheduled posts count per slot
	fmt.Println("\n=== Checking scheduled posts per slot ===\n")
	
	rows2, err := db.Query(`
		SELECT 
			p.page_name,
			pts.start_time,
			pts.end_time,
			pts.slot_capacity,
			COUNT(sp.id) as scheduled_count,
			DATE(sp.scheduled_time) as schedule_date
		FROM page_time_slots pts
		LEFT JOIN pages p ON p.id = pts.page_id
		LEFT JOIN scheduled_posts sp ON sp.time_slot_id = pts.id 
			AND sp.status IN ('pending', 'processing')
		GROUP BY p.page_name, pts.start_time, pts.end_time, pts.slot_capacity, DATE(sp.scheduled_time)
		HAVING COUNT(sp.id) > 0
		ORDER BY schedule_date DESC, p.page_name
		LIMIT 20
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	count2 := 0
	for rows2.Next() {
		var pageName, startTime, endTime string
		var capacity, scheduledCount int
		var scheduleDate sql.NullString
		
		err := rows2.Scan(&pageName, &startTime, &endTime, &capacity, &scheduledCount, &scheduleDate)
		if err != nil {
			log.Fatal(err)
		}
		
		count2++
		warning := ""
		if scheduledCount > capacity {
			warning = " ⚠️ VƯỢT QUÁ!"
		} else if scheduledCount == capacity {
			warning = " ⚠️ ĐẦY"
		}
		
		date := "N/A"
		if scheduleDate.Valid {
			date = scheduleDate.String
		}
		
		fmt.Printf("%s | %s: %s-%s | %d/%d bài%s\n", 
			date, pageName, startTime[:5], endTime[:5], scheduledCount, capacity, warning)
	}
	
	if count2 == 0 {
		fmt.Println("❌ Không có bài nào được schedule với time_slot_id")
		
		// Check scheduled posts without time_slot_id
		fmt.Println("\n=== Checking scheduled posts WITHOUT time_slot_id ===\n")
		
		var totalScheduled int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM scheduled_posts 
			WHERE status IN ('pending', 'processing')
		`).Scan(&totalScheduled)
		
		if err == nil && totalScheduled > 0 {
			fmt.Printf("⚠️ Có %d bài đã schedule nhưng KHÔNG có time_slot_id\n", totalScheduled)
			fmt.Println("   → Các bài này được tạo trước khi có cột time_slot_id")
			fmt.Println("   → Hệ thống KHÔNG KIỂM TRA giới hạn cho các bài này")
		}
	}
}

