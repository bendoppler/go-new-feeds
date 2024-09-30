package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"news-feed/internal/entity"
	"news-feed/pkg/logger"
	"time"
)

type PostRepositoryInterface interface {
	CreatePost(post entity.Post) (*entity.Post, error)
	GetPostByID(id int) (*entity.Post, error)
	UpdatePost(post entity.Post) (*entity.Post, error)
	DeletePost(id int) error
	CreateComment(comment entity.Comment) (*entity.Comment, error)
	AddLike(postID int, userID int) (*entity.Like, error)
	GetPostsByUserID(userID int, limit int, cursor int) ([]entity.Post, int, error)
	GetAllPosts() ([]entity.Post, error)
	GetComments(postID int, cursor int, limit int) ([]entity.Comment, int, error)
	GetLikes(postID int, cursor time.Time, limit int) ([]entity.Like, *time.Time, error)
	GetLikeCount(postID int) (int, error)
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

func (r *PostRepository) GetPostByID(id int) (*entity.Post, error) {
	var post entity.Post
	row := r.db.QueryRow(
		`
		SELECT id, content_text, content_image_path, fk_user_id 
		FROM post 
		WHERE id = ?`, id,
	)
	err := row.Scan(&post.ID, &post.ContentText, &post.ContentImagePath, &post.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) UpdatePost(post entity.Post) (*entity.Post, error) {
	_, err := r.db.Exec(
		`
		UPDATE post 
		SET content_text = ?
		WHERE id = ?`,
		post.ContentText, post.ID,
	)

	if err != nil {
		return nil, err
	}

	// Retrieve the updated post if the update was successful
	var updatedPost entity.Post
	err = r.db.QueryRow(
		`
		SELECT id, fk_user_id, content_text, content_image_path, created_at 
		FROM post 
		WHERE id = ?`,
		post.ID,
	).Scan(
		&updatedPost.ID,
		&updatedPost.UserID,
		&updatedPost.ContentText,
		&updatedPost.ContentImagePath,
		&updatedPost.CreatedAt,
	)

	if err != nil {
		// Return error if retrieval fails
		return nil, err
	}
	// Return the updated post and nil for error
	return &updatedPost, nil
}

func (r *PostRepository) DeletePost(id int) error {
	_, err := r.db.Exec(`DELETE FROM post WHERE id = ?`, id)
	return err
}

func (r *PostRepository) CreateComment(comment entity.Comment) (*entity.Comment, error) {
	// Execute the insert query
	result, err := r.db.Exec(
		`INSERT INTO comment (fk_post_id, fk_user_id, content) VALUES (?, ?, ?)`,
		comment.PostID, comment.UserID, comment.Content,
	)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error while creating comment: %v", err))
		return nil, err
	}

	// Get the last inserted comment ID
	commentID, err := result.LastInsertId()
	if err != nil {
		logger.LogError(fmt.Sprintf("Error getting last inserted comment ID: %v", err))
		return nil, err
	}

	// Query the inserted comment using the retrieved comment ID
	query := `SELECT id, fk_post_id, fk_user_id, content, created_at FROM comment WHERE id = ?`
	err = r.db.QueryRow(query, commentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
	)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error retrieving created comment: %v", err))
		return nil, err
	}

	// Return the populated comment entity
	return &comment, nil
}

func (r *PostRepository) AddLike(postID int, userID int) (*entity.Like, error) {
	_, err := r.db.Exec(
		`
		INSERT INTO like (fk_post_id, fk_user_id) 
		VALUES (?, ?)`,
		postID, userID,
	)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error while inserting new like: %v", err))
		return nil, err
	}
	// Query the newly inserted like from the database
	like := &entity.Like{}
	err = r.db.QueryRow(
		`
		SELECT fk_post_id, fk_user_id, created_at 
		FROM likes 
		WHERE fk_post_id = ? AND fk_user_id = ?
		ORDER BY created_at DESC
		LIMIT 1`,
		postID, userID,
	).Scan(&like.PostID, &like.UserID, &like.CreatedAt)

	if err != nil {
		logger.LogError(fmt.Sprintf("Error while retrieving like: %v", err))
		return nil, err
	}

	return like, nil
}

func (r *PostRepository) GetPostsByUserID(userID int, limit int, cursor int) ([]entity.Post, int, error) {
	rows, err := r.db.Query(
		"SELECT id, content_text, content_image_path FROM post p WHERE p.fk_user_id = ? AND p.id > ? ORDER BY id ASC LIMIT ?",
		userID, cursor, limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
			return
		}
	}(rows)

	var posts []entity.Post
	var nextCursor int
	for rows.Next() {
		var post entity.Post
		if err := rows.Scan(&post.ID, &post.ContentText, &post.ContentImagePath); err != nil {
			return nil, 0, err
		}
		posts = append(posts, post)
		nextCursor = post.ID // Update nextCursor with the last post id
	}
	return posts, nextCursor, nil
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

func (r *PostRepository) GetComments(postID int, cursor int, limit int) ([]entity.Comment, int, error) {
	rows, err := r.db.Query(
		`SELECT id, fk_post_id, fk_user_id, content FROM comment WHERE fk_post_id = ? AND id > ? ORDER BY id ASC LIMIT ?`,
		postID, cursor, limit,
	)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error while retrieving comments for post %d: %v", postID, err))
		return nil, 0, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
			return
		}
	}(rows)

	var comments []entity.Comment
	var nextCursor int
	for rows.Next() {
		var comment entity.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content); err != nil {
			logger.LogError(fmt.Sprintf("Error while scanning comment: %v", err))
			return nil, 0, err
		}
		comments = append(comments, comment)
		nextCursor = max(nextCursor, comment.ID)
	}
	return comments, nextCursor, nil
}

func (r *PostRepository) GetLikes(postID int, cursor time.Time, limit int) ([]entity.Like, *time.Time, error) {
	rows, err := r.db.Query(
		`SELECT fk_post_id, fk_user_id, created_at FROM like WHERE fk_post_id=? AND created_at > ? ORDER BY created_at ASC LIMIT ?`,
		postID, cursor, limit,
	)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error while retrieving likes for post %d: %v", postID, err))
		return nil, nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
			return
		}
	}(rows)

	var likes []entity.Like
	var nextCursor time.Time
	for rows.Next() {
		var like entity.Like
		if err := rows.Scan(&like.UserID, &like.PostID, &like.CreatedAt); err != nil {
			logger.LogError(fmt.Sprintf("Error while scanning likes: %v", err))
			return nil, nil, err
		}
		likes = append(likes, like)
		nextCursor = likes[len(likes)-1].CreatedAt
	}
	return likes, &nextCursor, nil
}

func (r *PostRepository) GetLikeCount(postID int) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM likes WHERE fk_post_id = ?`,
		postID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
