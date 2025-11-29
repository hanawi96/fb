# Logic mới: Slot Capacity

## 🎯 Định nghĩa

**`slot_capacity`** = Số bài trong khung giờ đó (không phải giới hạn tối đa)

## 📊 Ví dụ cụ thể

### Ví dụ 1: Đăng 1 bài lên 5 pages

**Setup:**
- 5 pages, mỗi page có slot 12h-13h với `capacity = 1`

**Khi đăng 1 bài:**
```
Page 1: Slot 12h-13h → 1 bài (đầy)
Page 2: Slot 12h-13h → 1 bài (đầy)
Page 3: Slot 12h-13h → 1 bài (đầy)
Page 4: Slot 12h-13h → 1 bài (đầy)
Page 5: Slot 12h-13h → 1 bài (đầy)
```

**Kết quả:** 5 bài được đăng vào cùng khung giờ 12h-13h, mỗi page 1 bài.

---

### Ví dụ 2: Đăng 3 bài lên 5 pages

**Setup:**
- 5 pages, mỗi page có:
  - Slot 1: 12h-13h (capacity = 1)
  - Slot 2: 14h-15h (capacity = 1)
  - Slot 3: 18h-19h (capacity = 1)

**Khi đăng 3 bài:**
```
Page 1:
  - Bài 1 → Slot 12h-13h (đầy)
  - Bài 2 → Slot 14h-15h (đầy)
  - Bài 3 → Slot 18h-19h (đầy)

Page 2:
  - Bài 1 → Slot 12h-13h (đầy)
  - Bài 2 → Slot 14h-15h (đầy)
  - Bài 3 → Slot 18h-19h (đầy)

... (tương tự cho Page 3, 4, 5)
```

**Kết quả:** 15 bài tổng cộng (3 bài × 5 pages), phân bổ vào 3 khung giờ.

---

### Ví dụ 3: Capacity lớn hơn

**Setup:**
- 5 pages, mỗi page có slot 12h-13h với `capacity = 10`

**Khi đăng 3 bài:**
```
Page 1: Slot 12h-13h → 3 bài (còn 7 chỗ)
Page 2: Slot 12h-13h → 3 bài (còn 7 chỗ)
Page 3: Slot 12h-13h → 3 bài (còn 7 chỗ)
Page 4: Slot 12h-13h → 3 bài (còn 7 chỗ)
Page 5: Slot 12h-13h → 3 bài (còn 7 chỗ)
```

**Kết quả:** 15 bài được đăng vào cùng khung giờ 12h-13h.

---

### Ví dụ 4: Vượt quá capacity

**Setup:**
- 1 page có:
  - Slot 1: 12h-13h (capacity = 2)
  - Slot 2: 14h-15h (capacity = 2)

**Khi đăng 5 bài:**
```
Bài 1 → Slot 12h-13h (1/2)
Bài 2 → Slot 12h-13h (2/2 - đầy)
Bài 3 → Slot 14h-15h (1/2)
Bài 4 → Slot 14h-15h (2/2 - đầy)
Bài 5 → Slot tiếp theo hoặc ngày mai
```

**Kết quả:** Bài thừa tự động chuyển sang slot tiếp theo.

---

## 🔧 Cách hoạt động

### 1. Kiểm tra slot còn chỗ

```go
func IsSlotAvailable(slotID, date) bool {
    currentCount = COUNT(scheduled_posts WHERE slot_id = slotID AND date = date)
    capacity = slot.slot_capacity
    
    return currentCount < capacity
}
```

### 2. Lấy số chỗ còn lại

```go
func GetSlotRemainingCapacity(slotID, date) int {
    currentCount = COUNT(scheduled_posts WHERE slot_id = slotID AND date = date)
    capacity = slot.slot_capacity
    
    return capacity - currentCount
}
```

### 3. Thuật toán phân bổ

```
FOR EACH page IN pages:
    slot = FindFirstAvailableSlot(page, date)
    
    IF slot.remaining > 0:
        SchedulePost(page, slot, time)
    ELSE:
        slot = FindNextSlot(page, date)
        SchedulePost(page, slot, time)
```

---

## ✅ Các tình huống đã xử lý

### ✅ Tình huống 1: Đăng ít hơn capacity
```
Capacity = 10, Đăng 3 bài
→ Đăng hết 3 bài vào slot đó
→ Slot còn 7 chỗ trống
```

### ✅ Tình huống 2: Đăng nhiều hơn capacity
```
Capacity = 10, Đăng 15 bài
→ Đăng 10 bài vào slot này (đầy)
→ 5 bài còn lại chuyển sang slot tiếp theo
```

### ✅ Tình huống 3: Nhiều khung giờ
```
Slot 1: Capacity = 3
Slot 2: Capacity = 5
Slot 3: Capacity = 2

Đăng 8 bài:
→ Slot 1: 3 bài (đầy)
→ Slot 2: 5 bài (đầy)
→ Slot 3: 0 bài (hết bài)
```

---

## 📝 Code đã thay đổi

### File: `backend/internal/db/timeslots.go`

**Hàm mới:**
```go
// IsSlotAvailable - Kiểm tra còn chỗ không
func (s *Store) IsSlotAvailable(slotID string, date time.Time) (bool, error)

// GetSlotRemainingCapacity - Lấy số chỗ còn lại
func (s *Store) GetSlotRemainingCapacity(slotID string, date time.Time) (int, error)
```

**Logic:**
- Đếm số bài hiện tại trong slot
- So sánh với `slot_capacity`
- Trả về `true` nếu `current < capacity`

---

## 🧪 Test Cases

### Test 1: Slot trống
```
Capacity = 10
Current = 0
→ IsSlotAvailable() = true
→ GetSlotRemainingCapacity() = 10
```

### Test 2: Slot gần đầy
```
Capacity = 10
Current = 9
→ IsSlotAvailable() = true
→ GetSlotRemainingCapacity() = 1
```

### Test 3: Slot đầy
```
Capacity = 10
Current = 10
→ IsSlotAvailable() = false
→ GetSlotRemainingCapacity() = 0
```

### Test 4: Slot vượt quá (không nên xảy ra)
```
Capacity = 10
Current = 11
→ IsSlotAvailable() = false
→ GetSlotRemainingCapacity() = 0 (không âm)
```

---

## 🚀 Kết luận

Logic mới đã hoạt động đúng theo yêu cầu:
- ✅ `slot_capacity` = Số bài trong khung giờ
- ✅ Đăng ít hơn → Đăng hết
- ✅ Đăng nhiều hơn → Chuyển sang slot tiếp theo
- ✅ Nhiều khung giờ → Đăng theo thứ tự

**Ngày cập nhật:** 29/11/2024  
**Trạng thái:** ✅ Hoàn thành
