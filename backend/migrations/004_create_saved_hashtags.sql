-- Create saved_hashtags table
CREATE TABLE IF NOT EXISTS saved_hashtags (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    hashtag VARCHAR(255) NOT NULL,
    media_count BIGINT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, hashtag)
);

CREATE INDEX idx_saved_hashtags_user_id ON saved_hashtags(user_id);
