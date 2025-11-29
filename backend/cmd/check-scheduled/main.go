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

	fmt.Println("=== Checking scheduled posts ===\n")

	rows, err := db.Query(`
		SELECT 
			sp.id,
			p.page_name,
			sp.scheduled_time,
			sp.time_slot_id,
			sp.status
		FROM scheduled_posts sp
		LEFT JOIN pages p ON p.id = sp.page_id
		WHERE sp.status IN ('pending', 'processing')
		ORDER BY sp.scheduled_time
		LIMIT 20
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	withSlot := 0
	withoutSlot := 0

	for rows.Next() {
		var id, pageName, status string
		var scheduledTime string
		var timeSlotID sql.NullString

		err := rows.Scan(&id, &pageName, &scheduledTime, &timeSlotID, &status)
		if err != nil {
			log.Fatal(err)
		}

		count++
		slotInfo := "❌ NO SLOT"
		if timeSlotID.Valid {
			slotInfo = "✅ " + timeSlotID.String[:8]
			withSlot++
		} else {
			withoutSlot++
		}

		fmt.Printf("%s | %s | %s | %s\n",
			scheduledTime[:16], pageName, slotInfo, status)
	}

	fmt.Printf("\n📊 Tổng cộng: %d bài\n", count)
	fmt.Printf("   ✅ Có time_slot_id: %d bài\n", withSlot)
	fmt.Printf("   ❌ KHÔNG có time_slot_id: %d bài\n", withoutSlot)

	if withoutSlot > 0 {
		fmt.Println("\n⚠️ CÁC BÀI KHÔNG CÓ time_slot_id SẼ KHÔNG BỊ KIỂM TRA GIỚI HẠN!")
	}
}
