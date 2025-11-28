package db

import "time"

// ============================================
// PAGE ACCOUNT ASSIGNMENTS METHODS
// ============================================

// GetAssignmentsByAccount lấy danh sách pages của 1 account
func (s *Store) GetAssignmentsByAccount(accountID string) ([]PageAccountAssignment, error) {
	query := `
		SELECT 
			paa.id, paa.page_id, paa.account_id, paa.is_primary,
			paa.posts_count, paa.last_post_at, paa.created_at,
			p.id, p.page_id, p.page_name, p.category, 
			p.profile_picture_url, p.is_active
		FROM page_account_assignments paa
		JOIN pages p ON p.id = paa.page_id
		WHERE paa.account_id = $1
		ORDER BY p.page_name
	`

	rows, err := s.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []PageAccountAssignment
	for rows.Next() {
		var a PageAccountAssignment
		var p Page

		err := rows.Scan(
			&a.ID, &a.PageID, &a.AccountID, &a.IsPrimary,
			&a.PostsCount, &a.LastPostAt, &a.CreatedAt,
			&p.ID, &p.PageID, &p.PageName, &p.Category,
			&p.ProfilePictureURL, &p.IsActive,
		)
		if err != nil {
			return nil, err
		}

		a.Page = &p
		assignments = append(assignments, a)
	}

	return assignments, nil
}

// GetAssignmentsByPage lấy danh sách accounts quản lý 1 page
func (s *Store) GetAssignmentsByPage(pageID string) ([]PageAccountAssignment, error) {
	query := `
		SELECT 
			paa.id, paa.page_id, paa.account_id, paa.is_primary,
			paa.posts_count, paa.last_post_at, paa.created_at,
			fa.id, fa.fb_user_id, fa.fb_user_name, fa.status,
			fa.posts_today, fa.max_posts_per_day
		FROM page_account_assignments paa
		JOIN facebook_accounts fa ON fa.id = paa.account_id
		WHERE paa.page_id = $1
		ORDER BY paa.is_primary DESC, fa.fb_user_name
	`

	rows, err := s.db.Query(query, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []PageAccountAssignment
	for rows.Next() {
		var a PageAccountAssignment
		var acc FacebookAccount

		err := rows.Scan(
			&a.ID, &a.PageID, &a.AccountID, &a.IsPrimary,
			&a.PostsCount, &a.LastPostAt, &a.CreatedAt,
			&acc.ID, &acc.FbUserID, &acc.FbUserName, &acc.Status,
			&acc.PostsToday, &acc.MaxPostsPerDay,
		)
		if err != nil {
			return nil, err
		}

		a.Account = &acc
		assignments = append(assignments, a)
	}

	return assignments, nil
}

// AssignPageToAccount gán page vào account
func (s *Store) AssignPageToAccount(pageID, accountID string, isPrimary bool) error {
	// Nếu là primary, bỏ primary của các assignment khác
	if isPrimary {
		_, err := s.db.Exec(`
			UPDATE page_account_assignments 
			SET is_primary = false 
			WHERE page_id = $1 AND account_id != $2
		`, pageID, accountID)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO page_account_assignments (page_id, account_id, is_primary)
		VALUES ($1, $2, $3)
		ON CONFLICT (page_id, account_id) 
		DO UPDATE SET is_primary = $3
	`

	_, err := s.db.Exec(query, pageID, accountID, isPrimary)
	return err
}

// UnassignPageFromAccount bỏ gán page khỏi account
func (s *Store) UnassignPageFromAccount(pageID, accountID string) error {
	_, err := s.db.Exec(`
		DELETE FROM page_account_assignments 
		WHERE page_id = $1 AND account_id = $2
	`, pageID, accountID)
	return err
}

// SetPrimaryAccount đặt account làm primary cho page
func (s *Store) SetPrimaryAccount(pageID, accountID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Bỏ primary cũ
	_, err = tx.Exec(`
		UPDATE page_account_assignments 
		SET is_primary = false 
		WHERE page_id = $1
	`, pageID)
	if err != nil {
		return err
	}

	// Set primary mới
	_, err = tx.Exec(`
		UPDATE page_account_assignments 
		SET is_primary = true 
		WHERE page_id = $1 AND account_id = $2
	`, pageID, accountID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetPrimaryAccountForPage lấy account primary của page
func (s *Store) GetPrimaryAccountForPage(pageID string) (*FacebookAccount, error) {
	query := `
		SELECT 
			fa.id, fa.fb_user_id, fa.fb_user_name, fa.access_token,
			fa.token_expires_at, fa.max_pages, fa.max_posts_per_day,
			fa.status, fa.rate_limit_until, fa.posts_today,
			fa.last_post_at, fa.last_error_at, fa.consecutive_failures,
			fa.notes, fa.created_at, fa.updated_at
		FROM facebook_accounts fa
		JOIN page_account_assignments paa ON paa.account_id = fa.id
		WHERE paa.page_id = $1 AND paa.is_primary = true
	`

	var a FacebookAccount
	err := s.db.QueryRow(query, pageID).Scan(
		&a.ID, &a.FbUserID, &a.FbUserName, &a.AccessToken,
		&a.TokenExpiresAt, &a.MaxPages, &a.MaxPostsPerDay,
		&a.Status, &a.RateLimitUntil, &a.PostsToday,
		&a.LastPostAt, &a.LastErrorAt, &a.ConsecutiveFailures,
		&a.Notes, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// CountPagesByAccount đếm số page của 1 account
func (s *Store) CountPagesByAccount(accountID string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM page_account_assignments WHERE account_id = $1
	`, accountID).Scan(&count)
	return count, err
}

// CanAssignMorePages kiểm tra account còn có thể nhận thêm page không
func (s *Store) CanAssignMorePages(accountID string) (bool, error) {
	var count, maxPages int
	err := s.db.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM page_account_assignments WHERE account_id = $1),
			(SELECT max_pages FROM facebook_accounts WHERE id = $1)
	`, accountID).Scan(&count, &maxPages)
	if err != nil {
		return false, err
	}
	return count < maxPages, nil
}

// GetUnassignedPages lấy danh sách pages chưa được gán cho account nào
func (s *Store) GetUnassignedPages() ([]Page, error) {
	query := `
		SELECT p.id, p.page_id, p.page_name, p.category, 
			p.profile_picture_url, p.is_active, p.created_at
		FROM pages p
		LEFT JOIN page_account_assignments paa ON paa.page_id = p.id
		WHERE paa.id IS NULL
		ORDER BY p.page_name
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []Page
	for rows.Next() {
		var p Page
		var createdAt time.Time
		err := rows.Scan(
			&p.ID, &p.PageID, &p.PageName, &p.Category,
			&p.ProfilePictureURL, &p.IsActive, &createdAt,
		)
		if err != nil {
			return nil, err
		}
		p.CreatedAt = createdAt
		pages = append(pages, p)
	}

	return pages, nil
}
