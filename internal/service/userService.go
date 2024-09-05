package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"news-feed/internal/entity"
	"news-feed/internal/repository"
	"news-feed/pkg/logger"
	"news-feed/pkg/middleware"
	"time"
)

// UserServiceInterface defines methods for user-related business logic.
type UserServiceInterface interface {
	Login(username, password string) (string, error)
	Signup(user entity.User) (string, error)
	EditProfile(user entity.User) error
}

// UserService is a concrete implementation of UserServiceInterface.
type UserService struct {
	userRepo    repository.UserRepositoryInterface
	redisClient *redis.Client
}

func (s *UserService) Signup(user entity.User) (string, error) {
	// Generate salt
	salt := generateSalt()

	// Hash the password with the salt
	hashedPassword := hashPassword(user.Password, salt)

	user.HashedPassword = hashedPassword
	user.Salt = salt

	err := s.userRepo.CreateUser(user)
	if err != nil {
		return "", err
	}
	// Generate JWT
	jwtToken, err := middleware.GenerateJWT(string(rune(user.ID)))
	if err != nil {
		return "", fmt.Errorf("could not generate JWT: %v", err)
	}

	// Store the JWT in Redis with a TTL (e.g., 24 hours)
	err = s.redisClient.Set(context.Background(), jwtToken, user.Username, 24*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("could not store JWT in Redis: %v", err)
	}

	return jwtToken, nil
}

func (s *UserService) Login(username, password string) (string, error) {
	// Define the Redis key for the user
	redisKey := fmt.Sprintf("user:%s", username)

	// Check if user data is cached in Redis
	cachedUserData, err := s.redisClient.HGetAll(context.Background(), redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) || len(cachedUserData) == 0 {
			// User not in cache, fetch from DB and cache it
			localCachedUser, s2, err2 := s.getUserFromDBAndCache(username)
			if err2 != nil {
				return s2, err2
			}
			cachedUserData = map[string]string{
				"hashedPassword": localCachedUser.HashedPassword,
				"salt":           localCachedUser.Salt,
			}
		} else {
			logger.LogError(fmt.Sprintf("Error when getting user data from Redis: %v", err))
			return "", fmt.Errorf("could not retrieve user data from Redis: %v", err)
		}
	}

	// Extract user fields from the cached data
	hashedPassword, passwordExists := cachedUserData["hashedPassword"]
	salt, saltExists := cachedUserData["salt"]

	// If some fields are missing, fetch from the database
	if !passwordExists || !saltExists {
		localCachedUser, s2, err2 := s.getUserFromDBAndCache(username)
		if err2 != nil {
			return s2, err2
		}
		hashedPassword = localCachedUser.HashedPassword
		salt = localCachedUser.Salt
	}

	// Verify the password
	if !verifyPassword(password, hashedPassword, salt) {
		logger.LogError(fmt.Sprintf("Error when verifying password: %v", err))
		return "", fmt.Errorf("invalid credentials")
	}

	// Generate JWT
	jwtToken, err := middleware.GenerateJWT(username)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error when generate JWT: %v", err))
		return "", fmt.Errorf("could not generate JWT: %v", err)
	}
	return jwtToken, nil
}

func (s *UserService) getUserFromDBAndCache(username string) (entity.User, string, error) {
	// Cache miss, fetch user from the database
	user, err := s.userRepo.GetByUserName(username)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error when get user from db %v", err))
		return entity.User{}, "", err
	}

	// Cache the user data in Redis using a hash set
	redisKey := fmt.Sprintf("user:%s", username)
	userCacheData := map[string]interface{}{
		"hashedPassword": user.HashedPassword,
		"salt":           user.Salt,
	}

	err = s.redisClient.HSet(context.Background(), redisKey, userCacheData).Err()
	if err != nil {
		logger.LogError(fmt.Sprintf("Error when caching user data in Redis: %v", err))
		return user, "Could not cache user data", err
	}

	return user, "", nil
}

// EditProfile updates a user's profile.
func (s *UserService) EditProfile(user entity.User) error {
	existingUser, err := s.userRepo.GetByUserName(user.Username)
	if err != nil {
		return err
	}
	if (existingUser == entity.User{}) {
		return fmt.Errorf("user does not exist")
	}
	return s.userRepo.UpdateUser(user)
}

func generateSalt() string {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		log.Fatalf("Failed to generate salt: %v", err)
	}
	return base64.StdEncoding.EncodeToString(salt)
}

func hashPassword(password, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func verifyPassword(password, hashedPassword, salt string) bool {
	return hashPassword(password, salt) == hashedPassword
}
