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

	fmt.Println("=== Renaming column: max_posts_per_slot → slot_capacity ===\n")

	// Rename column
	_, err = db.Exec(`
		ALTER TABLE page_time_slots 
		RENAME COLUMN max_posts_per_slot TO slot_capacity;
	`)
	if err != nil {
		log.Fatal("Rename failed:", err)
	}

	fmt.Println("✅ Column renamed successfully!")
	fmt.Println("   max_posts_per_slot → slot_capacity")
}
