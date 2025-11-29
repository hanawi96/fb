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

	fmt.Println("=== Checking scheduled_posts columns ===\n")
	
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'scheduled_posts'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	hasTimeSlotId := false
	for rows.Next() {
		var colName, dataType, nullable string
		err := rows.Scan(&colName, &dataType, &nullable)
		if err != nil {
			log.Fatal(err)
		}
		
		status := "✅"
		if colName == "time_slot_id" {
			hasTimeSlotId = true
			status = "🎯"
		}
		
		fmt.Printf("%s %s (%s) - nullable: %s\n", status, colName, dataType, nullable)
	}
	
	fmt.Println()
	if hasTimeSlotId {
		fmt.Println("✅ Column 'time_slot_id' EXISTS!")
	} else {
		fmt.Println("❌ Column 'time_slot_id' MISSING!")
		fmt.Println("   → Run migration 005_smart_scheduling.sql")
	}
}
