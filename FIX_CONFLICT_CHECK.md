# Fix Lỗi Conflict Check API

## ❌ Lỗi
```
Database error: sql: converting argument $1 type: unsupported type []string, a slice of string
```

## ✅ Đã Fix

### File: `backend/internal/api/schedule_conflict.go`

**Thêm import:**
```go
import (
    ...
    "github.com/lib/pq"
)
```

**Sửa query:**
```go
// Trước (SAI):
rows, err := h.db.Query(query, req.PageIDs, scheduledUTC)

// Sau (ĐÚNG):
rows, err := h.db.Query(query, pq.Array(req.PageIDs), scheduledUTC)
```

## 🔧 Nguyên Nhân

PostgreSQL không hỗ trợ trực tiếp `[]string` từ Go. Phải dùng `pq.Array()` để convert sang PostgreSQL array type.

## 🚀 Cách Áp Dụng

### 1. Code đã được fix
File `backend/internal/api/schedule_conflict.go` đã được sửa.

### 2. Build lại backend
```bash
cd backend
go build -o server.exe ./cmd/server
```

### 3. Restart backend server
- Stop server hiện tại (Ctrl+C)
- Start lại:
```bash
./server.exe
# hoặc
go run cmd/server/main.go
```

### 4. Test lại
```bash
cd backend/cmd/test-conflict-check
go run main.go
```

## ✅ Kết Quả Mong Đợi

```
🧪 Test 1: Thời gian không xung đột
   Time: 2025-11-29 19:00:00
   Expected: Không xung đột
   Has Conflict: false
   ✅ PASSED

🧪 Test 2: Tạo bài và check xung đột
   Time: 2025-11-29 17:00:00
   Expected: Có xung đột
   Has Conflict: true
   Conflict Pages:
      - Page A
   ✅ PASSED
```

## 📝 Lưu Ý

- Backend PHẢI restart để áp dụng thay đổi
- Frontend không cần thay đổi gì
- API endpoint vẫn giữ nguyên: `POST /api/schedule/check-conflict`

## 🎯 Test Từ Frontend

Sau khi restart backend:
1. Vào trang tạo bài
2. Chọn "Hẹn giờ"
3. Chọn thời gian trùng với bài cũ
4. Click "Đăng bài"
5. Modal sẽ xuất hiện nếu có xung đột!
