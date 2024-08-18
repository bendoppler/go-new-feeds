package service

import "news-feed/internal/repository"

type ServiceFactoryInterface interface {
	CreateUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface
	CreatePostService(repo repository.PostRepository) PostServiceInterface
}

type ServiceFactory struct{}

func (*ServiceFactory) CreateUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface {
	return &UserService{userRepo: userRepo}
}

func (*ServiceFactory) CreatePostService(repo repository.PostRepository) PostServiceInterface {
	return &PostService{postRepo: repo}
}
