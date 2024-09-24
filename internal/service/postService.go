package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"news-feed/internal/entity"
	"news-feed/internal/repository"
	"news-feed/internal/storage"
	"news-feed/pkg/logger"
	"time"
)

type PostServiceInterface interface {
	CreatePost(text string, fileName string, userID int) (*entity.Post, error)
	GetPost(postID int) (entity.Post, error)
	EditPost(post entity.Post) error
	DeletePost(postID int) error
	CommentOnPost(postID int, comment string) error
	LikePost(postID int) error
	UploadImage(fileName string, file io.Reader) (string, error)
}

type PostService struct {
	postRepo    repository.PostRepositoryInterface
	storage     storage.MinioStorageInterface
	redisClient *redis.Client
}

func (s *PostService) CreatePost(text string, fileName string, userID int) (*entity.Post, error) {
	var preSignedURL string
	if text == "" {
		err := errors.New("empty post text")
		logger.LogError("Cannot create post without text")
		return nil, err
	}
	if fileName != "" {
		var err error
		preSignedURL, err = s.storage.GenerateFileURL(fileName)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to generate pre signed url %v", err))
			return nil, err
		}
	}

	post := entity.Post{
		ContentText:      text,
		ContentImagePath: fileName,
		UserID:           userID,
	}

	createdPost, err := s.postRepo.CreatePost(post)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to create post: %v", err))
		return nil, err
	}
	createdPost.ContentImagePath = preSignedURL
	go func() {
		ctx := context.Background()
		postCacheKey := fmt.Sprintf("post:%d", createdPost.ID)    // Cache key for the post
		userPostsCacheKey := fmt.Sprintf("user_posts:%d", userID) // Cache key for the user's posts

		// 1. Cache the post itself in Redis (using post ID as key)
		_, err := s.redisClient.HSet(
			ctx, postCacheKey, map[string]interface{}{
				"id":                createdPost.ID,
				"content_text":      createdPost.ContentText,
				"content_image_url": createdPost.ContentImagePath,
				"user_id":           createdPost.UserID,
				"created_at":        createdPost.CreatedAt.Format(time.RFC3339), // Store created_at as string
			},
		).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to cache post with ID %d: %v", createdPost.ID, err))
			return
		}

		// 2. Add the post ID to the user's list of post IDs (sorted set with post ID as the score)
		_, err = s.redisClient.ZAdd(
			ctx,
			userPostsCacheKey,
			redis.Z{
				Score:  float64(createdPost.ID),
				Member: createdPost.ID,
			},
		).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to cache post ID %d for user %d: %v", createdPost.ID, userID, err))
			return
		}

		// Optionally, set TTL on the cache entries to expire
		// Set expiration for user post list cache (e.g., 24 hours)
		s.redisClient.Expire(ctx, userPostsCacheKey, 24*time.Hour)
		// Set expiration for the post itself (e.g., 24 hours)
		s.redisClient.Expire(ctx, postCacheKey, 24*time.Hour)

		logger.LogInfo(fmt.Sprintf("Successfully cached post %d for user %d", createdPost.ID, userID))
	}()

	return createdPost, nil
}

func (s *PostService) UploadImage(fileName string, file io.Reader) (string, error) {
	// Call the storage interface's UploadFile method to upload the image
	imageURL, err := s.storage.UploadFile(fileName, file)
	if err != nil {
		return "", fmt.Errorf("could not upload image: %w", err)
	}

	return imageURL, nil
}

func (s *PostService) GetPost(postID int) (entity.Post, error) {
	return s.postRepo.GetPostByID(postID)
}

func (s *PostService) EditPost(post entity.Post) error {
	return s.postRepo.UpdatePost(post)
}

func (s *PostService) DeletePost(postID int) error {
	return s.postRepo.DeletePost(postID)
}

func (s *PostService) CommentOnPost(postID int, comment string) error {
	commentEntity := entity.Comment{Content: comment}
	return s.postRepo.CreateComment(postID, commentEntity)
}

func (s *PostService) LikePost(postID int) error {
	// Assuming we have some logic to get the current user's ID
	userID := 1 // Placeholder for user ID
	return s.postRepo.AddLike(postID, userID)
}
