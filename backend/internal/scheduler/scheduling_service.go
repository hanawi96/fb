package scheduler

import (
	"sync"
	"time"

	"fbscheduler/internal/db"
)

// ============================================
// SCHEDULING SERVICE
// Xử lý queue và lock khi schedule nhiều bài
// ============================================

// SchedulingService service quản lý việc schedule
type SchedulingService struct {
	store     *db.Store
	algorithm *SmartScheduler
	mu        sync.Mutex
}

// NewSchedulingService tạo scheduling service mới
func NewSchedulingService(store *db.Store) *SchedulingService {
	return &SchedulingService{
		store:     store,
		algorithm: NewSmartScheduler(store),
	}
}

// SchedulePostToPages schedule 1 bài lên nhiều pages
func (s *SchedulingService) SchedulePostToPages(postID string, pageIDs []string, preferredDate time.Time) (*SchedulePreview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Bước 1: Tính toán thời gian
	req := ScheduleRequest{
		PostID:        postID,
		PageIDs:       pageIDs,
		PreferredDate: preferredDate,
		UseTimeSlots:  true,
	}

	preview, err := s.algorithm.CalculateSchedule(req)
	if err != nil {
		return nil, err
	}

	return preview, nil
}

// ConfirmSchedule xác nhận và tạo scheduled posts
func (s *SchedulingService) ConfirmSchedule(postID string, results []ScheduleResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range results {
		if r.Error != nil {
			continue
		}

		// Tạo scheduled post
		sp := &db.ScheduledPost{
			PostID:        postID,
			PageID:        r.PageID,
			ScheduledTime: r.ScheduledTime,
			Status:        "pending",
			MaxRetries:    3,
		}

		err := s.store.CreateScheduledPost(sp)
		if err != nil {
			return err
		}

		// Cập nhật thêm thông tin account và time slot
		if r.AccountID != "" || r.TimeSlotID != "" {
			err = s.updateScheduledPostDetails(sp.ID, r.AccountID, r.TimeSlotID, r.RandomOffset)
			if err != nil {
				// Log error but continue
				continue
			}
		}
	}

	return nil
}

// updateScheduledPostDetails cập nhật chi tiết scheduled post
func (s *SchedulingService) updateScheduledPostDetails(spID, accountID, timeSlotID string, randomOffset int) error {
	query := `
		UPDATE scheduled_posts 
		SET account_id = NULLIF($2, ''),
			time_slot_id = NULLIF($3, ''),
			random_offset_seconds = $4,
			calculated_time = scheduled_time
		WHERE id = $1
	`

	_, err := s.store.DB().Exec(query, spID, accountID, timeSlotID, randomOffset)
	return err
}

// PreviewSchedule chỉ preview, không tạo scheduled posts
func (s *SchedulingService) PreviewSchedule(postID string, pageIDs []string, preferredDate time.Time) (*SchedulePreview, error) {
	req := ScheduleRequest{
		PostID:        postID,
		PageIDs:       pageIDs,
		PreferredDate: preferredDate,
		UseTimeSlots:  true,
	}

	return s.algorithm.CalculateSchedule(req)
}

// SchedulePostToSinglePage schedule 1 bài lên 1 page với thời gian cụ thể
func (s *SchedulingService) SchedulePostToSinglePage(postID, pageID string, scheduledTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Lấy account tốt nhất
	account, _ := s.store.GetBestAccountForPage(pageID)
	accountID := ""
	if account != nil {
		accountID = account.ID
	}

	// Thêm random offset
	randomOffset := generateRandomOffset()
	adjustedTime := scheduledTime.Add(time.Duration(randomOffset) * time.Second)

	// Tạo scheduled post
	sp := &db.ScheduledPost{
		PostID:        postID,
		PageID:        pageID,
		ScheduledTime: adjustedTime,
		Status:        "pending",
		MaxRetries:    3,
	}

	err := s.store.CreateScheduledPost(sp)
	if err != nil {
		return err
	}

	// Cập nhật account
	if accountID != "" {
		s.updateScheduledPostDetails(sp.ID, accountID, "", randomOffset)
	}

	return nil
}

// GetScheduleStats lấy thống kê schedule
func (s *SchedulingService) GetScheduleStats(date time.Time) (*ScheduleStats, error) {
	stats := &ScheduleStats{
		Date: date,
	}

	// Đếm số bài pending
	pending, err := s.countScheduledPostsByStatus(date, "pending")
	if err != nil {
		return nil, err
	}
	stats.PendingCount = pending

	// Đếm số bài completed
	completed, err := s.countScheduledPostsByStatus(date, "completed")
	if err != nil {
		return nil, err
	}
	stats.CompletedCount = completed

	// Đếm số bài failed
	failed, err := s.countScheduledPostsByStatus(date, "failed")
	if err != nil {
		return nil, err
	}
	stats.FailedCount = failed

	stats.TotalCount = pending + completed + failed

	return stats, nil
}

// countScheduledPostsByStatus đếm số bài theo status
func (s *SchedulingService) countScheduledPostsByStatus(date time.Time, status string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM scheduled_posts 
		WHERE DATE(scheduled_time) = $1 AND status = $2
	`

	var count int
	err := s.store.DB().QueryRow(query, date.Format("2006-01-02"), status).Scan(&count)
	return count, err
}

// ScheduleStats thống kê schedule
type ScheduleStats struct {
	Date           time.Time `json:"date"`
	TotalCount     int       `json:"total_count"`
	PendingCount   int       `json:"pending_count"`
	CompletedCount int       `json:"completed_count"`
	FailedCount    int       `json:"failed_count"`
}
