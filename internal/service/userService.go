package service

import (
	"fmt"
	"news-feed/internal/entity"
	"news-feed/internal/repository"
)

// UserServiceInterface defines methods for user-related business logic.
type UserServiceInterface interface {
	Login(userName, password string) (entity.User, error)
	Signup(user entity.User) error
	EditProfile(user entity.User) error
}

// UserService is a concrete implementation of UserServiceInterface.
type UserService struct {
	userRepo repository.UserRepositoryInterface
}

// Login authenticates a user.
func (s *UserService) Login(userName, password string) (entity.User, error) {
	user, err := s.userRepo.GetByUserName(userName)
	if err != nil {
		return entity.User{}, err
	}
	if user.Password != password {
		return entity.User{}, fmt.Errorf("invalid credentials")
	}
	return user, nil
}

// Signup registers a new user.
func (s *UserService) Signup(user entity.User) error {
	existingUser, err := s.userRepo.GetByUserName(user.UserName)
	if err != nil {
		return err
	}
	if (existingUser != entity.User{}) {
		return fmt.Errorf("user already exists")
	}
	return s.userRepo.CreateUser(user)
}

// EditProfile updates a user's profile.
func (s *UserService) EditProfile(user entity.User) error {
	existingUser, err := s.userRepo.GetByUserName(user.UserName)
	if err != nil {
		return err
	}
	if (existingUser == entity.User{}) {
		return fmt.Errorf("user does not exist")
	}
	return s.userRepo.UpdateUser(user)
}
