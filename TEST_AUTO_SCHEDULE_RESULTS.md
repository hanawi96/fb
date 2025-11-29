# Test Auto Schedule - Kết Quả

## Mục đích
Test chức năng đăng bài lên lịch tự động với các quy tắc:
- Mỗi page chỉ được đăng **1 bài/slot** (slot_capacity = 1)
- Nếu slot đã đầy → chuyển sang slot tiếp theo trong ngày
- Nếu hết slot trong ngày → chuyển sang ngày mai

## Cấu hình Test
- **Số pages**: 3 (Test Page 1, 2, 3)
- **Số bài posts**: 5
- **Time slots**: 2 khung giờ/page
  - Slot 1: 19h-20h (slot_capacity = 1)
  - Slot 2: 21h-22h (slot_capacity = 1)
- **Ngày test**: 29/11/2025

## Kết Quả

### ✅ Test PASSED - Không có vi phạm

**Bài 1**: 3 pages đăng vào slot 19h-20h ngày 29/11
- Test Page 1: 19:10:54
- Test Page 2: 19:45:53  
- Test Page 3: 19:17:05

**Bài 2**: 3 pages đăng vào slot 21h-22h ngày 29/11 (slot 19h-20h đã đầy)
- Test Page 1: 21:42:04
- Test Page 2: 21:04:32
- Test Page 3: 21:45:06

**Bài 3**: 3 pages đăng vào slot 19h-20h ngày 30/11 (cả 2 slot ngày 29/11 đã đầy)
- Test Page 1: 19:19:00
- Test Page 2: 19:26:06
- Test Page 3: 19:48:53

**Bài 4**: 3 pages đăng vào slot 21h-22h ngày 30/11
- Test Page 1: 21:23:05
- Test Page 2: 21:05:10
- Test Page 3: 21:34:04

## Xác Nhận Logic

✅ **Mỗi page chỉ đăng 1 bài/slot** - ĐÚNG
- Không có slot nào có >1 bài của cùng 1 page

✅ **Slot đầy thì chuyển sang slot tiếp theo** - ĐÚNG
- Bài 2 tự động chuyển từ slot 19h-20h sang 21h-22h

✅ **Hết slot trong ngày thì chuyển sang ngày mai** - ĐÚNG
- Bài 3 và 4 tự động chuyển sang ngày 30/11

## Cách Chạy Test

```bash
cd backend/cmd/test-auto-schedule
go run main.go
```

## Database Verification

Script tự động:
1. Tạo test data (account, pages, posts, time slots)
2. Schedule các bài lên nhiều pages
3. Kiểm tra database
4. Báo cáo violations (nếu có)

## Bug Fix

Đã sửa lỗi trong `backend/internal/db/scheduled.go`:
- Xử lý NULL values cho `link_url` field
- Tránh lỗi 500 khi gọi API `/api/schedule`
