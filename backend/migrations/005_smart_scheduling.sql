-- ============================================
-- SMART SCHEDULING SYSTEM
-- ============================================

-- Bảng cấu hình khung giờ đăng bài cho từng page
CREATE TABLE IF NOT EXISTS page_time_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    slot_name VARCHAR(100), -- "Sáng sớm", "Trưa", "Chiều", "Tối"
    start_time TIME NOT NULL, -- 13:00:00
    end_time TIME NOT NULL, -- 15:00:00
    days_of_week INTEGER[] DEFAULT ARRAY[1,2,3,4,5,6,7], -- 1=Monday, 7=Sunday
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 5, -- 1-10, cao = ưu tiên hơn
    max_posts_per_slot INTEGER DEFAULT 1, -- Số bài tối đa trong 1 slot/ngày
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Bảng cấu hình chung cho hệ thống scheduling
CREATE TABLE IF NOT EXISTS scheduling_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_name VARCHAR(100) UNIQUE DEFAULT 'default',
    min_interval_minutes INTEGER DEFAULT 15, -- Khoảng cách tối thiểu giữa các bài (toàn hệ thống)
    min_interval_same_content_hours INTEGER DEFAULT 4, -- Bài giống nhau phải cách nhau ít nhất
    max_posts_per_page_per_day INTEGER DEFAULT 10,
    distribution_strategy VARCHAR(50) DEFAULT 'balanced', -- balanced, random, sequential
    auto_adjust_on_conflict BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default config
INSERT INTO scheduling_config (config_name) VALUES ('default') ON CONFLICT DO NOTHING;

-- Thêm cột vào scheduled_posts để hỗ trợ smart scheduling
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS time_slot_id UUID REFERENCES page_time_slots(id) ON DELETE SET NULL;
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64); -- MD5 hash của content
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS auto_scheduled BOOLEAN DEFAULT false;
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS original_time TIMESTAMP; -- Thời gian gốc user muốn
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS adjusted_time TIMESTAMP; -- Thời gian đã điều chỉnh
ALTER TABLE scheduled_posts ADD COLUMN IF NOT EXISTS adjustment_reason TEXT; -- Lý do điều chỉnh

-- Bảng theo dõi lịch sử đăng bài (để tránh trùng lặp)
CREATE TABLE IF NOT EXISTS posting_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    content_hash VARCHAR(64),
    scheduled_time TIMESTAMP NOT NULL,
    posted_at TIMESTAMP,
    time_slot_id UUID REFERENCES page_time_slots(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_page_time_slots_page ON page_time_slots(page_id) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_page_time_slots_time ON page_time_slots(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_scheduled_posts_content_hash ON scheduled_posts(content_hash);
CREATE INDEX IF NOT EXISTS idx_scheduled_posts_time_slot ON scheduled_posts(time_slot_id);
CREATE INDEX IF NOT EXISTS idx_posting_history_page_time ON posting_history(page_id, scheduled_time);
CREATE INDEX IF NOT EXISTS idx_posting_history_content ON posting_history(content_hash);

-- Trigger để tự động cập nhật updated_at
CREATE TRIGGER IF NOT EXISTS update_page_time_slots_updated_at 
BEFORE UPDATE ON page_time_slots
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function để kiểm tra xung đột thời gian
CREATE OR REPLACE FUNCTION check_time_slot_conflict(
    p_page_id UUID,
    p_scheduled_time TIMESTAMP,
    p_min_interval_minutes INTEGER
) RETURNS TABLE(has_conflict BOOLEAN, conflicting_times TIMESTAMP[]) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) > 0 as has_conflict,
        ARRAY_AGG(scheduled_time) as conflicting_times
    FROM scheduled_posts
    WHERE page_id = p_page_id
        AND status IN ('pending', 'processing')
        AND ABS(EXTRACT(EPOCH FROM (scheduled_time - p_scheduled_time))) < (p_min_interval_minutes * 60);
END;
$$ LANGUAGE plpgsql;

-- Function để tìm slot thời gian khả dụng tiếp theo
CREATE OR REPLACE FUNCTION find_next_available_slot(
    p_page_id UUID,
    p_time_slot_id UUID,
    p_preferred_date DATE,
    p_min_interval_minutes INTEGER
) RETURNS TIMESTAMP AS $$
DECLARE
    v_start_time TIME;
    v_end_time TIME;
    v_current_time TIMESTAMP;
    v_end_datetime TIMESTAMP;
    v_interval INTERVAL;
    v_has_conflict BOOLEAN;
BEGIN
    -- Lấy thông tin time slot
    SELECT start_time, end_time INTO v_start_time, v_end_time
    FROM page_time_slots
    WHERE id = p_time_slot_id;
    
    -- Tạo timestamp bắt đầu và kết thúc
    v_current_time := p_preferred_date + v_start_time;
    v_end_datetime := p_preferred_date + v_end_time;
    v_interval := (p_min_interval_minutes || ' minutes')::INTERVAL;
    
    -- Tìm slot trống
    WHILE v_current_time <= v_end_datetime LOOP
        -- Kiểm tra xung đột
        SELECT has_conflict INTO v_has_conflict
        FROM check_time_slot_conflict(p_page_id, v_current_time, p_min_interval_minutes);
        
        IF NOT v_has_conflict THEN
            RETURN v_current_time;
        END IF;
        
        v_current_time := v_current_time + v_interval;
    END LOOP;
    
    -- Không tìm thấy slot trống
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
