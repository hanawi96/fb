-- Migration 007: Optimize Scheduling Performance
-- Thêm indexes để tối ưu query tìm slot trống

-- Index cho scheduled_posts: tìm bài đã schedule theo slot và ngày
CREATE INDEX IF NOT EXISTS idx_scheduled_posts_slot_date_status 
ON scheduled_posts(time_slot_id, scheduled_time, status) 
WHERE status IN ('pending', 'processing');

-- Index cho scheduled_posts: tìm bài theo page và thời gian
CREATE INDEX IF NOT EXISTS idx_scheduled_posts_page_time_status 
ON scheduled_posts(page_id, scheduled_time DESC, status) 
WHERE status IN ('pending', 'processing');

-- Index cho page_time_slots: tìm slots active của page
CREATE INDEX IF NOT EXISTS idx_page_time_slots_page_active 
ON page_time_slots(page_id, is_active) 
WHERE is_active = true;

-- Index cho page_time_slots: tìm theo days_of_week
CREATE INDEX IF NOT EXISTS idx_page_time_slots_days 
ON page_time_slots USING GIN(days_of_week);

-- Analyze tables để update statistics
ANALYZE scheduled_posts;
ANALYZE page_time_slots;
