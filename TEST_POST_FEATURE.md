# Hướng dẫn Test Chức Năng Đăng Bài

## ✅ Đã hoàn thành

### Backend:
1. ✅ Tạo endpoint mới `/api/posts/publish` (POST) - Đăng bài ngay lập tức
2. ✅ Endpoint nhận:
   - `content`: Nội dung bài viết
   - `media_urls`: Mảng URL ảnh/video
   - `media_type`: Loại media (photo/video/text)
   - `page_ids`: Mảng ID các trang cần đăng
   - `privacy`: Quyền riêng tư (public/private)

3. ✅ Endpoint sẽ:
   - Tạo bản ghi post trong database
   - Đăng lên Facebook cho từng page được chọn
   - Trả về kết quả chi tiết cho từng page (thành công/thất bại)
   - Lưu log vào database

### Frontend:
1. ✅ Thêm method `publishPost()` vào API client
2. ✅ Cập nhật logic đăng bài:
   - "Đăng bài" (scheduled) → Gọi `/api/posts/publish` - đăng ngay
   - "Lưu nháp" (draft) → Gọi `/api/posts` - lưu draft
   - "Lên lịch" (later) → Gọi `/api/posts` + `/api/schedule` - lên lịch

## 🧪 Cách Test

### Bước 1: Kiểm tra Server
- Backend: http://localhost:8080 ✅ Đang chạy
- Frontend: http://localhost:5174 ✅ Đang chạy

### Bước 2: Truy cập trang đăng bài
1. Mở trình duyệt: http://localhost:5174/posts/new
2. Đăng nhập nếu chưa đăng nhập

### Bước 3: Test đăng bài cơ bản (Text + 1 ảnh)

#### Test Case 1: Đăng text đơn giản
1. Chọn 1 hoặc nhiều fanpage từ danh sách bên trái
2. Nhập nội dung: "Test đăng bài từ FB Scheduler 🎉"
3. Nhấn "Đăng bài"
4. ✅ Kỳ vọng: 
   - Hiện thông báo "Đã đăng bài thành công lên X trang!"
   - Bài viết xuất hiện trên fanpage Facebook

#### Test Case 2: Đăng text + 1 ảnh
1. Chọn fanpage
2. Nhập nội dung: "Test đăng ảnh 📸"
3. Click vào khu vực "Thêm ảnh/video" hoặc nút Upload
4. Chọn 1 file ảnh từ máy tính
5. Đợi upload xong (sẽ thấy preview ảnh)
6. Nhấn "Đăng bài"
7. ✅ Kỳ vọng:
   - Hiện thông báo thành công
   - Bài viết có cả text và ảnh trên Facebook

#### Test Case 3: Đăng nhiều ảnh (carousel)
1. Chọn fanpage
2. Nhập nội dung: "Test album ảnh 🖼️"
3. Upload 2-10 ảnh
4. Nhấn "Đăng bài"
5. ✅ Kỳ vọng:
   - Bài viết có album ảnh trên Facebook

#### Test Case 4: Đăng lên nhiều page cùng lúc
1. Chọn 2-3 fanpage
2. Nhập nội dung + upload ảnh
3. Nhấn "Đăng bài"
4. ✅ Kỳ vọng:
   - Bài viết xuất hiện trên TẤT CẢ các page đã chọn

### Bước 4: Kiểm tra kết quả

#### Trên Frontend:
- Xem thông báo toast (góc trên bên phải)
- Mở Console (F12) để xem log chi tiết

#### Trên Facebook:
- Truy cập từng fanpage đã chọn
- Kiểm tra bài viết mới nhất
- Xác nhận nội dung và ảnh đúng

#### Trong Database:
```sql
-- Xem posts đã tạo
SELECT * FROM posts ORDER BY created_at DESC LIMIT 5;

-- Xem logs
SELECT * FROM post_logs ORDER BY created_at DESC LIMIT 10;
```

## 🐛 Các lỗi có thể gặp

### Lỗi 1: "Page not found"
- **Nguyên nhân**: Page ID không tồn tại hoặc đã bị xóa
- **Giải pháp**: Kiểm tra lại danh sách pages, refresh lại trang

### Lỗi 2: "Failed to post to Facebook"
- **Nguyên nhân**: 
  - Access token hết hạn
  - Thiếu quyền đăng bài
  - URL ảnh không hợp lệ
- **Giải pháp**: 
  - Đăng nhập lại Facebook
  - Kiểm tra quyền của app
  - Đảm bảo ảnh đã upload thành công

### Lỗi 3: "Upload failed"
- **Nguyên nhân**: File quá lớn hoặc định dạng không hỗ trợ
- **Giải pháp**: 
  - Chỉ upload ảnh JPG, PNG (< 5MB)
  - Kiểm tra endpoint upload có hoạt động không

### Lỗi 4: "At least one page is required"
- **Nguyên nhân**: Chưa chọn page nào
- **Giải pháp**: Chọn ít nhất 1 page từ danh sách bên trái

## 📊 Kiểm tra Log

### Backend Log:
```bash
# Xem log trong terminal backend
# Sẽ thấy:
✅ Successfully posted to page 123456789: 123456789_987654321
```

### Frontend Console:
```javascript
// Mở Console (F12) sẽ thấy:
Publish result: {
  post_id: "uuid",
  results: [
    {
      page_id: "...",
      page_name: "...",
      status: "success",
      facebook_post_id: "..."
    }
  ]
}
```

## 🎯 Checklist Test

- [ ] Đăng text đơn giản
- [ ] Đăng text + 1 ảnh
- [ ] Đăng text + nhiều ảnh (2-10)
- [ ] Đăng lên 1 page
- [ ] Đăng lên nhiều page cùng lúc
- [ ] Kiểm tra bài viết trên Facebook
- [ ] Kiểm tra log trong database
- [ ] Test lưu nháp
- [ ] Test xem trước bài viết

## 🔧 Debug

Nếu có lỗi, kiểm tra:

1. **Backend log** (terminal chạy Go):
   ```
   Xem chi tiết lỗi từ Facebook API
   ```

2. **Frontend console** (F12):
   ```javascript
   // Xem request/response
   Network tab → Filter: /api/posts/publish
   ```

3. **Database**:
   ```sql
   -- Xem lỗi chi tiết
   SELECT * FROM post_logs WHERE status = 'failed' ORDER BY created_at DESC;
   ```

## 📝 Ghi chú

- Upload ảnh sẽ lưu vào thư mục `backend/uploads/`
- Mỗi lần đăng sẽ tạo 1 record trong bảng `posts`
- Mỗi page sẽ có 1 record trong bảng `post_logs`
- Nếu đăng lên 3 page → sẽ có 3 records trong `post_logs`

## 🚀 Tính năng đã có

✅ Đăng text
✅ Đăng 1 ảnh
✅ Đăng nhiều ảnh (carousel)
✅ Đăng lên nhiều page
✅ Lưu nháp
✅ Xem trước
✅ Log chi tiết
✅ Xử lý lỗi

## 🔜 Tính năng chưa có (có thể thêm sau)

- ⏰ Lên lịch đăng bài (đã có endpoint nhưng chưa test)
- 🎥 Đăng video
- 🔗 Đăng link với preview
- 😊 Emoji picker
- 📍 Thêm location
- 🏷️ Tag người/page
- 🔒 Cài đặt privacy chi tiết
- 📊 Thống kê engagement
