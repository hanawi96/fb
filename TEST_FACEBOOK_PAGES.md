# 🧪 Hướng dẫn Test Facebook Pages Connection

## Vấn đề hiện tại
Khi đã kết nối 2 fanpage và muốn kết nối thêm fanpage thứ 3, 4:
- Bấm "Kết nối thêm Facebook"
- Facebook hiện popup xác nhận quyền
- Bấm "Tiếp tục" → popup đóng và trang refresh
- **KHÔNG** hiện modal để chọn thêm/bỏ chọn pages

## Các nguyên nhân có thể

### 1. Facebook chỉ trả về pages đã được authorize
- Facebook API `/me/accounts` chỉ trả về các pages mà user đã cấp quyền cho app
- Nếu user không chọn thêm pages trong lần re-auth, API vẫn trả về số pages cũ

### 2. Scope không đủ
- Cần scope `pages_manage_metadata` để có thể re-request và quản lý pages
- Đã thêm scope này vào code

### 3. Facebook không hiện modal chọn pages
- Có thể do `auth_type=rerequest` không đủ
- Cần thêm tham số khác như `auth_nonce` hoặc `extras`

## Cách test chi tiết

### Bước 1: Khởi động backend với logging
```bash
cd backend
go run cmd/server/main.go
```

Backend sẽ log chi tiết:
- 🔗 Auth URL được tạo
- 📥 Code nhận được từ callback
- ✅ Access token
- 📊 Số pages Facebook trả về
- 💾 Pages được lưu vào DB

### Bước 2: Mở test page
```bash
# Mở file test-facebook-pages.html trong browser
start test-facebook-pages.html
```

### Bước 3: Test kết nối
1. **Bấm "Kết nối Facebook"**
   - Xem log trong test page
   - Xem log trong backend terminal
   - Kiểm tra popup Facebook có hiện modal chọn pages không

2. **Xem Pages từ Database**
   - Bấm "Tải Pages từ DB"
   - So sánh số pages trong DB vs số pages bạn có trên Facebook

3. **Debug Facebook API**
   - Lấy access token từ backend log (sau khi kết nối)
   - Paste vào ô "Nhập access token"
   - Bấm "Test Facebook API"
   - Xem Facebook trả về bao nhiêu pages

## Kiểm tra trong Facebook Developer Console

### Xem permissions đã cấp
1. Vào https://developers.facebook.com/tools/explorer/
2. Chọn app "Test Scheduler"
3. Get User Access Token
4. Xem các permissions đã được cấp

### Test Graph API trực tiếp
```
GET /me/accounts?fields=id,name,access_token,category,picture
```

Xem response trả về bao nhiêu pages.

## Giải pháp có thể

### Giải pháp 1: Thêm `extras` parameter
```go
params.Add("extras", `{"setup":{"channel":"IG_API_ONBOARDING"}}`)
```

### Giải pháp 2: Sử dụng Business Manager
- Thay vì dùng `/me/accounts`
- Dùng Business Manager API để lấy tất cả pages

### Giải pháp 3: Hướng dẫn user
- Sau khi bấm "Tiếp tục" trong popup
- Hướng dẫn user vào Facebook Settings
- Business Integrations → Test Scheduler → Edit Settings
- Add or Remove Pages

### Giải pháp 4: Revoke và re-authorize
```go
// Thêm vào auth URL
params.Add("auth_type", "rerequest")
params.Add("reauthorize", "true")
```

## Log mẫu khi thành công

Backend log:
```
🔗 Generated Auth URL: https://www.facebook.com/v18.0/dialog/oauth?...
📥 Received callback with code: AQCnxdE-WGUN6WzqNiVM...
✅ Got user access token: EABAUslZC41BUBQIB0vF...
📊 Received 3 pages from Facebook
  Page 1: ID=147785061761510, Name=Ánh Lê, Category=Đồ em bé/Đồ trẻ em
  Page 2: ID=111268468685496, Name=Vòng Dâu Tằm By Vui, Category=Trẻ em
  Page 3: ID=123456789012345, Name=Page mới, Category=...
💾 Saved page: Ánh Lê (ID: 147785061761510)
💾 Saved page: Vòng Dâu Tằm By Vui (ID: 111268468685496)
💾 Saved page: Page mới (ID: 123456789012345)
✅ Successfully saved 3 pages to database
```

## Kết luận

Sau khi test, chúng ta sẽ biết chính xác:
1. Facebook có hiện modal chọn pages không?
2. Facebook API trả về bao nhiêu pages?
3. Có pages nào bị thiếu không?
4. Nguyên nhân chính xác là gì?
