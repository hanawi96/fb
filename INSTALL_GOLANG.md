# Cài đặt Golang

## Bước 1: Download Golang

1. Mở trình duyệt và vào: https://go.dev/dl/
2. Download file: **go1.21.x.windows-amd64.msi** (phiên bản mới nhất)
3. Chạy file .msi vừa download
4. Click "Next" > "Next" > "Install"
5. Đợi cài đặt xong (2-3 phút)

## Bước 2: Kiểm tra cài đặt

Sau khi cài xong, mở PowerShell mới và chạy:

```powershell
go version
```

Nếu thấy: `go version go1.21.x windows/amd64` là thành công!

## Bước 3: Test Go

```powershell
cd backend
go mod download
```

Nếu không có lỗi là OK!

---

**Link download:** https://go.dev/dl/

**Thời gian:** ~5 phút
