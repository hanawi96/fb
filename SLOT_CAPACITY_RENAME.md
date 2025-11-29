# Đổi tên: max_posts_per_slot → slot_capacity

## 📋 Tổng quan

Đã đổi tên và thay đổi ý nghĩa của trường từ **"Max Posts"** (giới hạn tối đa) sang **"Slot Capacity"** (số bài trong khung giờ).

## 🔄 Thay đổi ý nghĩa

### Trước (Max Posts):
```
Khung 12h-13h: Max = 10 bài
→ Có thể đăng: 0, 1, 2, ... 10 bài (tùy ý)
→ Vượt 10 → Báo lỗi
```

### Sau (Slot Capacity):
```
Khung 12h-13h: Capacity = 5 bài
→ Khung giờ này chứa 5 bài
→ Ít hơn 5 → Đăng hết bài có
→ Nhiều hơn 5 → Chuyển bài thừa sang khung tiếp theo
```

## ✅ Những gì đã làm

### 1. Database
```sql
ALTER TABLE page_time_slots 
RENAME COLUMN max_posts_per_slot TO slot_capacity;
```

### 2. Backend Models (Go)

**File: `backend/internal/db/timeslots.go`**
```go
type PageTimeSlot struct {
    // ...
    SlotCapacity int `json:"slot_capacity"` // Đổi từ MaxPostsPerSlot
    // ...
}
```

Cập nhật tất cả queries:
- `GetTimeSlotsByPage()`
- `GetTimeSlotByID()`
- `CreateTimeSlot()`
- `UpdateTimeSlot()`
- `IsSlotAvailable()`

**File: `backend/internal/api/timeslots.go`**
```go
// Request struct
type CreateTimeSlotRequest struct {
    SlotCapacity int `json:"slot_capacity"` // Đổi từ max_posts_per_slot
}
```

### 3. Frontend UI

**File: `frontend/src/routes/timeslots/+page.svelte`**

Thay đổi:
- Tất cả `max_posts_per_slot` → `slot_capacity`
- Label: "Max:" → "Số bài:"

**Chế độ "Hằng ngày":**
```svelte
<label>Số bài:</label>
<input type="number" bind:value={slot.slot_capacity} />
<span>bài</span>
```

**Chế độ "Tùy chỉnh theo ngày":**
```svelte
<label>Số bài:</label>
<input type="number" bind:value={slot.slot_capacity} />
```

## 🎯 Logic mới

### Tình huống 1: Đăng ít hơn capacity
```
Khung 12h-13h: Capacity = 5 bài
User đăng 3 bài
→ Đăng hết 3 bài vào khung 12h-13h
→ Không báo lỗi
```

### Tình huống 2: Đăng nhiều hơn capacity
```
Khung 12h-13h: Capacity = 5 bài
User đăng 7 bài
→ Đăng 5 bài vào khung 12h-13h
→ 2 bài còn lại chuyển sang khung tiếp theo
```

### Tình huống 3: Nhiều khung giờ
```
Page có 3 khung:
- 9h-10h: Capacity = 3 bài
- 12h-13h: Capacity = 5 bài
- 18h-19h: Capacity = 2 bài

User đăng 8 bài:
→ 9h-10h: 3 bài (đầy)
→ 12h-13h: 5 bài (đầy)
→ 18h-19h: 0 bài (hết bài)
```

## 📊 UI Preview

### Trước:
```
[12:00 ▼] → [13:00 ▼]  Max: [10] bài  [🗑️]
```

### Sau:
```
[12:00 ▼] → [13:00 ▼]  Số bài: [10]  [🗑️]
```

## 🔧 API Changes

### Request (Create/Update):
```json
{
  "start_time": "12:00",
  "end_time": "13:00",
  "slot_capacity": 10  // ← Đổi từ max_posts_per_slot
}
```

### Response:
```json
{
  "id": "...",
  "start_time": "12:00:00",
  "end_time": "13:00:00",
  "slot_capacity": 10  // ← Đổi từ max_posts_per_slot
}
```

## ✅ Verification

Chạy lệnh kiểm tra:
```bash
cd backend
go run cmd/verify-rename/main.go
```

Kết quả mong đợi:
```
✅ Column 'slot_capacity' EXISTS!
✅ Old column 'max_posts_per_slot' removed
```

## 📝 Files đã thay đổi

### Backend:
1. `backend/internal/db/timeslots.go` - Models & queries
2. `backend/internal/api/timeslots.go` - API handlers
3. `backend/cmd/rename-column/main.go` - Migration script
4. `backend/cmd/verify-rename/main.go` - Verification script

### Frontend:
1. `frontend/src/routes/timeslots/+page.svelte` - UI & logic

### Database:
1. Column renamed: `max_posts_per_slot` → `slot_capacity`

## 🚀 Bước tiếp theo

Cần cập nhật thuật toán scheduling (`backend/internal/scheduler/algorithm.go`) để:
1. Đọc `slot_capacity` thay vì `max_posts_per_slot`
2. Phân bổ bài theo logic mới:
   - Lấp đầy khung giờ theo thứ tự
   - Chuyển bài thừa sang khung tiếp theo
   - Không báo lỗi khi ít hơn capacity

---

**Ngày cập nhật:** 29/11/2024  
**Trạng thái:** ✅ Hoàn thành (Database + Backend + Frontend)  
**Còn lại:** Cập nhật thuật toán scheduling
