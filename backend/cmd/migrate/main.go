package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/fbscheduler?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully!")

	// Run migration
	migration := `
-- Create saved_hashtags table
CREATE TABLE IF NOT EXISTS saved_hashtags (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    hashtag VARCHAR(255) NOT NULL,
    media_count BIGINT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, hashtag)
);

CREATE INDEX IF NOT EXISTS idx_saved_hashtags_user_id ON saved_hashtags(user_id);
`

	fmt.Println("Running migration...")
	_, err = db.Exec(migration)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println("Table 'saved_hashtags' created.")
}
