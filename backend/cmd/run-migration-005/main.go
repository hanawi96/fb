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

	fmt.Println("=== Running Migration 005: Smart Scheduling ===\n")

	// Chỉ chạy phần ALTER TABLE để thêm cột time_slot_id
	migration := `
-- Thêm cột vào scheduled_posts để hỗ trợ smart scheduling
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS time_slot_id UUID REFERENCES page_time_slots(id) ON DELETE SET NULL;
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS auto_scheduled BOOLEAN DEFAULT false;
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS original_time TIMESTAMP;
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS adjusted_time TIMESTAMP;
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS adjustment_reason TEXT;

-- Index
CREATE INDEX IF NOT EXISTS idx_scheduled_posts_time_slot ON scheduled_posts(time_slot_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_posts_content_hash ON scheduled_posts(content_hash);
`

	_, err = db.Exec(migration)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("✅ Migration 005 completed successfully!")
	fmt.Println("   - Added time_slot_id column")
	fmt.Println("   - Added content_hash column")
	fmt.Println("   - Added auto_scheduled column")
	fmt.Println("   - Added original_time column")
	fmt.Println("   - Added adjusted_time column")
	fmt.Println("   - Added adjustment_reason column")
	fmt.Println("   - Created indexes")
}
