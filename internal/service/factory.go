package service

import "news-feed/internal/repository"

type ServiceFactoryInterface interface {
	CreateUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface
}

type ServiceFactory struct{}

func (*ServiceFactory) CreateUserService(userRepo repository.UserRepositoryInterface) UserServiceInterface {
	return &UserService{userRepo: userRepo}
}
