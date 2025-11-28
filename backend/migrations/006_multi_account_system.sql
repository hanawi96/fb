-- ============================================
-- PHASE 1: MULTI-ACCOUNT SYSTEM
-- Quản lý nhiều nick Facebook để phân tán, chống spam
-- ============================================

-- ============================================
-- 1.1 BẢNG FACEBOOK_ACCOUNTS (Nick Facebook)
-- ============================================
CREATE TABLE IF NOT EXISTS facebook_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Thông tin Facebook account
    fb_user_id VARCHAR(100) NOT NULL UNIQUE,
    fb_user_name VARCHAR(255),
    access_token TEXT NOT NULL,
    token_expires_at TIMESTAMP,
    
    -- Cấu hình giới hạn
    max_pages INTEGER DEFAULT 5,
    max_posts_per_day INTEGER DEFAULT 20,
    
    -- Trạng thái
    status VARCHAR(20) DEFAULT 'active',  -- active, rate_limited, disabled, token_expired
    rate_limit_until TIMESTAMP,
    
    -- Thống kê realtime
    posts_today INTEGER DEFAULT 0,
    last_post_at TIMESTAMP,
    last_error_at TIMESTAMP,
    consecutive_failures INTEGER DEFAULT 0,
    
    -- Metadata
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 1.2 BẢNG PAGE_ACCOUNT_ASSIGNMENTS
-- Liên kết Page với Account
-- ============================================
CREATE TABLE IF NOT EXISTS page_account_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES facebook_accounts(id) ON DELETE CASCADE,
    
    is_primary BOOLEAN DEFAULT true,
    
    -- Thống kê
    posts_count INTEGER DEFAULT 0,
    last_post_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(page_id, account_id)
);

-- ============================================
-- 1.3 BẢNG NOTIFICATIONS
-- Thông báo trong app
-- ============================================
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    type VARCHAR(50) NOT NULL,  -- rate_limit, token_expiring, daily_limit, post_failed
    title VARCHAR(255) NOT NULL,
    message TEXT,
    
    -- Liên kết (optional)
    account_id UUID REFERENCES facebook_accounts(id) ON DELETE CASCADE,
    page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    
    -- Trạng thái
    is_read BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 1.4 CẬP NHẬT BẢNG SCHEDULED_POSTS
-- Thêm các cột mới cho smart scheduling
-- ============================================
ALTER TABLE scheduled_posts 
    ADD COLUMN IF NOT EXISTS account_id UUID REFERENCES facebook_accounts(id) ON DELETE SET NULL;

ALTER TABLE scheduled_posts 
    ADD COLUMN IF NOT EXISTS time_slot_id UUID REFERENCES page_time_slots(id) ON DELETE SET NULL;

ALTER TABLE scheduled_posts 
    ADD COLUMN IF NOT EXISTS calculated_time TIMESTAMP;

ALTER TABLE scheduled_posts 
    ADD COLUMN IF NOT EXISTS random_offset_seconds INTEGER DEFAULT 0;

-- ============================================
-- INDEXES CHO PERFORMANCE
-- ============================================

-- facebook_accounts indexes
CREATE INDEX IF NOT EXISTS idx_fb_accounts_status 
    ON facebook_accounts(status) WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_fb_accounts_posts_today 
    ON facebook_accounts(posts_today);

-- page_account_assignments indexes
CREATE INDEX IF NOT EXISTS idx_page_assignments_page 
    ON page_account_assignments(page_id);

CREATE INDEX IF NOT EXISTS idx_page_assignments_account 
    ON page_account_assignments(account_id);

CREATE INDEX IF NOT EXISTS idx_page_assignments_primary 
    ON page_account_assignments(page_id, is_primary) WHERE is_primary = true;

-- notifications indexes
CREATE INDEX IF NOT EXISTS idx_notifications_unread 
    ON notifications(is_read, created_at DESC) WHERE is_read = false;

CREATE INDEX IF NOT EXISTS idx_notifications_account 
    ON notifications(account_id);

-- scheduled_posts new indexes
CREATE INDEX IF NOT EXISTS idx_scheduled_posts_account 
    ON scheduled_posts(account_id);

CREATE INDEX IF NOT EXISTS idx_scheduled_posts_calculated_time 
    ON scheduled_posts(calculated_time) WHERE status = 'pending';

-- ============================================
-- TRIGGER: Auto-update updated_at
-- ============================================
CREATE TRIGGER IF NOT EXISTS update_facebook_accounts_updated_at 
    BEFORE UPDATE ON facebook_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- FUNCTIONS
-- ============================================

-- Function: Reset posts_today hàng ngày (gọi bởi cron job)
CREATE OR REPLACE FUNCTION reset_daily_post_counts() 
RETURNS void AS $$
BEGIN
    UPDATE facebook_accounts SET posts_today = 0;
END;
$$ LANGUAGE plpgsql;

-- Function: Ghi nhận post thành công
CREATE OR REPLACE FUNCTION record_successful_post(
    p_account_id UUID,
    p_page_id UUID
) RETURNS void AS $$
BEGIN
    -- Update account stats
    UPDATE facebook_accounts 
    SET 
        posts_today = posts_today + 1,
        last_post_at = NOW(),
        consecutive_failures = 0
    WHERE id = p_account_id;
    
    -- Update assignment stats
    UPDATE page_account_assignments
    SET 
        posts_count = posts_count + 1,
        last_post_at = NOW()
    WHERE account_id = p_account_id AND page_id = p_page_id;
END;
$$ LANGUAGE plpgsql;

-- Function: Ghi nhận lỗi
CREATE OR REPLACE FUNCTION record_post_failure(
    p_account_id UUID,
    p_is_rate_limit BOOLEAN DEFAULT false
) RETURNS void AS $$
BEGIN
    UPDATE facebook_accounts 
    SET 
        consecutive_failures = consecutive_failures + 1,
        last_error_at = NOW(),
        status = CASE 
            WHEN p_is_rate_limit THEN 'rate_limited'
            WHEN consecutive_failures + 1 >= 3 THEN 'disabled'
            ELSE status
        END,
        rate_limit_until = CASE 
            WHEN p_is_rate_limit THEN NOW() + INTERVAL '30 minutes'
            ELSE rate_limit_until
        END
    WHERE id = p_account_id;
END;
$$ LANGUAGE plpgsql;

-- Function: Lấy account tốt nhất cho 1 page
CREATE OR REPLACE FUNCTION get_best_account_for_page(
    p_page_id UUID
) RETURNS UUID AS $$
DECLARE
    v_account_id UUID;
BEGIN
    SELECT pa.account_id INTO v_account_id
    FROM page_account_assignments pa
    JOIN facebook_accounts fa ON fa.id = pa.account_id
    WHERE pa.page_id = p_page_id
        AND fa.status = 'active'
        AND (fa.rate_limit_until IS NULL OR fa.rate_limit_until < NOW())
        AND fa.posts_today < fa.max_posts_per_day
    ORDER BY 
        pa.is_primary DESC,           -- Ưu tiên primary
        fa.posts_today ASC,           -- Ưu tiên nick ít bài
        fa.last_error_at ASC NULLS FIRST  -- Ưu tiên nick không lỗi
    LIMIT 1;
    
    RETURN v_account_id;
END;
$$ LANGUAGE plpgsql;

-- Function: Tạo notification
CREATE OR REPLACE FUNCTION create_notification(
    p_type VARCHAR(50),
    p_title VARCHAR(255),
    p_message TEXT,
    p_account_id UUID DEFAULT NULL,
    p_page_id UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_id UUID;
BEGIN
    INSERT INTO notifications (type, title, message, account_id, page_id)
    VALUES (p_type, p_title, p_message, p_account_id, p_page_id)
    RETURNING id INTO v_id;
    
    RETURN v_id;
END;
$$ LANGUAGE plpgsql;
