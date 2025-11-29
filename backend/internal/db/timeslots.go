package db

import (
	"time"
	
	"github.com/lib/pq"
)

// ============================================
// MODELS
// ============================================

// PageTimeSlot khung giờ đăng bài của page
type PageTimeSlot struct {
	ID           string    `json:"id"`
	PageID       string    `json:"page_id"`
	SlotName     string    `json:"slot_name"`
	StartTime    string    `json:"start_time"` // "13:00:00"
	EndTime      string    `json:"end_time"`   // "15:00:00"
	DaysOfWeek   []int     `json:"days_of_week"`
	IsActive     bool      `json:"is_active"`
	Priority     int       `json:"priority"`
	SlotCapacity int       `json:"slot_capacity"` // Số bài trong khung giờ này
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ============================================
// TIME SLOTS METHODS
// ============================================

// GetTimeSlotsByPage lấy danh sách khung giờ của 1 page
func (s *Store) GetTimeSlotsByPage(pageID string) ([]PageTimeSlot, error) {
	query := `
		SELECT id, page_id, slot_name, start_time::text, end_time::text,
			days_of_week, is_active, priority, slot_capacity,
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
			&slot.SlotCapacity, &slot.CreatedAt, &slot.UpdatedAt,
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
			days_of_week, is_active, priority, slot_capacity,
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
		&slot.SlotCapacity, &slot.CreatedAt, &slot.UpdatedAt,
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
			days_of_week, is_active, priority, slot_capacity
		) VALUES ($1, $2, $3::time, $4::time, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	return s.db.QueryRow(
		query,
		slot.PageID, slot.SlotName, slot.StartTime, slot.EndTime,
		intArrayToPostgres(slot.DaysOfWeek), slot.IsActive,
		slot.Priority, slot.SlotCapacity,
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
			slot_capacity = $8
		WHERE id = $1
	`

	_, err := s.db.Exec(
		query,
		slot.ID, slot.SlotName, slot.StartTime, slot.EndTime,
		intArrayToPostgres(slot.DaysOfWeek), slot.IsActive,
		slot.Priority, slot.SlotCapacity,
	)
	return err
}

// DeleteTimeSlot xóa khung giờ
func (s *Store) DeleteTimeSlot(id string) error {
	_, err := s.db.Exec("DELETE FROM page_time_slots WHERE id = $1", id)
	return err
}

// IsSlotAvailable kiểm tra khung giờ còn chỗ không (cho ngày cụ thể)
// Logic mới: slot_capacity = số bài trong khung giờ
// Trả về true nếu số bài hiện tại < slot_capacity
func (s *Store) IsSlotAvailable(slotID string, date time.Time) (bool, error) {
	query := `
		SELECT 
			COALESCE(COUNT(sp.id), 0) as current_count,
			pts.slot_capacity
		FROM page_time_slots pts
		LEFT JOIN scheduled_posts sp ON sp.time_slot_id = pts.id
			AND DATE(sp.scheduled_time) = $2
			AND sp.status IN ('pending', 'processing')
		WHERE pts.id = $1
		GROUP BY pts.slot_capacity
	`

	var currentCount, capacity int
	err := s.db.QueryRow(query, slotID, date.Format("2006-01-02")).Scan(&currentCount, &capacity)
	if err != nil {
		// Nếu không có record nào, slot còn trống
		return true, nil
	}
	
	// Còn chỗ nếu số bài hiện tại < capacity
	return currentCount < capacity, nil
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

// GetSlotRemainingCapacity lấy số bài còn có thể đăng trong slot
func (s *Store) GetSlotRemainingCapacity(slotID string, date time.Time) (int, error) {
	query := `
		SELECT 
			pts.slot_capacity - COALESCE(COUNT(sp.id), 0) as remaining
		FROM page_time_slots pts
		LEFT JOIN scheduled_posts sp ON sp.time_slot_id = pts.id
			AND DATE(sp.scheduled_time) = $2
			AND sp.status IN ('pending', 'processing')
		WHERE pts.id = $1
		GROUP BY pts.slot_capacity
	`

	var remaining int
	err := s.db.QueryRow(query, slotID, date.Format("2006-01-02")).Scan(&remaining)
	if err != nil {
		// Nếu không có record, trả về capacity của slot
		slot, err := s.GetTimeSlotByID(slotID)
		if err != nil {
			return 0, err
		}
		return slot.SlotCapacity, nil
	}
	
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
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

// ============================================
// OPTIMIZED SLOT FINDING
// ============================================

// NextAvailableSlotResult kết quả tìm slot trống
type NextAvailableSlotResult struct {
	SlotID    string
	Date      time.Time
	StartTime string
	EndTime   string
	Capacity  int
	UsedCount int
}

// FindNextAvailableSlot tìm slot trống tiếp theo cho 1 page (OPTIMIZED)
// Sử dụng 1 query thay vì loop nhiều lần
func (s *Store) FindNextAvailableSlot(pageID string, startDate time.Time, maxDays int) (*NextAvailableSlotResult, error) {
	query := `
		WITH RECURSIVE date_series AS (
			-- Generate series of dates to check
			SELECT $2::date as check_date
			UNION ALL
			SELECT (check_date + INTERVAL '1 day')::date
			FROM date_series
			WHERE check_date < ($2::date + ($3 || ' days')::interval)::date
		),
		slot_availability AS (
			-- Calculate availability for each slot on each date
			SELECT 
				pts.id as slot_id,
				ds.check_date,
				pts.start_time::text,
				pts.end_time::text,
				pts.slot_capacity,
				pts.days_of_week,
				COALESCE(COUNT(sp.id), 0) as used_count
			FROM page_time_slots pts
			CROSS JOIN date_series ds
			LEFT JOIN scheduled_posts sp 
				ON sp.time_slot_id = pts.id 
				AND DATE(sp.scheduled_time) = ds.check_date
				AND sp.status IN ('pending', 'processing')
			WHERE pts.page_id = $1 
				AND pts.is_active = true
			GROUP BY pts.id, ds.check_date, pts.start_time, pts.end_time, 
					 pts.slot_capacity, pts.days_of_week, pts.priority
			HAVING COALESCE(COUNT(sp.id), 0) < pts.slot_capacity
			ORDER BY ds.check_date, pts.start_time
		)
		-- Get first available slot that matches day of week and not in the past
		SELECT 
			slot_id,
			check_date,
			start_time,
			end_time,
			slot_capacity,
			used_count
		FROM slot_availability
		WHERE EXTRACT(ISODOW FROM check_date)::int = ANY(days_of_week)
			AND (
				check_date > CURRENT_DATE
				OR (check_date = CURRENT_DATE AND end_time::time > CURRENT_TIME)
			)
		ORDER BY check_date, start_time
		LIMIT 1
	`

	var result NextAvailableSlotResult
	err := s.db.QueryRow(query, pageID, startDate.Format("2006-01-02"), maxDays).Scan(
		&result.SlotID,
		&result.Date,
		&result.StartTime,
		&result.EndTime,
		&result.Capacity,
		&result.UsedCount,
	)
	
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FindNextAvailableSlotsForPages tìm slot trống cho nhiều pages cùng lúc (BATCH)
func (s *Store) FindNextAvailableSlotsForPages(pageIDs []string, startDate time.Time, maxDays int) (map[string]*NextAvailableSlotResult, error) {
	query := `
		WITH RECURSIVE date_series AS (
			SELECT $2::date as check_date
			UNION ALL
			SELECT (check_date + INTERVAL '1 day')::date
			FROM date_series
			WHERE check_date < ($2::date + ($3 || ' days')::interval)::date
		),
		slot_availability AS (
			SELECT 
				pts.page_id,
				pts.id as slot_id,
				ds.check_date,
				pts.start_time::text,
				pts.end_time::text,
				pts.slot_capacity,
				pts.days_of_week,
				COALESCE(COUNT(sp.id), 0) as used_count,
				ROW_NUMBER() OVER (PARTITION BY pts.page_id ORDER BY ds.check_date, pts.start_time) as rn
			FROM page_time_slots pts
			CROSS JOIN date_series ds
			LEFT JOIN scheduled_posts sp 
				ON sp.time_slot_id = pts.id 
				AND DATE(sp.scheduled_time) = ds.check_date
				AND sp.status IN ('pending', 'processing')
			WHERE pts.page_id = ANY($1)
				AND pts.is_active = true
				AND EXTRACT(ISODOW FROM ds.check_date)::int = ANY(pts.days_of_week)
				AND (
					ds.check_date > CURRENT_DATE
					OR (ds.check_date = CURRENT_DATE AND pts.end_time::time > CURRENT_TIME)
				)
			GROUP BY pts.page_id, pts.id, ds.check_date, pts.start_time, 
					 pts.end_time, pts.slot_capacity, pts.days_of_week
			HAVING COALESCE(COUNT(sp.id), 0) < pts.slot_capacity
		)
		SELECT 
			page_id,
			slot_id,
			check_date,
			start_time,
			end_time,
			slot_capacity,
			used_count
		FROM slot_availability
		WHERE rn = 1
	`

	rows, err := s.db.Query(query, pq.Array(pageIDs), startDate.Format("2006-01-02"), maxDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]*NextAvailableSlotResult)
	for rows.Next() {
		var pageID string
		var result NextAvailableSlotResult
		
		err := rows.Scan(
			&pageID,
			&result.SlotID,
			&result.Date,
			&result.StartTime,
			&result.EndTime,
			&result.Capacity,
			&result.UsedCount,
		)
		if err != nil {
			return nil, err
		}
		
		results[pageID] = &result
	}

	return results, nil
}
