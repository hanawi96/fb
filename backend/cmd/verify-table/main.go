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

	// Check if table exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'saved_hashtags'
		)
	`).Scan(&exists)

	if err != nil {
		log.Fatal(err)
	}

	if exists {
		fmt.Println("✅ Table 'saved_hashtags' exists!")
		
		// Get table structure
		rows, err := db.Query(`
			SELECT column_name, data_type 
			FROM information_schema.columns 
			WHERE table_name = 'saved_hashtags'
			ORDER BY ordinal_position
		`)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		fmt.Println("\nTable structure:")
		for rows.Next() {
			var colName, dataType string
			rows.Scan(&colName, &dataType)
			fmt.Printf("  - %s (%s)\n", colName, dataType)
		}
	} else {
		fmt.Println("❌ Table 'saved_hashtags' does not exist!")
	}
}
