# Fix: Đăng Ngay Không Hiển Thị Trong Lịch Đăng

## Vấn đề
Khi đăng bài ở chế độ "Đăng ngay", bài đăng không hiển thị trong trang lịch đăng.

## Nguyên nhân
API `PublishPost` (đăng ngay) chỉ tạo:
- `post` record
- `post_log` record

Nhưng KHÔNG tạo `scheduled_post` record.

Trang lịch đăng chỉ query từ bảng `scheduled_posts`, không query `post_log`.

## Giải pháp

### File: `backend/internal/api/posts.go`

Thêm logic tạo `scheduled_post` sau khi đăng thành công/thất bại:

```go
// Thêm import
import (
    ...
    "time"
)

// Trong hàm PublishPost, sau khi tạo post_log:

if result.err != nil {
    // ... tạo post_log ...
    
    // Tạo scheduled_post với status failed
    account, _ := h.store.GetPrimaryAccountForPage(result.pageID)
    scheduledPost := &db.ScheduledPost{
        PostID:        post.ID,
        PageID:        result.pageID,
        ScheduledTime: time.Now(),
        Status:        "failed",
        MaxRetries:    0,
    }
    if account != nil {
        scheduledPost.AccountID = &account.ID
    }
    h.store.CreateScheduledPost(scheduledPost)
    
} else {
    // ... tạo post_log ...
    
    // Tạo scheduled_post với status completed
    account, _ := h.store.GetPrimaryAccountForPage(result.pageID)
    scheduledPost := &db.ScheduledPost{
        PostID:        post.ID,
        PageID:        result.pageID,
        ScheduledTime: time.Now(),
        Status:        "completed",
        MaxRetries:    0,
    }
    if account != nil {
        scheduledPost.AccountID = &account.ID
    }
    h.store.CreateScheduledPost(scheduledPost)
}
```

## Lợi ích

1. ✅ Bài đăng ngay hiển thị trong trang lịch đăng
2. ✅ Có đầy đủ thông tin: page, account, thời gian
3. ✅ Status "completed" cho bài thành công
4. ✅ Status "failed" cho bài thất bại
5. ✅ Thống nhất với bài hẹn giờ

## Cách test

### 1. Restart Backend
```bash
# Stop backend nếu đang chạy
# Start lại backend
cd backend
go run cmd/server/main.go
```

### 2. Test từ UI
1. Vào trang tạo bài mới
2. Nhập nội dung
3. Chọn "Đăng ngay"
4. Chọn pages
5. Click "Đăng bài"
6. Vào trang /schedule
7. Kiểm tra bài vừa đăng có hiển thị không

### 3. Test bằng script
```bash
cd backend/cmd/test-publish-now
go run main.go
```

Kết quả mong đợi:
```
✅ TEST PASSED!
   Post published now appears in schedule page
```

## Kết quả

Sau khi fix:
- ✅ Bài đăng ngay hiển thị trong lịch đăng
- ✅ Có avatar + tên account
- ✅ Status "completed" (màu xanh)
- ✅ Thời gian đăng chính xác

## Notes

- `scheduled_time` = thời gian đăng thực tế (NOW)
- `status` = "completed" (thành công) hoặc "failed" (thất bại)
- `max_retries` = 0 (không retry vì đã đăng rồi)
- `account_id` = primary account của page
