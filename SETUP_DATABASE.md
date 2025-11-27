# Setup Database

## Sau khi PostgreSQL cài xong

### Bước 1: Mở PowerShell

Mở PowerShell mới (để load PATH mới)

### Bước 2: Kiểm tra PostgreSQL

```powershell
psql --version
```

Nếu thấy version là OK!

### Bước 3: Tạo Database

```powershell
# Đăng nhập vào PostgreSQL
psql -U postgres

# Nhập password bạn đã đặt khi cài
```

Trong psql prompt, chạy:

```sql
CREATE DATABASE fbscheduler;
\q
```

### Bước 4: Chạy Migrations

```powershell
cd D:\FB\backend
psql -U postgres -d fbscheduler -f migrations/001_init.sql
```

Nhập password khi được hỏi.

Nếu thấy:
```
CREATE TABLE
CREATE TABLE
CREATE TABLE
CREATE TABLE
CREATE INDEX
...
```

Là thành công!

### Bước 5: Kiểm tra Database

```powershell
psql -U postgres -d fbscheduler
```

Trong psql:
```sql
\dt
```

Bạn sẽ thấy 4 tables:
- pages
- posts
- scheduled_posts
- post_logs

```sql
\q
```

---

## ✅ Checklist

- [ ] PostgreSQL đã cài xong
- [ ] Đã tạo database `fbscheduler`
- [ ] Đã chạy migrations
- [ ] Đã kiểm tra 4 tables tồn tại

---

## ⚠️ Nếu gặp lỗi "psql not found"

PostgreSQL chưa được thêm vào PATH. Thử:

```powershell
# Tìm đường dẫn PostgreSQL
cd "C:\Program Files\PostgreSQL\13\bin"
.\psql --version
```

Hoặc restart máy để load PATH mới.

---

**Thời gian:** ~5 phút
