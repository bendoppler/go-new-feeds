package service

import (
	"fmt"
	"io"
	"news-feed/internal/entity"
	"news-feed/internal/repository"
	"news-feed/internal/storage"
	"news-feed/pkg/logger"
)

type PostServiceInterface interface {
	CreatePost(text string, fileName string, userID int) (string, bool, error)
	GetPost(postID int) (entity.Post, error)
	EditPost(post entity.Post) error
	DeletePost(postID int) error
	CommentOnPost(postID int, comment string) error
	LikePost(postID int) error
	UploadImage(fileName string, file io.Reader) (string, error)
}

type PostService struct {
	postRepo repository.PostRepositoryInterface
	storage  storage.MinioStorageInterface
}

func (s *PostService) CreatePost(text string, fileName string, userID int) (string, bool, error) {
	var preSignedURL string
	if text == "" {
		logger.LogError("Cannot create post without text")
		return "", false, nil
	}
	if fileName != "" {
		var err error
		preSignedURL, err = s.storage.GenerateFileURL(fileName)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to generate pre signed url %v", err))
			return "", false, err
		}
	}

	post := entity.Post{
		ContentText:      text,
		ContentImagePath: fileName,
		UserID:           userID,
	}

	err := s.postRepo.CreatePost(post)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to create post: %v", err))
		return "", false, err
	}

	return preSignedURL, true, nil
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
