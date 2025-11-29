-- ============================================
-- MIGRATION 008: Fix timezone cho scheduled_posts
-- Đổi từ TIMESTAMP sang TIMESTAMPTZ để lưu đúng UTC
-- ============================================

-- Đổi cột scheduled_time sang TIMESTAMPTZ
ALTER TABLE scheduled_posts 
    ALTER COLUMN scheduled_time TYPE TIMESTAMPTZ 
    USING scheduled_time AT TIME ZONE 'UTC';

-- Đổi các cột timestamp khác nếu cần
ALTER TABLE scheduled_posts 
    ALTER COLUMN created_at TYPE TIMESTAMPTZ 
    USING created_at AT TIME ZONE 'UTC';

ALTER TABLE scheduled_posts 
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ 
    USING updated_at AT TIME ZONE 'UTC';

-- Set timezone mặc định cho database (optional nhưng recommended)
-- ALTER DATABASE fbscheduler SET timezone TO 'UTC';
