# Debug Upload Ảnh

## ✅ Đã sửa

1. ✅ Thêm log chi tiết trong upload handler
2. ✅ Mở rộng danh sách file type được chấp nhận:
   - image/jpeg, image/jpg, image/png
   - image/gif, image/webp
   - video/mp4, video/quicktime
3. ✅ Thêm default BACKEND_URL nếu không có trong env
4. ✅ Thêm route serve static files từ `/uploads/`

## 🧪 Test lại

### Bước 1: Mở trang đăng bài
- URL: http://localhost:5174/posts/new

### Bước 2: Upload ảnh
1. Click vào khu vực "Thêm ảnh/video"
2. Chọn 1 file ảnh (JPG, PNG, GIF, WebP)
3. Xem Console (F12) và Backend log

### Bước 3: Kiểm tra log

#### Backend log sẽ hiện:
```
📤 Uploading file: example.jpg, Content-Type: image/jpeg
✅ File uploaded successfully: http://localhost:8080/uploads/1732723762_abc123.jpg
```

#### Nếu lỗi sẽ hiện:
```
❌ Parse multipart form error: ...
hoặc
❌ FormFile error: ...
hoặc
❌ Invalid content type: ...
```

## 🐛 Các lỗi có thể gặp

### Lỗi 1: "No file uploaded"
**Nguyên nhân**: Frontend gửi sai field name
**Giải pháp**: Đảm bảo FormData append với key là `'image'`

### Lỗi 2: "Invalid file type"
**Nguyên nhân**: File không phải ảnh hoặc video
**Giải pháp**: Chỉ upload JPG, PNG, GIF, WebP, MP4

### Lỗi 3: "File too large"
**Nguyên nhân**: File > 10MB
**Giải pháp**: Nén ảnh hoặc tăng limit trong code

### Lỗi 4: CORS error
**Nguyên nhân**: Frontend URL không trong whitelist
**Giải pháp**: Kiểm tra FRONTEND_URL trong .env

## 📝 Kiểm tra

1. **Backend log** (terminal Go):
   - Xem log upload chi tiết
   - Xem Content-Type của file

2. **Frontend Console** (F12):
   ```javascript
   // Network tab → Filter: /api/upload
   // Xem Request Headers, Form Data
   ```

3. **Thư mục uploads**:
   ```bash
   # Kiểm tra file đã được lưu
   ls backend/uploads/
   ```

4. **Test URL ảnh**:
   - Mở: http://localhost:8080/uploads/[filename]
   - Phải thấy ảnh hiển thị

## 🔧 Nếu vẫn lỗi

Hãy gửi cho tôi:
1. Backend log (terminal Go)
2. Frontend Console error (F12)
3. Network tab → Request/Response của /api/upload
