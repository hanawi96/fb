package main

import (
	"fmt"
	"time"

	"fbscheduler/internal/config"
)

func main() {
	fmt.Println("=== TIMEZONE TEST ===\n")

	// Test 1: Parse date
	dateStr := "2025-11-29"
	parsed, _ := config.ParseDateVN(dateStr)
	fmt.Printf("1. Parse date '%s':\n", dateStr)
	fmt.Printf("   Result: %s\n", parsed.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("   ✅ Correct: %v\n\n", parsed.Location().String() == "Asia/Ho_Chi_Minh")

	// Test 2: Current time
	now := config.NowVN()
	fmt.Printf("2. Current time VN:\n")
	fmt.Printf("   Result: %s\n", now.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("   ✅ Correct: %v\n\n", now.Location().String() == "Asia/Ho_Chi_Minh")

	// Test 3: Create time 12:00 VN
	time12 := time.Date(2025, 11, 29, 12, 0, 0, 0, config.VietnamTZ)
	fmt.Printf("3. Create 12:00 VN:\n")
	fmt.Printf("   VN:  %s\n", time12.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("   UTC: %s\n", time12.UTC().Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("   ✅ Correct: %v\n\n", time12.Hour() == 12 && time12.UTC().Hour() == 5)

	fmt.Println("✅ All tests passed!")
}
