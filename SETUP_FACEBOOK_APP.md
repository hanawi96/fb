# Tạo Facebook App

## Bước 1: Truy cập Facebook Developers

1. Mở trình duyệt
2. Vào: **https://developers.facebook.com/**
3. Đăng nhập bằng tài khoản Facebook của bạn

## Bước 2: Tạo App mới

1. Click **"My Apps"** (góc trên bên phải)
2. Click **"Create App"**
3. Chọn **"Business"** type
4. Click **"Next"**

## Bước 3: Điền thông tin App

```
App Name: FB Scheduler (hoặc tên bạn thích)
App Contact Email: email của bạn
Business Account: Chọn hoặc tạo mới
```

5. Click **"Create App"**
6. Xác nhận Security Check (nếu có)

## Bước 4: Lấy App ID và App Secret

1. Vào **Settings > Basic** (sidebar trái)
2. Copy **App ID** - Lưu lại
3. Click **"Show"** ở App Secret
4. Copy **App Secret** - Lưu lại

⚠️ **LƯU Ý**: Giữ App Secret bí mật!

## Bước 5: Thêm Facebook Login

1. Vào **Dashboard**
2. Tìm **"Facebook Login"** product
3. Click **"Set Up"**
4. Chọn **"Web"**
5. Site URL: `http://localhost:5173`
6. Click **"Save"**

## Bước 6: Config OAuth Redirect URIs

1. Vào **Facebook Login > Settings** (sidebar trái)
2. Trong **"Valid OAuth Redirect URIs"**, thêm:
   ```
   http://localhost:5173/auth/callback
   http://localhost:8080/api/auth/facebook/callback
   ```
3. Click **"Save Changes"**

## Bước 7: Thêm Test Users (Quan trọng!)

Vì app đang ở Development mode, chỉ admin/tester mới dùng được.

1. Vào **Roles > Test Users** (sidebar trái)
2. Click **"Add"**
3. Tạo 1-2 test users
4. Hoặc thêm tài khoản Facebook thật của bạn vào **Roles > Administrators**

## Bước 8: Xin Permissions

1. Vào **App Review > Permissions and Features**
2. Tìm và request các permissions:
   - `pages_show_list`
   - `pages_read_engagement`
   - `pages_manage_posts`

⚠️ **LƯU Ý**: Trong Development mode, không cần approval. Chỉ cần request.

## Bước 9: Lưu thông tin

Tạo file `backend/.env` với nội dung:

```env
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@localhost:5432/fbscheduler?sslmode=disable
FACEBOOK_APP_ID=YOUR_APP_ID_HERE
FACEBOOK_APP_SECRET=YOUR_APP_SECRET_HERE
FACEBOOK_REDIRECT_URI=http://localhost:5173/auth/callback
PORT=8080
FRONTEND_URL=http://localhost:5173
```

Thay:
- `YOUR_PASSWORD` = password PostgreSQL bạn đã đặt
- `YOUR_APP_ID_HERE` = App ID vừa copy
- `YOUR_APP_SECRET_HERE` = App Secret vừa copy

---

## ✅ Checklist

- [ ] Đã tạo Facebook App
- [ ] Đã lấy App ID và App Secret
- [ ] Đã thêm Facebook Login product
- [ ] Đã config OAuth Redirect URIs
- [ ] Đã thêm test users hoặc admin
- [ ] Đã request permissions
- [ ] Đã tạo file backend/.env

---

**Thời gian:** ~10 phút

**Link:** https://developers.facebook.com/
