package db

import (
	"time"

	"github.com/lib/pq"
)

func (s *Store) CreateScheduledPost(sp *ScheduledPost) error {
	query := `
		INSERT INTO scheduled_posts (post_id, page_id, account_id, scheduled_time, status, max_retries, time_slot_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at, retry_count
	`
	
	return s.db.QueryRow(
		query,
		sp.PostID,
		sp.PageID,
		sp.AccountID,
		sp.ScheduledTime,
		sp.Status,
		sp.MaxRetries,
		sp.TimeSlotID,
	).Scan(&sp.ID, &sp.CreatedAt, &sp.UpdatedAt, &sp.RetryCount)
}

func (s *Store) GetScheduledPosts(status string, limit, offset int) ([]ScheduledPost, error) {
	query := `
		SELECT 
			sp.id, sp.post_id, sp.page_id, sp.account_id, sp.scheduled_time, sp.status, 
			sp.retry_count, sp.max_retries, sp.created_at, sp.updated_at,
			p.content, p.media_urls, p.media_type, p.link_url,
			pg.id, pg.page_name, pg.profile_picture_url,
			fa.id, fa.fb_user_name, fa.profile_picture_url
		FROM scheduled_posts sp
		JOIN posts p ON sp.post_id = p.id
		JOIN pages pg ON sp.page_id = pg.id
		LEFT JOIN facebook_accounts fa ON sp.account_id = fa.id
		WHERE ($1 = '' OR sp.status = $1)
		ORDER BY sp.scheduled_time DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := s.db.Query(query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scheduled := make([]ScheduledPost, 0)
	for rows.Next() {
		var sp ScheduledPost
		sp.Post = &Post{}
		sp.Page = &Page{}
		
		var accountID, accountName, accountPicture *string
		var linkURL *string
		
		err := rows.Scan(
			&sp.ID, &sp.PostID, &sp.PageID, &sp.AccountID, &sp.ScheduledTime, &sp.Status,
			&sp.RetryCount, &sp.MaxRetries, &sp.CreatedAt, &sp.UpdatedAt,
			&sp.Post.Content, pq.Array(&sp.Post.MediaURLs), &sp.Post.MediaType, &linkURL,
			&sp.Page.ID, &sp.Page.PageName, &sp.Page.ProfilePictureURL,
			&accountID, &accountName, &accountPicture,
		)
		if err != nil {
			return nil, err
		}
		
		// Set link URL if exists
		if linkURL != nil {
			sp.Post.LinkURL = *linkURL
		}
		
		// Set account if exists
		if accountID != nil {
			sp.Account = &FacebookAccount{
				ID:                *accountID,
				FbUserName:        *accountName,
			}
			if accountPicture != nil {
				sp.Account.ProfilePictureURL = *accountPicture
			}
		}
		
		scheduled = append(scheduled, sp)
	}
	
	return scheduled, nil
}

func (s *Store) GetPendingScheduledPosts() ([]ScheduledPost, error) {
	// Truyền UTC time từ Go để đảm bảo so sánh chính xác
	// Không phụ thuộc vào timezone của PostgreSQL server
	nowUTC := time.Now().UTC()

	query := `
		SELECT 
			sp.id, sp.post_id, sp.page_id, sp.scheduled_time, sp.status, 
			sp.retry_count, sp.max_retries,
			p.content, p.media_urls, p.media_type, p.link_url,
			pg.page_id, pg.access_token
		FROM scheduled_posts sp
		JOIN posts p ON sp.post_id = p.id
		JOIN pages pg ON sp.page_id = pg.id
		WHERE sp.status = 'pending' 
		  AND sp.scheduled_time <= $1
		  AND pg.is_active = true
		ORDER BY sp.scheduled_time ASC
	`

	rows, err := s.db.Query(query, nowUTC)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scheduled := make([]ScheduledPost, 0)
	for rows.Next() {
		var sp ScheduledPost
		sp.Post = &Post{}
		sp.Page = &Page{}
		
		var linkURL *string
		
		err := rows.Scan(
			&sp.ID, &sp.PostID, &sp.PageID, &sp.ScheduledTime, &sp.Status,
			&sp.RetryCount, &sp.MaxRetries,
			&sp.Post.Content, pq.Array(&sp.Post.MediaURLs), &sp.Post.MediaType, &linkURL,
			&sp.Page.PageID, &sp.Page.AccessToken,
		)
		if err != nil {
			return nil, err
		}
		
		// Set link URL if exists
		if linkURL != nil {
			sp.Post.LinkURL = *linkURL
		}
		
		scheduled = append(scheduled, sp)
	}
	
	return scheduled, nil
}

func (s *Store) UpdateScheduledPostStatus(id, status string) error {
	_, err := s.db.Exec("UPDATE scheduled_posts SET status = $1 WHERE id = $2", status, id)
	return err
}

func (s *Store) IncrementRetryCount(id string) error {
	_, err := s.db.Exec("UPDATE scheduled_posts SET retry_count = retry_count + 1 WHERE id = $1", id)
	return err
}

func (s *Store) DeleteScheduledPost(id string) error {
	_, err := s.db.Exec("DELETE FROM scheduled_posts WHERE id = $1", id)
	return err
}


func (s *Store) UpdateScheduledPostAccount(id, accountID string) error {
	_, err := s.db.Exec("UPDATE scheduled_posts SET account_id = $1 WHERE id = $2", accountID, id)
	return err
}
