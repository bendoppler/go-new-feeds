package service

import (
	"news-feed/internal/entity"
	"news-feed/internal/repository"
)

type NewsFeedServiceInterface interface {
	GetNewsfeedPosts() ([]entity.Post, error)
}

type NewsFeedService struct {
	postRepo repository.PostRepositoryInterface
}

// GetNewsfeedPosts retrieves all posts for the newsfeed.
func (s *NewsFeedService) GetNewsfeedPosts() ([]entity.Post, error) {
	return s.postRepo.GetAllPosts()
}
