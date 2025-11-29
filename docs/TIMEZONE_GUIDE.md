# Timezone Handling Guide

## Overview

This application is designed for Vietnamese users and uses **Vietnam timezone (UTC+7)** throughout the system.

All date/time operations are handled consistently to ensure:
- Scheduled posts display correct time (12:00 VN, not 19:00)
- Time slots work according to local time
- No confusion for users

## Architecture

### 1. Backend (Go)

**Centralized timezone configuration:**

```go
// internal/config/timezone.go
package config

import "time"

var VietnamTZ = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

func ParseDateVN(dateStr string) (time.Time, error) {
    return time.ParseInLocation("2006-01-02", dateStr, VietnamTZ)
}

func NowVN() time.Time {
    return time.Now().In(VietnamTZ)
}

func ToVN(t time.Time) time.Time {
    return t.In(VietnamTZ)
}
```

**Usage in API handlers:**

```go
import "fbscheduler/internal/config"

// Parse date from frontend
preferredDate := config.NowVN()
if req.PreferredDate != "" {
    parsed, err := config.ParseDateVN(req.PreferredDate)
    if err == nil {
        preferredDate = parsed
    }
}

// Create specific time (e.g., 12:00 VN)
dateInVN := config.ToVN(preferredDate)
time12VN := time.Date(
    dateInVN.Year(), 
    dateInVN.Month(), 
    dateInVN.Day(), 
    12, 0, 0, 0, 
    config.VietnamTZ
)
```

**Best practices:**
- ✅ Use `config.ParseDateVN()` for parsing dates
- ✅ Use `config.NowVN()` for current time
- ✅ Use `config.ToVN()` for timezone conversion
- ❌ Never use `time.Parse()` without timezone
- ❌ Never hardcode timezone in multiple places

### 2. Database (PostgreSQL)

**Nguyên tắc:** Lưu timestamp với timezone

```sql
-- Cột scheduled_time nên là TIMESTAMPTZ (có timezone)
CREATE TABLE scheduled_posts (
    id UUID PRIMARY KEY,
    scheduled_time TIMESTAMPTZ NOT NULL,  -- ← Có TZ
    ...
);

-- Khi query, PostgreSQL tự động convert
SELECT scheduled_time FROM scheduled_posts;
-- Kết quả: 2025-11-29 12:00:00+07 (đúng VN timezone)
```

### 3. Frontend (JavaScript/Svelte)

**Centralized datetime utilities:**

```javascript
// lib/utils/datetime.js
const VN_TIMEZONE = 'Asia/Ho_Chi_Minh';

export function formatTimeVN(dateStr) {
    if (!dateStr) return '--:--';
    const date = new Date(dateStr);
    return date.toLocaleTimeString('vi-VN', {
        hour: '2-digit',
        minute: '2-digit',
        timeZone: VN_TIMEZONE
    });
}

export function formatDateTimeVN(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleString('vi-VN', {
        timeZone: VN_TIMEZONE
    });
}

export function getTodayVN() {
    const now = new Date();
    return now.toLocaleDateString('en-CA', { timeZone: VN_TIMEZONE });
}
```

**Usage in components:**

```svelte
<script>
    import { formatTimeVN, formatDateTimeVN } from '$lib/utils/datetime';
    
    // Display time
    $: displayTime = formatTimeVN(post.scheduled_time);
    
    // Display full datetime
    $: displayDateTime = formatDateTimeVN(account.last_post_at);
</script>
```

**Best practices:**
- ✅ Use utility functions from `lib/utils/datetime.js`
- ✅ Send dates as "YYYY-MM-DD" format to backend
- ✅ Backend handles timezone conversion
- ❌ Never format dates without timezone parameter
- ❌ Never duplicate timezone logic in components

## Flow hoàn chỉnh

### Tạo lịch đăng bài

1. **Frontend:** User chọn ngày "29/11/2025"
   ```javascript
   const dateStr = "2025-11-29"; // Gửi lên backend
   ```

2. **Backend:** Parse với VN timezone
   ```go
   parsed, _ := time.ParseInLocation("2006-01-02", "2025-11-29", vietnamTZ)
   // → 2025-11-29 00:00:00 +07:00
   ```

3. **Backend:** Tạo thời gian trong khung giờ (12:00-13:00)
   ```go
   time12VN := time.Date(2025, 11, 29, 12, 30, 0, 0, vietnamTZ)
   // → 2025-11-29 12:30:00 +07:00
   ```

4. **Database:** Lưu với timezone
   ```sql
   INSERT INTO scheduled_posts (scheduled_time) 
   VALUES ('2025-11-29 12:30:00+07');
   ```

5. **Frontend:** Hiển thị
   ```javascript
   formatTime("2025-11-29T12:30:00+07:00")
   // → "12:30" (đúng giờ VN)
   ```

## Kiểm tra

### Test Backend

```bash
cd backend
go run cmd/test-schedule-api/main.go
```

Kết quả mong đợi:
```
Scheduled Time (VN): 2025-11-29 12:xx:xx Asia/Ho_Chi_Minh
✅ Đúng khung giờ 12-13h VN
```

### Test Frontend

Mở browser console:
```javascript
const date = new Date("2025-11-29T12:30:00+07:00");
console.log(date.toLocaleString('vi-VN', { 
    timeZone: 'Asia/Ho_Chi_Minh' 
}));
// → "29/11/2025, 12:30:00"
```

## Lỗi thường gặp

### ❌ Lỗi 1: Hiển thị 19:00 thay vì 12:00

**Nguyên nhân:** Backend parse date với UTC thay vì VN timezone

```go
// SAI
parsed, _ := time.Parse("2006-01-02", "2025-11-29")
// → 2025-11-29 00:00:00 UTC

time12 := time.Date(2025, 11, 29, 12, 0, 0, 0, vietnamTZ)
// → 2025-11-29 12:00:00 +07:00
// Nhưng vì parsed là UTC, nên logic bị sai
```

**Sửa:**
```go
// ĐÚNG
parsed, _ := time.ParseInLocation("2006-01-02", "2025-11-29", vietnamTZ)
// → 2025-11-29 00:00:00 +07:00
```

### ❌ Lỗi 2: Frontend hiển thị sai giờ

**Nguyên nhân:** Không chỉ định timezone khi format

```javascript
// SAI
date.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })
// → Hiển thị theo timezone của máy user (có thể không phải VN)
```

**Sửa:**
```javascript
// ĐÚNG
date.toLocaleTimeString('vi-VN', { 
    hour: '2-digit', 
    minute: '2-digit',
    timeZone: 'Asia/Ho_Chi_Minh'  // ← Bắt buộc
})
```

## Best Practices

1. **Luôn sử dụng Vietnam timezone** trong toàn bộ hệ thống
2. **Không dựa vào server timezone** - có thể là UTC hoặc khác
3. **Test kỹ** với nhiều khung giờ khác nhau
4. **Document rõ ràng** timezone được sử dụng
5. **Validate** thời gian trước khi lưu database

## Summary

| Component | Implementation | Notes |
|-----------|---------------|-------|
| Backend Config | `config.VietnamTZ` | Centralized timezone constant |
| Backend Parse | `config.ParseDateVN()` | Utility function |
| Backend Current Time | `config.NowVN()` | Utility function |
| Database | `TIMESTAMPTZ` | Stores timezone automatically |
| Frontend Utils | `lib/utils/datetime.js` | Centralized formatting |
| Frontend Display | `formatTimeVN()`, `formatDateTimeVN()` | Utility functions |
| Frontend Send | `"YYYY-MM-DD"` | Date only, no time |

## Files Structure

```
backend/
├── internal/
│   ├── config/
│   │   └── timezone.go          # Centralized timezone config
│   ├── api/
│   │   └── schedule_preview.go  # Uses config.ParseDateVN()
│   └── scheduler/
│       └── algorithm.go         # Uses config.VietnamTZ

frontend/
└── src/
    └── lib/
        └── utils/
            └── datetime.js      # Centralized datetime utilities
```

## Migration

Database columns must use `TIMESTAMPTZ`:

```sql
-- Migration 008: Fix timezone
ALTER TABLE scheduled_posts 
    ALTER COLUMN scheduled_time TYPE TIMESTAMPTZ 
    USING scheduled_time AT TIME ZONE 'UTC';
```

---

**Result:** System works correctly for Vietnamese users with consistent timezone handling throughout the stack.
