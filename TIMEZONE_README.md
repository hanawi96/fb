# Timezone Implementation

## Quick Reference

This application uses **Vietnam timezone (UTC+7)** consistently across all layers.

### Backend

```go
import "fbscheduler/internal/config"

// Parse date from frontend
date, _ := config.ParseDateVN("2025-11-29")

// Get current time in VN
now := config.NowVN()

// Convert to VN timezone
vnTime := config.ToVN(someTime)
```

### Frontend

```javascript
import { formatTimeVN, formatDateTimeVN } from '$lib/utils/datetime';

// Display time (12:30)
const time = formatTimeVN(post.scheduled_time);

// Display full datetime (29/11/2025, 12:30:45)
const datetime = formatDateTimeVN(account.last_post_at);
```

### Database

All timestamp columns use `TIMESTAMPTZ`:

```sql
CREATE TABLE scheduled_posts (
    scheduled_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Key Files

- `backend/internal/config/timezone.go` - Timezone utilities
- `frontend/src/lib/utils/datetime.js` - Datetime formatting
- `docs/TIMEZONE_GUIDE.md` - Detailed documentation

## Testing

```bash
# Backend
cd backend
go run cmd/server/main.go

# Test API returns correct timezone
curl http://localhost:8080/api/schedule/preview \
  -d '{"page_ids":["..."],"preferred_date":"2025-11-29"}'

# Should return: "ScheduledTime":"2025-11-29T12:xx:xx+07:00"
```

## Common Issues

**Issue:** Time displays as 19:00 instead of 12:00

**Solution:** 
1. Check database column is `TIMESTAMPTZ` (not `TIMESTAMP`)
2. Verify frontend uses `formatTimeVN()` utility
3. Ensure backend uses `config.ParseDateVN()`

---

For detailed documentation, see [TIMEZONE_GUIDE.md](docs/TIMEZONE_GUIDE.md)
