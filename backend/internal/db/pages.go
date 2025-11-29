package db

import (
	"database/sql"
)

func (s *Store) CreateOrUpdatePage(page *Page) error {
	query := `
		INSERT INTO pages (page_id, page_name, access_token, token_expires_at, category, profile_picture_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (page_id) 
		DO UPDATE SET 
			page_name = EXCLUDED.page_name,
			access_token = EXCLUDED.access_token,
			token_expires_at = EXCLUDED.token_expires_at,
			category = EXCLUDED.category,
			profile_picture_url = EXCLUDED.profile_picture_url,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, page_id, page_name, category, profile_picture_url, is_active, created_at, updated_at
	`
	
	return s.db.QueryRow(
		query,
		page.PageID,
		page.PageName,
		page.AccessToken,
		page.TokenExpiresAt,
		page.Category,
		page.ProfilePictureURL,
	).Scan(&page.ID, &page.PageID, &page.PageName, &page.Category, &page.ProfilePictureURL, &page.IsActive, &page.CreatedAt, &page.UpdatedAt)
}

func (s *Store) GetPages() ([]Page, error) {
	query := `SELECT id, page_id, page_name, category, profile_picture_url, is_active, created_at, updated_at 
	          FROM pages ORDER BY created_at DESC`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := make([]Page, 0)
	for rows.Next() {
		var p Page
		err := rows.Scan(&p.ID, &p.PageID, &p.PageName, &p.Category, &p.ProfilePictureURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	
	return pages, nil
}

func (s *Store) GetPageByID(id string) (*Page, error) {
	query := `SELECT id, page_id, page_name, access_token, token_expires_at, category, profile_picture_url, is_active, created_at, updated_at 
	          FROM pages WHERE id = $1`
	
	var p Page
	err := s.db.QueryRow(query, id).Scan(
		&p.ID, &p.PageID, &p.PageName, &p.AccessToken, &p.TokenExpiresAt, 
		&p.Category, &p.ProfilePictureURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (s *Store) DeletePage(id string) error {
	// Lấy account_id của page trước khi xóa
	var accountID sql.NullString
	s.db.QueryRow(`
		SELECT account_id FROM page_account_assignments WHERE page_id = $1 LIMIT 1
	`, id).Scan(&accountID)

	// Xóa page (cascade sẽ xóa assignment)
	_, err := s.db.Exec("DELETE FROM pages WHERE id = $1", id)
	if err != nil {
		return err
	}

	// Nếu có account, kiểm tra còn page nào không
	if accountID.Valid && accountID.String != "" {
		var count int
		s.db.QueryRow(`
			SELECT COUNT(*) FROM page_account_assignments WHERE account_id = $1
		`, accountID.String).Scan(&count)

		// Nếu không còn page nào, xóa account
		if count == 0 {
			s.db.Exec("DELETE FROM facebook_accounts WHERE id = $1", accountID.String)
		}
	}

	return nil
}

func (s *Store) TogglePage(id string) error {
	_, err := s.db.Exec("UPDATE pages SET is_active = NOT is_active WHERE id = $1", id)
	return err
}

func (s *Store) GetActivePages() ([]Page, error) {
	query := `SELECT id, page_id, page_name, access_token, category, profile_picture_url, created_at, updated_at 
	          FROM pages WHERE is_active = true`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := make([]Page, 0)
	for rows.Next() {
		var p Page
		err := rows.Scan(&p.ID, &p.PageID, &p.PageName, &p.AccessToken, &p.Category, &p.ProfilePictureURL, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		p.IsActive = true
		pages = append(pages, p)
	}
	
	return pages, nil
}

// DeletePageByPageID xóa page theo Facebook page_id
func (s *Store) DeletePageByPageID(pageID string) error {
	// Lấy internal id trước
	var id string
	err := s.db.QueryRow("SELECT id FROM pages WHERE page_id = $1", pageID).Scan(&id)
	if err == sql.ErrNoRows {
		return nil // Page không tồn tại, không cần xóa
	}
	if err != nil {
		return err
	}

	// Dùng DeletePage để xóa (đã có logic xử lý cascade)
	return s.DeletePage(id)
}

// PageWithAccount là Page kèm thông tin account quản lý
type PageWithAccount struct {
	Page
	AccountID         *string `json:"account_id,omitempty"`
	AccountName       *string `json:"account_name,omitempty"`
	AccountPictureURL *string `json:"account_picture_url,omitempty"`
}

// GetPagesWithAccount lấy danh sách pages kèm thông tin account primary
func (s *Store) GetPagesWithAccount() ([]PageWithAccount, error) {
	query := `
		SELECT 
			p.id, p.page_id, p.page_name, p.category, 
			p.profile_picture_url, p.is_active, p.created_at, p.updated_at,
			fa.id, fa.fb_user_name, fa.profile_picture_url
		FROM pages p
		LEFT JOIN page_account_assignments paa ON paa.page_id = p.id AND paa.is_primary = true
		LEFT JOIN facebook_accounts fa ON fa.id = paa.account_id
		ORDER BY p.created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := make([]PageWithAccount, 0)
	for rows.Next() {
		var p PageWithAccount
		var accountID, accountName, accountPicture sql.NullString

		err := rows.Scan(
			&p.ID, &p.PageID, &p.PageName, &p.Category,
			&p.ProfilePictureURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&accountID, &accountName, &accountPicture,
		)
		if err != nil {
			return nil, err
		}

		if accountID.Valid {
			p.AccountID = &accountID.String
		}
		if accountName.Valid {
			p.AccountName = &accountName.String
		}
		if accountPicture.Valid {
			p.AccountPictureURL = &accountPicture.String
		}

		pages = append(pages, p)
	}

	return pages, nil
}
