package handler

import "news-feed/internal/service"

type HandlerFactoryInterface interface {
	CreateUserHandler(userService service.UserServiceInterface) *UserHandler
}

type HandlerFactory struct{}

func (*HandlerFactory) CreateUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
