package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := "postgresql://postgres:yendev96@localhost:5432/fbscheduler?sslmode=disable"

	fmt.Println("Connecting to database...")

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Run migration 002
	migration := `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (username, password_hash) 
VALUES ('admin', '$2a$10$VT5rjDMF5nTuON1h8epin.eZpeaWwFruCfM6KddqBeKRYkHatwQgW')
ON CONFLICT (username) DO NOTHING;
`

	_, err = db.Exec(migration)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println("✅ Default user created: admin / admin123")
}
