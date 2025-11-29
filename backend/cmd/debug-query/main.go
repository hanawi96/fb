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

	// Lấy page Ánh Lê
	var pageID string
	err = db.QueryRow("SELECT id FROM pages WHERE page_name = 'Ánh Lê'").Scan(&pageID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Page ID: %s\n\n", pageID)

	// Test query
	query := `
		WITH RECURSIVE date_series AS (
			SELECT $2::date as check_date
			UNION ALL
			SELECT (check_date + INTERVAL '1 day')::date
			FROM date_series
			WHERE check_date < ($2::date + ($3 || ' days')::interval)::date
		),
		slot_availability AS (
			SELECT 
				pts.id as slot_id,
				ds.check_date,
				pts.start_time::text,
				pts.end_time::text,
				pts.slot_capacity,
				pts.days_of_week,
				COALESCE(COUNT(sp.id), 0) as used_count
			FROM page_time_slots pts
			CROSS JOIN date_series ds
			LEFT JOIN scheduled_posts sp 
				ON sp.time_slot_id = pts.id 
				AND DATE(sp.scheduled_time) = ds.check_date
				AND sp.status IN ('pending', 'processing')
			WHERE pts.page_id = $1 
				AND pts.is_active = true
			GROUP BY pts.id, ds.check_date, pts.start_time, pts.end_time, 
					 pts.slot_capacity, pts.days_of_week, pts.priority
			HAVING COALESCE(COUNT(sp.id), 0) < pts.slot_capacity
			ORDER BY ds.check_date, pts.start_time
		)
		SELECT 
			slot_id,
			check_date,
			start_time,
			end_time,
			slot_capacity,
			used_count,
			days_of_week
		FROM slot_availability
		WHERE EXTRACT(ISODOW FROM check_date)::int = ANY(days_of_week)
		LIMIT 10
	`

	startDate := time.Now().Format("2006-01-02")
	rows, err := db.Query(query, pageID, startDate, 5)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Kết quả query:")
	fmt.Println("─────────────────────────────────────────────────────────")
	
	count := 0
	for rows.Next() {
		var slotID, startTime, endTime string
		var checkDate time.Time
		var capacity, usedCount int
		var daysOfWeek []byte
		
		rows.Scan(&slotID, &checkDate, &startTime, &endTime, &capacity, &usedCount, &daysOfWeek)
		
		count++
		fmt.Printf("%d. Date: %s | Time: %s-%s | Used: %d/%d | Slot: %s\n",
			count,
			checkDate.Format("2006-01-02"),
			startTime[:5],
			endTime[:5],
			usedCount,
			capacity,
			slotID[:8])
	}

	if count == 0 {
		fmt.Println("❌ Không tìm thấy slot nào!")
	}
}
