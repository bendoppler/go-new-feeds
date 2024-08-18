package repository

import (
	"database/sql"
	"fmt"
	"news-feed/internal/entity"
)

type PostRepositoryInterface interface {
	CreatePost(post entity.Post) error
	GetPostByID(id int) (entity.Post, error)
	UpdatePost(post entity.Post) error
	DeletePost(id int) error
	CreateComment(postID int, comment entity.Comment) error
	AddLike(postID int, userID int) error
}

type PostRepository struct {
	db *sql.DB
}

func (r *PostRepository) CreatePost(post entity.Post) error {
	_, err := r.db.Exec(
		`
		INSERT INTO post (content_text, content_image_path, user_id) 
		VALUES (?, ?, ?)`,
		post.ContentText, post.ContentImagePath, post.UserID,
	)
	return err
}

func (r *PostRepository) GetPostByID(id int) (entity.Post, error) {
	var post entity.Post
	row := r.db.QueryRow(
		`
		SELECT id, text, image, user_id 
		FROM post 
		WHERE id = ?`, id,
	)
	err := row.Scan(&post.ID, &post.ContentText, &post.ContentImagePath, &post.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return post, fmt.Errorf("post not found")
		}
		return post, err
	}
	return post, nil
}

func (r *PostRepository) UpdatePost(post entity.Post) error {
	_, err := r.db.Exec(
		`
		UPDATE post 
		SET content_text = ?, content_image_path = ? 
		WHERE id = ?`,
		post.ContentText, post.ContentImagePath, post.ID,
	)
	return err
}

func (r *PostRepository) DeletePost(id int) error {
	_, err := r.db.Exec(`DELETE FROM post WHERE id = ?`, id)
	return err
}

func (r *PostRepository) CreateComment(postID int, comment entity.Comment) error {
	_, err := r.db.Exec(
		`
		INSERT INTO comment (post_id, content) 
		VALUES (?, ?)`,
		postID, comment.Content,
	)
	return err
}

func (r *PostRepository) AddLike(postID int, userID int) error {
	_, err := r.db.Exec(
		`
		INSERT INTO like (post_id, user_id) 
		VALUES (?, ?)`,
		postID, userID,
	)
	return err
}
