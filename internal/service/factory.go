package service

import (
	"news-feed/internal/cache"
	"news-feed/internal/repository"
	"news-feed/internal/storage"
)

type ServiceFactoryInterface interface {
	CreateUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface
	CreatePostService(repo repository.PostRepositoryInterface, storage storage.MinioStorageInterface) PostServiceInterface
	CreateFriendsService(
		friendsRepo repository.FriendsRepositoryInterface,
		postRepo repository.PostRepositoryInterface,
		userRepo repository.UserRepositoryInterface) FriendsServiceInterface
	CreateNewsFeedService(postRepo repository.PostRepositoryInterface) NewsFeedServiceInterface
}

type ServiceFactory struct{}

func (*ServiceFactory) CreateUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface {
	return &UserService{userRepo: userRepo, redisClient: cache.GetRedisClient()}
}

func (*ServiceFactory) CreatePostService(repo repository.PostRepositoryInterface, storage storage.MinioStorageInterface) PostServiceInterface {
	return &PostService{postRepo: repo, storage: storage}
}

func (*ServiceFactory) CreateFriendsService(
	friendsRepo repository.FriendsRepositoryInterface,
	postRepo repository.PostRepositoryInterface,
	userRepo repository.UserRepositoryInterface) FriendsServiceInterface {
	return &FriendsService{friendsRepo: friendsRepo, postRepo: postRepo, redisClient: cache.GetRedisClient(), userRepo: userRepo}
}

func (*ServiceFactory) CreateNewsFeedService(postRepo repository.PostRepositoryInterface) NewsFeedServiceInterface {
	return &NewsFeedService{postRepo: postRepo}
}
