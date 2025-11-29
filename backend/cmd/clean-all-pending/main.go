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

	result, err := db.Exec("DELETE FROM scheduled_posts WHERE status = 'pending'")
	if err != nil {
		log.Fatal(err)
	}

	rows, _ := result.RowsAffected()
	fmt.Printf("✅ Deleted %d pending posts\n", rows)
}
