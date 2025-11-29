# Cập nhật: Thêm trường Max Posts Per Slot

## 📋 Tổng quan

Đã thêm khả năng tùy chỉnh số bài tối đa trong mỗi khung giờ (`max_posts_per_slot`) trực tiếp từ UI.

## ✅ Những gì đã làm

### 1. **Frontend UI** (frontend/src/routes/timeslots/+page.svelte)

Thêm input `max_posts_per_slot` vào cả 2 chế độ:

#### Chế độ "Hằng ngày" (Daily):
```svelte
<div class="flex items-center gap-2 ml-2">
  <label class="text-xs text-gray-500 font-medium">Max:</label>
  <input 
    type="number" 
    bind:value={slot.max_posts_per_slot}
    min="1" 
    max="100"
    class="w-16 px-2 py-2 bg-gray-50 border rounded-lg text-sm"
  />
  <span class="text-xs text-gray-500">bài</span>
</div>
```

#### Chế độ "Tùy chỉnh theo ngày" (Custom):
```svelte
<div class="flex items-center gap-1.5 ml-1">
  <label class="text-xs text-gray-500">Max:</label>
  <input 
    type="number" 
    bind:value={slot.max_posts_per_slot}
    min="1" 
    max="100"
    class="w-14 px-1.5 py-1.5 bg-white border rounded-lg text-xs"
  />
</div>
```

### 2. **Giá trị mặc định**

- Khi tạo slot mới: `max_posts_per_slot = 10`
- Khi copy slot: Giữ nguyên giá trị từ slot nguồn
- Khi apply preset: `max_posts_per_slot = 10`

### 3. **Database Migration**

Đã chạy migration 005 để thêm cột `time_slot_id` vào bảng `scheduled_posts`:

```sql
ALTER TABLE scheduled_posts 
ADD COLUMN IF NOT EXISTS time_slot_id UUID 
REFERENCES page_time_slots(id) ON DELETE SET NULL;
```

### 4. **Cập nhật dữ liệu hiện có**

Đã cập nhật tất cả 10 time slots hiện có từ `max_posts_per_slot = 1` lên `10`:

```sql
UPDATE page_time_slots 
SET max_posts_per_slot = 10 
WHERE max_posts_per_slot = 1;
```

## 🎯 Cách sử dụng

1. Vào trang **Khung giờ đăng bài** (http://localhost:5173/timeslots)
2. Click "Chỉnh sửa" trên page bất kỳ
3. Bên cạnh phần chọn giờ, bạn sẽ thấy trường **"Max: [10] bài"**
4. Thay đổi số lượng tùy ý (1-100)
5. Click "Lưu thay đổi"

## 📊 Ví dụ

### Trước:
```
Khung giờ: 12:00 → 13:00
Max: 1 bài (cố định, không thể thay đổi)
```

### Sau:
```
Khung giờ: 12:00 → 13:00
Max: [10] bài ← Có thể chỉnh sửa
```

## ⚠️ Lưu ý quan trọng

### Vấn đề với bài đã schedule trước đây:

Hiện có **10 bài đã schedule KHÔNG có `time_slot_id`**:
- Các bài này được tạo trước khi có cột `time_slot_id`
- Hệ thống **KHÔNG KIỂM TRA** giới hạn cho các bài này
- Chúng không bị đếm vào `max_posts_per_slot`

### Giải pháp:

Từ bây giờ, khi tạo scheduled post mới, cần đảm bảo:
1. Gán `time_slot_id` cho scheduled post
2. Hệ thống sẽ kiểm tra giới hạn qua hàm `IsSlotAvailable()`

## 🔧 Backend API

API đã hỗ trợ sẵn `max_posts_per_slot`:

```go
// POST /api/pages/:id/timeslots
{
  "start_time": "12:00",
  "end_time": "13:00",
  "days_of_week": [1,2,3,4,5,6,7],
  "max_posts_per_slot": 10  // ← Đã hỗ trợ
}
```

## 📝 Files đã thay đổi

1. `frontend/src/routes/timeslots/+page.svelte` - Thêm UI input
2. `backend/cmd/run-migration-005/main.go` - Migration script
3. `backend/cmd/update-max-posts/main.go` - Update existing data
4. `backend/cmd/check-slots/main.go` - Verification script

## ✨ Kết quả

- ✅ UI đã có trường chỉnh sửa `max_posts_per_slot`
- ✅ Database đã có cột `time_slot_id` trong `scheduled_posts`
- ✅ Tất cả slots hiện có đã được cập nhật lên max = 10
- ✅ Backend API đã hỗ trợ đầy đủ

## 🚀 Bước tiếp theo

Cần đảm bảo khi tạo scheduled post mới:
1. Gọi `CalculateSchedule()` để lấy `time_slot_id`
2. Lưu `time_slot_id` vào `scheduled_posts`
3. Hệ thống sẽ tự động kiểm tra giới hạn

---

**Ngày cập nhật:** 29/11/2024
**Trạng thái:** ✅ Hoàn thành
