package service

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"news-feed/internal/entity"
	"news-feed/internal/repository"
	"news-feed/pkg/logger"
	"strconv"
)

type FriendsServiceInterface interface {
	GetFriends(userID int, limit int, cursor int) ([]entity.User, int, error)
	FollowUser(currentUserID int, followedUserID int) (string, error)
	UnfollowUser(currentUserID int, unfollowedUserID int) (string, error)
	GetUserPosts(userID int, limit int, cursor int) ([]entity.Post, int, error)
}

type FriendsService struct {
	friendsRepo repository.FriendsRepositoryInterface
	postRepo    repository.PostRepositoryInterface
	userRepo    repository.UserRepositoryInterface
	redisClient *redis.Client
}

// GetFriends retrieves the list of friends for a user.
func (s *FriendsService) GetFriends(userID int, limit int, cursor int) ([]entity.User, int, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("friends:%d", userID)

	// Get followers from cache
	followerIDs, err := s.redisClient.ZRangeByScore(
		context.Background(), cacheKey, &redis.ZRangeBy{
			Min:    fmt.Sprintf("%d", cursor),
			Max:    "+inf",
			Offset: 0,
			Count:  int64(limit),
		},
	).Result()

	if err != nil {
		return nil, 0, err
	}

	// Initialize a slice for User entities and find the maximum ID
	var followers []entity.User
	maxID := 0

	for _, followerID := range followerIDs {
		userID, err := strconv.Atoi(followerID)
		if err != nil {
			logger.LogError(fmt.Sprintf("Error converting follower ID %s to int: %v", followerID, err))
			continue
		}
		userKey := fmt.Sprintf("user-data:%d", userID)
		userData, err := s.redisClient.HGetAll(context.Background(), userKey).Result()
		follower := entity.User{
			ID:             userID,
			HashedPassword: userData["hashed_password"],
			Salt:           userData["salt"],
			FirstName:      userData["first_name"],
			LastName:       userData["last_name"],
			Email:          userData["email"],
			Username:       userData["username"],
		}
		followers = append(followers, follower) // Append user to the slice
		if follower.ID > maxID {
			maxID = follower.ID
		}
	}

	if len(followers) > 0 {
		return followers, maxID, nil // next cursor = maxID+1
	}

	followers, nextCursor, err := s.friendsRepo.GetFriends(userID, limit, cursor)

	if err != nil {
		return nil, 0, err
	}

	go func() {
		// Prepare data for Redis sorted set
		for _, follower := range followers {

			// Cache user data in a hash.
			userKey := fmt.Sprintf("user-data:%d", follower.ID)
			_, err = s.redisClient.HSet(
				context.Background(), userKey, map[string]interface{}{
					"hashed_password": follower.HashedPassword,
					"salt":            follower.Salt,
					"first_name":      follower.FirstName,
					"last_name":       follower.LastName,
					"birthday":        follower.Birthday,
					"email":           follower.Email,
					"username":        follower.Username,
				},
			).Result()
			if err != nil {
				logger.LogError(fmt.Sprintf("Error when caching user %d: %v", follower.ID, err))
				return
			}

			_, err = s.redisClient.ZAdd(
				context.Background(),
				cacheKey,
				redis.Z{
					Score:  float64(follower.ID),
					Member: follower.ID,
				},
			).Result()
			if err != nil {
				logger.LogError(fmt.Sprintf("Error adding follower to cache: %v", err))
				return
			}
		}
	}()
	return followers, nextCursor, nil
}

// FollowUser follows a user and returns a message.
func (s *FriendsService) FollowUser(currentUserID int, followedUserID int) (string, error) {
	err := s.friendsRepo.FollowUser(currentUserID, followedUserID)
	if err != nil {
		return "Failed to follow user", err
	}

	go func() {
		user, err := s.userRepo.GetByUserID(followedUserID)
		if err != nil {
			logger.LogError(fmt.Sprintf("Error when get followed user %d with error %v", followedUserID, err))
			return
		}

		// Cache user data in a hash.
		userKey := fmt.Sprintf("user-data:%d", user.ID)
		_, err = s.redisClient.HSet(
			context.Background(), userKey, map[string]interface{}{
				"hashed_password": user.HashedPassword,
				"salt":            user.Salt,
				"first_name":      user.FirstName,
				"last_name":       user.LastName,
				"birthday":        user.Birthday,
				"email":           user.Email,
				"username":        user.Username,
			},
		).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Error when caching user %d: %v", user.ID, err))
			return
		}

		cacheKey := fmt.Sprintf("%d", currentUserID)
		_, err = s.redisClient.ZAdd(
			context.Background(),
			cacheKey,
			redis.Z{
				Score:  float64(followedUserID),
				Member: followedUserID,
			},
		).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Error when add followed user to cache %d with error %v", followedUserID, err))
			return
		}
	}()
	return "Successfully followed user", nil
}

// UnfollowUser unfollows a user and returns a message.
func (s *FriendsService) UnfollowUser(currentUserID int, unfollowedUserID int) (string, error) {
	err := s.friendsRepo.UnfollowUser(currentUserID, unfollowedUserID)
	if err != nil {
		return "Failed to unfollow user", err
	}
	go func() {
		cacheKey := fmt.Sprintf("%d", currentUserID)
		_, err = s.redisClient.ZRem(
			context.Background(),
			cacheKey,
			unfollowedUserID,
		).Result()
		if err != nil {
			logger.LogError(
				fmt.Sprintf(
					"Error when add followed user to cache %d with error %v", unfollowedUserID, err,
				),
			)
			return
		}
	}()
	return "Successfully unfollowed user", nil
}

// GetUserPosts retrieves the posts by a user.
func (s *FriendsService) GetUserPosts(userID int, limit int, cursor int) ([]entity.Post, int, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("posts:%d", userID)

	// Get posts from cache
	postIDs, err := s.redisClient.ZRangeByScore(
		context.Background(), cacheKey, &redis.ZRangeBy{
			Min:    fmt.Sprintf("%d", cursor),
			Max:    "+inf",
			Offset: 0,
			Count:  int64(limit),
		},
	).Result()

	if err != nil {
		return nil, 0, err
	}

	// Initialize a slice for User entities and find the maximum ID
	var posts []entity.Post
	maxID := 0

	for _, postIDStr := range postIDs {
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			logger.LogError(fmt.Sprintf("Error converting post ID %s to int: %v", postID, err))
			continue
		}
		postKey := fmt.Sprintf("user-posts:%d", postID)
		postData, err := s.redisClient.HGetAll(context.Background(), postKey).Result()
		if err != nil {
			logger.LogError(fmt.Sprintf("Error when get user posts for user %d: %v", userID, err))
			continue
		}
		post := entity.Post{
			ID:               postID,
			UserID:           userID,
			ContentText:      postData["content_text"],
			ContentImagePath: postData["content_image_path"],
		}
		posts = append(posts, post) // Append user to the slice
		if post.ID > maxID {
			maxID = post.ID
		}
	}

	if len(posts) > 0 {
		return posts, maxID, nil // next cursor = maxID+1
	}

	posts, nextCursor, err := s.postRepo.GetPostsByUserID(userID, limit, cursor)

	if err != nil {
		return nil, 0, err
	}

	go func() {
		// Prepare data for Redis sorted set
		for _, post := range posts {

			// Cache user data in a hash.
			postKey := fmt.Sprintf("user-posts:%d", post.ID)
			_, err = s.redisClient.HSet(
				context.Background(), postKey, map[string]interface{}{
					"id":                 post.ID,
					"content_text":       post.ContentText,
					"content_image_path": post.ContentImagePath,
				},
			).Result()
			if err != nil {
				logger.LogError(fmt.Sprintf("Error when caching post %d: %v", post.ID, err))
				return
			}

			_, err = s.redisClient.ZAdd(
				context.Background(),
				cacheKey,
				redis.Z{
					Score:  float64(post.ID),
					Member: post.ID,
				},
			).Result()
			if err != nil {
				logger.LogError(fmt.Sprintf("Error adding post to cache: %v", err))
				return
			}
		}
	}()
	return posts, nextCursor, nil
}
