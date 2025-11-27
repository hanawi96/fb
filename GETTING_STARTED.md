# 🚀 BẮT ĐẦU NHANH

## Tổng quan dự án

Hệ thống đăng bài hàng loạt và hẹn giờ cho 50 Facebook Pages.

**Tech Stack:**
- Frontend: SvelteKit + TailwindCSS
- Backend: Golang
- Database: PostgreSQL
- Scheduler: Go cron

## Cài đặt nhanh (5 phút)

### 1. Clone & Install
```bash
# Backend
cd backend
go mod download

# Frontend
cd frontend
npm install
```

### 2. Setup Database
```bash
psql -U postgres -c "CREATE DATABASE fbscheduler;"
psql -U postgres -d fbscheduler -f backend/migrations/001_init.sql
```

### 3. Config Environment
```bash
# Backend
cd backend
cp .env.example .env
# Sửa DATABASE_URL, FACEBOOK_APP_ID, FACEBOOK_APP_SECRET

# Frontend
cd frontend
cp .env.example .env
# Giữ nguyên hoặc sửa nếu cần
```

### 4. Chạy
```bash
# Terminal 1 - Backend
cd backend
go run cmd/server/main.go

# Terminal 2 - Frontend
cd frontend
npm run dev
```

### 5. Mở trình duyệt
http://localhost:5173

## Tài liệu chi tiết

- **Setup đầy đủ**: `docs/SETUP.md`
- **Kế hoạch 14 ngày**: `docs/PLAN.md`
- **API Documentation**: `docs/API.md`
- **Deploy Production**: `docs/DEPLOYMENT.md`

## Cấu trúc thư mục

```
├── backend/              # Golang API
│   ├── cmd/server/       # Entry point
│   ├── internal/
│   │   ├── api/          # HTTP handlers
│   │   ├── db/           # Database queries
│   │   ├── facebook/     # Facebook API client
│   │   └── scheduler/    # Cron scheduler
│   └── migrations/       # SQL migrations
│
├── frontend/             # SvelteKit app
│   └── src/
│       ├── routes/       # Pages
│       └── lib/          # Components & utils
│
└── docs/                 # Documentation
```

## Features chính

✅ OAuth login với Facebook
✅ Quản lý 50 pages
✅ Tạo bài viết với text + ảnh (max 10 ảnh)
✅ Đăng bài hàng loạt
✅ Hẹn giờ đăng bài
✅ Retry tự động khi thất bại
✅ Lịch sử đăng bài & logs
✅ UI đẹp, responsive

## Giao diện

### Dashboard
- Thống kê tổng quan
- Quick start guide

### Quản lý Pages
- Kết nối Facebook
- Bật/tắt pages
- Xóa pages

### Tạo bài mới
- Editor với character counter
- Upload nhiều ảnh
- Preview ảnh

### Lịch đăng bài
- Danh sách bài có sẵn
- Chọn pages để đăng
- DateTime picker
- Xem lịch đã hẹn

### Lịch sử
- Table với logs
- Status badges
- Link đến bài đăng Facebook

## Đánh giá độ khó: 6/10

**Dễ:**
- Database schema đơn giản
- CRUD operations
- UI với TailwindCSS

**Vừa:**
- Facebook OAuth
- Scheduler logic
- Multi-page posting

**Khó:**
- Facebook API quirks
- Token management
- Rate limits handling

## Timeline

- **Full-time**: 10-14 ngày
- **Part-time**: 3-4 tuần
- **Mới học**: 6-8 tuần

## Lưu ý quan trọng

1. **Facebook App Review**: Cần review nếu muốn dùng production với user khác
2. **Token Expiration**: Long-lived token chỉ 60 ngày, cần re-auth
3. **Rate Limits**: Facebook giới hạn API calls, đã implement retry logic
4. **50 Pages**: Đủ với PostgreSQL, không cần Redis

## Support

Nếu gặp vấn đề:
1. Đọc `docs/SETUP.md` - Troubleshooting section
2. Check logs: Backend terminal & Browser console
3. Verify Facebook App settings
4. Test với Postman trước

## Next Steps

Sau khi setup xong:
1. Đọc `docs/PLAN.md` để biết roadmap
2. Test từng feature một
3. Deploy lên production (xem `docs/DEPLOYMENT.md`)

Good luck! 🎉
