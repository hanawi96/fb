package db

import (
	"time"
)

// ============================================
// MODELS
// ============================================

// PageTimeSlot khung giờ đăng bài của page
type PageTimeSlot struct {
	ID              string    `json:"id"`
	PageID          string    `json:"page_id"`
	SlotName        string    `json:"slot_name"`
	StartTime       string    `json:"start_time"` // "13:00:00"
	EndTime         string    `json:"end_time"`   // "15:00:00"
	DaysOfWeek      []int     `json:"days_of_week"`
	IsActive        bool      `json:"is_active"`
	Priority        int       `json:"priority"`
	MaxPostsPerSlot int       `json:"max_posts_per_slot"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ============================================
// TIME SLOTS METHODS
// ============================================

// GetTimeSlotsByPage lấy danh sách khung giờ của 1 page
func (s *Store) GetTimeSlotsByPage(pageID string) ([]PageTimeSlot, error) {
	query := `
		SELECT id, page_id, slot_name, start_time::text, end_time::text,
			days_of_week, is_active, priority, max_posts_per_slot,
			created_at, updated_at
		FROM page_time_slots
		WHERE page_id = $1 AND is_active = true
		ORDER BY start_time
	`

	rows, err := s.db.Query(query, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []PageTimeSlot
	for rows.Next() {
		var slot PageTimeSlot
		var daysOfWeek []byte

		err := rows.Scan(
			&slot.ID, &slot.PageID, &slot.SlotName,
			&slot.StartTime, &slot.EndTime,
			&daysOfWeek, &slot.IsActive, &slot.Priority,
			&slot.MaxPostsPerSlot, &slot.CreatedAt, &slot.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse PostgreSQL array
		slot.DaysOfWeek = parseIntArray(daysOfWeek)
		slots = append(slots, slot)
	}

	return slots, nil
}

// GetTimeSlotByID lấy 1 khung giờ theo ID
func (s *Store) GetTimeSlotByID(id string) (*PageTimeSlot, error) {
	query := `
		SELECT id, page_id, slot_name, start_time::text, end_time::text,
			days_of_week, is_active, priority, max_posts_per_slot,
			created_at, updated_at
		FROM page_time_slots
		WHERE id = $1
	`

	var slot PageTimeSlot
	var daysOfWeek []byte

	err := s.db.QueryRow(query, id).Scan(
		&slot.ID, &slot.PageID, &slot.SlotName,
		&slot.StartTime, &slot.EndTime,
		&daysOfWeek, &slot.IsActive, &slot.Priority,
		&slot.MaxPostsPerSlot, &slot.CreatedAt, &slot.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	slot.DaysOfWeek = parseIntArray(daysOfWeek)
	return &slot, nil
}

// CreateTimeSlot tạo khung giờ mới
func (s *Store) CreateTimeSlot(slot *PageTimeSlot) error {
	query := `
		INSERT INTO page_time_slots (
			page_id, slot_name, start_time, end_time,
			days_of_week, is_active, priority, max_posts_per_slot
		) VALUES ($1, $2, $3::time, $4::time, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	return s.db.QueryRow(
		query,
		slot.PageID, slot.SlotName, slot.StartTime, slot.EndTime,
		intArrayToPostgres(slot.DaysOfWeek), slot.IsActive,
		slot.Priority, slot.MaxPostsPerSlot,
	).Scan(&slot.ID, &slot.CreatedAt, &slot.UpdatedAt)
}

// UpdateTimeSlot cập nhật khung giờ
func (s *Store) UpdateTimeSlot(slot *PageTimeSlot) error {
	query := `
		UPDATE page_time_slots SET
			slot_name = $2,
			start_time = $3::time,
			end_time = $4::time,
			days_of_week = $5,
			is_active = $6,
			priority = $7,
			max_posts_per_slot = $8
		WHERE id = $1
	`

	_, err := s.db.Exec(
		query,
		slot.ID, slot.SlotName, slot.StartTime, slot.EndTime,
		intArrayToPostgres(slot.DaysOfWeek), slot.IsActive,
		slot.Priority, slot.MaxPostsPerSlot,
	)
	return err
}

// DeleteTimeSlot xóa khung giờ
func (s *Store) DeleteTimeSlot(id string) error {
	_, err := s.db.Exec("DELETE FROM page_time_slots WHERE id = $1", id)
	return err
}

// IsSlotAvailable kiểm tra khung giờ còn trống không (cho ngày cụ thể)
func (s *Store) IsSlotAvailable(slotID string, date time.Time) (bool, error) {
	query := `
		SELECT COUNT(*) < pts.max_posts_per_slot
		FROM page_time_slots pts
		LEFT JOIN scheduled_posts sp ON sp.time_slot_id = pts.id
			AND DATE(sp.scheduled_time) = $2
			AND sp.status IN ('pending', 'processing')
		WHERE pts.id = $1
		GROUP BY pts.max_posts_per_slot
	`

	var available bool
	err := s.db.QueryRow(query, slotID, date.Format("2006-01-02")).Scan(&available)
	if err != nil {
		// Nếu không có record nào, slot còn trống
		return true, nil
	}
	return available, nil
}

// GetPostsCountInSlot đếm số bài đã schedule trong slot
func (s *Store) GetPostsCountInSlot(slotID string, date time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM scheduled_posts
		WHERE time_slot_id = $1
			AND DATE(scheduled_time) = $2
			AND status IN ('pending', 'processing', 'completed')
	`

	var count int
	err := s.db.QueryRow(query, slotID, date.Format("2006-01-02")).Scan(&count)
	return count, err
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// parseIntArray parse PostgreSQL integer array
func parseIntArray(data []byte) []int {
	if len(data) == 0 {
		return []int{1, 2, 3, 4, 5, 6, 7}
	}

	// PostgreSQL array format: {1,2,3,4,5,6,7}
	str := string(data)
	if len(str) < 2 {
		return []int{1, 2, 3, 4, 5, 6, 7}
	}

	// Remove { and }
	str = str[1 : len(str)-1]
	if str == "" {
		return []int{1, 2, 3, 4, 5, 6, 7}
	}

	var result []int
	for _, s := range splitString(str, ',') {
		if n := parseInt(s); n > 0 {
			result = append(result, n)
		}
	}

	if len(result) == 0 {
		return []int{1, 2, 3, 4, 5, 6, 7}
	}
	return result
}

// intArrayToPostgres convert Go slice to PostgreSQL array string
func intArrayToPostgres(arr []int) string {
	if len(arr) == 0 {
		return "{1,2,3,4,5,6,7}"
	}

	result := "{"
	for i, n := range arr {
		if i > 0 {
			result += ","
		}
		result += intToString(n)
	}
	result += "}"
	return result
}

// splitString simple string split
func splitString(s string, sep rune) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == sep {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// parseInt simple string to int
func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

// intToString simple int to string
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
