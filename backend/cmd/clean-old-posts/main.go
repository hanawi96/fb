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

	fmt.Println("=== Cleaning old scheduled posts without time_slot_id ===\n")

	result, err := db.Exec(`
		DELETE FROM scheduled_posts 
		WHERE time_slot_id IS NULL 
		AND status IN ('pending', 'processing')
	`)
	if err != nil {
		log.Fatal("Delete failed:", err)
	}

	rowsAffected, _ := result.RowsAffected()
	
	fmt.Printf("✅ Deleted %d old scheduled posts\n", rowsAffected)
	fmt.Println("   → Các bài này không có time_slot_id nên không bị kiểm tra giới hạn")
	fmt.Println("\nBây giờ hệ thống sẽ kiểm tra đúng slot_capacity!")
}
