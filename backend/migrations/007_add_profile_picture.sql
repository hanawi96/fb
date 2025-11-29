-- ============================================
-- MIGRATION 007: Add profile_picture_url to facebook_accounts
-- ============================================

ALTER TABLE facebook_accounts 
    ADD COLUMN IF NOT EXISTS profile_picture_url TEXT;
