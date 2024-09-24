package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"news-feed/internal/entity"
	"news-feed/pkg/logger"
)

type PostRepositoryInterface interface {
	CreatePost(post entity.Post) (*entity.Post, error)
	GetPostByID(id int) (entity.Post, error)
	UpdatePost(post entity.Post) error
	DeletePost(id int) error
	CreateComment(postID int, comment entity.Comment) error
	AddLike(postID int, userID int) error
	GetPostsByUserID(userID int) ([]entity.Post, error)
	GetAllPosts() ([]entity.Post, error)
}

type PostRepository struct {
	db *sql.DB
}

func (r *PostRepository) CreatePost(post entity.Post) (*entity.Post, error) {
	// Insert the post without using RETURNING
	result, err := r.db.Exec(
		`
		INSERT INTO post (content_text, content_image_path, fk_user_id) VALUES (?, ?, ?)`,
		post.ContentText, post.ContentImagePath, post.UserID,
	)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error while inserting new post: %v", err))
		return nil, err
	}

	// Retrieve the last inserted post ID using LAST_INSERT_ID()
	postID, err := result.LastInsertId()
	if err != nil {
		logger.LogError(fmt.Sprintf("Error while retrieving last inserted ID: %v", err))
		return nil, err
	}

	// Query the inserted post to get full details, including created_at
	var createdPost entity.Post
	err = r.db.QueryRow(
		`SELECT id, content_text, content_image_path, fk_user_id, created_at 
		FROM post WHERE id = ?`, postID,
	).Scan(
		&createdPost.ID, &createdPost.ContentText, &createdPost.ContentImagePath, &createdPost.UserID,
		&createdPost.CreatedAt,
	)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error while retrieving created post: %v", err))
		return nil, err
	}

	// Return the created post with all details
	return &createdPost, nil
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
		if errors.Is(err, sql.ErrNoRows) {
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

func (r *PostRepository) GetPostsByUserID(userID int) ([]entity.Post, error) {
	rows, err := r.db.Query("SELECT id, content_text, content_image_path FROM post WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
			return
		}
	}(rows)

	var posts []entity.Post
	for rows.Next() {
		var post entity.Post
		if err := rows.Scan(&post.ID, &post.ContentText, &post.ContentImagePath); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *PostRepository) GetAllPosts() ([]entity.Post, error) {
	rows, err := r.db.Query("SELECT id, content_text, content_image_path FROM post")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
			return
		}
	}(rows)

	var posts []entity.Post
	for rows.Next() {
		var post entity.Post
		if err := rows.Scan(&post.ID, &post.ContentText, &post.ContentImagePath); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
