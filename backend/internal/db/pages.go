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
	_, err := s.db.Exec("DELETE FROM pages WHERE id = $1", id)
	return err
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
