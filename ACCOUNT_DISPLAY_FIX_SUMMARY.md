# Account Display Fix - Summary

## Vấn đề
Cột "Người đăng" trong bảng lịch đăng hiển thị trống (--) thay vì tên account.

## Nguyên nhân
1. `scheduled_posts.account_id` không được lưu khi tạo scheduled post
2. Query `CreateScheduledPost` thiếu field `account_id`
3. API `/api/schedule` không auto-assign account

## Giải pháp đã thực hiện

### 1. Database Fix
```sql
-- Update tất cả scheduled posts cũ
UPDATE scheduled_posts sp
SET account_id = (
    SELECT paa.account_id
    FROM page_account_assignments paa
    WHERE paa.page_id = sp.page_id
    AND paa.is_primary = true
    LIMIT 1
)
WHERE sp.account_id IS NULL;
```

### 2. Backend Fixes

#### File: `backend/internal/db/scheduled.go`
```go
// Thêm account_id vào INSERT query
query := `
    INSERT INTO scheduled_posts (post_id, page_id, account_id, scheduled_time, status, max_retries, time_slot_id)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id, created_at, updated_at, retry_count
`

// Thêm profile_picture_url vào SELECT query
SELECT 
    ...
    fa.id, fa.fb_user_name, fa.profile_picture_url
FROM scheduled_posts sp
...
```

#### File: `backend/internal/api/schedule.go`
```go
// Auto-assign primary account khi schedule
account, err := h.store.GetPrimaryAccountForPage(pageID)
if err == nil && account != nil {
    sp.AccountID = &account.ID
}
```

#### File: `backend/internal/scheduler/scheduling_service.go`
```go
// Set account_id trước khi tạo scheduled post
if r.AccountID != "" {
    sp.AccountID = &r.AccountID
}
if r.TimeSlotID != "" {
    sp.TimeSlotID = &r.TimeSlotID
}
```

### 3. Frontend Enhancement

#### File: `frontend/src/lib/components/schedule/SchedulePostRow.svelte`
```svelte
<!-- Hiển thị avatar + tên account -->
{#if post.account?.fb_user_name}
    <div class="flex items-center gap-1.5 justify-center">
        {#if post.account?.profile_picture_url}
            <img src={post.account.profile_picture_url} 
                 class="w-5 h-5 rounded-full" />
        {:else}
            <div class="w-5 h-5 rounded-full bg-blue-500">
                {post.account.fb_user_name.charAt(0).toUpperCase()}
            </div>
        {/if}
        <span class="text-xs text-gray-700 truncate">
            {post.account.fb_user_name}
        </span>
    </div>
{:else}
    <div class="text-xs text-gray-400 text-center">--</div>
{/if}
```

## Test Results

### ✅ Test 1: Database Check
```
Posts with account: 23/23 (100%)
Posts without account: 0
```

### ✅ Test 2: API Schedule
```
POST /api/schedule
- Created 2 scheduled posts
- Both have account_id in database
- Both have account info in response
```

### ✅ Test 3: API Get
```
GET /api/schedule
- Returns account info correctly
- Includes profile_picture_url
```

## Cách test

### Backend Test
```bash
cd backend/cmd/test-api-schedule
go run main.go
```

### Frontend Test
1. Tạo bài đăng mới từ UI
2. Schedule lên nhiều pages
3. Vào trang /schedule
4. Kiểm tra cột "Người đăng" có hiển thị avatar + tên

## Kết quả
✅ Tất cả scheduled posts mới đều có account_id
✅ Cột "Người đăng" hiển thị đầy đủ avatar + tên
✅ Scheduled posts cũ đã được update
✅ API hoạt động đúng

## Notes
- Nếu frontend vẫn hiển thị trống, hãy hard refresh (Ctrl+F5)
- Nếu vẫn không được, restart backend server
- Check browser console có lỗi không
