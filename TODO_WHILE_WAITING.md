# ✅ VIỆC CẦN LÀM (Trong khi chờ PostgreSQL)

## 📋 CHECKLIST

### ⏳ Đang chờ PostgreSQL cài đặt...

Trong khi chờ, làm 2 việc này:

---

## 1️⃣ CÀI GOLANG (5 phút) ⚡

### Bước nhanh:
1. Mở trình duyệt
2. Vào: **https://go.dev/dl/**
3. Download: **go1.21.x.windows-amd64.msi**
4. Chạy file .msi
5. Next > Next > Install
6. Đợi 2-3 phút

### Kiểm tra:
```powershell
# Mở PowerShell mới
go version
```

✅ Thấy `go version go1.21.x` là OK!

📄 **Chi tiết:** Xem file `INSTALL_GOLANG.md`

---

## 2️⃣ TẠO FACEBOOK APP (10 phút) 📱

### Bước nhanh:
1. Vào: **https://developers.facebook.com/**
2. Đăng nhập Facebook
3. My Apps > Create App
4. Chọn "Business" type
5. Điền tên app: "FB Scheduler"
6. Lấy **App ID** và **App Secret**
7. Thêm Facebook Login product
8. Config OAuth Redirect URIs:
   - `http://localhost:5173/auth/callback`
   - `http://localhost:8080/api/auth/facebook/callback`
9. Request permissions:
   - `pages_show_list`
   - `pages_read_engagement`
   - `pages_manage_posts`

### Lưu thông tin:
Tạo file `backend/.env`:
```env
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@localhost:5432/fbscheduler?sslmode=disable
FACEBOOK_APP_ID=YOUR_APP_ID
FACEBOOK_APP_SECRET=YOUR_APP_SECRET
FACEBOOK_REDIRECT_URI=http://localhost:5173/auth/callback
PORT=8080
FRONTEND_URL=http://localhost:5173
```

📄 **Chi tiết:** Xem file `SETUP_FACEBOOK_APP.md`

---

## 3️⃣ SAU KHI POSTGRESQL CÀI XONG (5 phút) 🗄️

### Bước nhanh:
```powershell
# Tạo database
psql -U postgres
CREATE DATABASE fbscheduler;
\q

# Chạy migrations
cd D:\FB\backend
psql -U postgres -d fbscheduler -f migrations/001_init.sql

# Kiểm tra
psql -U postgres -d fbscheduler
\dt
\q
```

📄 **Chi tiết:** Xem file `SETUP_DATABASE.md`

---

## 📊 TIẾN ĐỘ

```
[⏳] PostgreSQL đang cài...
[ ] Golang đã cài
[ ] Facebook App đã tạo
[ ] Database đã setup
[ ] Backend .env đã tạo
```

---

## 🎯 SAU KHI XONG TẤT CẢ

### Test Backend:
```powershell
cd backend
go run cmd/server/main.go
```

Nếu thấy:
```
✅ Connected to PostgreSQL
✅ Scheduler started
🚀 Server running on http://localhost:8080
```

Là thành công! 🎉

### Test Frontend:
Frontend đã chạy sẵn rồi: **http://localhost:5173**

---

## ⏱️ TỔNG THỜI GIAN

- PostgreSQL: ~10 phút (đang chờ)
- Golang: ~5 phút
- Facebook App: ~10 phút
- Database setup: ~5 phút

**Tổng: ~30 phút**

---

## 📚 TÀI LIỆU CHI TIẾT

1. `INSTALL_GOLANG.md` - Hướng dẫn cài Golang
2. `SETUP_FACEBOOK_APP.md` - Hướng dẫn tạo Facebook App
3. `SETUP_DATABASE.md` - Hướng dẫn setup database

---

## 💡 TIPS

- Mở nhiều tab trình duyệt để làm song song
- Lưu App ID và App Secret vào Notepad
- Nhớ password PostgreSQL bạn đã đặt
- Restart PowerShell sau khi cài Golang

---

**BẮT ĐẦU NGAY!** ⚡

1. Mở tab mới: https://go.dev/dl/
2. Mở tab mới: https://developers.facebook.com/
3. Đợi PostgreSQL cài xong
