package db

import (
	"database/sql"
	"time"
)

// ============================================
// MODELS
// ============================================

type FacebookAccount struct {
	ID                  string     `json:"id"`
	FbUserID            string     `json:"fb_user_id"`
	FbUserName          string     `json:"fb_user_name"`
	ProfilePictureURL   string     `json:"profile_picture_url"`
	AccessToken         string     `json:"-"` // Don't expose
	TokenExpiresAt      *time.Time `json:"token_expires_at"`
	MaxPages            int        `json:"max_pages"`
	MaxPostsPerDay      int        `json:"max_posts_per_day"`
	Status              string     `json:"status"`
	RateLimitUntil      *time.Time `json:"rate_limit_until"`
	PostsToday          int        `json:"posts_today"`
	LastPostAt          *time.Time `json:"last_post_at"`
	LastErrorAt         *time.Time `json:"last_error_at"`
	ConsecutiveFailures int        `json:"consecutive_failures"`
	Notes               string     `json:"notes"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	// Computed fields
	PagesCount    int  `json:"pages_count"`
	TokenDaysLeft int  `json:"token_days_left"`
	IsWarning     bool `json:"is_warning"`  // >= 80% daily limit
	IsAtLimit     bool `json:"is_at_limit"` // >= 100% daily limit
}

type PageAccountAssignment struct {
	ID         string     `json:"id"`
	PageID     string     `json:"page_id"`
	AccountID  string     `json:"account_id"`
	IsPrimary  bool       `json:"is_primary"`
	PostsCount int        `json:"posts_count"`
	LastPostAt *time.Time `json:"last_post_at"`
	CreatedAt  time.Time  `json:"created_at"`

	// Joined fields
	Page    *Page            `json:"page,omitempty"`
	Account *FacebookAccount `json:"account,omitempty"`
}

type Notification struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	AccountID *string   `json:"account_id"`
	PageID    *string   `json:"page_id"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================
// FACEBOOK ACCOUNTS METHODS
// ============================================

func (s *Store) GetAllAccounts() ([]FacebookAccount, error) {
	query := `
		SELECT 
			fa.id, fa.fb_user_id, fa.fb_user_name, COALESCE(fa.profile_picture_url, ''),
			fa.access_token, fa.token_expires_at, fa.max_pages, fa.max_posts_per_day,
			fa.status, fa.rate_limit_until, fa.posts_today,
			fa.last_post_at, fa.last_error_at, fa.consecutive_failures,
			fa.notes, fa.created_at, fa.updated_at,
			COUNT(paa.id) as pages_count
		FROM facebook_accounts fa
		LEFT JOIN page_account_assignments paa ON paa.account_id = fa.id
		GROUP BY fa.id
		ORDER BY fa.created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []FacebookAccount
	for rows.Next() {
		var a FacebookAccount
		err := rows.Scan(
			&a.ID, &a.FbUserID, &a.FbUserName, &a.ProfilePictureURL,
			&a.AccessToken, &a.TokenExpiresAt, &a.MaxPages, &a.MaxPostsPerDay,
			&a.Status, &a.RateLimitUntil, &a.PostsToday,
			&a.LastPostAt, &a.LastErrorAt, &a.ConsecutiveFailures,
			&a.Notes, &a.CreatedAt, &a.UpdatedAt,
			&a.PagesCount,
		)
		if err != nil {
			return nil, err
		}

		a.computeFields()
		accounts = append(accounts, a)
	}

	return accounts, nil
}

func (s *Store) GetAccountByID(id string) (*FacebookAccount, error) {
	query := `
		SELECT 
			fa.id, fa.fb_user_id, fa.fb_user_name, COALESCE(fa.profile_picture_url, ''),
			fa.access_token, fa.token_expires_at, fa.max_pages, fa.max_posts_per_day,
			fa.status, fa.rate_limit_until, fa.posts_today,
			fa.last_post_at, fa.last_error_at, fa.consecutive_failures,
			fa.notes, fa.created_at, fa.updated_at,
			COUNT(paa.id) as pages_count
		FROM facebook_accounts fa
		LEFT JOIN page_account_assignments paa ON paa.account_id = fa.id
		WHERE fa.id = $1
		GROUP BY fa.id
	`

	var a FacebookAccount
	err := s.db.QueryRow(query, id).Scan(
		&a.ID, &a.FbUserID, &a.FbUserName, &a.ProfilePictureURL,
		&a.AccessToken, &a.TokenExpiresAt, &a.MaxPages, &a.MaxPostsPerDay,
		&a.Status, &a.RateLimitUntil, &a.PostsToday,
		&a.LastPostAt, &a.LastErrorAt, &a.ConsecutiveFailures,
		&a.Notes, &a.CreatedAt, &a.UpdatedAt,
		&a.PagesCount,
	)
	if err != nil {
		return nil, err
	}

	a.computeFields()
	return &a, nil
}

func (s *Store) GetAccountByFbUserID(fbUserID string) (*FacebookAccount, error) {
	query := `
		SELECT id, fb_user_id, fb_user_name, COALESCE(profile_picture_url, ''),
			access_token, token_expires_at, max_pages, max_posts_per_day,
			status, rate_limit_until, posts_today,
			last_post_at, last_error_at, consecutive_failures,
			notes, created_at, updated_at
		FROM facebook_accounts
		WHERE fb_user_id = $1
	`

	var a FacebookAccount
	err := s.db.QueryRow(query, fbUserID).Scan(
		&a.ID, &a.FbUserID, &a.FbUserName, &a.ProfilePictureURL,
		&a.AccessToken, &a.TokenExpiresAt, &a.MaxPages, &a.MaxPostsPerDay,
		&a.Status, &a.RateLimitUntil, &a.PostsToday,
		&a.LastPostAt, &a.LastErrorAt, &a.ConsecutiveFailures,
		&a.Notes, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (s *Store) CreateAccount(a *FacebookAccount) error {
	query := `
		INSERT INTO facebook_accounts (
			fb_user_id, fb_user_name, profile_picture_url, access_token, token_expires_at,
			max_pages, max_posts_per_day, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	return s.db.QueryRow(
		query,
		a.FbUserID, a.FbUserName, a.ProfilePictureURL, a.AccessToken, a.TokenExpiresAt,
		a.MaxPages, a.MaxPostsPerDay, a.Notes,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (s *Store) UpdateAccount(a *FacebookAccount) error {
	query := `
		UPDATE facebook_accounts SET
			fb_user_name = $2,
			profile_picture_url = $3,
			access_token = $4,
			token_expires_at = $5,
			max_pages = $6,
			max_posts_per_day = $7,
			status = $8,
			notes = $9
		WHERE id = $1
	`

	_, err := s.db.Exec(
		query,
		a.ID, a.FbUserName, a.ProfilePictureURL, a.AccessToken, a.TokenExpiresAt,
		a.MaxPages, a.MaxPostsPerDay, a.Status, a.Notes,
	)
	return err
}

func (s *Store) DeleteAccount(id string) error {
	_, err := s.db.Exec("DELETE FROM facebook_accounts WHERE id = $1", id)
	return err
}

// RecordSuccessfulPost ghi nhận post thành công
func (s *Store) RecordSuccessfulPost(accountID, pageID string) error {
	_, err := s.db.Exec("SELECT record_successful_post($1, $2)", accountID, pageID)
	return err
}

// RecordPostFailure ghi nhận lỗi
func (s *Store) RecordPostFailure(accountID string, isRateLimit bool) error {
	_, err := s.db.Exec("SELECT record_post_failure($1, $2)", accountID, isRateLimit)
	return err
}

// GetBestAccountForPage lấy account tốt nhất để đăng bài
func (s *Store) GetBestAccountForPage(pageID string) (*FacebookAccount, error) {
	var accountID sql.NullString
	err := s.db.QueryRow("SELECT get_best_account_for_page($1)", pageID).Scan(&accountID)
	if err != nil {
		return nil, err
	}

	if !accountID.Valid {
		return nil, nil // No available account
	}

	return s.GetAccountByID(accountID.String)
}

// ResetDailyPostCounts reset counter hàng ngày
func (s *Store) ResetDailyPostCounts() error {
	_, err := s.db.Exec("SELECT reset_daily_post_counts()")
	return err
}

// Helper: compute derived fields
func (a *FacebookAccount) computeFields() {
	// Token days left
	if a.TokenExpiresAt != nil {
		days := int(time.Until(*a.TokenExpiresAt).Hours() / 24)
		if days < 0 {
			days = 0
		}
		a.TokenDaysLeft = days
	}

	// Warning thresholds
	if a.MaxPostsPerDay > 0 {
		percentage := float64(a.PostsToday) / float64(a.MaxPostsPerDay) * 100
		a.IsWarning = percentage >= 80
		a.IsAtLimit = percentage >= 100
	}
}
