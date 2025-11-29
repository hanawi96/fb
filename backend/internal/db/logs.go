package db

import "github.com/lib/pq"

func (s *Store) CreatePostLog(log *PostLog) error {
	query := `
		INSERT INTO post_logs (scheduled_post_id, post_id, page_id, facebook_post_id, status, error_message, response_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, posted_at
	`

	// Đảm bảo response_data là JSON hợp lệ (null nếu empty)
	responseData := log.ResponseData
	if responseData == "" {
		responseData = "{}"
	}

	return s.db.QueryRow(
		query,
		log.ScheduledPostID,
		log.PostID,
		log.PageID,
		log.FacebookPostID,
		log.Status,
		log.ErrorMessage,
		responseData,
	).Scan(&log.ID, &log.PostedAt)
}

func (s *Store) GetPostLogs(limit, offset int) ([]PostLog, error) {
	query := `
		SELECT 
			pl.id, pl.scheduled_post_id, pl.post_id, pl.page_id, 
			pl.facebook_post_id, pl.status, pl.error_message, pl.posted_at,
			p.content, p.media_urls,
			pg.page_name, pg.profile_picture_url
		FROM post_logs pl
		JOIN posts p ON pl.post_id = p.id
		JOIN pages pg ON pl.page_id = pg.id
		ORDER BY pl.posted_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]PostLog, 0)
	for rows.Next() {
		var log PostLog
		log.Post = &Post{}
		log.Page = &Page{}
		
		err := rows.Scan(
			&log.ID, &log.ScheduledPostID, &log.PostID, &log.PageID,
			&log.FacebookPostID, &log.Status, &log.ErrorMessage, &log.PostedAt,
			&log.Post.Content, pq.Array(&log.Post.MediaURLs),
			&log.Page.PageName, &log.Page.ProfilePictureURL,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}
