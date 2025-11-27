package db

import (
	"database/sql"

	"github.com/lib/pq"
)

func (s *Store) CreatePost(post *Post) error {
	query := `
		INSERT INTO posts (content, media_urls, media_type, link_url, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	
	return s.db.QueryRow(
		query,
		post.Content,
		pq.Array(post.MediaURLs),
		post.MediaType,
		post.LinkURL,
		post.Status,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (s *Store) GetPosts(limit, offset int) ([]Post, error) {
	query := `SELECT id, content, media_urls, media_type, link_url, status, created_at, updated_at 
	          FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]Post, 0)
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Content, pq.Array(&p.MediaURLs), &p.MediaType, &p.LinkURL, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	
	return posts, nil
}

func (s *Store) GetPostByID(id string) (*Post, error) {
	query := `SELECT id, content, media_urls, media_type, link_url, status, created_at, updated_at 
	          FROM posts WHERE id = $1`
	
	var p Post
	err := s.db.QueryRow(query, id).Scan(
		&p.ID, &p.Content, pq.Array(&p.MediaURLs), &p.MediaType, &p.LinkURL, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (s *Store) UpdatePost(post *Post) error {
	query := `
		UPDATE posts 
		SET content = $1, media_urls = $2, media_type = $3, link_url = $4, status = $5
		WHERE id = $6
	`
	
	_, err := s.db.Exec(query, post.Content, pq.Array(post.MediaURLs), post.MediaType, post.LinkURL, post.Status, post.ID)
	return err
}

func (s *Store) DeletePost(id string) error {
	_, err := s.db.Exec("DELETE FROM posts WHERE id = $1", id)
	return err
}
