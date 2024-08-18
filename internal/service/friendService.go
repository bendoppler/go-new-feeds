package service

import (
	"news-feed/internal/entity"
	"news-feed/internal/repository"
)

type FriendsServiceInterface interface {
	GetFriends(userID int) ([]entity.User, error)
	FollowUser(userID int) (string, error)
	UnfollowUser(userID int) (string, error)
	GetUserPosts(userID int) ([]entity.Post, error)
}

type FriendsService struct {
	friendsRepo repository.FriendsRepositoryInterface
	postRepo    repository.PostRepositoryInterface
}

// GetFriends retrieves the list of friends for a user.
func (s *FriendsService) GetFriends(userID int) ([]entity.User, error) {
	return s.friendsRepo.GetFriends(userID)
}

// FollowUser follows a user and returns a message.
func (s *FriendsService) FollowUser(userID int) (string, error) {
	err := s.friendsRepo.FollowUser(userID)
	if err != nil {
		return "Failed to follow user", err
	}
	return "Successfully followed user", nil
}

// UnfollowUser unfollows a user and returns a message.
func (s *FriendsService) UnfollowUser(userID int) (string, error) {
	err := s.friendsRepo.UnfollowUser(userID)
	if err != nil {
		return "Failed to unfollow user", err
	}
	return "Successfully unfollowed user", nil
}

// GetUserPosts retrieves the posts by a user.
func (s *FriendsService) GetUserPosts(userID int) ([]entity.Post, error) {
	return s.postRepo.GetPostsByUserID(userID)
}
