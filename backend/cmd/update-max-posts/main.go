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

	fmt.Println("=== Updating slot_capacity for existing slots ===\n")

	// Update all existing slots to have slot_capacity = 10
	result, err := db.Exec(`
		UPDATE page_time_slots 
		SET slot_capacity = 10 
		WHERE slot_capacity = 1
	`)
	if err != nil {
		log.Fatal("Update failed:", err)
	}

	rowsAffected, _ := result.RowsAffected()
	
	fmt.Printf("✅ Updated %d time slots\n", rowsAffected)
	fmt.Println("   - Changed slot_capacity from 1 to 10")
	fmt.Println("\nBây giờ mỗi khung giờ chứa 10 bài!")
}
