package db

import (
	"database/sql"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// DB returns the underlying database connection
func (s *Store) DB() *sql.DB {
	return s.db
}

// Models
type Page struct {
	ID                string     `json:"id"`
	PageID            string     `json:"page_id"`
	PageName          string     `json:"page_name"`
	AccessToken       string     `json:"-"` // Don't expose in JSON
	TokenExpiresAt    *time.Time `json:"token_expires_at"`
	Category          string     `json:"category"`
	ProfilePictureURL string     `json:"profile_picture_url"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Post struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	MediaURLs []string  `json:"media_urls"`
	MediaType string    `json:"media_type"`
	LinkURL   string    `json:"link_url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ScheduledPost struct {
	ID            string    `json:"id"`
	PostID        string    `json:"post_id"`
	PageID        string    `json:"page_id"`
	ScheduledTime time.Time `json:"scheduled_time"`
	Status        string    `json:"status"`
	RetryCount    int       `json:"retry_count"`
	MaxRetries    int       `json:"max_retries"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	// Joined fields
	Post     *Post `json:"post,omitempty"`
	Page     *Page `json:"page,omitempty"`
}

type PostLog struct {
	ID               string    `json:"id"`
	ScheduledPostID  string    `json:"scheduled_post_id"`
	PostID           string    `json:"post_id"`
	PageID           string    `json:"page_id"`
	FacebookPostID   string    `json:"facebook_post_id"`
	Status           string    `json:"status"`
	ErrorMessage     string    `json:"error_message"`
	ResponseData     string    `json:"response_data"`
	PostedAt         time.Time `json:"posted_at"`
	
	// Joined fields
	Post *Post `json:"post,omitempty"`
	Page *Page `json:"page,omitempty"`
}
