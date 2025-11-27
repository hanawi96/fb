# Cách Lấy Facebook App Secret Đúng

## Vấn đề
Lỗi: **"Error validating client secret"** - App Secret không đúng

## Cách Fix

### Bước 1: Lấy App Secret từ Facebook Developer Console

1. Vào: https://developers.facebook.com/apps/4526355974247445/settings/basic/
2. Tìm mục **"App Secret"**
3. Click **"Show"** (có thể cần nhập mật khẩu Facebook)
4. Copy App Secret mới

### Bước 2: Cập nhật file .env

File: `backend/.env`

```env
FACEBOOK_APP_SECRET=<paste_app_secret_ở_đây>
```

### Bước 3: Restart Backend

Sau khi cập nhật .env, restart backend server để load config mới.

## Lưu ý

- App Secret rất nhạy cảm, không được share công khai
- Nếu App Secret bị lộ, hãy reset trong Facebook Developer Console
- Đảm bảo không có khoảng trắng thừa khi copy/paste
