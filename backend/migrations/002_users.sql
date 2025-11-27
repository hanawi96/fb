-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default admin user (password: admin123)
-- Hash generated with bcrypt cost 10
INSERT OR IGNORE INTO users (username, password_hash) 
VALUES ('admin', '$2a$10$VT5rjDMF5nTuON1h8epin.eZpeaWwFruCfM6KddqBeKRYkHatwQgW');
