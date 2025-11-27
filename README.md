# Facebook Multi-Page Scheduler

Hệ thống đăng bài hàng loạt và hẹn giờ cho nhiều Facebook Pages.

## Tech Stack

- **Frontend**: SvelteKit + TailwindCSS
- **Backend**: Golang
- **Database**: PostgreSQL
- **Storage**: Cloudinary

## Project Structure

```
├── backend/          # Golang API server
├── frontend/         # SvelteKit app
├── docs/            # Documentation
└── README.md
```

## Quick Start

### Backend
```bash
cd backend
go mod download
go run cmd/server/main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

## Environment Variables

### Backend (.env)
```
DATABASE_URL=postgresql://user:pass@localhost:5432/fbscheduler
FACEBOOK_APP_ID=your_app_id
FACEBOOK_APP_SECRET=your_app_secret
CLOUDINARY_URL=cloudinary://key:secret@cloud_name
PORT=8080
```

### Frontend (.env)
```
PUBLIC_API_URL=http://localhost:8080
```

## Database Setup

```bash
cd backend
# Run migrations
psql -U postgres -d fbscheduler -f migrations/001_init.sql
```

## Features

- ✅ OAuth login với Facebook
- ✅ Quản lý nhiều Facebook Pages
- ✅ Tạo bài viết với text + images
- ✅ Đăng bài hàng loạt lên nhiều pages
- ✅ Hẹn giờ đăng bài
- ✅ Lịch sử đăng bài & logs
- ✅ Retry tự động khi thất bại

## Development Timeline

- Week 1: Backend + Database
- Week 2: Frontend + Integration
