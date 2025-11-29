# Tính Năng Kiểm Tra Xung Đột Thời Gian

## 🎯 Mục Đích
Khi user chọn "Hẹn giờ" đăng bài, hệ thống sẽ kiểm tra xem thời gian đó đã có bài nào được lên lịch chưa. Nếu có xung đột, hiển thị cảnh báo và cho user 3 lựa chọn.

## ✨ Tính Năng

### 1. **Kiểm Tra Xung Đột Tự Động**
- Check chính xác đến **phút** (13:15)
- Check **từng page riêng biệt**
- Chỉ check posts có status `pending` hoặc `processing`

### 2. **Modal Cảnh Báo**
Khi phát hiện xung đột, hiển thị modal với:
- Thời gian xung đột
- Danh sách pages bị xung đột
- 3 lựa chọn xử lý

### 3. **Ba Lựa Chọn**

#### **A. "Có, đăng luôn"**
- Cho phép đăng trùng thời gian
- Tự động thêm random offset 5-30 giây
- Tránh spam Facebook

#### **B. "Không"**
- Đóng modal
- Giữ nguyên form
- User có thể:
  - Chọn lại thời gian khác
  - Bỏ chọn pages xung đột

#### **C. "Lịch tự động"**
- Dùng Smart Scheduling
- Tự động tìm slot trống gần nhất
- Phân bổ thời gian tối ưu

## 🔧 Cấu Trúc Code

### Backend

**File: `backend/internal/api/schedule_conflict.go`**
```go
// API check conflict
POST /api/schedule/check-conflict

Request:
{
  "page_ids": ["page-1", "page-2"],
  "scheduled_time": "2025-11-29T13:15:00+07:00"
}

Response:
{
  "has_conflict": true,
  "conflict_pages": [
    {"page_id": "page-1", "page_name": "Page A"}
  ],
  "no_conflict_pages": [
    {"page_id": "page-2", "page_name": "Page B"}
  ]
}
```

**Query tối ưu:**
- Sử dụng `DATE_TRUNC('minute', ...)` để so sánh chính xác
- Index trên `scheduled_time` và `page_id`
- Chỉ query pages cần thiết

### Frontend

**File: `frontend/src/lib/components/ConflictWarningModal.svelte`**
- Component modal gọn gàng, chuyên nghiệp
- Animation mượt mà
- Responsive design

**File: `frontend/src/routes/posts/new/+page.svelte`**
- Tích hợp check conflict vào flow đăng bài
- Xử lý 3 lựa chọn
- Random offset khi cho phép trùng

## 📊 Flow Hoạt Động

```
1. User chọn: Pages [A, B], Time: 13:15 29/11
   ↓
2. Click "Đăng bài"
   ↓
3. Frontend gọi API check-conflict
   ↓
4a. Không xung đột → Đăng ngay
   ↓
4b. Có xung đột → Hiện modal
   ↓
5. User chọn:
   
   5a. "Có" → Đăng với random offset
       - Page A: 13:15:17
       - Page B: 13:15:23
   
   5b. "Không" → Đóng modal
       - User chọn lại thời gian
       - Hoặc bỏ page xung đột
   
   5c. "Lịch tự động" → Smart schedule
       - Page A: 19:08 (slot trống)
       - Page B: 19:12 (slot trống)
```

## 🎨 UI/UX

### Modal Design
- **Header**: Icon warning + tiêu đề
- **Content**: 
  - Thời gian xung đột (highlight)
  - Danh sách pages (compact list)
  - Hướng dẫn ngắn gọn
- **Actions**: 3 buttons rõ ràng
- **Animation**: Scale-in mượt mà

### Colors
- Warning: Yellow (#EAB308)
- Primary: Blue (#2563EB)
- Danger: Red (#DC2626)
- Success: Green (#16A34A)

## ⚡ Tối Ưu

### Backend
- Query chỉ lấy `page_id` và `page_name`
- Không lấy content bài cũ (không cần thiết)
- Sử dụng `DISTINCT` để tránh duplicate
- Index trên `(page_id, scheduled_time, status)`

### Frontend
- Check conflict chỉ khi cần (scheduleType === 'later')
- Debounce nếu user thay đổi thời gian liên tục
- Lazy load modal component
- Minimal re-renders

## 🧪 Test Cases

### Test 1: Không xung đột
```
Input: Pages [A, B], Time: 13:15
Database: Không có bài nào vào 13:15
Expected: Đăng thành công ngay
```

### Test 2: Xung đột một phần
```
Input: Pages [A, B, C], Time: 13:15
Database: Page A có bài vào 13:15
Expected: Hiện modal, list Page A
```

### Test 3: Xung đột toàn bộ
```
Input: Pages [A, B], Time: 13:15
Database: Cả A và B đều có bài vào 13:15
Expected: Hiện modal, list cả A và B
```

### Test 4: Chọn "Có"
```
Action: User chọn "Có, đăng luôn"
Expected: 
- Đăng với random offset
- Toast success
- Reset form
```

### Test 5: Chọn "Không"
```
Action: User chọn "Không"
Expected:
- Đóng modal
- Giữ nguyên form
- User có thể chỉnh sửa
```

### Test 6: Chọn "Lịch tự động"
```
Action: User chọn "Lịch tự động"
Expected:
- Gọi smart scheduling
- Tìm slot trống
- Toast với thời gian mới
```

## 📝 Lưu Ý

1. **Timezone**: Tất cả thời gian đều xử lý theo Vietnam timezone
2. **Random Offset**: Chỉ áp dụng khi user chọn "Có"
3. **Validation**: Vẫn giữ các validation cũ (quá khứ, 30 ngày)
4. **Error Handling**: Graceful fallback nếu API lỗi

## 🚀 Cách Sử Dụng

### Cho User
1. Tạo bài mới
2. Chọn "Hẹn giờ"
3. Chọn ngày giờ
4. Chọn pages
5. Click "Đăng bài"
6. Nếu có xung đột → Chọn 1 trong 3 options
7. Done!

### Cho Developer
```bash
# Backend
cd backend
go run cmd/server/main.go

# Frontend
cd frontend
npm run dev

# Test API
curl -X POST http://localhost:8080/api/schedule/check-conflict \
  -H "Content-Type: application/json" \
  -d '{
    "page_ids": ["page-id-1"],
    "scheduled_time": "2025-11-29T13:15:00+07:00"
  }'
```

## ✅ Checklist

- [x] Backend API check-conflict
- [x] Frontend API helper
- [x] Modal component
- [x] Tích hợp vào trang tạo bài
- [x] Xử lý 3 lựa chọn
- [x] Random offset
- [x] Smart scheduling fallback
- [x] Error handling
- [x] UI/UX polish
- [x] Documentation

## 🎉 Kết Quả

Tính năng hoàn chỉnh, chạy mượt mà, UI đẹp, code tối ưu!
