# THIẾT KẾ HỆ THỐNG SMART SCHEDULING & MULTI-ACCOUNT

## 📋 TỔNG QUAN

Hệ thống quản lý đăng bài Facebook với:
- 500 page hiện tại, mở rộng lên 5000 page
- Nhiều nick Facebook để phân tán, chống spam
- Thuật toán phân bổ thời gian thông minh

---

## PHASE 1: DATABASE & MULTI-ACCOUNT

### 1.1 Bảng `facebook_accounts` (Nick Facebook)
```
- id, user_id
- fb_user_id, fb_user_name
- access_token, token_expires_at
- max_pages (default: 5)
- max_posts_per_day (default: 20)
- status: active | rate_limited | disabled | token_expired
- posts_today, posts_this_hour
- last_used_at, last_error_at
- consecutive_failures
```

### 1.2 Bảng `page_account_assignments`
```
- page_id, account_id
- is_primary (boolean)
- posts_count, last_post_at
```

### 1.3 Cập nhật bảng `scheduled_posts`
```
- account_id (nick nào đăng)
- calculated_time (thời gian sau khi tính toán)
- time_slot_id (khung giờ nào)
- random_offset_seconds (offset ngẫu nhiên)
```

### 1.4 Bảng `notifications`
```
- user_id, type, title, message
- is_read, created_at
```

**Checklist Phase 1:**
- [x] Tạo migration cho các bảng mới (`006_multi_account_system.sql`)
- [x] Tạo indexes cho performance
- [x] Tạo functions: reset daily/hourly counts
- [x] Go models: `db/accounts.go`
- [x] Go models: `db/assignments.go`
- [x] Go models: `db/notifications.go`

---

## PHASE 2: SMART SCHEDULING ALGORITHM

### 2.1 Quy tắc cơ bản
- Mỗi khung giờ/page = tối đa 1 bài
- Chọn khung giờ gần nhất còn trống
- Phân tán đều: `interval = thời_lượng / số_page`
- Cùng nick cách nhau tối thiểu 5 phút
- Random offset ±1-3 phút cho tự nhiên

### 2.2 Xử lý chồng lấn khung giờ
- Xem toàn bộ timeline, không chỉ từng khung riêng
- Ưu tiên phân tán ra các khoảng không chồng lấn

### 2.3 Khi khung giờ đầy
- Đẩy sang khung tiếp theo trong ngày
- Hết khung trong ngày → đẩy sang ngày mai
- Trả về cảnh báo cho user

### 2.4 Queue xử lý
- Lock khi tính toán slot (tránh race condition)
- Xử lý tuần tự các request schedule

**Checklist Phase 2:**
- [x] Go models: `db/timeslots.go` - CRUD time slots
- [x] Algorithm: `scheduler/algorithm.go` - Smart scheduling algorithm
  - [x] `CalculateSchedule()` - Tính toán thời gian cho nhiều pages
  - [x] `collectPageTimeSlots()` - Thu thập time slots
  - [x] `findNearestAvailableSlot()` - Tìm slot gần nhất còn trống
  - [x] `groupPagesByOverlappingSlots()` - Nhóm pages theo khung giờ chồng lấn
  - [x] `distributeTimesInGroup()` - Phân bổ thời gian trong nhóm
- [x] Random offset ±60-180 giây (1-3 phút)
- [x] Service: `scheduler/scheduling_service.go` - Queue/Lock mechanism
  - [x] `SchedulePostToPages()` - Schedule với lock
  - [x] `PreviewSchedule()` - Preview trước khi schedule
  - [x] `ConfirmSchedule()` - Xác nhận và tạo scheduled posts
- [x] Cập nhật `db/store.go` - Thêm method `DB()`

---

## PHASE 3: POSTING ENGINE

### 3.1 Cooldown & Rate Limiting
- Sau mỗi bài cùng nick: chờ 30 giây
- Cùng nick cách nhau tối thiểu 5 phút (schedule)
- Gọi API song song tối đa 3 request/nick

### 3.2 Retry thông minh
- Fail lần 1: chờ 2 phút
- Fail lần 2: chờ 5 phút
- Fail lần 3: dừng, thông báo user

### 3.3 Chọn nick tối ưu
Ưu tiên theo thứ tự:
1. Nick ít bài hôm nay nhất
2. Nick không có lỗi gần đây
3. Nick được gán primary cho page

**Checklist Phase 3:**
- [x] `scheduler/posting_engine.go` - Posting Engine mới
  - [x] Cooldown 30s giữa các bài cùng nick (`waitForCooldown()`)
  - [x] Retry với delay tăng dần (2 phút → 5 phút → dừng)
  - [x] Rate limit detection (`isRateLimitError()`)
  - [x] Semaphore giới hạn 3 concurrent/nick (`getAccountSemaphore()`)
  - [x] Tự động tạo notification khi rate limit/fail
  - [x] Check warning threshold 80%/100%
- [x] `scheduler/scheduler.go` - Cập nhật scheduler
  - [x] Tích hợp PostingEngine
  - [x] Group posts by account
  - [x] Daily reset job (00:00)
  - [x] Graceful shutdown

---

## PHASE 4: BACKEND API

### 4.1 API Quản lý Nick
```
GET    /api/accounts           - Danh sách nick
POST   /api/accounts/connect   - Thêm nick (OAuth)
DELETE /api/accounts/:id       - Xóa nick
GET    /api/accounts/:id/pages - Pages của nick
PUT    /api/accounts/:id       - Cập nhật giới hạn
```

### 4.2 API Gán Page
```
POST   /api/pages/:id/assign   - Gán page vào nick
DELETE /api/pages/:id/assign   - Bỏ gán
```

### 4.3 API Schedule (cập nhật)
```
POST   /api/schedule/preview   - Preview trước khi schedule
POST   /api/schedule           - Tạo schedule (có account_id)
```

### 4.4 API Notifications
```
GET    /api/notifications      - Danh sách thông báo
PUT    /api/notifications/:id/read - Đánh dấu đã đọc
```

**Checklist Phase 4:**
- [x] `api/accounts.go` - CRUD Facebook Accounts
  - [x] GET /api/accounts - Danh sách nick
  - [x] POST /api/accounts - Tạo nick mới
  - [x] GET /api/accounts/:id - Chi tiết nick
  - [x] PUT /api/accounts/:id - Cập nhật nick
  - [x] DELETE /api/accounts/:id - Xóa nick
  - [x] GET /api/accounts/:id/pages - Pages của nick
  - [x] POST /api/accounts/:id/refresh - Refresh token
- [x] `api/assignments.go` - API gán page vào nick
  - [x] GET /api/pages/:id/assignments - Danh sách accounts của page
  - [x] POST /api/pages/:id/assign - Gán page vào account
  - [x] DELETE /api/pages/:id/assign/:accountId - Bỏ gán
  - [x] PUT /api/pages/:id/primary - Đặt primary account
  - [x] GET /api/pages/unassigned - Pages chưa gán
- [x] `api/notifications.go` - API Notifications
  - [x] GET /api/notifications - Danh sách thông báo
  - [x] GET /api/notifications/count - Số chưa đọc
  - [x] PUT /api/notifications/:id/read - Đánh dấu đã đọc
  - [x] PUT /api/notifications/read-all - Đánh dấu tất cả
  - [x] DELETE /api/notifications/:id - Xóa thông báo
- [x] `api/schedule_preview.go` - API Preview Schedule
  - [x] POST /api/schedule/preview - Preview trước khi schedule
  - [x] POST /api/schedule/smart - Schedule với smart algorithm
  - [x] GET /api/schedule/stats - Thống kê schedule
- [x] `api/timeslots.go` - API Time Slots
  - [x] GET /api/pages/:id/timeslots - Danh sách khung giờ
  - [x] POST /api/pages/:id/timeslots - Tạo khung giờ
  - [x] PUT /api/timeslots/:id - Cập nhật khung giờ
  - [x] DELETE /api/timeslots/:id - Xóa khung giờ
- [x] Cập nhật `cmd/server/main.go` - Đăng ký routes mới

---

## PHASE 5: FRONTEND UI

### 5.1 Trang Quản lý Nick (`/accounts`)
Bảng hiển thị:
| Nick | Pages | Hôm nay | Trạng thái | Token | Actions |
|------|-------|---------|------------|-------|---------|
| Nick A | 4/5 | 12/20 | ✅ Active | 25 ngày | Edit/Delete |

Tính năng:
- Nút "Thêm Nick" → OAuth popup
- Xem chi tiết pages của nick
- Cảnh báo màu vàng khi đạt 80%
- Cảnh báo màu đỏ khi đạt 100%

### 5.2 Cập nhật trang Pages
- Hiển thị nick đang quản lý page
- Dropdown chọn nick khi thêm page

### 5.3 Preview Schedule
Modal hiển thị trước khi confirm:
- Danh sách page + thời gian dự kiến
- Nick nào đăng page nào
- Cảnh báo nếu đẩy sang ngày mai

### 5.4 Notifications
- Icon bell với badge số chưa đọc
- Dropdown danh sách thông báo
- Click để đánh dấu đã đọc

**Checklist Phase 5:**
- [x] `lib/api.js` - Thêm 25+ API methods mới
- [x] `routes/accounts/+page.svelte` - Trang quản lý Nick
  - [x] Danh sách nick với stats (pages, bài/ngày, token)
  - [x] Progress bar hiển thị % giới hạn
  - [x] Expand để xem chi tiết
  - [x] Xóa nick
- [x] `lib/components/NotificationBell.svelte` - Component thông báo
  - [x] Badge số chưa đọc
  - [x] Dropdown danh sách thông báo
  - [x] Đánh dấu đã đọc / xóa
  - [x] Auto-poll mỗi 30s
- [x] `lib/components/SchedulePreviewModal.svelte` - Preview schedule
  - [x] Hiển thị thời gian dự kiến từng page
  - [x] Hiển thị nick đăng
  - [x] Cảnh báo nếu đẩy sang ngày mai
  - [x] Xác nhận schedule
- [x] `lib/components/TimeSlotEditor.svelte` - Cấu hình khung giờ
  - [x] Danh sách khung giờ hiện có
  - [x] Thêm khung giờ mới
  - [x] Chọn ngày trong tuần
  - [x] Bật/tắt khung giờ
- [x] Cập nhật `routes/+layout.svelte` - Thêm menu Accounts + NotificationBell
- [x] Cập nhật `routes/pages/+page.svelte` - Thêm nút cấu hình khung giờ

---

## PHASE 6: MONITORING & ALERTS

### 6.1 Cảnh báo tự động
- Nick đạt 80% giới hạn bài/ngày
- Nick đạt 100% giới hạn
- Token còn < 7 ngày
- Nick bị rate limit
- Bài đăng fail 3 lần

### 6.2 Analytics cơ bản
Lưu trữ:
- Số bài thành công/fail theo ngày
- Số lần rate limit theo nick
- Thời gian đăng thực tế vs dự kiến

**Checklist Phase 6:**
- [ ] Background job check cảnh báo
- [ ] Tạo notifications tự động
- [ ] Bảng analytics đơn giản
- [ ] Dashboard hiển thị stats

---

## 📊 TỔNG KẾT PHASES

| Phase | Nội dung | Độ ưu tiên |
|-------|----------|------------|
| 1 | Database & Multi-Account | 🔴 Cao |
| 2 | Smart Scheduling Algorithm | 🔴 Cao |
| 3 | Posting Engine | 🔴 Cao |
| 4 | Backend API | 🔴 Cao |
| 5 | Frontend UI | 🟡 Trung bình |
| 6 | Monitoring & Alerts | 🟡 Trung bình |

---

## 🔧 CẤU HÌNH MẶC ĐỊNH

```
MAX_PAGES_PER_ACCOUNT = 5
MAX_POSTS_PER_ACCOUNT_PER_DAY = 20
MIN_INTERVAL_SAME_ACCOUNT = 5 phút
COOLDOWN_AFTER_POST = 30 giây
RANDOM_OFFSET_RANGE = ±1-3 phút
RETRY_DELAYS = [2 phút, 5 phút, dừng]
WARNING_THRESHOLD = 80%
MAX_CONCURRENT_POSTS_PER_ACCOUNT = 3
```

---

## 📝 GHI CHÚ

- User tự phân bổ page vào nick trước khi đăng nhập
- Khi rate limit → thông báo, user xử lý thủ công
- Preview là optional, có thể bỏ qua để đăng nhanh
