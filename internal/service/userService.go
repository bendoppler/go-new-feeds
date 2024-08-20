package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"news-feed/internal/entity"
	"news-feed/internal/repository"
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
	// Check if user data is cached in Redis
	cachedUser, err := s.redisClient.Get(context.Background(), username).Result()
	if errors.Is(err, redis.Nil) {
		// Cache miss, fetch user from the database
		user, err := s.userRepo.GetByUserName(username)
		if err != nil {
			return "", err
		}

		// Cache the user data in Redis for future requests
		userData, err := json.Marshal(user)
		if err != nil {
			return "", fmt.Errorf("could not marshal user data: %v", err)
		}

		err = s.redisClient.Set(context.Background(), username, userData, 24*time.Hour).Err()
		if err != nil {
			return "", fmt.Errorf("could not store user data in Redis: %v", err)
		}

		cachedUser = string(userData)
	} else if err != nil {
		return "", fmt.Errorf("could not retrieve user from Redis: %v", err)
	}

	// Unmarshal the cached user data
	var user entity.User
	err = json.Unmarshal([]byte(cachedUser), &user)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal cached user data: %v", err)
	}

	// Verify the password
	if !verifyPassword(password, user.HashedPassword, user.Salt) {
		return "", fmt.Errorf("invalid credentials")
	}

	// Generate JWT
	jwtToken, err := middleware.GenerateJWT(string(rune(user.ID)))
	if err != nil {
		return "", fmt.Errorf("could not generate JWT: %v", err)
	}

	// Store the JWT in Redis with a TTL (e.g., 24 hours)
	err = s.redisClient.Set(context.Background(), jwtToken, username, 24*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("could not store JWT in Redis: %v", err)
	}

	return jwtToken, nil
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

func generateSessionToken() string {
	// Generate a random session token (this is a placeholder; consider using a more robust method)
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		log.Fatalf("Failed to generate session token: %v", err)
	}
	return base64.StdEncoding.EncodeToString(token)
}
