# Hướng dẫn Test Max Posts Per Slot

## 🧪 Các bước test

### 1. Mở trang Timeslots
```
http://localhost:5173/timeslots
```

### 2. Click "Chỉnh sửa" trên một page bất kỳ

### 3. Kiểm tra UI

Bạn sẽ thấy:

**Chế độ "Hằng ngày":**
```
┌─────────────────────────────────────────────────┐
│ [09:00 ▼] → [10:00 ▼]  Max: [10] bài  [🗑️]    │
│ [12:00 ▼] → [13:00 ▼]  Max: [10] bài  [🗑️]    │
└─────────────────────────────────────────────────┘
```

**Chế độ "Tùy chỉnh theo ngày":**
```
┌─────────────────────────────────────────────────┐
│ T2 - Thứ Hai                          2 khung giờ│
│ ├─ [09:00 ▼] → [10:00 ▼] Max: [10]  [🗑️]       │
│ └─ [12:00 ▼] → [13:00 ▼] Max: [10]  [🗑️]       │
└─────────────────────────────────────────────────┘
```

### 4. Thử thay đổi giá trị

- Click vào ô số "10"
- Thay đổi thành 20, 50, hoặc 100
- Click "Lưu thay đổi"

### 5. Kiểm tra database

Chạy lệnh:
```bash
cd backend
go run cmd/check-slots/main.go
```

Kết quả mong đợi:
```
✅ Page Name: 12:00 - 13:00 (max: 20 bài)  ← Đã thay đổi
```

## ✅ Test Cases

### Test 1: Tạo slot mới
1. Click "Thêm khung giờ"
2. Chọn giờ: 14:00 → 15:00
3. Kiểm tra: Max mặc định = 10 ✅

### Test 2: Copy slot
1. Click "Copy từ page khác"
2. Chọn page nguồn
3. Kiểm tra: Max được copy theo ✅

### Test 3: Apply preset
1. Click preset "Giờ vàng Facebook"
2. Kiểm tra: Tất cả slots có Max = 10 ✅

### Test 4: Bulk edit
1. Chọn nhiều pages
2. Click "Chỉnh sửa khung giờ"
3. Thêm slot với Max = 15
4. Lưu
5. Kiểm tra: Tất cả pages đã chọn có Max = 15 ✅

## 🐛 Troubleshooting

### Không thấy trường "Max"?
- Refresh trang (Ctrl+R)
- Clear cache (Ctrl+Shift+R)
- Kiểm tra console có lỗi không

### Lưu không thành công?
- Mở DevTools → Network tab
- Xem request POST /api/pages/:id/timeslots
- Kiểm tra payload có `max_posts_per_slot` không

### Giá trị không được lưu?
- Chạy: `go run cmd/check-slots/main.go`
- Xem database có cập nhật không
- Kiểm tra backend logs

## 📸 Screenshots

Chụp màn hình các vị trí sau:
1. Modal chỉnh sửa - Chế độ "Hằng ngày"
2. Modal chỉnh sửa - Chế độ "Tùy chỉnh theo ngày"
3. Sau khi thay đổi giá trị Max
4. Kết quả trong database

---

**Lưu ý:** Nếu gặp lỗi, check file `TIMESLOT_MAX_POSTS_UPDATE.md` để xem chi tiết.
