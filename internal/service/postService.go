package service

import (
	"news-feed/internal/api/model"
	"news-feed/internal/entity"
	"news-feed/internal/repository"
	"time"
)

type PostServiceInterface interface {
	CreatePost(post *model.Post) (string, bool, int)
}

type PostService struct {
	postRepo repository.PostRepository
}

func (s *PostService) CreatePost(post *model.Post) (string, bool, int) {
	postEntity := &entity.Post{
		UserID:           post.UserID,
		ContentText:      post.ContentText,
		ContentImagePath: post.ContentImagePath,
		CreatedAt:        time.Now(),
	}

	err := s.postRepo.CreatePost(postEntity)
	if err != nil {
		return "Post creation failed", false, 1
	}
	return "Post created successfully", true, 0
}
