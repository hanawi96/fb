package scheduler

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sort"
	"sync"
	"time"

	"fbscheduler/internal/config"
	"fbscheduler/internal/db"
)

// ============================================
// CONSTANTS
// ============================================

const (
	MinIntervalSameAccountMinutes = 5   // Khoảng cách tối thiểu cùng nick (phút)
	RandomOffsetMinSeconds        = 60  // Random offset tối thiểu (giây)
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
		} else {
			// Có scheduled time = success (kể cả có warning)
			preview.SuccessCount++
			if r.Warning != "" {
				preview.WarningCount++
				if isNextDay(r.ScheduledTime, req.PreferredDate) {
					preview.NextDayCount++
				}
			}
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
			// Page không có time slot, tạo default (9h-21h Vietnam time)
			dateVN := config.ToVN(date)
			result = append(result, pageSlotInfo{
				PageID:      pageID,
				PageName:    page.PageName,
				AccountID:   accountID,
				AccountName: accountName,
				Slot:        nil,
				StartTime:   time.Date(dateVN.Year(), dateVN.Month(), dateVN.Day(), 9, 0, 0, 0, config.VietnamTZ),
				EndTime:     time.Date(dateVN.Year(), dateVN.Month(), dateVN.Day(), 21, 0, 0, 0, config.VietnamTZ),
			})
			continue
		}

		// Sử dụng query tối ưu để tìm slot trống tiếp theo
		// Bắt đầu từ preferred date (không cần check bài muộn nhất)
		startDate := date
		nowVN := config.NowVN()
		if startDate.Before(nowVN) {
			startDate = nowVN
		}

		// Tìm slot trống tiếp theo bằng 1 query (thay vì loop 30 lần)
		slotResult, err := s.store.FindNextAvailableSlot(pageID, startDate, 30)
		
		if err == nil && slotResult != nil {
			// Tìm được slot trống
			slot, err := s.store.GetTimeSlotByID(slotResult.SlotID)
			if err == nil {
				startTime, endTime := s.parseSlotTimes(slot, slotResult.Date)
				result = append(result, pageSlotInfo{
					PageID:      pageID,
					PageName:    page.PageName,
					AccountID:   accountID,
					AccountName: accountName,
					Slot:        slot,
					StartTime:   startTime,
					EndTime:     endTime,
				})
				continue
			}
		}
		
		// Không tìm được slot trống, thêm vào với error
		result = append(result, pageSlotInfo{
			PageID:      pageID,
			PageName:    page.PageName,
			AccountID:   accountID,
			AccountName: accountName,
			Slot:        nil,
			StartTime:   time.Time{}, // Zero time để đánh dấu lỗi
			EndTime:     time.Time{},
		})
	}

	return result, nil
}

// findNearestAvailableSlot tìm slot gần nhất còn trống trong 1 ngày cụ thể
func (s *SmartScheduler) findNearestAvailableSlot(slots []db.PageTimeSlot, date time.Time) *db.PageTimeSlot {
	// Sử dụng thời gian Vietnam để so sánh
	nowVN := config.NowVN()
	dateVN := config.ToVN(date)
	
	dayOfWeek := int(dateVN.Weekday())
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
		if dateVN.Year() == nowVN.Year() && dateVN.YearDay() == nowVN.YearDay() {
			if endTime.Before(nowVN) {
				continue
			}
		}

		return slot
	}

	return nil
}

// parseSlotTimes parse start/end time từ slot
// Sử dụng Vietnam timezone vì khung giờ được setup theo giờ Việt Nam
func (s *SmartScheduler) parseSlotTimes(slot *db.PageTimeSlot, date time.Time) (time.Time, time.Time) {
	startHour, startMin := parseTimeString(slot.StartTime)
	endHour, endMin := parseTimeString(slot.EndTime)

	// Chuyển date sang Vietnam timezone
	dateInVN := config.ToVN(date)

	// Tạo thời gian theo Vietnam timezone
	startTime := time.Date(dateInVN.Year(), dateInVN.Month(), dateInVN.Day(), startHour, startMin, 0, 0, config.VietnamTZ)
	endTime := time.Date(dateInVN.Year(), dateInVN.Month(), dateInVN.Day(), endHour, endMin, 0, 0, config.VietnamTZ)

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
// Thuật toán mới: Phân bổ ngẫu nhiên rải đều trong toàn bộ khung giờ
func (s *SmartScheduler) distributeTimesInGroup(group []pageSlotInfo, preferredDate time.Time) []ScheduleResult {
	results := make([]ScheduleResult, 0, len(group))

	if len(group) == 0 {
		return results
	}
	
	// Kiểm tra và xử lý các page không tìm được slot (StartTime = zero)
	var validPages []pageSlotInfo
	for _, page := range group {
		if page.StartTime.IsZero() {
			// Page không tìm được slot trống, thêm vào result với error
			results = append(results, ScheduleResult{
				PageID:      page.PageID,
				PageName:    page.PageName,
				AccountID:   page.AccountID,
				AccountName: page.AccountName,
				Error:       errors.New("Không tìm được khung giờ trống trong 7 ngày tới"),
			})
		} else {
			validPages = append(validPages, page)
		}
	}
	
	// Nếu không còn page hợp lệ, return luôn
	if len(validPages) == 0 {
		return results
	}
	
	// Tiếp tục xử lý với các page hợp lệ
	group = validPages

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

	// Nhóm pages theo account
	accountPages := make(map[string][]pageSlotInfo)
	for _, page := range group {
		key := page.AccountID
		if key == "" {
			key = "no_account_" + page.PageID // Mỗi page không có account là 1 nhóm riêng
		}
		accountPages[key] = append(accountPages[key], page)
	}

	// Phân bổ thời gian cho từng account
	minInterval := time.Duration(MinIntervalSameAccountMinutes) * time.Minute

	for _, pages := range accountPages {
		accountResults := s.distributeTimesForAccount(pages, commonStart, commonEnd, minInterval, preferredDate)
		results = append(results, accountResults...)
	}

	// Sort kết quả theo thời gian
	sort.Slice(results, func(i, j int) bool {
		return results[i].ScheduledTime.Before(results[j].ScheduledTime)
	})

	return results
}

// distributeTimesForAccount phân bổ thời gian ngẫu nhiên rải đều cho 1 account
// Chia khung giờ thành N vùng, random trong mỗi vùng
func (s *SmartScheduler) distributeTimesForAccount(pages []pageSlotInfo, start, end time.Time, minInterval time.Duration, preferredDate time.Time) []ScheduleResult {
	results := make([]ScheduleResult, 0, len(pages))
	n := len(pages)

	if n == 0 {
		return results
	}

	// Tổng thời gian khung giờ
	totalDuration := end.Sub(start)

	// Chia thành N vùng
	zoneDuration := totalDuration / time.Duration(n)

	// Đảm bảo mỗi vùng đủ lớn cho khoảng cách tối thiểu
	if zoneDuration < minInterval {
		zoneDuration = minInterval
	}

	// Tạo danh sách thời gian cho từng page
	var scheduledTimes []time.Time
	var lastTime time.Time

	for i := 0; i < n; i++ {
		// Tính vùng thời gian cho page này
		zoneStart := start.Add(zoneDuration * time.Duration(i))
		zoneEnd := start.Add(zoneDuration * time.Duration(i+1))

		// Vùng cuối cùng kéo dài đến end
		if i == n-1 {
			zoneEnd = end
		}

		// Đảm bảo zoneStart >= lastTime + minInterval
		if i > 0 && zoneStart.Before(lastTime.Add(minInterval)) {
			zoneStart = lastTime.Add(minInterval)
		}

		// Nếu zoneStart >= zoneEnd, không còn chỗ
		if !zoneStart.Before(zoneEnd) {
			// Đặt ngay sau lastTime + minInterval
			scheduledTimes = append(scheduledTimes, lastTime.Add(minInterval))
			lastTime = lastTime.Add(minInterval)
			continue
		}

		// Random trong vùng [zoneStart, zoneEnd - buffer]
		// Buffer để đảm bảo không quá sát cuối vùng
		buffer := time.Duration(30) * time.Second
		availableDuration := zoneEnd.Sub(zoneStart) - buffer
		if availableDuration < 0 {
			availableDuration = 0
		}

		// Random offset trong vùng
		randomSeconds := secureRandomInt(int(availableDuration.Seconds()))
		scheduledTime := zoneStart.Add(time.Duration(randomSeconds) * time.Second)

		// Đảm bảo khoảng cách tối thiểu với bài trước
		if i > 0 && scheduledTime.Before(lastTime.Add(minInterval)) {
			scheduledTime = lastTime.Add(minInterval)
		}

		scheduledTimes = append(scheduledTimes, scheduledTime)
		lastTime = scheduledTime
	}

	// Tạo kết quả
	for i, page := range pages {
		scheduledTime := scheduledTimes[i]

		warning := ""
		if scheduledTime.After(end) {
			warning = "Thời gian đăng vượt quá khung giờ"
		}
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
			RandomOffset:  0, // Đã random trong vùng, không cần offset thêm
			Warning:       warning,
		})
	}

	return results
}

// secureRandomInt tạo số ngẫu nhiên từ 0 đến max-1
func secureRandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
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
