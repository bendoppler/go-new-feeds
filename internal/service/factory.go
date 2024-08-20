package service

import (
	"news-feed/internal/cache"
	"news-feed/internal/repository"
	"news-feed/internal/storage"
)

type ServiceFactoryInterface interface {
	CreateUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface
	CreatePostService(repo repository.PostRepositoryInterface, storage storage.MinioStorageInterface) PostServiceInterface
	CreateFriendsService(friendsRepo repository.FriendsRepositoryInterface, postRepo repository.PostRepositoryInterface) FriendsServiceInterface
	CreateNewsFeedService(postRepo repository.PostRepositoryInterface) NewsFeedServiceInterface
}

type ServiceFactory struct{}

func (*ServiceFactory) CreateUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface {
	return &UserService{userRepo: userRepo, redisClient: cache.GetRedisClient()}
}

func (*ServiceFactory) CreatePostService(repo repository.PostRepositoryInterface, storage storage.MinioStorageInterface) PostServiceInterface {
	return &PostService{postRepo: repo, storage: storage}
}

func (*ServiceFactory) CreateFriendsService(friendsRepo repository.FriendsRepositoryInterface, postRepo repository.PostRepositoryInterface) FriendsServiceInterface {
	return &FriendsService{friendsRepo: friendsRepo, postRepo: postRepo}
}

func (*ServiceFactory) CreateNewsFeedService(postRepo repository.PostRepositoryInterface) NewsFeedServiceInterface {
	return &NewsFeedService{postRepo: postRepo}
}
