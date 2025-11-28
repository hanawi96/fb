package scheduler

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sort"
	"sync"
	"time"

	"fbscheduler/internal/db"
)

// ============================================
// CONSTANTS
// ============================================

const (
	MinIntervalSameAccountMinutes = 5  // Khoảng cách tối thiểu cùng nick (phút)
	RandomOffsetMinSeconds        = 60 // Random offset tối thiểu (giây)
	RandomOffsetMaxSeconds        = 180 // Random offset tối đa (giây)
)

// ============================================
// TYPES
// ============================================

// ScheduleRequest yêu cầu schedule 1 bài lên nhiều page
type ScheduleRequest struct {
	PostID       string
	PageIDs      []string
	PreferredDate time.Time
	UseTimeSlots bool // true = dùng khung giờ của page, false = dùng thời gian cụ thể
}

// ScheduleResult kết quả schedule cho 1 page
type ScheduleResult struct {
	PageID        string
	PageName      string
	AccountID     string
	AccountName   string
	TimeSlotID    string
	ScheduledTime time.Time
	RandomOffset  int // seconds
	Warning       string // Cảnh báo nếu có (ví dụ: đẩy sang ngày mai)
	Error         error
}

// SchedulePreview preview trước khi schedule
type SchedulePreview struct {
	Results       []ScheduleResult
	TotalPages    int
	SuccessCount  int
	WarningCount  int
	ErrorCount    int
	NextDayCount  int // Số page bị đẩy sang ngày mai
}

// ============================================
// SMART SCHEDULER
// ============================================

// SmartScheduler thuật toán schedule thông minh
type SmartScheduler struct {
	store *db.Store
	mu    sync.Mutex // Lock để tránh race condition
}

// NewSmartScheduler tạo smart scheduler mới
func NewSmartScheduler(store *db.Store) *SmartScheduler {
	return &SmartScheduler{
		store: store,
	}
}

// CalculateSchedule tính toán thời gian đăng cho nhiều page
func (s *SmartScheduler) CalculateSchedule(req ScheduleRequest) (*SchedulePreview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	preview := &SchedulePreview{
		Results:    make([]ScheduleResult, 0, len(req.PageIDs)),
		TotalPages: len(req.PageIDs),
	}

	if len(req.PageIDs) == 0 {
		return preview, nil
	}

	// Bước 1: Thu thập thông tin tất cả pages và time slots
	pageSlots, err := s.collectPageTimeSlots(req.PageIDs, req.PreferredDate)
	if err != nil {
		return nil, err
	}

	// Bước 2: Nhóm pages theo khung giờ chồng lấn
	groups := s.groupPagesByOverlappingSlots(pageSlots)

	// Bước 3: Phân bổ thời gian cho từng nhóm
	for _, group := range groups {
		results := s.distributeTimesInGroup(group, req.PreferredDate)
		preview.Results = append(preview.Results, results...)
	}

	// Bước 4: Tính toán thống kê
	for _, r := range preview.Results {
		if r.Error != nil {
			preview.ErrorCount++
		} else if r.Warning != "" {
			preview.WarningCount++
			if isNextDay(r.ScheduledTime, req.PreferredDate) {
				preview.NextDayCount++
			}
		} else {
			preview.SuccessCount++
		}
	}

	return preview, nil
}

// ============================================
// INTERNAL METHODS
// ============================================

// pageSlotInfo thông tin page và slot
type pageSlotInfo struct {
	PageID     string
	PageName   string
	AccountID  string
	AccountName string
	Slot       *db.PageTimeSlot
	StartTime  time.Time
	EndTime    time.Time
}

// collectPageTimeSlots thu thập thông tin time slots của các pages
func (s *SmartScheduler) collectPageTimeSlots(pageIDs []string, date time.Time) ([]pageSlotInfo, error) {
	var result []pageSlotInfo

	for _, pageID := range pageIDs {
		// Lấy thông tin page
		page, err := s.store.GetPageByID(pageID)
		if err != nil {
			continue
		}

		// Lấy account tốt nhất cho page
		account, err := s.store.GetBestAccountForPage(pageID)
		accountID := ""
		accountName := ""
		if err == nil && account != nil {
			accountID = account.ID
			accountName = account.FbUserName
		}

		// Lấy time slots của page
		slots, err := s.store.GetTimeSlotsByPage(pageID)
		if err != nil || len(slots) == 0 {
			// Page không có time slot, tạo default
			result = append(result, pageSlotInfo{
				PageID:      pageID,
				PageName:    page.PageName,
				AccountID:   accountID,
				AccountName: accountName,
				Slot:        nil,
				StartTime:   time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location()),
				EndTime:     time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, date.Location()),
			})
			continue
		}

		// Tìm slot gần nhất còn trống
		slot := s.findNearestAvailableSlot(slots, date)
		if slot == nil {
			// Không có slot trống hôm nay, thử ngày mai
			nextDay := date.AddDate(0, 0, 1)
			slot = s.findNearestAvailableSlot(slots, nextDay)
			if slot != nil {
				date = nextDay
			}
		}

		if slot != nil {
			startTime, endTime := s.parseSlotTimes(slot, date)
			result = append(result, pageSlotInfo{
				PageID:      pageID,
				PageName:    page.PageName,
				AccountID:   accountID,
				AccountName: accountName,
				Slot:        slot,
				StartTime:   startTime,
				EndTime:     endTime,
			})
		}
	}

	return result, nil
}

// findNearestAvailableSlot tìm slot gần nhất còn trống
func (s *SmartScheduler) findNearestAvailableSlot(slots []db.PageTimeSlot, date time.Time) *db.PageTimeSlot {
	now := time.Now()
	dayOfWeek := int(date.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7 // Sunday = 7
	}

	// Sort theo start_time
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].StartTime < slots[j].StartTime
	})

	for i := range slots {
		slot := &slots[i]

		// Kiểm tra ngày trong tuần
		if !containsInt(slot.DaysOfWeek, dayOfWeek) {
			continue
		}

		// Kiểm tra slot còn trống không
		available, err := s.store.IsSlotAvailable(slot.ID, date)
		if err != nil || !available {
			continue
		}

		// Nếu là hôm nay, kiểm tra thời gian đã qua chưa
		_, endTime := s.parseSlotTimes(slot, date)
		if date.Year() == now.Year() && date.YearDay() == now.YearDay() {
			if endTime.Before(now) {
				continue
			}
		}

		return slot
	}

	return nil
}

// parseSlotTimes parse start/end time từ slot
func (s *SmartScheduler) parseSlotTimes(slot *db.PageTimeSlot, date time.Time) (time.Time, time.Time) {
	startHour, startMin := parseTimeString(slot.StartTime)
	endHour, endMin := parseTimeString(slot.EndTime)

	startTime := time.Date(date.Year(), date.Month(), date.Day(), startHour, startMin, 0, 0, date.Location())
	endTime := time.Date(date.Year(), date.Month(), date.Day(), endHour, endMin, 0, 0, date.Location())

	return startTime, endTime
}

// groupPagesByOverlappingSlots nhóm pages theo khung giờ chồng lấn
func (s *SmartScheduler) groupPagesByOverlappingSlots(pages []pageSlotInfo) [][]pageSlotInfo {
	if len(pages) == 0 {
		return nil
	}

	// Sort theo start time
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].StartTime.Before(pages[j].StartTime)
	})

	var groups [][]pageSlotInfo
	currentGroup := []pageSlotInfo{pages[0]}
	currentEnd := pages[0].EndTime

	for i := 1; i < len(pages); i++ {
		page := pages[i]

		// Kiểm tra có chồng lấn với group hiện tại không
		if page.StartTime.Before(currentEnd) {
			// Có chồng lấn, thêm vào group
			currentGroup = append(currentGroup, page)
			if page.EndTime.After(currentEnd) {
				currentEnd = page.EndTime
			}
		} else {
			// Không chồng lấn, tạo group mới
			groups = append(groups, currentGroup)
			currentGroup = []pageSlotInfo{page}
			currentEnd = page.EndTime
		}
	}

	// Thêm group cuối
	groups = append(groups, currentGroup)

	return groups
}

// distributeTimesInGroup phân bổ thời gian trong 1 nhóm
func (s *SmartScheduler) distributeTimesInGroup(group []pageSlotInfo, preferredDate time.Time) []ScheduleResult {
	results := make([]ScheduleResult, 0, len(group))

	if len(group) == 0 {
		return results
	}

	// Tìm khoảng thời gian chung
	commonStart := group[0].StartTime
	commonEnd := group[0].EndTime

	for _, p := range group {
		if p.StartTime.After(commonStart) {
			commonStart = p.StartTime
		}
		if p.EndTime.Before(commonEnd) {
			commonEnd = p.EndTime
		}
	}

	// Nếu không có khoảng chung, dùng khoảng rộng nhất
	if commonEnd.Before(commonStart) || commonEnd.Equal(commonStart) {
		commonStart = group[0].StartTime
		commonEnd = group[0].EndTime
		for _, p := range group {
			if p.StartTime.Before(commonStart) {
				commonStart = p.StartTime
			}
			if p.EndTime.After(commonEnd) {
				commonEnd = p.EndTime
			}
		}
	}

	// Tính interval
	duration := commonEnd.Sub(commonStart)
	interval := duration / time.Duration(len(group))

	// Đảm bảo interval tối thiểu cho cùng nick
	minInterval := time.Duration(MinIntervalSameAccountMinutes) * time.Minute

	// Nhóm theo account để đảm bảo khoảng cách
	accountLastTime := make(map[string]time.Time)

	for i, page := range group {
		scheduledTime := commonStart.Add(interval * time.Duration(i))

		// Kiểm tra khoảng cách với bài trước cùng nick
		if page.AccountID != "" {
			if lastTime, ok := accountLastTime[page.AccountID]; ok {
				minNextTime := lastTime.Add(minInterval)
				if scheduledTime.Before(minNextTime) {
					scheduledTime = minNextTime
				}
			}
			accountLastTime[page.AccountID] = scheduledTime
		}

		// Thêm random offset
		randomOffset := generateRandomOffset()
		scheduledTime = scheduledTime.Add(time.Duration(randomOffset) * time.Second)

		// Kiểm tra có vượt quá end time không
		warning := ""
		if scheduledTime.After(commonEnd) {
			warning = "Thời gian đăng vượt quá khung giờ"
		}

		// Kiểm tra có phải ngày mai không
		if isNextDay(scheduledTime, preferredDate) {
			warning = "Đẩy sang ngày mai do hết slot"
		}

		slotID := ""
		if page.Slot != nil {
			slotID = page.Slot.ID
		}

		results = append(results, ScheduleResult{
			PageID:        page.PageID,
			PageName:      page.PageName,
			AccountID:     page.AccountID,
			AccountName:   page.AccountName,
			TimeSlotID:    slotID,
			ScheduledTime: scheduledTime,
			RandomOffset:  randomOffset,
			Warning:       warning,
		})
	}

	return results
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// parseTimeString parse "13:00:00" to hour, minute
func parseTimeString(timeStr string) (int, int) {
	hour := 0
	min := 0

	if len(timeStr) >= 5 {
		hour = int(timeStr[0]-'0')*10 + int(timeStr[1]-'0')
		min = int(timeStr[3]-'0')*10 + int(timeStr[4]-'0')
	}

	return hour, min
}

// containsInt kiểm tra slice có chứa int không
func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// generateRandomOffset tạo random offset từ 60-180 giây
func generateRandomOffset() int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(RandomOffsetMaxSeconds-RandomOffsetMinSeconds)))
	if err != nil {
		return RandomOffsetMinSeconds
	}
	return int(n.Int64()) + RandomOffsetMinSeconds
}

// isNextDay kiểm tra có phải ngày mai không
func isNextDay(t time.Time, baseDate time.Time) bool {
	return t.Year() != baseDate.Year() || t.YearDay() != baseDate.YearDay()
}

// ErrNoAvailableSlot lỗi không có slot trống
var ErrNoAvailableSlot = errors.New("no available time slot")
