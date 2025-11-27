# Hệ thống đăng nhập đã được cài đặt

## ✅ Đã hoàn thành

### Backend
- ✅ Tạo bảng `users` trong database
- ✅ API đăng nhập `/api/login`
- ✅ API verify token `/api/verify`
- ✅ JWT authentication với bcrypt

### Frontend
- ✅ Trang đăng nhập `/login`
- ✅ Auth store với localStorage
- ✅ Protected routes (tự động redirect về login)
- ✅ Nút đăng xuất trong layout
- ✅ Token tự động gửi kèm mọi API request

## 🚀 Cách sử dụng

### 1. Chạy migration
```bash
# Migration sẽ tự động chạy khi start backend
cd backend
go run cmd/server/main.go
```

### 2. Đăng nhập
- Truy cập: http://localhost:5173/login
- **Username**: `admin`
- **Password**: `admin123`

### 3. Tạo user mới (nếu cần)
```bash
# Hash password
cd backend
go run cmd/hashpass/main.go your-password

# Thêm vào database
INSERT INTO users (username, password_hash) VALUES ('newuser', 'hash-từ-lệnh-trên');
```

## 🎯 Tính năng

- **Siêu nhanh**: JWT token lưu trong localStorage, không cần query DB mỗi request
- **Siêu nhẹ**: Chỉ check token khi cần, không middleware phức tạp
- **Tự động**: Redirect về login nếu chưa đăng nhập
- **Bảo mật**: Password hash với bcrypt, JWT với expiry 24h

## 📝 Lưu ý

- Token hết hạn sau 24 giờ
- Đổi `jwtSecret` trong `backend/internal/api/login.go` khi deploy production
- Migration 002_users.sql đã tạo user admin mặc định
