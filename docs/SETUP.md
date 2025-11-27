# Hướng dẫn Setup Project

## 1. Cài đặt Dependencies

### Backend (Golang)
```bash
cd backend
go mod download
```

### Frontend (SvelteKit)
```bash
cd frontend
npm install
```

## 2. Setup Database (PostgreSQL)

### Cài đặt PostgreSQL
- Windows: Download từ https://www.postgresql.org/download/windows/
- Mac: `brew install postgresql`
- Linux: `sudo apt-get install postgresql`

### Tạo Database
```bash
psql -U postgres
CREATE DATABASE fbscheduler;
\q
```

### Chạy Migrations
```bash
cd backend
psql -U postgres -d fbscheduler -f migrations/001_init.sql
```

## 3. Setup Facebook App

### Bước 1: Tạo Facebook App
1. Truy cập https://developers.facebook.com/
2. Đăng nhập và vào "My Apps"
3. Click "Create App"
4. Chọn "Business" type
5. Điền thông tin app

### Bước 2: Config App
1. Vào Settings > Basic
2. Copy App ID và App Secret
3. Thêm App Domain: `localhost`
4. Thêm Privacy Policy URL (bắt buộc)

### Bước 3: Setup OAuth
1. Vào Products > Facebook Login > Settings
2. Thêm Valid OAuth Redirect URIs:
   - `http://localhost:5173/auth/callback`
   - `http://localhost:8080/api/auth/facebook/callback`

### Bước 4: Xin Permissions
1. Vào App Review > Permissions and Features
2. Request các permissions:
   - `pages_show_list`
   - `pages_read_engagement`
   - `pages_manage_posts`

**Lưu ý**: Trong development mode, chỉ admin/developer/tester của app mới dùng được.

## 4. Setup Environment Variables

### Backend (.env)
```bash
cd backend
cp .env.example .env
```

Sửa file `.env`:
```
DATABASE_URL=postgresql://postgres:your_password@localhost:5432/fbscheduler?sslmode=disable
FACEBOOK_APP_ID=your_app_id_here
FACEBOOK_APP_SECRET=your_app_secret_here
FACEBOOK_REDIRECT_URI=http://localhost:5173/auth/callback
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
PORT=8080
FRONTEND_URL=http://localhost:5173
```

### Frontend (.env)
```bash
cd frontend
cp .env.example .env
```

Sửa file `.env`:
```
PUBLIC_API_URL=http://localhost:8080
```

## 5. Setup Cloudinary (Optional - cho upload ảnh)

1. Đăng ký tài khoản miễn phí tại https://cloudinary.com/
2. Vào Dashboard, copy:
   - Cloud Name
   - API Key
   - API Secret
3. Paste vào file `.env` của backend

**Lưu ý**: Nếu không dùng Cloudinary, có thể dùng local upload (đã implement sẵn).

## 6. Chạy Project

### Terminal 1 - Backend
```bash
cd backend
go run cmd/server/main.go
```

Server sẽ chạy tại: http://localhost:8080

### Terminal 2 - Frontend
```bash
cd frontend
npm run dev
```

Frontend sẽ chạy tại: http://localhost:5173

## 7. Test Flow

1. Mở http://localhost:5173
2. Vào "Quản lý Pages"
3. Click "Kết nối Facebook"
4. Đăng nhập và cho phép quyền
5. Chọn pages muốn quản lý
6. Vào "Tạo bài mới" để tạo bài
7. Vào "Lịch đăng bài" để hẹn giờ

## Troubleshooting

### Lỗi Database Connection
- Kiểm tra PostgreSQL đã chạy: `pg_isready`
- Kiểm tra password trong DATABASE_URL
- Kiểm tra database đã tạo: `psql -U postgres -l`

### Lỗi Facebook OAuth
- Kiểm tra App ID và App Secret
- Kiểm tra Redirect URI khớp với config
- Kiểm tra app đang ở Development mode
- Thêm tài khoản test vào App Roles > Roles

### Lỗi CORS
- Kiểm tra FRONTEND_URL trong backend .env
- Restart backend sau khi đổi .env

### Lỗi Upload Image
- Kiểm tra thư mục `backend/uploads` đã tạo
- Hoặc config Cloudinary credentials
