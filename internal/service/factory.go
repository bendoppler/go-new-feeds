package service

import (
	"github.com/go-redis/redis/v8"
	"news-feed/internal/repository"
)

type ServiceFactoryInterface interface {
	CreateUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface
	CreatePostService(repo repository.PostRepository) PostServiceInterface
}

type ServiceFactory struct{}

func (*ServiceFactory) CreateUserService(userRepo repository.UserRepositoryInterface, redisClient *redis.Client) UserServiceInterface {
	return &UserService{userRepo: userRepo, redisClient: redisClient}
}

func (*ServiceFactory) CreatePostService(repo repository.PostRepository) PostServiceInterface {
	return &PostService{postRepo: repo}
}
