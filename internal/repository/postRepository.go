package repository

import (
	"database/sql"
	"news-feed/internal/entity"
)

type PostRepositoryInterface interface {
	CreatePost(post *entity.Post) error
}

type PostRepository struct {
	db *sql.DB
}

func (r *PostRepository) CreatePost(post *entity.Post) error {
	query := `INSERT INTO post (fk_user_id, content_text, content_image_path, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, post.UserID, post.ContentText, post.ContentImagePath, post.CreatedAt)
	return err
}
