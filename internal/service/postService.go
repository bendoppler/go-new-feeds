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
	"strconv"
	"time"
)

type PostServiceInterface interface {
	CreatePost(text string, fileName string, userID int) (*entity.Post, error)
	GetPost(postID int) (*entity.Post, error)
	EditPost(post entity.Post) (*entity.Post, error)
	DeletePost(postID int, userID int) error
	CommentOnPost(postID int, userID int, comment string) (*entity.Comment, error)
	LikePost(postID int, userID int) error
	UploadImage(fileName string, file io.Reader) (string, error)
	GetComments(postID int, cursor int, limit int) ([]entity.Comment, int, error)
	GetLikes(postID int, cursor time.Time, limit int) ([]entity.User, *time.Time, error)
	GetLikeCount(postID int) (int, error)
}

type PostService struct {
	postRepo    repository.PostRepositoryInterface
	storage     storage.MinioStorageInterface
	redisClient *redis.Client
	userService UserServiceInterface
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
		userPostsCacheKey := fmt.Sprintf("user-posts:%d", userID) // Cache key for the user's posts

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

func (s *PostService) GetPost(postID int) (*entity.Post, error) {
	ctx := context.Background()
	postCacheKey := fmt.Sprintf("post:%d", postID)

	// Try to get the post from Redis cache first
	cachedPostData, err := s.redisClient.HGetAll(ctx, postCacheKey).Result()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to get post %d from cache: %v", postID, err))
	}

	// If cache is found and has data
	if len(cachedPostData) > 0 {
		var post entity.Post

		// Parse the cached data back into the Post struct
		post.ID, _ = strconv.Atoi(cachedPostData["id"])
		post.ContentText = cachedPostData["content_text"]
		post.ContentImagePath = cachedPostData["content_image_url"]
		post.UserID, _ = strconv.Atoi(cachedPostData["user_id"])

		// Parse the created_at field into time.Time
		createdAt, err := time.Parse(time.RFC3339, cachedPostData["created_at"])
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to parse created_at for post %d: %v", postID, err))
			return nil, err
		}
		post.CreatedAt = createdAt

		logger.LogInfo(fmt.Sprintf("Successfully retrieved post %d from cache", postID))
		return &post, nil
	}
	return s.postRepo.GetPostByID(postID)
}

func (s *PostService) EditPost(post entity.Post) (*entity.Post, error) {
	// 1. Update the post in the database
	updatedPost, err := s.postRepo.UpdatePost(post)
	if err != nil {
		return nil, err
	}

	// 2. Update the post in Redis cache
	go func() {
		ctx := context.Background()
		postCacheKey := fmt.Sprintf("post:%d", post.ID)

		_, err := s.redisClient.HSet(
			ctx, postCacheKey, map[string]interface{}{
				"id":                updatedPost.ID,
				"content_text":      updatedPost.ContentText,
				"content_image_url": updatedPost.ContentImagePath,
				"user_id":           updatedPost.UserID,
				"created_at":        updatedPost.CreatedAt.Format(time.RFC3339), // Store created_at as string
			},
		).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to update cache for post ID %d: %v", updatedPost.ID, err))
		} else {
			logger.LogInfo(fmt.Sprintf("Successfully updated cache for post ID %d", updatedPost.ID))
		}
	}()

	return updatedPost, nil
}

func (s *PostService) DeletePost(postID int, userID int) error {
	// 1. Delete the post from the database
	err := s.postRepo.DeletePost(postID)
	if err != nil {
		return err
	}

	// 2. Remove the post from Redis cache
	go func() {
		ctx := context.Background()
		postCacheKey := fmt.Sprintf("post:%d", postID)
		userPostsCacheKey := fmt.Sprintf("user_posts:%d", userID)

		// Delete the post from the cache
		_, err := s.redisClient.Del(ctx, postCacheKey).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to delete cache for post ID %d: %v", postID, err))
		} else {
			logger.LogInfo(fmt.Sprintf("Successfully deleted cache for post ID %d", postID))
		}

		// Remove the post ID from the user's post list
		_, err = s.redisClient.ZRem(ctx, userPostsCacheKey, postID).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to remove post ID %d from user post list: %v", postID, err))
		} else {
			logger.LogInfo(fmt.Sprintf("Successfully removed post ID %d from user post list", postID))
		}
	}()

	return nil
}

func (s *PostService) CommentOnPost(postID int, userID int, comment string) (*entity.Comment, error) {
	commentEntity := entity.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: comment,
	}

	// 1. Add the comment to the database
	createdComment, err := s.postRepo.CreateComment(commentEntity)
	if err != nil {
		return nil, err
	}

	// 2. Update cache with the new comment using ZADD
	go func() {
		ctx := context.Background()
		postCommentsCacheKey := fmt.Sprintf("comments:post:%d", postID) // Cache key for post comments sorted set
		commentCacheKey := fmt.Sprintf("comment:%d", createdComment.ID) // Cache key for the comment hash

		// Add comment details to the comment cache (hash)
		_, err := s.redisClient.HSet(
			ctx, commentCacheKey, map[string]interface{}{
				"id":         createdComment.ID,
				"user_id":    createdComment.UserID,
				"post_id":    createdComment.PostID,
				"content":    createdComment.Content,
				"created_at": createdComment.CreatedAt.Format(time.RFC3339), // Store created_at as string
			},
		).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to cache comment ID %d: %v", createdComment.ID, err))
			return
		}

		// Add the comment ID to the post's comment list in the cache (sorted set)
		// Use `createdComment.CreatedAt.Unix()` as the score to sort by timestamp
		_, err = s.redisClient.ZAdd(
			ctx,
			postCommentsCacheKey,
			redis.Z{
				Score:  float64(createdComment.ID), // Use created_at timestamp as score
				Member: createdComment.ID,
			},
		).Result()
		if err != nil {
			logger.LogError(
				fmt.Sprintf(
					"Failed to add comment ID %d to post comments cache for post ID %d: %v", createdComment.ID, postID,
					err,
				),
			)
		}

		// Optionally, set TTL on the comment cache entry
		s.redisClient.Expire(ctx, commentCacheKey, 24*time.Hour)
		s.redisClient.Expire(ctx, postCommentsCacheKey, 24*time.Hour)

		logger.LogInfo(fmt.Sprintf("Successfully cached comment ID %d for post ID %d", createdComment.ID, postID))
	}()

	return createdComment, nil
}

func (s *PostService) LikePost(postID int, userID int) error {
	// Add the like in the repository (database)
	like, err := s.postRepo.AddLike(postID, userID)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to add like to post %d by user %d: %v", postID, userID, err))
		return err
	}

	// Update the cache asynchronously
	go func() {
		ctx := context.Background()
		postLikeKey := fmt.Sprintf("post_likes:%d", postID)  // Cache key for the post's likes set
		userLikesKey := fmt.Sprintf("user_likes:%d", postID) // Cache key for the user's liked posts sorted set

		// Cache the like itself in the sorted set for user's likes
		_, err := s.redisClient.ZAdd(
			ctx,
			userLikesKey,
			redis.Z{
				Score:  float64(like.CreatedAt.UnixMilli()), // Use a timestamp as the score
				Member: userID,
			},
		).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to cache like for post %d by user %d: %v", postID, userID, err))
			return
		}

		// Add the user ID to the set of likes for this post
		_, err = s.redisClient.SAdd(ctx, postLikeKey, userID).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to cache user %d liking post %d: %v", userID, postID, err))
			return
		}

		// Set expiration for user like list cache and post like count cache (e.g., 24 hours)
		s.redisClient.Expire(ctx, userLikesKey, 24*time.Hour)
		s.redisClient.Expire(ctx, postLikeKey, 24*time.Hour)

		logger.LogInfo(fmt.Sprintf("Successfully cached like for post %d by user %d", postID, userID))
	}()

	return nil
}

func (s *PostService) GetComments(postID int, cursor int, limit int) ([]entity.Comment, int, error) {
	postCommentsCacheKey := fmt.Sprintf("comments:post:%d", postID) // Cache key for post comments sorted set
	// Attempt to fetch comments from cache
	commentIDs, err := s.redisClient.ZRangeByScore(
		context.Background(),
		postCommentsCacheKey,
		&redis.ZRangeBy{
			Min:   fmt.Sprintf("%d", cursor), // Start from the cursor
			Max:   "+inf",                    // Up to the maximum value
			Count: int64(limit),              // Limit the number of comments
		},
	).Result()

	if err != nil {
		return nil, 0, err
	}

	var comments []entity.Comment
	maxID := 0

	for _, commentIDStr := range commentIDs {
		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to parse comment ID %s: %v", commentIDStr, err))
			continue
		}
		commentCacheKey := fmt.Sprintf("comment:%d", commentID) // Cache key for the comment hash
		commentData, err := s.redisClient.HGetAll(context.Background(), commentCacheKey).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to get comment data %d: %v", commentID, err))
			continue
		}
		comment := entity.Comment{
			PostID:  postID,
			ID:      commentID,
			Content: commentData["content"],
		}
		comments = append(comments, comment)
		maxID = max(maxID, commentID)
	}

	if len(comments) > 0 {
		return comments, maxID, nil
	}

	// If not found in cache, query the database
	comments, nextCursor, err := s.postRepo.GetComments(postID, cursor, limit)
	if err != nil {
		return nil, nextCursor, err
	}

	// Cache the retrieved comments
	go func() {
		for _, comment := range comments {
			commentCacheKey := fmt.Sprintf("comment:%d", comment.ID) // Cache key for the comment hash
			_, err := s.redisClient.HSet(
				context.Background(), commentCacheKey, map[string]interface{}{
					"id":      comment.ID,
					"post_id": comment.PostID,
					"content": comment.Content,
				},
			).Result()
			if err != nil {
				logger.LogError(fmt.Sprintf("Failed to cache comment %d: %v", comment.ID, err))
				return
			}
			_, err = s.redisClient.ZAdd(
				context.Background(),
				postCommentsCacheKey,
				redis.Z{
					Score:  float64(comment.ID),
					Member: comment.ID,
				},
			).Result()
			if err != nil {
				logger.LogError(fmt.Sprintf("Failed to cache comment %d for post %d: %v", comment.ID, postID, err))
			}
		}
	}()

	return comments, nextCursor, nil
}

func (s *PostService) GetLikes(postID int, cursor time.Time, limit int) ([]entity.User, *time.Time, error) {
	userLikesKey := fmt.Sprintf("user_likes:%d", postID) // Cache key for the user's liked posts sorted set
	// Attempt to fetch likes from cache
	cachedLikes, err := s.redisClient.ZRangeByScoreWithScores(
		context.Background(),
		userLikesKey,
		&redis.ZRangeBy{
			Min:   fmt.Sprintf("%d", cursor.UnixMilli()),
			Max:   "+inf",
			Count: int64(limit),
		},
	).Result()
	if err != nil {
		logger.LogError(fmt.Sprintf("Error while getting likes for post %d: %v", postID, err))
		return nil, nil, err
	}
	if len(cachedLikes) > 0 {
		// If likes are found in cache, convert them back to user structs
		users, err := s.convertCacheToUsers(cachedLikes)
		lastItem := cachedLikes[len(cachedLikes)-1]
		// Convert milliseconds to seconds and nanoseconds
		timestampInMillis := int64(lastItem.Score)
		seconds := timestampInMillis / 1000
		nanoseconds := (timestampInMillis % 1000) * int64(time.Millisecond)

		// Create time.Time object and store it as nextCursor
		nextCursor := time.Unix(seconds, nanoseconds).UTC()
		if err != nil {
			logger.LogError(fmt.Sprintf("Error while converting likes for post %d: %v", postID, err))
			return nil, nil, err
		}
		return users, &nextCursor, nil
	}

	// If not found in cache, query the database
	likes, nextCursor, err := s.postRepo.GetLikes(postID, cursor, limit)
	if err != nil {
		return nil, nil, err
	}

	// Cache the retrieved likes
	go func() {
		for _, like := range likes {
			_, err := s.redisClient.ZAdd(
				context.Background(),
				userLikesKey,
				redis.Z{
					Score:  float64(like.CreatedAt.UnixMilli()),
					Member: like.UserID, // Store user ID or a user struct
				},
			).Result()
			if err != nil {
				logger.LogError(fmt.Sprintf("Failed to cache like %d for post %d: %v", like.UserID, postID, err))
			}
		}
	}()
	users, err := s.convertCacheToUsers(cachedLikes)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error while converting likes for post %d: %v", postID, err))
		return nil, nil, err
	}
	return users, nextCursor, nil
}

func (s *PostService) convertCacheToUsers(cachedLikes []redis.Z) ([]entity.User, error) {
	// Implement conversion logic here
	var userIDs []int
	for _, cachedLike := range cachedLikes {
		// Convert each string to int
		userID, err := strconv.Atoi(cachedLike.Member.(string))
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to parse user ID %v", err))
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return s.userService.GetUsers(userIDs)
}

func (s *PostService) convertToUsers(likes []entity.Like) ([]entity.User, error) {
	// Implement conversion logic here
	var userIDs []int
	for _, like := range likes {
		userIDs = append(userIDs, like.UserID)
	}

	return s.userService.GetUsers(userIDs)
}

// GetLikeCount retrieves the like count for a specific post, first checking the cache, then the database if necessary.
func (s *PostService) GetLikeCount(postID int) (int, error) {
	ctx := context.Background()
	postLikeKey := fmt.Sprintf("post_likes:%d", postID) // Cache key for the post's likes set

	likeCount, err := s.redisClient.SCard(ctx, postLikeKey).Result()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to get like count from Redis for post ID %d: %v", postID, err))
		return 0, err
	}

	if likeCount == 0 {
		intLikeCount, err := s.postRepo.GetLikeCount(postID)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to retrieve like count from database for post ID %d: %v", postID, err))
			return 0, err
		}
		return intLikeCount, nil
	}

	return int(likeCount), nil
}
